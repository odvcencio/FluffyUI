package i18n

import "testing"

func TestCursorMap_PureLTR(t *testing.T) {
	text := "Hello"
	reordered := BidiReorder(text, DirectionLTR)
	m := BuildCursorMap(text, reordered)
	for i := 0; i < len([]rune(text)); i++ {
		if v := m.LogicalToVisual(i); v != i {
			t.Errorf("LTR: logical %d → visual %d, want %d", i, v, i)
		}
	}
}

func TestCursorMap_PureRTL(t *testing.T) {
	text := "\u05e9\u05dc\u05d5\u05dd" // שלום
	reordered := BidiReorder(text, DirectionRTL)
	m := BuildCursorMap(text, reordered)
	// RTL text is reversed, so logical 0 maps to visual end
	n := len([]rune(text))
	for i := 0; i < n; i++ {
		v := m.LogicalToVisual(i)
		if v != n-1-i {
			t.Errorf("RTL: logical %d → visual %d, want %d", i, v, n-1-i)
		}
	}
}

func TestCursorMap_OutOfBounds(t *testing.T) {
	m := BuildCursorMap("ab", "ab")
	if v := m.LogicalToVisual(99); v != 2 {
		t.Errorf("out of bounds: got %d, want 2", v)
	}
	if v := m.LogicalToVisual(-1); v != 0 {
		t.Errorf("negative: got %d, want 0", v)
	}
}

func TestCursorMap_Empty(t *testing.T) {
	m := CursorMap{}
	if v := m.LogicalToVisual(5); v != 5 {
		t.Errorf("empty map: got %d, want 5", v)
	}
}

func TestCursorMap_VisualToLogical_PureRTL(t *testing.T) {
	text := "\u05e9\u05dc\u05d5\u05dd" // שלום
	reordered := BidiReorder(text, DirectionRTL)
	m := BuildCursorMap(text, reordered)
	n := len([]rune(text))
	for i := 0; i < n; i++ {
		l := m.VisualToLogical(i)
		if l != n-1-i {
			t.Errorf("RTL: visual %d → logical %d, want %d", i, l, n-1-i)
		}
	}
}

func TestCursorMap_VisualToLogical_OutOfBounds(t *testing.T) {
	m := BuildCursorMap("ab", "ab")
	if v := m.VisualToLogical(99); v != 2 {
		t.Errorf("out of bounds: got %d, want 2", v)
	}
	if v := m.VisualToLogical(-1); v != 0 {
		t.Errorf("negative: got %d, want 0", v)
	}
}

func TestCursorMap_VisualToLogical_Empty(t *testing.T) {
	m := CursorMap{}
	if v := m.VisualToLogical(5); v != 5 {
		t.Errorf("empty map: got %d, want 5", v)
	}
}

func TestCursorMap_MismatchedLengths(t *testing.T) {
	// When lengths don't match, identity mapping is used
	m := BuildCursorMap("abc", "ab")
	if v := m.LogicalToVisual(0); v != 0 {
		t.Errorf("mismatched: got %d, want 0", v)
	}
	if v := m.LogicalToVisual(2); v != 2 {
		t.Errorf("mismatched: got %d, want 2", v)
	}
}
