package style

import (
	"fmt"
	"testing"
)

type mockNode struct {
	typ     string
	id      string
	classes []string
	state   WidgetState
}

func (m *mockNode) StyleType() string          { return m.typ }
func (m *mockNode) StyleID() string            { return m.id }
func (m *mockNode) StyleClasses() []string     { return m.classes }
func (m *mockNode) StyleState() WidgetState    { return m.state }

// BenchmarkResolveWithManyRules tests performance with a large stylesheet
func BenchmarkResolveWithManyRules(b *testing.B) {
	// Create a stylesheet with many rules (more realistic scenario)
	sheet := NewStylesheet()

	// Add 1000 rules with various selectors to simulate a large app stylesheet
	for i := 0; i < 1000; i++ {
		typeName := fmt.Sprintf("Widget%d", i%100)
		className := fmt.Sprintf("class%d", i%50)
		sheet.Add(Select(typeName), Style{Foreground: ColorRed})
		sheet.Add(Select(typeName).Class(className), Style{Foreground: ColorGreen})
		sheet.Add(Select(typeName).Pseudo(PseudoFocus), Style{Foreground: ColorBlue})
		sheet.Add(SelectID(fmt.Sprintf("id%d", i)), Style{Background: ColorWhite})
	}

	// Build indices
	sheet.mu.Lock()
	sheet.buildIndices()
	sheet.mu.Unlock()

	// Create test nodes
	node := &mockNode{
		typ:     "Widget50",
		classes: []string{"class25"},
		state:   WidgetState{},
	}

	var ancestors []Node

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}

// BenchmarkResolveWithManyRulesNoIndex tests performance without indices (for comparison)
func BenchmarkResolveWithManyRulesNoIndex(b *testing.B) {
	// Create a stylesheet with many rules
	sheet := NewStylesheet()

	// Add 1000 rules with various selectors to simulate a large app stylesheet
	for i := 0; i < 1000; i++ {
		typeName := fmt.Sprintf("Widget%d", i%100)
		className := fmt.Sprintf("class%d", i%50)
		sheet.Add(Select(typeName), Style{Foreground: ColorRed})
		sheet.Add(Select(typeName).Class(className), Style{Foreground: ColorGreen})
		sheet.Add(Select(typeName).Pseudo(PseudoFocus), Style{Foreground: ColorBlue})
		sheet.Add(SelectID(fmt.Sprintf("id%d", i)), Style{Background: ColorWhite})
	}

	// Don't build indices - force fallback to full scan
	sheet.mu.Lock()
	sheet.indicesBuilt = false
	sheet.indicesStale = false
	sheet.mu.Unlock()

	// Create test nodes
	node := &mockNode{
		typ:     "Widget50",
		classes: []string{"class25"},
		state:   WidgetState{},
	}

	var ancestors []Node

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}

// BenchmarkSpecificityCached tests specificity with caching (as used in parsed stylesheets)
func BenchmarkSpecificityCached(b *testing.B) {
	// Parse a stylesheet to get selectors with cached specificity
	fss := `
		Dialog Button.primary.large:focus:hover { foreground: red; }
	`
	sheet, err := Parse(fss)
	if err != nil {
		b.Fatalf("Failed to parse: %v", err)
	}

	sel := sheet.rules[0].Selector

	if !sel.specCached {
		b.Fatal("Selector specificity should be cached after parsing")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sel.specificity()
	}
}

// BenchmarkSpecificityUncached tests specificity without caching (as used in manually built selectors)
func BenchmarkSpecificityUncached(b *testing.B) {
	// Create selector without cached specificity (via builder API)
	sel := Select("Button").Class("primary", "large").Pseudo(PseudoFocus, PseudoHover).Inside("Dialog").Selector()

	if sel.specCached {
		b.Fatal("Selector built via API should not have cached specificity")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sel.specificity()
	}
}

// BenchmarkRuleSorting tests the performance impact of sorting rules with cached specificity
func BenchmarkRuleSorting(b *testing.B) {
	// Create many rules with varying specificity
	fss := ""
	for i := 0; i < 100; i++ {
		fss += fmt.Sprintf("Widget%d { foreground: red; }\n", i)
		fss += fmt.Sprintf("Widget%d.class%d { background: blue; }\n", i, i)
		fss += fmt.Sprintf("#id%d { margin: 1; }\n", i)
		fss += fmt.Sprintf("Panel Widget%d:focus { padding: 2; }\n", i)
	}

	sheet, err := Parse(fss)
	if err != nil {
		b.Fatalf("Failed to parse: %v", err)
	}

	// Create a node that matches many rules
	node := &mockNode{
		typ:     "Widget50",
		id:      "id50",
		classes: []string{"class50"},
		state:   WidgetState{Focused: true},
	}

	ancestors := []Node{&mockNode{typ: "Panel"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}
