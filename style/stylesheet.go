package style

import (
	"sort"
	"strings"
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
	rules     []Rule
	nextOrder int
	variables map[string]string
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
	if s.variables == nil {
		s.variables = make(map[string]string)
	}
	s.variables[key] = value
	return s
}

// GetVariable returns a variable value if set.
func (s *Stylesheet) GetVariable(name string) (string, bool) {
	if s == nil || s.variables == nil {
		return "", false
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
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

// Merge combines multiple stylesheets into one.
func Merge(sheets ...*Stylesheet) *Stylesheet {
	merged := NewStylesheet()
	for _, sheet := range sheets {
		if sheet == nil {
			continue
		}
		for _, rule := range sheet.rules {
			rule.order = merged.nextOrder
			merged.nextOrder++
			merged.rules = append(merged.rules, rule)
		}
		if len(sheet.variables) > 0 {
			if merged.variables == nil {
				merged.variables = make(map[string]string)
			}
			for key, value := range sheet.variables {
				merged.variables[key] = value
			}
		}
	}
	return merged
}
