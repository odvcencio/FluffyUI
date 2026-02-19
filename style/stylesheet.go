package style

import (
	"sort"
	"strings"
	"sync"
)

// Rule binds a selector to a style.
type Rule struct {
	Selector  Selector
	Style     Style
	Important Style
	Media     []MediaQuery

	order       int
	specificity specificity
}

// Stylesheet is a collection of style rules.
type Stylesheet struct {
	mu            sync.RWMutex
	rules         []Rule
	nextOrder     int
	variables     map[string]string
	// Indices for fast rule lookup
	typeIndex     map[string][]int // type name -> rule indices
	idIndex       map[string][]int // id -> rule indices
	classIndex    map[string][]int // class -> rule indices
	otherRules    []int            // rules that don't match a simple type/id/class
	indicesStale  bool             // whether indices need rebuilding
	indicesBuilt  bool             // whether indices have been built at least once
}

// NewStylesheet creates an empty stylesheet.
func NewStylesheet() *Stylesheet {
	return &Stylesheet{}
}

// SetVariable defines a stylesheet variable.
func (s *Stylesheet) SetVariable(name, value string) *Stylesheet {
	if s == nil {
		return s
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return s
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.variables == nil {
		s.variables = make(map[string]string)
	}
	s.variables[key] = value
	return s
}

// GetVariable returns a variable value if set.
func (s *Stylesheet) GetVariable(name string) (string, bool) {
	if s == nil {
		return "", false
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.variables == nil {
		return "", false
	}
	value, ok := s.variables[key]
	return value, ok
}

// Add appends a rule to the stylesheet.
func (s *Stylesheet) Add(selector *SelectorBuilder, style Style) *Stylesheet {
	if s == nil {
		return s
	}
	if selector == nil {
		return s
	}
	sel := selector.Selector()
	s.addRule(sel, style, Style{}, nil)
	return s
}

// AddWithMedia appends a rule gated by media queries.
func (s *Stylesheet) AddWithMedia(selector *SelectorBuilder, style Style, media []MediaQuery) *Stylesheet {
	if s == nil {
		return s
	}
	if selector == nil {
		return s
	}
	sel := selector.Selector()
	s.addRule(sel, style, Style{}, media)
	return s
}

func (s *Stylesheet) addRule(selector Selector, style Style, important Style, media []MediaQuery) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rule := Rule{
		Selector:    selector,
		Style:       style,
		Important:   important,
		Media:       media,
		order:       s.nextOrder,
		specificity: selector.specificity(),
	}
	s.nextOrder++
	s.rules = append(s.rules, rule)
	// Mark indices as stale when rules are added
	s.indicesStale = true
}

// buildIndices constructs lookup indices for fast rule matching.
// Must be called with mu held (write lock).
func (s *Stylesheet) buildIndices() {
	if s == nil {
		return
	}
	// Reset indices
	s.typeIndex = make(map[string][]int)
	s.idIndex = make(map[string][]int)
	s.classIndex = make(map[string][]int)
	s.otherRules = nil

	// Build indices from rules
	for i := range s.rules {
		rule := &s.rules[i]
		// Look at the first (rightmost) selector segment
		sel := &rule.Selector
		indexed := false

		// Index by type
		if sel.Type != "" && sel.Type != "*" {
			s.typeIndex[sel.Type] = append(s.typeIndex[sel.Type], i)
			indexed = true
		}

		// Index by ID
		if sel.ID != "" {
			s.idIndex[sel.ID] = append(s.idIndex[sel.ID], i)
			indexed = true
		}

		// Index by classes
		for _, class := range sel.Classes {
			s.classIndex[class] = append(s.classIndex[class], i)
			indexed = true
		}

		// If no simple index applies, add to otherRules
		if !indexed {
			s.otherRules = append(s.otherRules, i)
		}
	}

	s.indicesStale = false
	s.indicesBuilt = true
}

// Resolve returns the merged style for the given node.
func (s *Stylesheet) Resolve(node Node, ancestors []Node) Style {
	return s.ResolveWithContext(node, ancestors, MediaContext{})
}

// ResolveWithContext returns the merged style for the given node and media context.
func (s *Stylesheet) ResolveWithContext(node Node, ancestors []Node, ctx MediaContext) Style {
	if s == nil || node == nil {
		return Style{}
	}

	// Check if we need to rebuild indices (with read lock first for common case)
	s.mu.RLock()
	needsRebuild := s.indicesStale || !s.indicesBuilt
	s.mu.RUnlock()

	if needsRebuild {
		s.mu.Lock()
		// Double-check after acquiring write lock
		if s.indicesStale || !s.indicesBuilt {
			s.buildIndices()
		}
		s.mu.Unlock()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// If indices aren't built (shouldn't happen now), fall back to full scan
	if !s.indicesBuilt {
		var matches []Rule
		for _, rule := range s.rules {
			if rule.Selector.Matches(node, ancestors) && mediaMatches(rule.Media, ctx) {
				matches = append(matches, rule)
			}
		}
		if len(matches) == 0 {
			return Style{}
		}
		sort.SliceStable(matches, func(i, j int) bool {
			if matches[i].specificity == matches[j].specificity {
				return matches[i].order < matches[j].order
			}
			return matches[i].specificity.less(matches[j].specificity)
		})
		var resolved Style
		for _, rule := range matches {
			resolved = resolved.Merge(rule.Style)
		}
		for _, rule := range matches {
			resolved = resolved.Merge(rule.Important)
		}
		return resolved
	}

	// Collect candidate rule indices from node's type, id, and classes
	candidateSet := make(map[int]struct{})

	// Add rules matching node type
	if indices, ok := s.typeIndex[node.StyleType()]; ok {
		for _, idx := range indices {
			candidateSet[idx] = struct{}{}
		}
	}

	// Add rules matching node ID
	if nodeID := node.StyleID(); nodeID != "" {
		if indices, ok := s.idIndex[nodeID]; ok {
			for _, idx := range indices {
				candidateSet[idx] = struct{}{}
			}
		}
	}

	// Add rules matching any node class
	for _, class := range node.StyleClasses() {
		if indices, ok := s.classIndex[class]; ok {
			for _, idx := range indices {
				candidateSet[idx] = struct{}{}
			}
		}
	}

	// Always include otherRules (universal selectors, pseudo-only, etc.)
	for _, idx := range s.otherRules {
		candidateSet[idx] = struct{}{}
	}

	// Now test only the candidate rules
	var matches []Rule
	for idx := range candidateSet {
		rule := s.rules[idx]
		if rule.Selector.Matches(node, ancestors) && mediaMatches(rule.Media, ctx) {
			matches = append(matches, rule)
		}
	}

	if len(matches) == 0 {
		return Style{}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].specificity == matches[j].specificity {
			return matches[i].order < matches[j].order
		}
		return matches[i].specificity.less(matches[j].specificity)
	})
	var resolved Style
	for _, rule := range matches {
		resolved = resolved.Merge(rule.Style)
	}
	for _, rule := range matches {
		resolved = resolved.Merge(rule.Important)
	}
	return resolved
}

// classNode is a minimal Node implementation for ResolveClass.
type classNode struct {
	class string
}

func (n *classNode) StyleType() string        { return "" }
func (n *classNode) StyleID() string           { return "" }
func (n *classNode) StyleClasses() []string    { return []string{n.class} }
func (n *classNode) StyleState() WidgetState   { return WidgetState{} }

// ResolveClass resolves the style for a given class name.
// This is useful for mapping syntax highlight captures to FSS styles.
func (s *Stylesheet) ResolveClass(className string) Style {
	if s == nil || className == "" {
		return Style{}
	}
	return s.ResolveWithContext(&classNode{class: className}, nil, MediaContext{})
}

// RelayoutOnFocus reports whether focus-based rules can affect layout.
func (s *Stylesheet) RelayoutOnFocus() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rule := range s.rules {
		if !rule.Selector.HasPseudo(PseudoFocus) {
			continue
		}
		if rule.Style.AffectsLayout() || rule.Important.AffectsLayout() {
			return true
		}
	}
	return false
}

// Merge combines multiple stylesheets into one.
func Merge(sheets ...*Stylesheet) *Stylesheet {
	merged := NewStylesheet()
	for _, sheet := range sheets {
		if sheet == nil {
			continue
		}
		sheet.mu.RLock()
		rules := append([]Rule(nil), sheet.rules...)
		vars := make(map[string]string, len(sheet.variables))
		for key, value := range sheet.variables {
			vars[key] = value
		}
		sheet.mu.RUnlock()
		for _, rule := range rules {
			rule.order = merged.nextOrder
			merged.nextOrder++
			merged.rules = append(merged.rules, rule)
		}
		if len(vars) > 0 {
			if merged.variables == nil {
				merged.variables = make(map[string]string)
			}
			for key, value := range vars {
				merged.variables[key] = value
			}
		}
	}
	// Build indices after merging all rules
	merged.buildIndices()
	return merged
}
