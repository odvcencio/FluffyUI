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

	// Callbacks
	onSubmit func(text string)
	onChange func(text string)
}

// NewInput creates a new input widget.
func NewInput() *Input {
	input := &Input{
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Bold(true),
	}
	input.Base.Role = accessibility.RoleTextbox
	input.syncA11y()
	return input
}

// Bind attaches app services.
func (i *Input) Bind(services runtime.Services) {
	i.services = services
}

// Unbind releases app services.
func (i *Input) Unbind() {
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

// OnSubmit sets the callback for when Enter is pressed.
func (i *Input) OnSubmit(fn func(text string)) {
	i.onSubmit = fn
}

// OnChange sets the callback for when text changes.
func (i *Input) OnChange(fn func(text string)) {
	i.onChange = fn
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
	i.cursorPos = i.text.Len()
	i.syncA11y()
}

// Clear clears the input text.
func (i *Input) Clear() {
	i.text.Reset()
	i.cursorPos = 0
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
	if offset > i.text.Len() {
		offset = i.text.Len()
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

	// Show placeholder if empty and not focused
	if text == "" && !i.focused && i.placeholder != "" {
		placeholderStyle := style.Dim(true)
		display := i.placeholder
		if len(display) > content.Width {
			display = display[:content.Width]
		}
		ctx.Buffer.SetString(content.X, content.Y, display, placeholderStyle)
		return
	}

	// Calculate visible portion of text
	// Scroll so cursor is always visible
	visibleStart := 0
	if i.cursorPos >= content.Width {
		visibleStart = i.cursorPos - content.Width + 1
	}

	visibleEnd := visibleStart + content.Width
	if visibleEnd > len(text) {
		visibleEnd = len(text)
	}

	visible := ""
	if visibleStart < len(text) {
		visible = text[visibleStart:visibleEnd]
	}

	// Draw text with selection highlighting
	selectionStyle := style.Reverse(true)
	sel := i.selection.Normalize()
	hasSelection := !i.selection.IsEmpty()

	for idx, ch := range visible {
		textIdx := visibleStart + idx
		screenX := content.X + idx
		charStyle := style

		// Highlight if within selection
		if hasSelection && textIdx >= sel.Start && textIdx < sel.End {
			charStyle = selectionStyle
		}

		ctx.Buffer.Set(screenX, content.Y, ch, charStyle)
	}

	// Draw cursor if focused (by inverting the cell)
	if i.focused {
		cursorX := content.X + i.cursorPos - visibleStart
		if cursorX >= content.X && cursorX < content.X+content.Width {
			var cursorChar rune = ' '
			if i.cursorPos < len(text) {
				cursorChar = rune(text[i.cursorPos])
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
	case terminal.KeyCtrlC:
		if i.copyToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlX:
		if i.cutToClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyCtrlV:
		if i.pasteFromClipboard() {
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		if i.onSubmit != nil {
			text := i.text.String()
			i.onSubmit(text)
		}
		return runtime.WithCommand(runtime.Submit{Text: i.text.String()})

	case terminal.KeyBackspace:
		if i.cursorPos > 0 {
			text := i.text.String()
			i.text.Reset()
			i.text.WriteString(text[:i.cursorPos-1])
			i.text.WriteString(text[i.cursorPos:])
			i.cursorPos--
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyDelete:
		text := i.text.String()
		if i.cursorPos < len(text) {
			i.text.Reset()
			i.text.WriteString(text[:i.cursorPos])
			i.text.WriteString(text[i.cursorPos+1:])
			i.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if key.Ctrl {
			// Word left
			i.cursorPos = i.wordBoundaryLeft()
		} else if i.cursorPos > 0 {
			i.cursorPos--
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if key.Ctrl {
			// Word right
			i.cursorPos = i.wordBoundaryRight()
		} else if i.cursorPos < i.text.Len() {
			i.cursorPos++
		}
		return runtime.Handled()

	case terminal.KeyHome:
		i.cursorPos = 0
		return runtime.Handled()

	case terminal.KeyEnd:
		i.cursorPos = i.text.Len()
		return runtime.Handled()

	case terminal.KeyRune:
		// Insert character
		text := i.text.String()
		i.text.Reset()
		i.text.WriteString(text[:i.cursorPos])
		i.text.WriteRune(key.Rune)
		i.text.WriteString(text[i.cursorPos:])
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
	text := i.text.String()
	if sel.End > len(text) {
		sel.End = len(text)
	}
	if sel.Start > len(text) {
		sel.Start = len(text)
	}
	i.text.Reset()
	i.text.WriteString(text[:sel.Start])
	i.text.WriteString(text[sel.End:])
	i.cursorPos = sel.Start
	i.selection = Selection{}
	i.notifyChange()
}

// ClipboardPaste inserts text at the cursor.
func (i *Input) ClipboardPaste(text string) bool {
	if i == nil || text == "" {
		return false
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
	current := i.text.String()
	i.text.Reset()
	i.text.WriteString(current[:i.cursorPos])
	i.text.WriteString(text)
	i.text.WriteString(current[i.cursorPos:])
	i.cursorPos += len(text)
	i.notifyChange()
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
	textLen := i.text.Len()
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
	i.selection = Selection{Start: 0, End: i.text.Len()}
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
	start, end := findWordBoundaries(text, i.cursorPos)
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
	text := i.text.String()
	if sel.End > len(text) {
		sel.End = len(text)
	}
	if sel.Start > len(text) {
		return ""
	}
	return text[sel.Start:sel.End]
}

var _ Selectable = (*Input)(nil)

// findWordBoundaries returns the start and end positions of the word at pos.
func findWordBoundaries(text string, pos int) (start, end int) {
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
	text := i.text.String()
	pos := i.cursorPos - 1

	// Skip whitespace
	for pos > 0 && text[pos] == ' ' {
		pos--
	}
	// Skip word characters
	for pos > 0 && text[pos-1] != ' ' {
		pos--
	}
	return pos
}

func (i *Input) wordBoundaryRight() int {
	text := i.text.String()
	pos := i.cursorPos

	// Skip word characters
	for pos < len(text) && text[pos] != ' ' {
		pos++
	}
	// Skip whitespace
	for pos < len(text) && text[pos] == ' ' {
		pos++
	}
	return pos
}

func multilineWordBoundaryLeft(text string, cursor int) int {
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

func multilineWordBoundaryRight(text string, cursor int) int {
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

func isWordSeparator(ch byte) bool {
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
	input.syncA11y()
	return input
}

// StyleType returns the selector type name.
func (m *MultilineInput) StyleType() string {
	return "MultilineInput"
}

// Bind attaches app services.
func (m *MultilineInput) Bind(services runtime.Services) {
	m.services = services
}

// Unbind releases app services.
func (m *MultilineInput) Unbind() {
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
	m.cursorX = len(m.lines[m.cursorY])
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
		offset += len(m.lines[i]) + 1
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
	if x < 0 {
		x = 0
	}
	if x > len(line) {
		x = len(line)
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
	total := len(m.Text())
	if offset > total {
		offset = total
	}
	remaining := offset
	for i, line := range m.lines {
		lineLen := len(line)
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
	m.cursorX = len(m.lines[m.cursorY])
	m.ensureCursorVisible()
	m.services.Invalidate()
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (m *MultilineInput) CursorWordLeft() {
	if m == nil {
		return
	}
	offset := multilineWordBoundaryLeft(m.Text(), m.CursorOffset())
	m.SetCursorOffset(offset)
}

// CursorWordRight moves the cursor to the next word boundary.
func (m *MultilineInput) CursorWordRight() {
	if m == nil {
		return
	}
	offset := multilineWordBoundaryRight(m.Text(), m.CursorOffset())
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

// OnSubmit sets the callback (Ctrl+Enter to submit).
func (m *MultilineInput) OnSubmit(fn func(text string)) {
	m.onSubmit = fn
}

// OnChange sets the callback for when text changes.
func (m *MultilineInput) OnChange(fn func(text string)) {
	m.onChange = fn
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
	for i := 0; i < content.Height; i++ {
		lineIdx := m.scrollY + i
		if lineIdx >= len(m.lines) {
			break
		}

		line := m.lines[lineIdx]
		if len(line) > content.Width {
			line = line[:content.Width]
		}
		ctx.Buffer.SetString(content.X, content.Y+i, line, style)
	}

	// Draw cursor
	if m.focused {
		cursorScreenY := m.cursorY - m.scrollY
		if cursorScreenY >= 0 && cursorScreenY < content.Height {
			cursorX := content.X + m.cursorX
			if cursorX >= content.X && cursorX < content.X+content.Width {
				var ch rune = ' '
				if m.cursorY < len(m.lines) && m.cursorX < len(m.lines[m.cursorY]) {
					ch = rune(m.lines[m.cursorY][m.cursorX])
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
		if m.pasteFromClipboard() {
			return runtime.Handled()
		}

	case terminal.KeyEnter:
		if key.Ctrl && m.onSubmit != nil {
			m.onSubmit(m.Text())
			return runtime.WithCommand(runtime.Submit{Text: m.Text()})
		}
		// Insert newline
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX]
		newLine := line[m.cursorX:]
		m.lines = append(m.lines[:m.cursorY+1], append([]string{newLine}, m.lines[m.cursorY+1:]...)...)
		m.cursorY++
		m.cursorX = 0
		m.ensureCursorVisible()
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyBackspace:
		if m.cursorX > 0 {
			line := m.lines[m.cursorY]
			m.lines[m.cursorY] = line[:m.cursorX-1] + line[m.cursorX:]
			m.cursorX--
		} else if m.cursorY > 0 {
			// Join with previous line
			prevLine := m.lines[m.cursorY-1]
			m.cursorX = len(prevLine)
			m.lines[m.cursorY-1] = prevLine + m.lines[m.cursorY]
			m.lines = append(m.lines[:m.cursorY], m.lines[m.cursorY+1:]...)
			m.cursorY--
		}
		m.notifyChange()
		return runtime.Handled()

	case terminal.KeyUp:
		if m.cursorY > 0 {
			m.cursorY--
			if m.cursorX > len(m.lines[m.cursorY]) {
				m.cursorX = len(m.lines[m.cursorY])
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			if m.cursorX > len(m.lines[m.cursorY]) {
				m.cursorX = len(m.lines[m.cursorY])
			}
			m.ensureCursorVisible()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		if m.cursorX > 0 {
			m.cursorX--
		} else if m.cursorY > 0 {
			m.cursorY--
			m.cursorX = len(m.lines[m.cursorY])
		}
		return runtime.Handled()

	case terminal.KeyRight:
		if m.cursorX < len(m.lines[m.cursorY]) {
			m.cursorX++
		} else if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.cursorX = 0
		}
		return runtime.Handled()

	case terminal.KeyRune:
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX] + string(key.Rune) + line[m.cursorX:]
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
	text := m.Text()
	if sel.End > len(text) {
		sel.End = len(text)
	}
	if sel.Start > len(text) {
		sel.Start = len(text)
	}
	newText := text[:sel.Start] + text[sel.End:]
	m.SetText(newText)
	m.SetCursorOffset(sel.Start)
	m.selection = Selection{}
	m.notifyChange()
}

// ClipboardPaste inserts text at the cursor.
func (m *MultilineInput) ClipboardPaste(text string) bool {
	if m == nil || text == "" {
		return false
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
	prefix := line[:m.cursorX]
	suffix := line[m.cursorX:]

	if len(parts) == 1 {
		m.lines[m.cursorY] = prefix + parts[0] + suffix
		m.cursorX += len(parts[0])
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
	m.cursorX = len(parts[len(parts)-1])
	m.ensureCursorVisible()
	m.notifyChange()
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
	textLen := len(m.Text())
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
	m.selection = Selection{Start: 0, End: len(m.Text())}
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
	start, end := findWordBoundaries(text, offset)
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
		start += len(m.lines[i]) + 1 // +1 for newline
	}
	// Calculate end offset (end of current line)
	end := start + len(m.lines[m.cursorY])
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
	text := m.Text()
	if sel.End > len(text) {
		sel.End = len(text)
	}
	if sel.Start > len(text) {
		return ""
	}
	return text[sel.Start:sel.End]
}

var _ Selectable = (*MultilineInput)(nil)
