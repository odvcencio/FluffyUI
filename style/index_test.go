package style

import (
	"fmt"
	"testing"
)

// TestIndexingCorrectness verifies that indexed resolve produces the same results as non-indexed
func TestIndexingCorrectness(t *testing.T) {
	// Create a stylesheet with various rules
	sheet := NewStylesheet()

	// Add rules with different selector types
	sheet.Add(Select("Button"), Style{Foreground: ColorRed})
	sheet.Add(Select("Button").Class("primary"), Style{Background: ColorBlue})
	sheet.Add(Select("Button").Pseudo(PseudoFocus), Style{Bold: Bool(true)})
	sheet.Add(SelectID("submit"), Style{Foreground: ColorGreen})
	sheet.Add(Select("*").Class("active"), Style{Background: ColorYellow})
	sheet.Add(Select("Button").Inside("Dialog"), Style{Margin: Pad(2)})

	// Create test nodes
	tests := []struct {
		name     string
		node     Node
		hasStyle bool
	}{
		{
			name: "simple button",
			node: &mockNode{typ: "Button"},
			hasStyle: true,
		},
		{
			name: "button with class",
			node: &mockNode{typ: "Button", classes: []string{"primary"}},
			hasStyle: true,
		},
		{
			name: "button with id",
			node: &mockNode{typ: "Button", id: "submit"},
			hasStyle: true,
		},
		{
			name: "non-matching widget",
			node: &mockNode{typ: "Label"},
			hasStyle: false,
		},
		{
			name: "universal class selector",
			node: &mockNode{typ: "Label", classes: []string{"active"}},
			hasStyle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with indices built
			sheet.mu.Lock()
			sheet.buildIndices()
			sheet.mu.Unlock()

			resultWithIndex := sheet.Resolve(tt.node, nil)

			// Test without indices (force full scan)
			sheet.mu.Lock()
			sheet.indicesBuilt = false
			sheet.indicesStale = false
			sheet.mu.Unlock()

			resultNoIndex := sheet.Resolve(tt.node, nil)

			// Results should be identical
			if resultWithIndex != resultNoIndex {
				t.Errorf("Results differ:\nWith index: %+v\nNo index: %+v", resultWithIndex, resultNoIndex)
			}

			if tt.hasStyle && resultWithIndex.IsZero() {
				t.Errorf("Expected non-zero style but got zero")
			}
		})
	}
}

// TestIndexingWithComplexSelectors tests indexing with complex selector chains
func TestIndexingWithComplexSelectors(t *testing.T) {
	// Create a stylesheet
	fss := `
		Button { foreground: red; }
		Button.primary { background: blue; }
		Button:focus { bold: true; }
		#submit { foreground: green; }
		Dialog Button { margin: 2; }
		Panel > Button { padding: 1; }
	`

	sheet, err := Parse(fss)
	if err != nil {
		t.Fatalf("Failed to parse FSS: %v", err)
	}

	// Create test scenario with ancestors
	dialog := &mockNode{typ: "Dialog"}
	button := &mockNode{typ: "Button", id: "submit", classes: []string{"primary"}}
	ancestors := []Node{dialog}

	// Test with indices
	resultWithIndex := sheet.Resolve(button, ancestors)

	// Test without indices
	sheet.mu.Lock()
	sheet.indicesBuilt = false
	sheet.indicesStale = false
	sheet.mu.Unlock()

	resultNoIndex := sheet.Resolve(button, ancestors)

	// Results should be identical
	if resultWithIndex != resultNoIndex {
		t.Errorf("Results differ:\nWith index: %+v\nNo index: %+v", resultWithIndex, resultNoIndex)
	}

	// Should have styles from multiple matching rules
	if resultWithIndex.IsZero() {
		t.Error("Expected non-zero style")
	}
}

// TestIndexRebuildingOnAdd tests that indices are properly rebuilt when rules are added
func TestIndexRebuildingOnAdd(t *testing.T) {
	sheet := NewStylesheet()

	// Add initial rule and build indices
	sheet.Add(Select("Button"), Style{Foreground: ColorRed})
	sheet.mu.Lock()
	sheet.buildIndices()
	sheet.mu.Unlock()

	node := &mockNode{typ: "Button"}

	// Resolve should work
	result1 := sheet.Resolve(node, nil)
	if result1.Foreground != ColorRed {
		t.Errorf("Expected red foreground, got %v", result1.Foreground)
	}

	// Add another rule - indices should be marked stale
	sheet.Add(Select("Button"), Style{Background: ColorBlue})

	// Resolve should still work and include new rule
	result2 := sheet.Resolve(node, nil)
	if result2.Foreground != ColorRed {
		t.Errorf("Expected red foreground after adding rule, got %v", result2.Foreground)
	}
	if result2.Background != ColorBlue {
		t.Errorf("Expected blue background after adding rule, got %v", result2.Background)
	}
}

// TestIndexBuildingInParse tests that indices are built during parsing
func TestIndexBuildingInParse(t *testing.T) {
	fss := `
		Button { foreground: red; }
		Label { foreground: blue; }
	`

	sheet, err := Parse(fss)
	if err != nil {
		t.Fatalf("Failed to parse FSS: %v", err)
	}

	// Indices should be built
	if !sheet.indicesBuilt {
		t.Error("Indices should be built after Parse()")
	}

	if sheet.typeIndex == nil {
		t.Error("typeIndex should not be nil after Parse()")
	}

	// Check that type index contains entries
	if _, ok := sheet.typeIndex["Button"]; !ok {
		t.Error("typeIndex should contain 'Button'")
	}
	if _, ok := sheet.typeIndex["Label"]; !ok {
		t.Error("typeIndex should contain 'Label'")
	}
}

// TestIndexBuildingInMerge tests that indices are built during merge
func TestIndexBuildingInMerge(t *testing.T) {
	sheet1 := NewStylesheet()
	sheet1.Add(Select("Button"), Style{Foreground: ColorRed})

	sheet2 := NewStylesheet()
	sheet2.Add(Select("Label"), Style{Foreground: ColorBlue})

	merged := Merge(sheet1, sheet2)

	// Indices should be built
	if !merged.indicesBuilt {
		t.Error("Indices should be built after Merge()")
	}

	if merged.typeIndex == nil {
		t.Error("typeIndex should not be nil after Merge()")
	}

	// Check that both types are indexed
	if _, ok := merged.typeIndex["Button"]; !ok {
		t.Error("typeIndex should contain 'Button' after merge")
	}
	if _, ok := merged.typeIndex["Label"]; !ok {
		t.Error("typeIndex should contain 'Label' after merge")
	}
}

// TestSpecificityCaching tests that specificity is cached correctly when parsed
func TestSpecificityCaching(t *testing.T) {
	// Selectors built through the builder API don't have cached specificity
	// until they're parsed. This is by design - the cache is for stylesheet rules.
	sel := Select("Button").Class("primary").Pseudo(PseudoFocus).Inside("Dialog").Selector()

	// First call computes specificity (not cached since not parsed)
	spec1 := sel.specificity()

	// Second call should return same value
	spec2 := sel.specificity()
	if spec1 != spec2 {
		t.Errorf("Specificity differs: %+v vs %+v", spec1, spec2)
	}

	// Expected specificity: 1 type (Dialog), 1 type (Button), 2 classes (primary, focus)
	expectedClasses := 2
	expectedTypes := 2
	if spec1.classes != expectedClasses {
		t.Errorf("Expected %d classes, got %d", expectedClasses, spec1.classes)
	}
	if spec1.types != expectedTypes {
		t.Errorf("Expected %d types, got %d", expectedTypes, spec1.types)
	}
}

// TestSpecificityCachingInParse tests that specificity is cached at parse time
func TestSpecificityCachingInParse(t *testing.T) {
	fss := `
		Button.primary:focus { foreground: red; }
		Dialog > Button { background: blue; }
	`

	sheet, err := Parse(fss)
	if err != nil {
		t.Fatalf("Failed to parse FSS: %v", err)
	}

	// All selectors should have cached specificity
	for i, rule := range sheet.rules {
		if !rule.Selector.specCached {
			t.Errorf("Rule %d selector specificity should be cached", i)
		}
	}
}

// BenchmarkIndexEfficiency tests the efficiency of indexing with various scenarios
func BenchmarkIndexEfficiency(b *testing.B) {
	scenarios := []struct {
		name      string
		numRules  int
		numTypes  int
	}{
		{"small_10rules", 10, 5},
		{"medium_100rules", 100, 20},
		{"large_1000rules", 1000, 50},
		{"xlarge_5000rules", 5000, 100},
	}

	for _, scenario := range scenarios {
		b.Run(fmt.Sprintf("%s_indexed", scenario.name), func(b *testing.B) {
			sheet := NewStylesheet()
			for i := 0; i < scenario.numRules; i++ {
				typeName := fmt.Sprintf("Widget%d", i%scenario.numTypes)
				sheet.Add(Select(typeName), Style{Foreground: ColorRed})
			}
			sheet.mu.Lock()
			sheet.buildIndices()
			sheet.mu.Unlock()

			node := &mockNode{typ: fmt.Sprintf("Widget%d", scenario.numTypes/2)}
			var ancestors []Node

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = sheet.Resolve(node, ancestors)
			}
		})

		b.Run(fmt.Sprintf("%s_no_index", scenario.name), func(b *testing.B) {
			sheet := NewStylesheet()
			for i := 0; i < scenario.numRules; i++ {
				typeName := fmt.Sprintf("Widget%d", i%scenario.numTypes)
				sheet.Add(Select(typeName), Style{Foreground: ColorRed})
			}
			sheet.mu.Lock()
			sheet.indicesBuilt = false
			sheet.indicesStale = false
			sheet.mu.Unlock()

			node := &mockNode{typ: fmt.Sprintf("Widget%d", scenario.numTypes/2)}
			var ancestors []Node

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = sheet.Resolve(node, ancestors)
			}
		})
	}
}
