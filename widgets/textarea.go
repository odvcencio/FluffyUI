package widgets

import (
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// textAreaState represents the state of a text area for undo/redo.
type textAreaState struct {
	text   []rune
	cursor int
}

// TextArea is a multi-line text input widget.
type TextArea struct {
	FocusableBase

	text        []rune
	cursor      int
	scrollY     int
	label       string
	style       backend.Style
	focusStyle  backend.Style
	onChange    func(text string)
	services    runtime.Services
	styleSet    bool
	focusSet    bool
	validators  []forms.Validator
	valErrors   []forms.ValidationError
	valMessages []string

	// History (undo/redo)
	history *state.History[textAreaState]

	// Line metadata cache
	lineStarts     []int
	lineLengths    []int
	lineMetaDirty  bool
}

// NewTextArea creates a new text area.
func NewTextArea() *TextArea {
	ta := &TextArea{
		label:      "Text Area",
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	ta.history = state.NewHistory(textAreaState{}, state.WithGroupWindow(300*time.Millisecond))
	ta.Base.Role = accessibility.RoleTextbox
	ta.syncA11y()
	return ta
}

// Bind attaches app services.
func (t *TextArea) Bind(services runtime.Services) {
	t.services = services
}

// Unbind releases app services.
func (t *TextArea) Unbind() {
	t.services = runtime.Services{}
}

// SetText sets the text and moves the cursor to the end.
func (t *TextArea) SetText(text string) {
	if t == nil {
		return
	}
	t.text = []rune(text)
	t.cursor = len(t.text)
	t.lineMetaDirty = true
	t.syncValue()
	t.pushHistory(false) // non-grouped: explicit set is a distinct operation
}

// CursorOffset returns the cursor offset in the text.
func (t *TextArea) CursorOffset() int {
	if t == nil {
		return 0
	}
	return t.cursor
}

// CursorPosition returns the cursor coordinates within the text area.
func (t *TextArea) CursorPosition() (x, y int) {
	if t == nil {
		return 0, 0
	}
	lineStarts, lineLengths := t.lineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	return col, line
}

// SetCursorOffset moves the cursor to the given offset.
func (t *TextArea) SetCursorOffset(offset int) {
	if t == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(t.text) {
		offset = len(t.text)
	}
	t.cursor = offset
	t.services.Invalidate()
}

// SetCursorPosition moves the cursor to the given coordinates.
func (t *TextArea) SetCursorPosition(x, y int) {
	if t == nil {
		return
	}
	lineStarts, lineLengths := t.lineMeta()
	if len(lineStarts) == 0 {
		t.cursor = 0
		return
	}
	if y < 0 {
		y = 0
	}
	if y >= len(lineStarts) {
		y = len(lineStarts) - 1
	}
	lineLen := lineLengths[y]
	if x < 0 {
		x = 0
	}
	if x > lineLen {
		x = lineLen
	}
	t.cursor = lineStarts[y] + x
	t.services.Invalidate()
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (t *TextArea) CursorWordLeft() {
	if t == nil {
		return
	}
	t.cursor = textAreaWordBoundaryLeft(t.text, t.cursor)
	t.services.Invalidate()
}

// CursorWordRight moves the cursor to the next word boundary.
func (t *TextArea) CursorWordRight() {
	if t == nil {
		return
	}
	t.cursor = textAreaWordBoundaryRight(t.text, t.cursor)
	t.services.Invalidate()
}

// Text returns the current text.
func (t *TextArea) Text() string {
	if t == nil {
		return ""
	}
	return string(t.text)
}

// SetOnChange registers a callback for text changes.
func (t *TextArea) SetOnChange(fn func(text string)) {
	if t == nil {
		return
	}
	t.onChange = fn
}

// Deprecated: use SetOnChange instead.
func (t *TextArea) OnChange(fn func(text string)) {
	t.SetOnChange(fn)
}

// SetValidators updates validation rules for the text area.
func (t *TextArea) SetValidators(validators ...forms.Validator) {
	if t == nil {
		return
	}
	t.validators = validators
}

// Validate runs validation rules and returns validation errors.
func (t *TextArea) Validate() []forms.ValidationError {
	if t == nil {
		return nil
	}
	errs, messages := validateValue(t.Text(), t.validators)
	t.valErrors = errs
	t.valMessages = messages
	return errs
}

// Errors returns the latest validation error messages.
func (t *TextArea) Errors() []string {
	if t == nil {
		return nil
	}
	if len(t.validators) > 0 {
		t.Validate()
	}
	if len(t.valMessages) == 0 {
		return nil
	}
	out := make([]string, len(t.valMessages))
	copy(out, t.valMessages)
	return out
}

// Valid reports whether validation passes.
func (t *TextArea) Valid() bool {
	if t == nil {
		return true
	}
	return len(t.Validate()) == 0
}

// Undo reverts to the previous state.
// Returns true if undo was successful.
func (t *TextArea) Undo() bool {
	if t == nil || t.history == nil {
		return false
	}
	s, ok := t.history.Undo()
	if !ok {
		return false
	}
	t.text = s.text
	t.cursor = s.cursor
	t.lineMetaDirty = true
	t.syncValue()
	return true
}

// Redo reapplies a previously undone state.
// Returns true if redo was successful.
func (t *TextArea) Redo() bool {
	if t == nil || t.history == nil {
		return false
	}
	s, ok := t.history.Redo()
	if !ok {
		return false
	}
	t.text = s.text
	t.cursor = s.cursor
	t.lineMetaDirty = true
	t.syncValue()
	return true
}

// CanUndo returns true if undo is available.
func (t *TextArea) CanUndo() bool {
	if t == nil || t.history == nil {
		return false
	}
	return t.history.CanUndo()
}

// CanRedo returns true if redo is available.
func (t *TextArea) CanRedo() bool {
	if t == nil || t.history == nil {
		return false
	}
	return t.history.CanRedo()
}

// ClearHistory resets the undo/redo history.
func (t *TextArea) ClearHistory() {
	if t == nil || t.history == nil {
		return
	}
	t.history.Clear()
}

func (t *TextArea) pushHistory(grouped bool) {
	if t.history == nil {
		return
	}
	s := textAreaState{text: append([]rune(nil), t.text...), cursor: t.cursor}
	if grouped {
		t.history.PushGrouped(s)
	} else {
		t.history.Push(s)
	}
}

// SetLabel updates the accessibility label.
func (t *TextArea) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// SetStyle sets the normal style.
func (t *TextArea) SetStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.style = style
	t.styleSet = true
}

// SetFocusStyle sets the focused style.
func (t *TextArea) SetFocusStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.focusStyle = style
	t.focusSet = true
}

// StyleType returns the selector type name.
func (t *TextArea) StyleType() string {
	return "TextArea"
}

// Measure returns the desired size.
// TextArea returns minimal preferred size (1x1) to work correctly in flex layouts.
// When flex containers measure children with unbounded constraints (maxInt),
// returning maxInt would cause the flex shrink algorithm to shrink this widget to 0.
func (t *TextArea) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{Width: 1, Height: 1})
	})
}

// Render draws the text area.
func (t *TextArea) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	outer := t.bounds
	content := t.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	style := t.style
	resolved := ctx.ResolveStyle(t)
	if !resolved.IsZero() {
		final := resolved
		if t.styleSet {
			final = final.Merge(uistyle.FromBackend(t.style))
		}
		if t.focused && t.focusSet {
			final = final.Merge(uistyle.FromBackend(t.focusStyle))
		}
		style = final.ToBackend()
	} else if t.focused {
		style = t.focusStyle
	}
	ctx.Buffer.Fill(outer, ' ', style)

	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	lineStarts, lineLengths := t.lineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	t.scrollY = min(max(t.scrollY, 0), max(0, len(lineStarts)-1))
	if line < t.scrollY {
		t.scrollY = line
	} else if line >= t.scrollY+content.Height {
		t.scrollY = line - content.Height + 1
	}
	scrollX := 0
	if col >= content.Width {
		scrollX = col - content.Width + 1
	}

	for row := 0; row < content.Height; row++ {
		lineIndex := t.scrollY + row
		if lineIndex >= len(lineStarts) {
			break
		}
		lineText := t.lineText(lineIndex, lineStarts, lineLengths)
		if scrollX < len(lineText) {
			lineText = lineText[scrollX:]
		} else {
			lineText = ""
		}
		if len(lineText) > content.Width {
			lineText = lineText[:content.Width]
		}
		writePadded(ctx.Buffer, content.X, content.Y+row, content.Width, lineText, style)
	}

	if t.focused {
		cursorRow := line - t.scrollY
		cursorCol := col - scrollX
		if cursorRow >= 0 && cursorRow < content.Height && cursorCol >= 0 && cursorCol < content.Width {
			cursorX := content.X + cursorCol
			cursorY := content.Y + cursorRow
			ch := ' '
			lineText := t.lineText(line, lineStarts, lineLengths)
			if col < len(lineText) {
				ch = rune(lineText[col])
			}
			ctx.Buffer.Set(cursorX, cursorY, ch, style.Reverse(true))
		}
	}
}

// HandleMessage processes keyboard input.
func (t *TextArea) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlZ:
		if key.Shift {
			t.Redo()
		} else {
			t.Undo()
		}
		return runtime.Handled()
	case terminal.KeyCtrlY:
		t.Redo()
		return runtime.Handled()
	case terminal.KeyCtrlC:
		if t.copyToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlX:
		if t.cutToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlV:
		if t.pasteFromClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		t.pushHistory(true)
		t.insertRune('\n')
		return runtime.Handled()
	case terminal.KeyBackspace:
		if t.cursor > 0 {
			t.pushHistory(true)
			t.deleteRune(t.cursor - 1)
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if t.cursor < len(t.text) {
			t.pushHistory(true)
			t.deleteRune(t.cursor)
		}
		return runtime.Handled()
	case terminal.KeyLeft:
		if t.cursor > 0 {
			t.cursor--
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if t.cursor < len(t.text) {
			t.cursor++
		}
		return runtime.Handled()
	case terminal.KeyUp:
		t.moveVertical(-1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.moveVertical(1)
		return runtime.Handled()
	case terminal.KeyHome:
		t.moveLineBoundary(true)
		return runtime.Handled()
	case terminal.KeyEnd:
		t.moveLineBoundary(false)
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune != 0 {
			t.pushHistory(true)
			t.insertRune(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *TextArea) insertRune(r rune) {
	t.text = append(t.text[:t.cursor], append([]rune{r}, t.text[t.cursor:]...)...)
	t.cursor++
	t.lineMetaDirty = true
	t.syncValue()
}

func (t *TextArea) insertText(text string) {
	if text == "" {
		return
	}
	runes := []rune(text)
	t.text = append(t.text[:t.cursor], append(runes, t.text[t.cursor:]...)...)
	t.cursor += len(runes)
	t.lineMetaDirty = true
	t.syncValue()
}

func (t *TextArea) deleteRune(index int) {
	if index < 0 || index >= len(t.text) {
		return
	}
	t.text = append(t.text[:index], t.text[index+1:]...)
	if t.cursor > index {
		t.cursor--
	}
	t.lineMetaDirty = true
	t.syncValue()
}

func (t *TextArea) moveVertical(delta int) {
	lineStarts, lineLengths := t.lineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	target := line + delta
	if target < 0 || target >= len(lineStarts) {
		return
	}
	targetLen := lineLengths[target]
	if col > targetLen {
		col = targetLen
	}
	t.cursor = lineStarts[target] + col
}

func (t *TextArea) moveLineBoundary(start bool) {
	lineStarts, lineLengths := t.lineMeta()
	line, _ := t.cursorLineCol(lineStarts, lineLengths)
	if line < 0 || line >= len(lineStarts) {
		return
	}
	if start {
		t.cursor = lineStarts[line]
		return
	}
	t.cursor = lineStarts[line] + lineLengths[line]
}

func (t *TextArea) lineMeta() ([]int, []int) {
	if t == nil {
		return []int{0}, []int{0}
	}
	if !t.lineMetaDirty && t.lineStarts != nil {
		return t.lineStarts, t.lineLengths
	}
	starts := []int{0}
	var lengths []int
	for i, r := range t.text {
		if r == '\n' {
			lengths = append(lengths, i-starts[len(starts)-1])
			starts = append(starts, i+1)
		}
	}
	lastStart := starts[len(starts)-1]
	lengths = append(lengths, len(t.text)-lastStart)
	t.lineStarts = starts
	t.lineLengths = lengths
	t.lineMetaDirty = false
	return t.lineStarts, t.lineLengths
}

func (t *TextArea) lineText(line int, starts []int, lengths []int) string {
	if line < 0 || line >= len(starts) {
		return ""
	}
	start := starts[line]
	end := start + lengths[line]
	if start > len(t.text) || end > len(t.text) || start > end {
		return ""
	}
	return string(t.text[start:end])
}

func (t *TextArea) cursorLineCol(starts []int, lengths []int) (int, int) {
	if len(starts) == 0 {
		return 0, 0
	}
	for i, start := range starts {
		end := start + lengths[i]
		if t.cursor <= end {
			return i, t.cursor - start
		}
	}
	last := len(starts) - 1
	return last, lengths[last]
}

func textAreaWordBoundaryLeft(text []rune, cursor int) int {
	if cursor <= 0 {
		return 0
	}
	if cursor > len(text) {
		cursor = len(text)
	}
	pos := cursor - 1
	for pos > 0 && isTextAreaSeparator(text[pos]) {
		pos--
	}
	for pos > 0 && !isTextAreaSeparator(text[pos-1]) {
		pos--
	}
	return pos
}

func textAreaWordBoundaryRight(text []rune, cursor int) int {
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(text) {
		return len(text)
	}
	pos := cursor
	for pos < len(text) && !isTextAreaSeparator(text[pos]) {
		pos++
	}
	for pos < len(text) && isTextAreaSeparator(text[pos]) {
		pos++
	}
	return pos
}

func isTextAreaSeparator(r rune) bool {
	switch r {
	case ' ', '\n', '\t':
		return true
	default:
		return false
	}
}

func (t *TextArea) syncValue() {
	t.syncA11y()
	if t.onChange != nil {
		t.onChange(t.Text())
	}
}

func (t *TextArea) syncA11y() {
	if t == nil {
		return
	}
	label := strings.TrimSpace(t.label)
	if label == "" {
		label = "Text Area"
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTextbox
	}
	t.Base.Label = label
	t.Base.Value = &accessibility.ValueInfo{Text: t.Text()}
}

// ClipboardCopy returns the current text.
func (t *TextArea) ClipboardCopy() (string, bool) {
	if t == nil {
		return "", false
	}
	return t.Text(), true
}

// ClipboardCut returns the current text and clears it.
func (t *TextArea) ClipboardCut() (string, bool) {
	if t == nil {
		return "", false
	}
	t.pushHistory(true)
	text := t.Text()
	t.text = nil
	t.cursor = 0
	t.scrollY = 0
	t.lineMetaDirty = true
	t.syncValue()
	return text, true
}

// ClipboardPaste inserts text at the cursor.
func (t *TextArea) ClipboardPaste(text string) bool {
	if t == nil || text == "" {
		return false
	}
	// Large pastes (> 100 chars) are not grouped
	grouped := len(text) <= 100
	t.pushHistory(grouped)
	t.insertText(text)
	return true
}

func (t *TextArea) copyToClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := t.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (t *TextArea) cutToClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := t.ClipboardCut()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (t *TextArea) pasteFromClipboard() bool {
	cb := t.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, err := cb.Read()
	if err != nil || text == "" {
		return false
	}
	return t.ClipboardPaste(text)
}

var _ clipboard.Target = (*TextArea)(nil)

var _ runtime.Widget = (*TextArea)(nil)
var _ runtime.Focusable = (*TextArea)(nil)
var _ Validatable = (*TextArea)(nil)
