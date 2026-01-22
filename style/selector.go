package style

import "strings"

// WidgetState captures pseudo-class state.
type WidgetState struct {
	Focused    bool
	Disabled   bool
	Hovered    bool
	Active     bool
	FirstChild bool
	LastChild  bool
}

// PseudoClass identifies a widget pseudo-class.
type PseudoClass string

const (
	PseudoFocus    PseudoClass = "focus"
	PseudoDisabled PseudoClass = "disabled"
	PseudoHover    PseudoClass = "hover"
	PseudoActive   PseudoClass = "active"
	PseudoFirst    PseudoClass = "first-child"
	PseudoLast     PseudoClass = "last-child"
)

// Node exposes style metadata for selector matching.
type Node interface {
	StyleType() string
	StyleID() string
	StyleClasses() []string
	StyleState() WidgetState
}

// Selector defines a CSS-like selector.
type Selector struct {
	Type    string
	ID      string
	Classes []string
	Pseudo  []PseudoClass
	Parent  *Selector
}

// SelectorBuilder builds selectors fluently.
type SelectorBuilder struct {
	sel Selector
}

// Select creates a new selector for a widget type.
func Select(widgetType string) *SelectorBuilder {
	return &SelectorBuilder{sel: Selector{Type: strings.TrimSpace(widgetType)}}
}

// SelectID creates a new selector for an ID.
func SelectID(id string) *SelectorBuilder {
	return &SelectorBuilder{sel: Selector{Type: "*", ID: strings.TrimSpace(id)}}
}

// ID sets the selector ID.
func (b *SelectorBuilder) ID(id string) *SelectorBuilder {
	if b == nil {
		return b
	}
	b.sel.ID = strings.TrimSpace(id)
	return b
}

// Class adds one or more class selectors.
func (b *SelectorBuilder) Class(classes ...string) *SelectorBuilder {
	if b == nil {
		return b
	}
	for _, class := range classes {
		name := strings.TrimSpace(class)
		if name == "" {
			continue
		}
		b.sel.Classes = append(b.sel.Classes, name)
	}
	return b
}

// Pseudo adds pseudo-classes to the selector.
func (b *SelectorBuilder) Pseudo(pseudos ...PseudoClass) *SelectorBuilder {
	if b == nil {
		return b
	}
	for _, pseudo := range pseudos {
		if pseudo == "" {
			continue
		}
		b.sel.Pseudo = append(b.sel.Pseudo, pseudo)
	}
	return b
}

// Inside adds a descendant selector.
func (b *SelectorBuilder) Inside(parentType string) *SelectorBuilder {
	if b == nil {
		return b
	}
	parent := &Selector{Type: strings.TrimSpace(parentType)}
	if b.sel.Parent == nil {
		b.sel.Parent = parent
		return b
	}
	cursor := b.sel.Parent
	for cursor.Parent != nil {
		cursor = cursor.Parent
	}
	cursor.Parent = parent
	return b
}

// Selector returns the built selector.
func (b *SelectorBuilder) Selector() Selector {
	if b == nil {
		return Selector{}
	}
	return b.sel
}

type specificity struct {
	ids     int
	classes int
	types   int
}

func (s Selector) specificity() specificity {
	spec := specificity{}
	if s.ID != "" {
		spec.ids++
	}
	if len(s.Classes) > 0 {
		spec.classes += len(s.Classes)
	}
	if len(s.Pseudo) > 0 {
		spec.classes += len(s.Pseudo)
	}
	if s.Type != "" && s.Type != "*" {
		spec.types++
	}
	if s.Parent != nil {
		parentSpec := s.Parent.specificity()
		spec.ids += parentSpec.ids
		spec.classes += parentSpec.classes
		spec.types += parentSpec.types
	}
	return spec
}

func (s specificity) less(other specificity) bool {
	if s.ids != other.ids {
		return s.ids < other.ids
	}
	if s.classes != other.classes {
		return s.classes < other.classes
	}
	return s.types < other.types
}

// Matches reports whether the selector matches the node.
func (s Selector) Matches(node Node, ancestors []Node) bool {
	if node == nil {
		return false
	}
	if !s.matchesSelf(node) {
		return false
	}
	if s.Parent == nil {
		return true
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		if s.Parent.Matches(ancestors[i], ancestors[:i]) {
			return true
		}
	}
	return false
}

func (s Selector) matchesSelf(node Node) bool {
	if node == nil {
		return false
	}
	typ := node.StyleType()
	if s.Type != "" && s.Type != "*" && typ != s.Type {
		return false
	}
	if s.ID != "" && node.StyleID() != s.ID {
		return false
	}
	if len(s.Classes) > 0 && !hasAllClasses(node.StyleClasses(), s.Classes) {
		return false
	}
	state := node.StyleState()
	for _, pseudo := range s.Pseudo {
		if !stateHasPseudo(state, pseudo) {
			return false
		}
	}
	return true
}

func hasAllClasses(nodeClasses, selectorClasses []string) bool {
	if len(selectorClasses) == 0 {
		return true
	}
	if len(nodeClasses) == 0 {
		return false
	}
	for _, sel := range selectorClasses {
		found := false
		for _, cls := range nodeClasses {
			if cls == sel {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func stateHasPseudo(state WidgetState, pseudo PseudoClass) bool {
	switch pseudo {
	case PseudoFocus:
		return state.Focused
	case PseudoDisabled:
		return state.Disabled
	case PseudoHover:
		return state.Hovered
	case PseudoActive:
		return state.Active
	case PseudoFirst:
		return state.FirstChild
	case PseudoLast:
		return state.LastChild
	default:
		return false
	}
}
