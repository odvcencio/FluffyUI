package widgets

import "testing"

func TestSelection(t *testing.T) {
	t.Run("IsEmpty", func(t *testing.T) {
		empty := Selection{Start: 5, End: 5}
		if !empty.IsEmpty() {
			t.Error("Selection with Start==End should be empty")
		}

		nonEmpty := Selection{Start: 0, End: 5}
		if nonEmpty.IsEmpty() {
			t.Error("Selection with Start!=End should not be empty")
		}
	})

	t.Run("Length", func(t *testing.T) {
		sel := Selection{Start: 2, End: 8}
		if sel.Length() != 6 {
			t.Errorf("Length = %d, want 6", sel.Length())
		}

		// Reversed selection should also work
		reversed := Selection{Start: 8, End: 2}
		if reversed.Length() != 6 {
			t.Errorf("Reversed Length = %d, want 6", reversed.Length())
		}
	})

	t.Run("Normalize", func(t *testing.T) {
		reversed := Selection{Start: 10, End: 5}
		normalized := reversed.Normalize()
		if normalized.Start != 5 || normalized.End != 10 {
			t.Errorf("Normalize = (%d, %d), want (5, 10)", normalized.Start, normalized.End)
		}

		// Already normalized should stay the same
		normal := Selection{Start: 5, End: 10}
		sameNormalized := normal.Normalize()
		if sameNormalized.Start != 5 || sameNormalized.End != 10 {
			t.Errorf("Already normalized changed to (%d, %d)", sameNormalized.Start, sameNormalized.End)
		}
	})
}

func TestInputSelection(t *testing.T) {
	input := NewInput()
	input.SetText("hello world test")

	t.Run("SelectAll", func(t *testing.T) {
		input.SelectAll()
		if !input.HasSelection() {
			t.Error("HasSelection should be true after SelectAll")
		}
		sel := input.GetSelection()
		if sel.Start != 0 || sel.End != 16 {
			t.Errorf("SelectAll selection = (%d, %d), want (0, 16)", sel.Start, sel.End)
		}
	})

	t.Run("SelectNone", func(t *testing.T) {
		input.SelectAll()
		input.SelectNone()
		if input.HasSelection() {
			t.Error("HasSelection should be false after SelectNone")
		}
	})

	t.Run("SetSelection", func(t *testing.T) {
		input.SetSelection(Selection{Start: 6, End: 11})
		text := input.GetSelectedText()
		if text != "world" {
			t.Errorf("GetSelectedText = %q, want %q", text, "world")
		}
	})

	t.Run("SetSelection clamps to bounds", func(t *testing.T) {
		input.SetSelection(Selection{Start: -5, End: 100})
		sel := input.GetSelection()
		if sel.Start != 0 {
			t.Errorf("Start should be clamped to 0, got %d", sel.Start)
		}
		if sel.End != 16 {
			t.Errorf("End should be clamped to text length, got %d", sel.End)
		}
	})

	t.Run("SelectWord", func(t *testing.T) {
		// Set cursor in middle of "world"
		input.cursorPos = 8
		input.SelectWord()
		text := input.GetSelectedText()
		if text != "world" {
			t.Errorf("SelectWord at pos 8 = %q, want %q", text, "world")
		}
	})

	t.Run("SelectLine", func(t *testing.T) {
		input.SelectLine()
		text := input.GetSelectedText()
		if text != "hello world test" {
			t.Errorf("SelectLine = %q, want %q", text, "hello world test")
		}
	})
}

func TestMultilineInputSelection(t *testing.T) {
	input := NewMultilineInput()
	input.SetText("line one\nline two\nline three")

	t.Run("SelectAll", func(t *testing.T) {
		input.SelectAll()
		if !input.HasSelection() {
			t.Error("HasSelection should be true after SelectAll")
		}
		text := input.GetSelectedText()
		if text != "line one\nline two\nline three" {
			t.Errorf("SelectAll text = %q", text)
		}
	})

	t.Run("SelectNone", func(t *testing.T) {
		input.SelectAll()
		input.SelectNone()
		if input.HasSelection() {
			t.Error("HasSelection should be false after SelectNone")
		}
	})

	t.Run("SetSelection", func(t *testing.T) {
		input.SetSelection(Selection{Start: 0, End: 8})
		text := input.GetSelectedText()
		if text != "line one" {
			t.Errorf("GetSelectedText = %q, want %q", text, "line one")
		}
	})

	t.Run("SelectLine", func(t *testing.T) {
		input.cursorY = 1 // Second line
		input.SelectLine()
		text := input.GetSelectedText()
		if text != "line two" {
			t.Errorf("SelectLine = %q, want %q", text, "line two")
		}
	})
}

func TestFindWordBoundaries(t *testing.T) {
	tests := []struct {
		text      string
		pos       int
		wantStart int
		wantEnd   int
	}{
		{"hello world", 0, 0, 5},   // Start of first word
		{"hello world", 2, 0, 5},   // Middle of first word
		{"hello world", 5, 0, 5},   // End of first word
		{"hello world", 6, 6, 11},  // Start of second word
		{"hello world", 8, 6, 11},  // Middle of second word
		{"hello world", 11, 6, 11}, // End of text
		{"", 0, 0, 0},              // Empty string
		{"word", 2, 0, 4},          // Single word
	}

	for _, tt := range tests {
		start, end := findWordBoundaries([]rune(tt.text), tt.pos)
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("findWordBoundaries(%q, %d) = (%d, %d), want (%d, %d)",
				tt.text, tt.pos, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}

// Verify Input implements Selectable
var _ Selectable = (*Input)(nil)

// Verify MultilineInput implements Selectable
var _ Selectable = (*MultilineInput)(nil)
