package style

import "sort"

// Rule binds a selector to a style.
type Rule struct {
	Selector Selector
	Style    Style

	order       int
	specificity specificity
}

// Stylesheet is a collection of style rules.
type Stylesheet struct {
	rules     []Rule
	nextOrder int
}

// NewStylesheet creates an empty stylesheet.
func NewStylesheet() *Stylesheet {
	return &Stylesheet{}
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
	rule := Rule{
		Selector:    sel,
		Style:       style,
		order:       s.nextOrder,
		specificity: sel.specificity(),
	}
	s.nextOrder++
	s.rules = append(s.rules, rule)
	return s
}

// Resolve returns the merged style for the given node.
func (s *Stylesheet) Resolve(node Node, ancestors []Node) Style {
	if s == nil || node == nil {
		return Style{}
	}
	var matches []Rule
	for _, rule := range s.rules {
		if rule.Selector.Matches(node, ancestors) {
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
	}
	return merged
}
