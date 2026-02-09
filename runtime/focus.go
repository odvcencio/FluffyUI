package runtime

// FocusRestore saves focused widgets on a stack for later restoration.
// This is used by modal dialogs, popovers, and dropdowns to return focus
// to the element that triggered them when they close.
type FocusRestore struct {
	saved []Focusable
}

// Push saves a focusable widget onto the stack.
func (fr *FocusRestore) Push(current Focusable) {
	fr.saved = append(fr.saved, current)
}

// Pop removes and returns the most recently saved widget.
// Returns nil if the stack is empty.
func (fr *FocusRestore) Pop() Focusable {
	if len(fr.saved) == 0 {
		return nil
	}
	last := fr.saved[len(fr.saved)-1]
	fr.saved = fr.saved[:len(fr.saved)-1]
	return last
}

// Len returns the number of saved entries.
func (fr *FocusRestore) Len() int {
	return len(fr.saved)
}

// AutoFocusPolicy controls how FocusScope handles initial focus.
type AutoFocusPolicy int

const (
	// AutoFocusFirst focuses the first registered focusable widget (default).
	AutoFocusFirst AutoFocusPolicy = iota
	// AutoFocusLast focuses the last registered focusable widget.
	AutoFocusLast
	// AutoFocusNone disables auto-focus; apps must call SetFocus explicitly.
	AutoFocusNone
)

// FocusScope manages focus within a layer/context.
// Each modal layer has its own FocusScope, so overlays trap focus.
type FocusScope struct {
	widgets         []Focusable
	widgetSet       map[Focusable]struct{}
	current         int // Index of focused widget, -1 if none
	onChange        func(prev Focusable, next Focusable)
	autoFocusPolicy AutoFocusPolicy
}

// NewFocusScope creates a new empty focus scope with default auto-focus policy.
func NewFocusScope() *FocusScope {
	return &FocusScope{current: -1, autoFocusPolicy: AutoFocusFirst}
}

// NewFocusScopeWithPolicy creates a focus scope with the specified auto-focus policy.
func NewFocusScopeWithPolicy(policy AutoFocusPolicy) *FocusScope {
	return &FocusScope{current: -1, autoFocusPolicy: policy}
}

// SetAutoFocusPolicy changes the auto-focus policy.
func (f *FocusScope) SetAutoFocusPolicy(policy AutoFocusPolicy) {
	if f == nil {
		return
	}
	f.autoFocusPolicy = policy
}

// SetOnChange registers a focus change callback.
func (f *FocusScope) SetOnChange(fn func(prev Focusable, next Focusable)) {
	if f == nil {
		return
	}
	f.onChange = fn
}

// Register adds a focusable widget to the scope.
// Auto-focus behavior depends on the scope's AutoFocusPolicy.
func (f *FocusScope) Register(w Focusable) {
	// Initialize map lazily
	if f.widgetSet == nil {
		f.widgetSet = make(map[Focusable]struct{})
	}

	// Check if already registered (O(1) lookup)
	if _, exists := f.widgetSet[w]; exists {
		return
	}

	f.widgets = append(f.widgets, w)
	f.widgetSet[w] = struct{}{}

	// Auto-focus based on policy
	if f.autoFocusPolicy == AutoFocusNone {
		return
	}
	if f.current != -1 || !w.CanFocus() {
		return
	}
	if f.autoFocusPolicy == AutoFocusLast {
		return
	}
	// AutoFocusFirst: focus this widget (it's the first focusable)
	f.current = len(f.widgets) - 1
	w.Focus()
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

			// Remove from slice and map
			f.widgets = append(f.widgets[:i], f.widgets[i+1:]...)
			if f.widgetSet != nil {
				delete(f.widgetSet, w)
			}

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
	f.widgetSet = nil
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
