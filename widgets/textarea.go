package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// TextArea is a multi-line text input widget.
type TextArea struct {
	FocusableBase

	text       []rune
	cursor     int
	scrollY    int
	label      string
	style      backend.Style
	focusStyle backend.Style
	onChange   func(text string)
	services   runtime.Services
	styleSet   bool
	focusSet   bool
}

// NewTextArea creates a new text area.
func NewTextArea() *TextArea {
	ta := &TextArea{
		label:      "Text Area",
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
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
	t.syncValue()
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

// OnChange registers a callback for text changes.
func (t *TextArea) OnChange(fn func(text string)) {
	if t == nil {
		return
	}
	t.onChange = fn
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
func (t *TextArea) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: contentConstraints.MaxHeight})
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
		t.insertRune('\n')
		return runtime.Handled()
	case terminal.KeyBackspace:
		if t.cursor > 0 {
			t.deleteRune(t.cursor - 1)
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if t.cursor < len(t.text) {
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
			t.insertRune(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *TextArea) insertRune(r rune) {
	t.text = append(t.text[:t.cursor], append([]rune{r}, t.text[t.cursor:]...)...)
	t.cursor++
	t.syncValue()
}

func (t *TextArea) insertText(text string) {
	if text == "" {
		return
	}
	runes := []rune(text)
	t.text = append(t.text[:t.cursor], append(runes, t.text[t.cursor:]...)...)
	t.cursor += len(runes)
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
	return starts, lengths
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
	text := t.Text()
	t.text = nil
	t.cursor = 0
	t.scrollY = 0
	t.syncValue()
	return text, true
}

// ClipboardPaste inserts text at the cursor.
func (t *TextArea) ClipboardPaste(text string) bool {
	if t == nil || text == "" {
		return false
	}
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
