package widgets

// Selection represents a text selection range.
type Selection struct {
	Start int
	End   int
}

// IsEmpty returns true if no text is selected.
func (s Selection) IsEmpty() bool { return s.Start == s.End }

// Length returns the length of the selection.
func (s Selection) Length() int {
	if s.Start > s.End {
		return s.Start - s.End
	}
	return s.End - s.Start
}

// Normalize returns a Selection with Start <= End.
func (s Selection) Normalize() Selection {
	if s.Start > s.End {
		return Selection{Start: s.End, End: s.Start}
	}
	return s
}

// Selectable is implemented by widgets that support text selection.
type Selectable interface {
	// GetSelection returns the current selection range.
	GetSelection() Selection

	// SetSelection sets the selection range.
	SetSelection(sel Selection)

	// SelectAll selects all text.
	SelectAll()

	// SelectNone clears the selection.
	SelectNone()

	// SelectWord selects the word at the cursor position.
	SelectWord()

	// SelectLine selects the current line.
	SelectLine()

	// HasSelection returns true if text is selected.
	HasSelection() bool

	// GetSelectedText returns the currently selected text.
	GetSelectedText() string
}
