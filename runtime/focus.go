package runtime

// FocusScope manages focus within a layer/context.
// Each modal layer has its own FocusScope, so overlays trap focus.
type FocusScope struct {
	widgets  []Focusable
	current  int // Index of focused widget, -1 if none
	onChange func(prev Focusable, next Focusable)
}

// NewFocusScope creates a new empty focus scope.
func NewFocusScope() *FocusScope {
	return &FocusScope{current: -1}
}

// SetOnChange registers a focus change callback.
func (f *FocusScope) SetOnChange(fn func(prev Focusable, next Focusable)) {
	if f == nil {
		return
	}
	f.onChange = fn
}

// Register adds a focusable widget to the scope.
// The first registered widget receives focus if nothing is focused.
func (f *FocusScope) Register(w Focusable) {
	// Check if already registered
	for _, existing := range f.widgets {
		if existing == w {
			return
		}
	}
	f.widgets = append(f.widgets, w)

	// Auto-focus first widget
	if f.current == -1 && w.CanFocus() {
		f.current = len(f.widgets) - 1
		w.Focus()
	}
}

// Unregister removes a widget from the scope.
// If it was focused, focus moves to the next available widget.
func (f *FocusScope) Unregister(w Focusable) {
	for i, existing := range f.widgets {
		if existing == w {
			// Blur if focused
			if f.current == i {
				w.Blur()
				f.current = -1
			} else if f.current > i {
				f.current--
			}

			// Remove from slice
			f.widgets = append(f.widgets[:i], f.widgets[i+1:]...)

			// Find next focusable if needed
			if f.current == -1 && len(f.widgets) > 0 {
				f.FocusFirst()
			}
			return
		}
	}
}

// Current returns the currently focused widget, or nil.
func (f *FocusScope) Current() Focusable {
	if f.current >= 0 && f.current < len(f.widgets) {
		return f.widgets[f.current]
	}
	return nil
}

// SetFocus focuses a specific widget.
// Returns true if focus changed.
func (f *FocusScope) SetFocus(w Focusable) bool {
	for i, existing := range f.widgets {
		if existing == w && w.CanFocus() {
			return f.focusIndex(i)
		}
	}
	return false
}

// FocusFirst focuses the first focusable widget.
func (f *FocusScope) FocusFirst() bool {
	for i, w := range f.widgets {
		if w.CanFocus() {
			return f.focusIndex(i)
		}
	}
	return false
}

// FocusLast focuses the last focusable widget.
func (f *FocusScope) FocusLast() bool {
	for i := len(f.widgets) - 1; i >= 0; i-- {
		if f.widgets[i].CanFocus() {
			return f.focusIndex(i)
		}
	}
	return false
}

// FocusNext moves focus to the next focusable widget.
// Wraps around to the first widget if at the end.
// Returns true if focus changed.
func (f *FocusScope) FocusNext() bool {
	if len(f.widgets) == 0 {
		return false
	}

	start := f.current
	if start < 0 {
		start = -1
	}

	// Search forward, wrapping around
	for i := 1; i <= len(f.widgets); i++ {
		idx := (start + i) % len(f.widgets)
		if f.widgets[idx].CanFocus() {
			return f.focusIndex(idx)
		}
	}
	return false
}

// FocusPrev moves focus to the previous focusable widget.
// Wraps around to the last widget if at the beginning.
// Returns true if focus changed.
func (f *FocusScope) FocusPrev() bool {
	if len(f.widgets) == 0 {
		return false
	}

	start := f.current
	if start < 0 {
		start = len(f.widgets)
	}

	// Search backward, wrapping around
	for i := 1; i <= len(f.widgets); i++ {
		idx := (start - i + len(f.widgets)) % len(f.widgets)
		if f.widgets[idx].CanFocus() {
			return f.focusIndex(idx)
		}
	}
	return false
}

// ClearFocus removes focus from the current widget.
func (f *FocusScope) ClearFocus() {
	var prev Focusable
	if f.current >= 0 && f.current < len(f.widgets) {
		prev = f.widgets[f.current]
		prev.Blur()
	}
	f.current = -1
	if f.onChange != nil {
		f.onChange(prev, nil)
	}
}

// Reset clears focus and forgets all registered widgets.
func (f *FocusScope) Reset() {
	if f == nil {
		return
	}
	f.ClearFocus()
	f.widgets = nil
	f.current = -1
}

// Count returns the number of registered widgets.
func (f *FocusScope) Count() int {
	return len(f.widgets)
}

// focusIndex changes focus to the widget at index i.
func (f *FocusScope) focusIndex(i int) bool {
	if i == f.current {
		return false
	}

	// Blur current
	var prev Focusable
	if f.current >= 0 && f.current < len(f.widgets) {
		prev = f.widgets[f.current]
		prev.Blur()
	}

	// Focus new
	f.current = i
	var next Focusable
	if i >= 0 && i < len(f.widgets) {
		next = f.widgets[i]
		next.Focus()
	}
	if f.onChange != nil {
		f.onChange(prev, next)
	}
	return true
}
