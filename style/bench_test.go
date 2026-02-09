package style

import (
	"fmt"
	"testing"
)

// BenchmarkSpecificityCalculation measures the cost of computing specificity
// for a complex selector built via the builder API (uncached).
func BenchmarkSpecificityCalculation(b *testing.B) {
	sel := Select("Button").Class("primary", "large").Pseudo(PseudoFocus, PseudoHover).Inside("Dialog").Selector()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sel.specificity()
	}
}

// BenchmarkSelectorMatches measures the cost of matching a selector
// against a node with ancestors.
func BenchmarkSelectorMatches(b *testing.B) {
	sel := Select("Button").Class("primary").Pseudo(PseudoFocus).Inside("Dialog").Selector()
	node := &mockNode{
		typ:     "Button",
		classes: []string{"primary"},
		state:   WidgetState{Focused: true},
	}
	ancestors := []Node{&mockNode{typ: "Dialog"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sel.Matches(node, ancestors)
	}
}

// BenchmarkResolve1Rule measures Resolve with a minimal stylesheet (1 rule).
func BenchmarkResolve1Rule(b *testing.B) {
	sheet := NewStylesheet()
	sheet.Add(Select("Button"), Style{Foreground: ColorRed})
	sheet.mu.Lock()
	sheet.buildIndices()
	sheet.mu.Unlock()

	node := &mockNode{typ: "Button"}
	var ancestors []Node
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}

// BenchmarkResolve100Rules measures Resolve with 100 rules.
func BenchmarkResolve100Rules(b *testing.B) {
	sheet := NewStylesheet()
	for i := 0; i < 100; i++ {
		typeName := fmt.Sprintf("Widget%d", i%20)
		sheet.Add(Select(typeName), Style{Foreground: ColorRed})
	}
	sheet.mu.Lock()
	sheet.buildIndices()
	sheet.mu.Unlock()

	node := &mockNode{typ: "Widget10", classes: []string{"active"}}
	var ancestors []Node
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}

// BenchmarkSelectorParse measures parsing a single selector string from FSS.
func BenchmarkSelectorParse(b *testing.B) {
	fss := `Button.primary:focus { foreground: red; }`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(fss)
	}
}

// BenchmarkRuleParseLarge measures parsing a realistic stylesheet
// with many rules and varied selectors.
func BenchmarkRuleParseLarge(b *testing.B) {
	var fss string
	for i := 0; i < 50; i++ {
		fss += fmt.Sprintf("Widget%d { foreground: red; background: blue; }\n", i)
		fss += fmt.Sprintf("Widget%d.class%d:focus { foreground: green; }\n", i, i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(fss)
	}
}

// BenchmarkStyleMerge measures the cost of merging two styles.
func BenchmarkStyleMerge(b *testing.B) {
	base := Style{Foreground: ColorRed, Bold: Bool(true), Padding: Pad(2)}
	override := Style{Foreground: ColorGreen, Italic: Bool(true)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Merge(override)
	}
}

// BenchmarkResolveNoMatch measures Resolve when no rules match the node,
// testing the fast path through index lookups with zero candidates.
func BenchmarkResolveNoMatch(b *testing.B) {
	sheet := NewStylesheet()
	for i := 0; i < 100; i++ {
		sheet.Add(Select(fmt.Sprintf("Widget%d", i)), Style{Foreground: ColorRed})
	}
	sheet.mu.Lock()
	sheet.buildIndices()
	sheet.mu.Unlock()

	node := &mockNode{typ: "NoSuchWidget"}
	var ancestors []Node
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sheet.Resolve(node, ancestors)
	}
}

// BenchmarkStyleIsZero measures the IsZero check on an empty style,
// which is the hot path for merge short-circuiting.
func BenchmarkStyleIsZero(b *testing.B) {
	s := Style{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.IsZero()
	}
}
