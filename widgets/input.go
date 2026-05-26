package widgets

import (
	"strings"
	"time"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/clipboard"
	"m31labs.dev/fluffyui/forms"
	"m31labs.dev/fluffyui/i18n"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	uistyle "m31labs.dev/fluffyui/style"
	"m31labs.dev/fluffyui/terminal"
)

// inputState represents the state of an input for undo/redo.
type inputState struct {
	Text           string
	CursorPos      int
	SelectionStart int
	SelectionEnd   int
}

// Input is a text input widget with cursor support.
type Input struct {
	FocusableBase

	text        strings.Builder
	cursorPos   int
	selection   Selection
	label       string
	style       backend.Style
	focusStyle  backend.Style
	placeholder string
	services    runtime.Services
	styleSet    bool
	focusSet    bool
	validators  []forms.Validator
	valErrors   []forms.ValidationError
	valMessages []string

	// History (undo/redo)
	history        *state.History[inputState]
	historyEnabled bool

	// Callbacks
	onSubmit func(text string)
	onChange func(text string)

	// Rune cache
	runeCache      []rune
	runeCacheDirty bool
}

// defaultInputGroupWindow is the default time window for grouping rapid edits.
const defaultInputGroupWindow = 500 * time.Millisecond

// NewInput creates a single-line text input with cursor, selection, and undo/redo.
// Configure with SetPlaceholder, SetOnSubmit, and SetOnChange.
func NewInput() *Input {
	input := &Input{
		style:          backend.DefaultStyle(),
		focusStyle:     backend.DefaultStyle().Bold(true),
		historyEnabled: true,
	}
	// Initialize history with default settings
	input.history = state.NewHistory(inputState{}, state.WithGroupWindow(defaultInputGroupWindow))
	input.Base.Role = accessibility.RoleTextbox
	input.syncA11y()
	return input
}

// Bind attaches app services.
func (i *Input) Bind(services runtime.Services) {
	if i == nil {
		return
	}
	i.services = services
}

// Unbind releases app services.
func (i *Input) Unbind() {
	if i == nil {
		return
	}
	i.services = runtime.Services{}
}

// SetPlaceholder sets the placeholder text shown when empty.
func (i *Input) SetPlaceholder(text string) {
	i.placeholder = text
	i.syncA11y()
}

// SetStyle sets the normal style.
func (i *Input) SetStyle(style backend.Style) {
	i.style = style
	i.styleSet = true
}

// SetFocusStyle sets the focused style.
func (i *Input) SetFocusStyle(style backend.Style) {
	i.focusStyle = style
	i.focusSet = true
}

// StyleType returns the selector type name.
func (i *Input) StyleType() string {
	return "Input"
}

// SetOnSubmit sets the callback for when Enter is pressed.
func (i *Input) SetOnSubmit(fn func(text string)) {
	if i == nil {
		return
	}
	i.onSubmit = fn
}

// OnSubmit sets the callback for when Enter is pressed.
//
// Deprecated: Use SetOnSubmit instead. This method will be removed in v1.0.
func (i *Input) OnSubmit(fn func(text string)) {
	i.SetOnSubmit(fn)
}

// SetOnChange sets the callback for when text changes.
func (i *Input) SetOnChange(fn func(text string)) {
	if i == nil {
		return
	}
	i.onChange = fn
}

// OnChange sets the callback for when the input text changes.
//
// Deprecated: Use SetOnChange instead. This method will be removed in v1.0.
func (i *Input) OnChange(fn func(text string)) {
	i.SetOnChange(fn)
}

// SetValidators updates validation rules for the input.
func (i *Input) SetValidators(validators ...forms.Validator) {
	if i == nil {
		return
	}
	i.validators = validators
}

// Validate runs validation rules and returns validation errors.
func (i *Input) Validate() []forms.ValidationError {
	if i == nil {
		return nil
	}
	errs, messages := validateValue(i.Text(), i.validators)
	i.valErrors = errs
	i.valMessages = messages
	return errs
}

// Errors returns the latest validation error messages.
func (i *Input) Errors() []string {
	if i == nil {
		return nil
	}
	if len(i.validators) > 0 {
		i.Validate()
	}
	if len(i.valMessages) == 0 {
		return nil
	}
	out := make([]string, len(i.valMessages))
	copy(out, i.valMessages)
	return out
}

// Valid reports whether validation passes.
func (i *Input) Valid() bool {
	if i == nil {
		return true
	}
	return len(i.Validate()) == 0
}

// Undo reverts to the previous state.
// Returns true if undo was successful.
func (i *Input) Undo() bool {
	if i == nil || !i.historyEnabled || i.history == nil {
		return false
	}
	s, ok := i.history.Undo()
	if !ok {
		return false
	}
	i.restoreState(s)
	return true
}

// Redo reapplies a previously undone state.
// Returns true if redo was successful.
func (i *Input) Redo() bool {
	if i == nil || !i.historyEnabled || i.history == nil {
		return false
	}
	s, ok := i.history.Redo()
	if !ok {
		return false
	}
	i.restoreState(s)
	return true
}

// CanUndo returns true if undo is available.
func (i *Input) CanUndo() bool {
	if i == nil || !i.historyEnabled || i.history == nil {
		return false
	}
	return i.history.CanUndo()
}

// CanRedo returns true if redo is available.
func (i *Input) CanRedo() bool {
	if i == nil || !i.historyEnabled || i.history == nil {
		return false
	}
	return i.history.CanRedo()
}

// SetHistoryEnabled enables or disables undo/redo support.
func (i *Input) SetHistoryEnabled(enabled bool) {
	if i == nil {
		return
	}
	i.historyEnabled = enabled
}

// SetHistoryOptions reconfigures the history with new options.
func (i *Input) SetHistoryOptions(opts ...state.HistoryOption) {
	if i == nil {
		return
	}
	// Create new history with options, preserving current state
	currentState := i.captureState()
	i.history = state.NewHistory(currentState, opts...)
}

// ClearHistory resets the undo/redo history.
func (i *Input) ClearHistory() {
	if i == nil || i.history == nil {
		return
	}
	i.history.Clear()
}

func (i *Input) captureState() inputState {
	return inputState{
		Text:           i.text.String(),
		CursorPos:      i.cursorPos,
		SelectionStart: i.selection.Start,
		SelectionEnd:   i.selection.End,
	}
}

func (i *Input) restoreState(s inputState) {
	i.text.Reset()
	i.text.WriteString(s.Text)
	i.cursorPos = s.CursorPos
	i.selection = Selection{Start: s.SelectionStart, End: s.SelectionEnd}
	i.runeCacheDirty = true
	i.syncA11y()
	i.services.Invalidate()
}

func (i *Input) pushHistory(grouped bool) {
	if !i.historyEnabled || i.history == nil {
		return
	}
	s := i.captureState()
	if grouped {
		i.history.PushGrouped(s)
	} else {
		i.history.Push(s)
	}
}

// SetLabel sets the accessibility label for the input.
func (i *Input) SetLabel(label string) {
	i.label = label
	i.syncA11y()
}

// Text returns the current input text.
func (i *Input) Text() string {
	return i.text.String()
}

// SetText sets the input text and moves cursor to end.
func (i *Input) SetText(text string) {
	i.text.Reset()
	i.text.WriteString(text)
	i.cursorPos = runeCount(text)
	i.runeCacheDirty = true
	i.syncA11y()
}

// Clear clears the input text.
func (i *Input) Clear() {
	i.text.Reset()
	i.cursorPos = 0
	i.runeCacheDirty = true
	i.syncA11y()
}

// CursorPos returns the current cursor position.
func (i *Input) CursorPos() int {
	return i.cursorPos
}

// CursorOffset returns the current cursor offset (alias for CursorPos).
func (i *Input) CursorOffset() int {
	if i == nil {
		return 0
	}
	return i.cursorPos
}

// CursorPosition returns the cursor coordinates within the input.
func (i *Input) CursorPosition() (x, y int) {
	if i == nil {
		return 0, 0
	}
	return i.cursorPos, 0
}

// SetCursorOffset moves the cursor to the given offset.
func (i *Input) SetCursorOffset(offset int) {
	if i == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	textLen := len(i.textRunes())
	if offset > textLen {
		offset = textLen
	}
	i.cursorPos = offset
	i.services.Invalidate()
}

// SetCursorPosition moves the cursor to the given coordinates.
func (i *Input) SetCursorPosition(x, y int) {
	_ = y
	i.SetCursorOffset(x)
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (i *Input) CursorWordLeft() {
	if i == nil {
		return
	}
	i.cursorPos = i.wordBoundaryLeft()
	i.services.Invalidate()
}

// CursorWordRight moves the cursor to the next word boundary.
func (i *Input) CursorWordRight() {
	if i == nil {
		return
	}
	i.cursorPos = i.wordBoundaryRight()
	i.services.Invalidate()
}

// Measure returns the size needed for the input.
func (i *Input) Measure(constraints runtime.Constraints) runtime.Size {
	return i.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Input is typically 1 line tall, fills available width
		return runtime.Size{
			Width:  contentConstraints.MaxWidth,
			Height: 1,
		}
	})
}

// Render draws the input field.
func (i *Input) Render(ctx runtime.RenderContext) {
	outer := i.bounds
	content := i.ContentBounds()
	if outer.Width == 0 || outer.Height == 0 {
		return
	}

	style := i.style
	resolved := ctx.ResolveStyle(i)
	if !resolved.IsZero() {
		final := resolved
		if i.styleSet {
			final = final.Merge(uistyle.FromBackend(i.style))
		}
		if i.focused && i.focusSet {
			final = final.Merge(uistyle.FromBackend(i.focusStyle))
		}
		style = final.ToBackend()
	} else if i.focused {
		style = i.focusStyle
	}

	// Clear the input area
	ctx.Buffer.Fill(outer, ' ', style)

	if content.Width == 0 || content.Height == 0 {
		return
	}

	text := i.text.String()
	runes := []rune(text)
	textLen := len(runes)

	// Show placeholder if empty and not focused
	if text == "" && !i.focused && i.placeholder != "" {
		placeholderStyle := style.Dim(true)
		display := i.placeholder
		if runeCount(display) > content.Width {
			displayRunes := []rune(display)
			display = string(displayRunes[:content.Width])
		}
		ctx.Buffer.SetString(content.X, content.Y, display, placeholderStyle)
		return
	}

	// Detect text direction for BiDi support
	dir := i18n.DetectDirection(text)

	// Calculate visible portion of text
	// Scroll so cursor is always visible
	visibleStart := 0
	if i.cursorPos >= content.Width {
		visibleStart = i.cursorPos - content.Width + 1
	}

	visibleEnd := visibleStart + content.Width
	if visibleEnd > textLen {
		visibleEnd = textLen
	}

	var visibleRunes []rune
	if visibleStart < textLen {
		visibleRunes = runes[visibleStart:visibleEnd]
	}

	// Apply BiDi reordering for display
	displayRunes := visibleRunes
	var cursorMap i18n.CursorMap
	if dir == i18n.DirectionRTL && len(visibleRunes) > 0 {
		logicalStr := string(visibleRunes)
		displayStr := i18n.BidiReorder(logicalStr, i18n.DirectionRTL)
		displayRunes = []rune(displayStr)
		cursorMap = i18n.BuildCursorMap(logicalStr, displayStr)
	}

	// Right-align RTL text
	xOffset := 0
	if dir == i18n.DirectionRTL {
		displayW := textWidth(string(displayRunes))
		if displayW < content.Width {
			xOffset = content.Width - displayW
		}
	}

	// Draw text with selection highlighting
	selectionStyle := style.Reverse(true)
	sel := i.selection.Normalize()
	hasSelection := !i.selection.IsEmpty()

	for idx, ch := range displayRunes {
		screenX := content.X + xOffset + idx
		charStyle := style

		// Map display index back to logical index for selection highlighting
		logicalIdx := idx
		if dir == i18n.DirectionRTL {
			logicalIdx = cursorMap.VisualToLogical(idx)
		}
		textIdx := visibleStart + logicalIdx

		// Highlight if within selection
		if hasSelection && textIdx >= sel.Start && textIdx < sel.End {
			charStyle = selectionStyle
		}

		ctx.Buffer.Set(screenX, content.Y, ch, charStyle)
	}

	// Draw cursor if focused (by inverting the cell)
	if i.focused {
		visualCursorPos := i.cursorPos - visibleStart
		if dir == i18n.DirectionRTL && len(visibleRunes) > 0 {
			if visualCursorPos >= 0 && visualCursorPos < len(visibleRunes) {
				visualCursorPos = cursorMap.LogicalToVisual(visualCursorPos)
			}
			// For RTL, when cursor is at the end of text, place it at the left edge
			if i.cursorPos >= textLen {
				visualCursorPos = -1 // will be adjusted with xOffset below
				// Place cursor at the position before the first visual character
				if xOffset > 0 {
					cursorX := content.X + xOffset - 1
					if cursorX >= content.X && cursorX < content.X+content.Width {
						cursorStyle := style.Reverse(true)
						ctx.Buffer.Set(cursorX, content.Y, ' ', cursorStyle)
					}
					return
				}
			}
		}
		cursorX := content.X + xOffset + visualCursorPos
		if cursorX >= content.X && cursorX < content.X+content.Width {
			var cursorChar rune = ' '
			if dir == i18n.DirectionRTL && len(displayRunes) > 0 && visualCursorPos >= 0 && visualCursorPos < len(displayRunes) {
				cursorChar = displayRunes[visualCursorPos]
			} else if i.cursorPos < textLen {
				cursorChar = runes[i.cursorPos]
			}
			cursorStyle := style.Reverse(true)
			ctx.Buffer.Set(cursorX, content.Y, cursorChar, cursorStyle)
		}
	}
}

// HandleMessage processes keyboard input.
func (i *Input) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if !i.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlZ:
		// Ctrl+Shift+Z is redo
		if key.Shift {
			if i.Redo() {
				return runtime.Handled()
			}
		} else {
			if i.Undo() {
				return runtime.Handled()
			}
		}
		return runtime.Handled()
	case terminal.KeyCtrlY:
		if i.Redo() {
			return runtime.Handled()
		}
		return runtime.Handled()
	case terminal.KeyCtrlC:
		if i.copyToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlX:
		if i.cutToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlV:
		if i.HasSelection() {
			i.deleteSelection()
		}
		if i.pasteFromClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		if len(i.validators) > 0 {
			errs := i.Errors()
			if len(errs) > 0 {
				if announcer := i.services.Announcer(); announcer != nil {
					announcer.Announce(strings.Join(errs, ", "), accessibility.PriorityAssertive)
				}
			}
		}
		if i.onSubmit != nil {
			text := i.text.String()
			i.onSubmit(text)
		}
		return runtime.WithCommand(runtime.Submit{Text: i.text.String()})

	case terminal.KeyBackspace:
		if i.HasSelection() {
			i.pushHistory(true) // grouped
			i.deleteSelection()
			return runtime.Handled()
		}
		if i.cursorPos > 0 {
			i.pushHistory(true) // grouped
			runes := i.textRunes()
			if i.cursorPos > len(runes) {
				i.cursorPos = len(runes)
			}
			runes = append(runes[:i.cursorPos-1], runes[i.cursorPos:]...)
			i.setTextRunes(runes)
			i.cursorPos--
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyDelete:
		if i.HasSelection() {
			i.pushHistory(true) // grouped
			i.deleteSelection()
			return runtime.Handled()
		}
		runes := i.textRunes()
		if i.cursorPos > len(runes) {
			i.cursorPos = len(runes)
		}
		if i.cursorPos < len(runes) {
			i.pushHistory(true) // grouped
			runes = append(runes[:i.cursorPos], runes[i.cursorPos+1:]...)
			i.setTextRunes(runes)
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if i.collapseSelection(true) {
			return runtime.Handled()
		}
		if key.Ctrl {
			// Word left
			i.cursorPos = i.wordBoundaryLeft()
		} else if i.cursorPos > 0 {
			i.cursorPos--
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if i.collapseSelection(false) {
			return runtime.Handled()
		}
		if key.Ctrl {
			// Word right
			i.cursorPos = i.wordBoundaryRight()
		} else if i.cursorPos < len(i.textRunes()) {
			i.cursorPos++
		}
		return runtime.Handled()

	case terminal.KeyHome:
		if i.HasSelection() {
			i.selection = Selection{}
		}
		i.cursorPos = 0
		return runtime.Handled()

	case terminal.KeyEnd:
		if i.HasSelection() {
			i.selection = Selection{}
		}
		i.cursorPos = len(i.textRunes())
		return runtime.Handled()

	case terminal.KeyRune:
		// Insert character
		i.pushHistory(true) // grouped
		if i.HasSelection() {
			i.deleteSelection()
		}
		runes := i.textRunes()
		if i.cursorPos > len(runes) {
			i.cursorPos = len(runes)
		}
		runes = append(runes[:i.cursorPos], append([]rune{key.Rune}, runes[i.cursorPos:]...)...)
		i.setTextRunes(runes)
		i.cursorPos++
		i.notifyChange()
		return runtime.Handled()

	case terminal.KeyTab:
		// Tab might be focus navigation
		if key.Shift {
			return runtime.WithCommand(runtime.FocusPrev{})
		}
		return runtime.WithCommand(runtime.FocusNext{})

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

func (i *Input) notifyChange() {
	i.syncA11y()
	if i.onChange != nil {
		i.onChange(i.text.String())
	}
}

func (i *Input) syncA11y() {
	if i == nil {
		return
	}
	label := strings.TrimSpace(i.label)
	if label == "" {
		label = strings.TrimSpace(i.placeholder)
	}
	if label == "" {
		label = "Input"
	}
	if i.Base.Role == "" {
		i.Base.Role = accessibility.RoleTextbox
	}
	i.Base.Label = label
	i.Base.Placeholder = i.placeholder
	i.Base.Value = &accessibility.ValueInfo{Text: i.Text()}
}

// ClipboardCopy returns selected text, or all text if no selection.
func (i *Input) ClipboardCopy() (string, bool) {
	if i == nil {
		return "", false
	}
	if i.HasSelection() {
		return i.GetSelectedText(), true
	}
	return i.text.String(), true
}

// ClipboardCut returns selected text and deletes it, or all text if no selection.
func (i *Input) ClipboardCut() (string, bool) {
	if i == nil {
		return "", false
	}
	i.pushHistory(true) // grouped
	if i.HasSelection() {
		text := i.GetSelectedText()
		i.deleteSelection()
		return text, true
	}
	text := i.text.String()
	i.Clear()
	i.notifyChange()
	return text, true
}

// deleteSelection removes the selected text and clears the selection.
func (i *Input) deleteSelection() {
	if i == nil || i.selection.IsEmpty() {
		return
	}
	sel := i.selection.Normalize()
	runes := i.textRunes()
	if sel.End > len(runes) {
		sel.End = len(runes)
	}
	if sel.Start > len(runes) {
		sel.Start = len(runes)
	}
	runes = append(runes[:sel.Start], runes[sel.End:]...)
	i.setTextRunes(runes)
	i.cursorPos = sel.Start
	i.selection = Selection{}
	i.notifyChange()
}

// ClipboardPaste inserts text at the cursor.
func (i *Input) ClipboardPaste(text string) bool {
	if i == nil || text == "" {
		return false
	}
	// Large pastes (> 100 chars) are not grouped
	grouped := len(text) <= 100
	i.pushHistory(grouped)
	if i.HasSelection() {
		i.deleteSelection()
	}
	i.insertText(text)
	return true
}

func (i *Input) copyToClipboard() bool {
	cb := i.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := i.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (i *Input) cutToClipboard() bool {
	cb := i.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := i.ClipboardCut()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (i *Input) pasteFromClipboard() bool {
	cb := i.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, err := cb.Read()
	if err != nil || text == "" {
		return false
	}
	return i.ClipboardPaste(text)
}

func (i *Input) insertText(text string) {
	if text == "" {
		return
	}
	current := i.textRunes()
	if i.cursorPos > len(current) {
		i.cursorPos = len(current)
	}
	insertRunes := []rune(text)
	current = append(current[:i.cursorPos], append(insertRunes, current[i.cursorPos:]...)...)
	i.setTextRunes(current)
	i.cursorPos += len(insertRunes)
	i.notifyChange()
}

func (i *Input) collapseSelection(toStart bool) bool {
	if i == nil || i.selection.IsEmpty() {
		return false
	}
	sel := i.selection.Normalize()
	if toStart {
		i.cursorPos = sel.Start
	} else {
		i.cursorPos = sel.End
	}
	i.selection = Selection{}
	return true
}

func (i *Input) textRunes() []rune {
	if i == nil {
		return nil
	}
	if !i.runeCacheDirty && i.runeCache != nil {
		return i.runeCache
	}
	i.runeCache = []rune(i.text.String())
	i.runeCacheDirty = false
	return i.runeCache
}

func (i *Input) setTextRunes(runes []rune) {
	if i == nil {
		return
	}
	i.text.Reset()
	i.text.WriteString(string(runes))
	i.runeCacheDirty = true
}

var _ clipboard.Target = (*Input)(nil)

// Selection methods - implements Selectable interface

// GetSelection returns the current selection range.
func (i *Input) GetSelection() Selection {
	if i == nil {
		return Selection{}
	}
	return i.selection
}

// SetSelection sets the selection range.
func (i *Input) SetSelection(sel Selection) {
	if i == nil {
		return
	}
	textLen := len(i.textRunes())
	// Clamp to valid range
	if sel.Start < 0 {
		sel.Start = 0
	}
	if sel.End < 0 {
		sel.End = 0
	}
	if sel.Start > textLen {
		sel.Start = textLen
	}
	if sel.End > textLen {
		sel.End = textLen
	}
	i.selection = sel
	i.services.Invalidate()
}

// SelectAll selects all text.
func (i *Input) SelectAll() {
	if i == nil {
		return
	}
	i.selection = Selection{Start: 0, End: len(i.textRunes())}
	i.services.Invalidate()
}

// SelectNone clears the selection.
func (i *Input) SelectNone() {
	if i == nil {
		return
	}
	i.selection = Selection{}
	i.services.Invalidate()
}

// SelectWord selects the word at the cursor position.
func (i *Input) SelectWord() {
	if i == nil {
		return
	}
	text := i.text.String()
	if len(text) == 0 {
		return
	}
	start, end := findWordBoundaries([]rune(text), i.cursorPos)
	i.selection = Selection{Start: start, End: end}
	i.services.Invalidate()
}

// SelectLine selects the entire line (all text for single-line input).
func (i *Input) SelectLine() {
	i.SelectAll() // Single-line input has only one line
}

// HasSelection returns true if text is selected.
func (i *Input) HasSelection() bool {
	if i == nil {
		return false
	}
	return !i.selection.IsEmpty()
}

// GetSelectedText returns the currently selected text.
func (i *Input) GetSelectedText() string {
	if i == nil || i.selection.IsEmpty() {
		return ""
	}
	sel := i.selection.Normalize()
	runes := i.textRunes()
	if sel.End > len(runes) {
		sel.End = len(runes)
	}
	if sel.Start > len(runes) {
		return ""
	}
	return string(runes[sel.Start:sel.End])
}

var _ Selectable = (*Input)(nil)

func runeCount(text string) int {
	return len([]rune(text))
}

// findWordBoundaries returns the start and end positions of the word at pos.
func findWordBoundaries(text []rune, pos int) (start, end int) {
	if len(text) == 0 {
		return 0, 0
	}
	if pos < 0 {
		pos = 0
	}
	if pos > len(text) {
		pos = len(text)
	}

	// Find start of word
	start = pos
	for start > 0 && !isWordSeparator(text[start-1]) {
		start--
	}

	// Find end of word
	end = pos
	for end < len(text) && !isWordSeparator(text[end]) {
		end++
	}

	return start, end
}

func (i *Input) wordBoundaryLeft() int {
	runes := i.textRunes()
	if len(runes) == 0 {
		return 0
	}
	pos := i.cursorPos - 1
	if pos < 0 {
		return 0
	}
	if pos >= len(runes) {
		pos = len(runes) - 1
	}

	// Skip whitespace
	for pos > 0 && runes[pos] == ' ' {
		pos--
	}
	// Skip word characters
	for pos > 0 && runes[pos-1] != ' ' {
		pos--
	}
	return pos
}

func (i *Input) wordBoundaryRight() int {
	runes := i.textRunes()
	pos := i.cursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	// Skip word characters
	for pos < len(runes) && runes[pos] != ' ' {
		pos++
	}
	// Skip whitespace
	for pos < len(runes) && runes[pos] == ' ' {
		pos++
	}
	return pos
}

func multilineWordBoundaryLeft(text []rune, cursor int) int {
	if cursor <= 0 {
		return 0
	}
	if cursor > len(text) {
		cursor = len(text)
	}
	pos := cursor - 1
	for pos > 0 && isWordSeparator(text[pos]) {
		pos--
	}
	for pos > 0 && !isWordSeparator(text[pos-1]) {
		pos--
	}
	return pos
}

func multilineWordBoundaryRight(text []rune, cursor int) int {
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(text) {
		return len(text)
	}
	pos := cursor
	for pos < len(text) && !isWordSeparator(text[pos]) {
		pos++
	}
	for pos < len(text) && isWordSeparator(text[pos]) {
		pos++
	}
	return pos
}

func isWordSeparator(ch rune) bool {
	switch ch {
	case ' ', '\n', '\t':
		return true
	default:
		return false
	}
}

// MultilineInput is a text input that supports multiple lines.
type MultilineInput struct {
	FocusableBase

	lines      []string
	cursorX    int
	cursorY    int
	scrollY    int // First visible line
	selection  Selection
	label      string
	style      backend.Style
	focusStyle backend.Style
	services   runtime.Services
	styleSet   bool
	focusSet   bool

	onSubmit func(text string)
	onChange func(text string)

	validators  []forms.Validator
	valErrors   []forms.ValidationError
	valMessages []string
}

// NewMultilineInput creates a new multiline input widget.
func NewMultilineInput() *MultilineInput {
	input := &MultilineInput{
		lines:      []string{""},
		label:      "Multiline Input",
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle(),
	}
	input.Base.Role = accessibility.RoleTextbox
	input.Base.State.Multiline = true
	input.syncA11y()
	return input
}

// StyleType returns the selector type name.
func (m *MultilineInput) StyleType() string {
	return "MultilineInput"
}

// Bind attaches app services.
func (m *MultilineInput) Bind(services runtime.Services) {
	if m == nil {
		return
	}
	m.services = services
}

// Unbind releases app services.
func (m *MultilineInput) Unbind() {
	if m == nil {
		return
	}
	m.services = runtime.Services{}
}

// Text returns the full text content.
func (m *MultilineInput) Text() string {
	return strings.Join(m.lines, "\n")
}

// SetText sets the content.
func (m *MultilineInput) SetText(text string) {
	m.lines = strings.Split(text, "\n")
	if len(m.lines) == 0 {
		m.lines = []string{""}
	}
	m.cursorY = len(m.lines) - 1
	m.cursorX = runeCount(m.lines[m.cursorY])
	m.syncA11y()
}

// CursorPosition returns the cursor coordinates within the input.
func (m *MultilineInput) CursorPosition() (x, y int) {
	if m == nil {
		return 0, 0
	}
	return m.cursorX, m.cursorY
}

// CursorOffset returns the cursor offset in the full text.
func (m *MultilineInput) CursorOffset() int {
	if m == nil {
		return 0
	}
	offset := 0
	for i := 0; i < m.cursorY && i < len(m.lines); i++ {
		offset += runeCount(m.lines[i]) + 1
	}
	offset += m.cursorX
	return offset
}

// SetCursorPosition moves the cursor to the given coordinates.
func (m *MultilineInput) SetCursorPosition(x, y int) {
	if m == nil || len(m.lines) == 0 {
		return
	}
	if y < 0 {
		y = 0
	}
	if y >= len(m.lines) {
		y = len(m.lines) - 1
	}
	line := m.lines[y]
	lineRunes := []rune(line)
	if x < 0 {
		x = 0
	}
	if x > len(lineRunes) {
		x = len(lineRunes)
	}
	m.cursorX = x
	m.cursorY = y
	m.ensureCursorVisible()
	m.services.Invalidate()
}

// SetCursorOffset moves the cursor to the given offset.
func (m *MultilineInput) SetCursorOffset(offset int) {
	if m == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	total := runeCount(m.Text())
	if offset > total {
		offset = total
	}
	remaining := offset
	for i, line := range m.lines {
		lineLen := runeCount(line)
		if remaining <= lineLen {
			m.cursorY = i
			m.cursorX = remaining
			m.ensureCursorVisible()
			m.services.Invalidate()
			return
		}
		remaining -= lineLen
		if remaining == 0 {
			m.cursorY = i
			m.cursorX = lineLen
			m.ensureCursorVisible()
			m.services.Invalidate()
			return
		}
		remaining--
	}
	m.cursorY = len(m.lines) - 1
	m.cursorX = runeCount(m.lines[m.cursorY])
	m.ensureCursorVisible()
	m.services.Invalidate()
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (m *MultilineInput) CursorWordLeft() {
	if m == nil {
		return
	}
	offset := multilineWordBoundaryLeft([]rune(m.Text()), m.CursorOffset())
	m.SetCursorOffset(offset)
}

// CursorWordRight moves the cursor to the next word boundary.
func (m *MultilineInput) CursorWordRight() {
	if m == nil {
		return
	}
	offset := multilineWordBoundaryRight([]rune(m.Text()), m.CursorOffset())
	m.SetCursorOffset(offset)
}

// Clear clears all content.
func (m *MultilineInput) Clear() {
	m.lines = []string{""}
	m.cursorX = 0
	m.cursorY = 0
	m.scrollY = 0
	m.syncA11y()
}

// SetOnSubmit sets the callback (Ctrl+Enter to submit).
func (m *MultilineInput) SetOnSubmit(fn func(text string)) {
	if m == nil {
		return
	}
	m.onSubmit = fn
}

// OnSubmit sets the callback for when Ctrl+Enter is pressed.
//
// Deprecated: Use SetOnSubmit instead. This method will be removed in v1.0.
func (m *MultilineInput) OnSubmit(fn func(text string)) {
	m.SetOnSubmit(fn)
}

// SetOnChange sets the callback for when text changes.
func (m *MultilineInput) SetOnChange(fn func(text string)) {
	if m == nil {
		return
	}
	m.onChange = fn
}

// OnChange sets the callback for when the multiline input text changes.
//
// Deprecated: Use SetOnChange instead. This method will be removed in v1.0.
func (m *MultilineInput) OnChange(fn func(text string)) {
	m.SetOnChange(fn)
}

// SetValidators updates validation rules for the multiline input.
func (m *MultilineInput) SetValidators(validators ...forms.Validator) {
	if m == nil {
		return
	}
	m.validators = validators
}

// Validate runs validation rules and returns validation errors.
func (m *MultilineInput) Validate() []forms.ValidationError {
	if m == nil {
		return nil
	}
	errs, messages := validateValue(m.Text(), m.validators)
	m.valErrors = errs
	m.valMessages = messages
	return errs
}

// Errors returns the latest validation error messages.
func (m *MultilineInput) Errors() []string {
	if m == nil {
		return nil
	}
	if len(m.validators) > 0 {
		m.Validate()
	}
	if len(m.valMessages) == 0 {
		return nil
	}
	out := make([]string, len(m.valMessages))
	copy(out, m.valMessages)
	return out
}

// Valid reports whether validation passes.
func (m *MultilineInput) Valid() bool {
	if m == nil {
		return true
	}
	return len(m.Validate()) == 0
}

// SetLabel sets the accessibility label.
func (m *MultilineInput) SetLabel(label string) {
	m.label = label
	m.syncA11y()
}

// SetStyle sets the normal style.
func (m *MultilineInput) SetStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.style = style
	m.styleSet = true
}

// SetFocusStyle sets the focused style.
func (m *MultilineInput) SetFocusStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.focusStyle = style
	m.focusSet = true
}

// Measure returns the preferred size.
func (m *MultilineInput) Measure(constraints runtime.Constraints) runtime.Size {
	return m.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Prefer to be at least 3 lines tall, up to content or max
		height := len(m.lines)
		if height < 3 {
			height = 3
		}
		return contentConstraints.Constrain(runtime.Size{
			Width:  contentConstraints.MaxWidth,
			Height: height,
		})
	})
}

// Render draws the multiline input.
func (m *MultilineInput) Render(ctx runtime.RenderContext) {
	outer := m.bounds
	content := m.ContentBounds()
	if outer.Width == 0 || outer.Height == 0 {
		return
	}

	style := m.style
	resolved := ctx.ResolveStyle(m)
	if !resolved.IsZero() {
		final := resolved
		if m.styleSet {
			final = final.Merge(uistyle.FromBackend(m.style))
		}
		if m.focused && m.focusSet {
			final = final.Merge(uistyle.FromBackend(m.focusStyle))
		}
		style = final.ToBackend()
	} else if m.focused {
		style = m.focusStyle
	}

	// Clear area
	ctx.Buffer.Fill(outer, ' ', style)

	if content.Width == 0 || content.Height == 0 {
		return
	}

	// Draw visible lines
	selectionStyle := style.Reverse(true)
	sel := m.selection.Normalize()
	hasSelection := !m.selection.IsEmpty()
	lineOffset := 0
	if m.scrollY > 0 {
		for i := 0; i < m.scrollY && i < len(m.lines); i++ {
			lineOffset += runeCount(m.lines[i]) + 1
		}
	}
	for i := 0; i < content.Height; i++ {
		lineIdx := m.scrollY + i
		if lineIdx >= len(m.lines) {
			break
		}

		line := m.lines[lineIdx]
		lineRunes := []rune(line)
		visibleRunes := lineRunes
		if len(visibleRunes) > content.Width {
			visibleRunes = visibleRunes[:content.Width]
		}
		for idx, ch := range visibleRunes {
			charStyle := style
			if hasSelection {
				textIdx := lineOffset + idx
				if textIdx >= sel.Start && textIdx < sel.End {
					charStyle = selectionStyle
				}
			}
			ctx.Buffer.Set(content.X+idx, content.Y+i, ch, charStyle)
		}
		lineOffset += len(lineRunes) + 1
	}

	// Draw cursor
	if m.focused {
		cursorScreenY := m.cursorY - m.scrollY
		if cursorScreenY >= 0 && cursorScreenY < content.Height {
			cursorX := content.X + m.cursorX
			if cursorX >= content.X && cursorX < content.X+content.Width {
				var ch rune = ' '
				if m.cursorY < len(m.lines) {
					lineRunes := []rune(m.lines[m.cursorY])
					if m.cursorX < len(lineRunes) {
						ch = lineRunes[m.cursorX]
					}
				}
				ctx.Buffer.Set(cursorX, content.Y+cursorScreenY, ch, style.Reverse(true))
			}
		}
	}
}

// HandleMessage processes input for multiline editing.
func (m *MultilineInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if !m.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlC:
		if m.copyToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlX:
		if m.cutToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlV:
		if m.HasSelection() {
			m.deleteSelection()
		}
		if m.pasteFromClipboard() {
			return runtime.Handled()
		}

	case terminal.KeyEnter:
		if m.HasSelection() {
			m.deleteSelection()
		}
		if key.Ctrl && m.onSubmit != nil {
			m.onSubmit(m.Text())
			return runtime.WithCommand(runtime.Submit{Text: m.Text()})
		}
		// Insert newline
		line := m.lines[m.cursorY]
		lineRunes := []rune(line)
		if m.cursorX > len(lineRunes) {
			m.cursorX = len(lineRunes)
		}
		m.lines[m.cursorY] = string(lineRunes[:m.cursorX])
		newLine := string(lineRunes[m.cursorX:])
		m.lines = append(m.lines[:m.cursorY+1], append([]string{newLine}, m.lines[m.cursorY+1:]...)...)
		m.cursorY++
		m.cursorX = 0
		m.ensureCursorVisible()
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyBackspace:
		if m.HasSelection() {
			m.deleteSelection()
			return runtime.Handled()
		}
		if m.cursorX > 0 {
			line := m.lines[m.cursorY]
			lineRunes := []rune(line)
			if m.cursorX > len(lineRunes) {
				m.cursorX = len(lineRunes)
			}
			lineRunes = append(lineRunes[:m.cursorX-1], lineRunes[m.cursorX:]...)
			m.lines[m.cursorY] = string(lineRunes)
			m.cursorX--
		} else if m.cursorY > 0 {
			// Join with previous line
			prevLine := m.lines[m.cursorY-1]
			m.cursorX = runeCount(prevLine)
			m.lines[m.cursorY-1] = prevLine + m.lines[m.cursorY]
			m.lines = append(m.lines[:m.cursorY], m.lines[m.cursorY+1:]...)
			m.cursorY--
		}
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyDelete:
		if m.HasSelection() {
			m.deleteSelection()
			return runtime.Handled()
		}
		line := m.lines[m.cursorY]
		lineRunes := []rune(line)
		if m.cursorX > len(lineRunes) {
			m.cursorX = len(lineRunes)
		}
		if m.cursorX < len(lineRunes) {
			lineRunes = append(lineRunes[:m.cursorX], lineRunes[m.cursorX+1:]...)
			m.lines[m.cursorY] = string(lineRunes)
		} else if m.cursorY < len(m.lines)-1 {
			m.lines[m.cursorY] = m.lines[m.cursorY] + m.lines[m.cursorY+1]
			m.lines = append(m.lines[:m.cursorY+1], m.lines[m.cursorY+2:]...)
		}
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyUp:
		if m.collapseSelection(true) {
			return runtime.Handled()
		}
		if m.cursorY > 0 {
			m.cursorY--
			lineLen := runeCount(m.lines[m.cursorY])
			if m.cursorX > lineLen {
				m.cursorX = lineLen
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if m.collapseSelection(false) {
			return runtime.Handled()
		}
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			lineLen := runeCount(m.lines[m.cursorY])
			if m.cursorX > lineLen {
				m.cursorX = lineLen
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if m.collapseSelection(true) {
			return runtime.Handled()
		}
		if m.cursorX > 0 {
			m.cursorX--
		} else if m.cursorY > 0 {
			m.cursorY--
			m.cursorX = runeCount(m.lines[m.cursorY])
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if m.collapseSelection(false) {
			return runtime.Handled()
		}
		if m.cursorX < runeCount(m.lines[m.cursorY]) {
			m.cursorX++
		} else if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.cursorX = 0
		}
		return runtime.Handled()

	case terminal.KeyRune:
		if m.HasSelection() {
			m.deleteSelection()
		}
		line := m.lines[m.cursorY]
		lineRunes := []rune(line)
		if m.cursorX > len(lineRunes) {
			m.cursorX = len(lineRunes)
		}
		lineRunes = append(lineRunes[:m.cursorX], append([]rune{key.Rune}, lineRunes[m.cursorX:]...)...)
		m.lines[m.cursorY] = string(lineRunes)
		m.cursorX++
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

func (m *MultilineInput) ensureCursorVisible() {
	content := m.ContentBounds()
	if content.Height <= 0 {
		m.scrollY = 0
		return
	}
	if m.cursorY < m.scrollY {
		m.scrollY = m.cursorY
	} else if m.cursorY >= m.scrollY+content.Height {
		m.scrollY = m.cursorY - content.Height + 1
	}
}

func (m *MultilineInput) notifyChange() {
	m.syncA11y()
	if m.onChange != nil {
		m.onChange(m.Text())
	}
}

func (m *MultilineInput) syncA11y() {
	if m == nil {
		return
	}
	label := strings.TrimSpace(m.label)
	if label == "" {
		label = "Multiline Input"
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleTextbox
	}
	m.Base.Label = label
	m.Base.Value = &accessibility.ValueInfo{Text: m.Text()}
}

// ClipboardCopy returns selected text, or all text if no selection.
func (m *MultilineInput) ClipboardCopy() (string, bool) {
	if m == nil {
		return "", false
	}
	if m.HasSelection() {
		return m.GetSelectedText(), true
	}
	return m.Text(), true
}

// ClipboardCut returns selected text and deletes it, or all text if no selection.
func (m *MultilineInput) ClipboardCut() (string, bool) {
	if m == nil {
		return "", false
	}
	if m.HasSelection() {
		text := m.GetSelectedText()
		m.deleteSelection()
		return text, true
	}
	text := m.Text()
	m.Clear()
	m.notifyChange()
	return text, true
}

// deleteSelection removes the selected text and clears the selection.
func (m *MultilineInput) deleteSelection() {
	if m == nil || m.selection.IsEmpty() {
		return
	}
	sel := m.selection.Normalize()
	runes := []rune(m.Text())
	if sel.End > len(runes) {
		sel.End = len(runes)
	}
	if sel.Start > len(runes) {
		sel.Start = len(runes)
	}
	runes = append(runes[:sel.Start], runes[sel.End:]...)
	m.SetText(string(runes))
	m.SetCursorOffset(sel.Start)
	m.selection = Selection{}
	m.notifyChange()
}

// ClipboardPaste inserts text at the cursor.
func (m *MultilineInput) ClipboardPaste(text string) bool {
	if m == nil || text == "" {
		return false
	}
	if m.HasSelection() {
		m.deleteSelection()
	}
	m.insertText(text)
	return true
}

func (m *MultilineInput) copyToClipboard() bool {
	cb := m.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := m.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (m *MultilineInput) cutToClipboard() bool {
	cb := m.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := m.ClipboardCut()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (m *MultilineInput) pasteFromClipboard() bool {
	cb := m.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, err := cb.Read()
	if err != nil || text == "" {
		return false
	}
	return m.ClipboardPaste(text)
}

func (m *MultilineInput) insertText(text string) {
	if text == "" {
		return
	}
	if len(m.lines) == 0 {
		m.lines = []string{""}
		m.cursorX = 0
		m.cursorY = 0
	}
	parts := strings.Split(text, "\n")
	line := m.lines[m.cursorY]
	lineRunes := []rune(line)
	if m.cursorX > len(lineRunes) {
		m.cursorX = len(lineRunes)
	}
	prefix := string(lineRunes[:m.cursorX])
	suffix := string(lineRunes[m.cursorX:])

	if len(parts) == 1 {
		m.lines[m.cursorY] = prefix + parts[0] + suffix
		m.cursorX += runeCount(parts[0])
		m.notifyChange()
		return
	}

	first := prefix + parts[0]
	last := parts[len(parts)-1] + suffix
	middle := parts[1 : len(parts)-1]

	newLines := make([]string, 0, len(m.lines)+len(parts)-1)
	newLines = append(newLines, m.lines[:m.cursorY]...)
	newLines = append(newLines, first)
	newLines = append(newLines, middle...)
	newLines = append(newLines, last)
	if m.cursorY+1 < len(m.lines) {
		newLines = append(newLines, m.lines[m.cursorY+1:]...)
	}
	m.lines = newLines
	m.cursorY += len(parts) - 1
	m.cursorX = runeCount(parts[len(parts)-1])
	m.ensureCursorVisible()
	m.notifyChange()
}

func (m *MultilineInput) collapseSelection(toStart bool) bool {
	if m == nil || m.selection.IsEmpty() {
		return false
	}
	sel := m.selection.Normalize()
	if toStart {
		m.SetCursorOffset(sel.Start)
	} else {
		m.SetCursorOffset(sel.End)
	}
	m.selection = Selection{}
	return true
}

var _ clipboard.Target = (*MultilineInput)(nil)

// Selection methods - implements Selectable interface

// GetSelection returns the current selection range (character offsets).
func (m *MultilineInput) GetSelection() Selection {
	if m == nil {
		return Selection{}
	}
	return m.selection
}

// SetSelection sets the selection range (character offsets).
func (m *MultilineInput) SetSelection(sel Selection) {
	if m == nil {
		return
	}
	textLen := runeCount(m.Text())
	// Clamp to valid range
	if sel.Start < 0 {
		sel.Start = 0
	}
	if sel.End < 0 {
		sel.End = 0
	}
	if sel.Start > textLen {
		sel.Start = textLen
	}
	if sel.End > textLen {
		sel.End = textLen
	}
	m.selection = sel
	m.services.Invalidate()
}

// SelectAll selects all text.
func (m *MultilineInput) SelectAll() {
	if m == nil {
		return
	}
	m.selection = Selection{Start: 0, End: runeCount(m.Text())}
	m.services.Invalidate()
}

// SelectNone clears the selection.
func (m *MultilineInput) SelectNone() {
	if m == nil {
		return
	}
	m.selection = Selection{}
	m.services.Invalidate()
}

// SelectWord selects the word at the cursor position.
func (m *MultilineInput) SelectWord() {
	if m == nil {
		return
	}
	text := m.Text()
	if len(text) == 0 {
		return
	}
	offset := m.CursorOffset()
	start, end := findWordBoundaries([]rune(text), offset)
	m.selection = Selection{Start: start, End: end}
	m.services.Invalidate()
}

// SelectLine selects the current line.
func (m *MultilineInput) SelectLine() {
	if m == nil || len(m.lines) == 0 {
		return
	}
	// Calculate start offset (beginning of current line)
	start := 0
	for i := 0; i < m.cursorY && i < len(m.lines); i++ {
		start += runeCount(m.lines[i]) + 1 // +1 for newline
	}
	// Calculate end offset (end of current line)
	end := start + runeCount(m.lines[m.cursorY])
	m.selection = Selection{Start: start, End: end}
	m.services.Invalidate()
}

// HasSelection returns true if text is selected.
func (m *MultilineInput) HasSelection() bool {
	if m == nil {
		return false
	}
	return !m.selection.IsEmpty()
}

// GetSelectedText returns the currently selected text.
func (m *MultilineInput) GetSelectedText() string {
	if m == nil || m.selection.IsEmpty() {
		return ""
	}
	sel := m.selection.Normalize()
	runes := []rune(m.Text())
	if sel.End > len(runes) {
		sel.End = len(runes)
	}
	if sel.Start > len(runes) {
		return ""
	}
	return string(runes[sel.Start:sel.End])
}

var _ Selectable = (*MultilineInput)(nil)

var _ runtime.Widget = (*Input)(nil)
var _ runtime.Focusable = (*Input)(nil)
var _ runtime.Bindable = (*Input)(nil)
var _ runtime.Unbindable = (*Input)(nil)
var _ Validatable = (*Input)(nil)
var _ runtime.Widget = (*MultilineInput)(nil)
var _ runtime.Focusable = (*MultilineInput)(nil)
var _ runtime.Bindable = (*MultilineInput)(nil)
var _ runtime.Unbindable = (*MultilineInput)(nil)
var _ Validatable = (*MultilineInput)(nil)
