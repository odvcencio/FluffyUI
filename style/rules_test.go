package style

import "testing"

func TestStylesheet_Rules(t *testing.T) {
	s := NewStylesheet()
	s.Add(Select("Label"), Style{Bold: Bool(true)})
	s.Add(Select("Chip"), Style{Italic: Bool(true)})

	rules := s.Rules()
	if len(rules) != 2 {
		t.Fatalf("Rules() returned %d rules, want 2", len(rules))
	}
	if rules[0].Selector.Type != "Label" {
		t.Errorf("rules[0].Selector.Type = %q, want Label", rules[0].Selector.Type)
	}
	if rules[1].Selector.Type != "Chip" {
		t.Errorf("rules[1].Selector.Type = %q, want Chip", rules[1].Selector.Type)
	}
}

func TestStylesheet_Rules_Nil(t *testing.T) {
	var s *Stylesheet
	rules := s.Rules()
	if rules != nil {
		t.Errorf("nil stylesheet Rules() should return nil")
	}
}
