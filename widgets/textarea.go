package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// TextAreaHighlight defines a styled range within the TextArea content.
// Ranges are in rune offsets (not byte offsets) to match TextArea's internal []rune storage.
type TextAreaHighlight struct {
	Start int           // rune offset, inclusive
	End   int           // rune offset, exclusive
	Style backend.Style // override style (typically just foreground color)
}

// textAreaState represents the state of a text area for undo/redo.
type textAreaState struct {
	text           []rune
	cursor         int
	selectionStart int
	selectionEnd   int
}

// TextArea is a multi-line text input widget.
type TextArea struct {
	FocusableBase

	text        []rune
	cursor      int
	selection   Selection
	scrollY     int
	scrollX     int
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

	// Highlight ranges for syntax coloring
	highlights []TextAreaHighlight

	// Line numbers
	showLineNumbers bool
	gutterStyle     backend.Style
	gutterStyleSet  bool

	// Tab configuration
	tabSize int
	useTabs bool

	// Line metadata cache
	lineStarts    []int
	lineLengths   []int
	lineMetaDirty bool

	// Wrapped line metadata cache.
	wrappedLineStarts    []int
	wrappedLineLengths   []int
	wrappedLineMetaDirty bool
	wrappedLineMetaWidth int

	// Wrapping state.
	wordWrap bool

	// Optional logical line filter used by editors to hide ranges (for example,
	// code folding). Nil means all logical lines are visible.
	visibleLines []int
}

// NewTextArea creates a multi-line text editor with line wrapping, selection,
// clipboard support, and undo/redo history.
func NewTextArea() *TextArea {
	ta := &TextArea{
		label:      "Text Area",
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
		tabSize:    4,
	}
	ta.history = state.NewHistory(textAreaState{}, state.WithGroupWindow(300*time.Millisecond))
	ta.Base.Role = accessibility.RoleTextbox
	ta.syncA11y()
	return ta
}

// Bind attaches app services.
func (t *TextArea) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *TextArea) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

// Blur removes focus and announces validation errors if any.
func (t *TextArea) Blur() {
	if t == nil {
		return
	}
	t.FocusableBase.Blur()
	if len(t.validators) > 0 {
		errs := t.Validate()
		if len(errs) > 0 && len(t.valMessages) > 0 {
			if announcer := t.services.Announcer(); announcer != nil {
				announcer.Announce(strings.Join(t.valMessages, ", "), accessibility.PriorityAssertive)
			}
		}
	}
}

// SetText sets the text and moves the cursor to the end.
func (t *TextArea) SetText(text string) {
	if t == nil {
		return
	}
	t.text = []rune(text)
	t.cursor = len(t.text)
	t.selection = Selection{}
	t.invalidateLineMeta()
	t.syncValue()
	t.pushHistory(false) // non-grouped: explicit set is a distinct operation
}

// SetWordWrap enables or disables soft wrapping for long lines.
func (t *TextArea) SetWordWrap(enabled bool) {
	if t == nil {
		return
	}
	if t.wordWrap == enabled {
		return
	}
	t.wordWrap = enabled
	t.invalidateLineMeta()
	t.services.Invalidate()
}

// SetVisibleLines limits the rendered/interactive display to the provided
// logical line indices. Passing nil or an empty slice restores all lines.
func (t *TextArea) SetVisibleLines(lines []int) {
	if t == nil {
		return
	}

	if len(lines) == 0 {
		if t.visibleLines == nil {
			return
		}
		t.visibleLines = nil
		t.invalidateLineMeta()
		t.services.Invalidate()
		return
	}

	filtered := make([]int, 0, len(lines))
	last := -1
	for _, line := range lines {
		if line < 0 || line == last {
			continue
		}
		if line < last {
			// Keep behavior deterministic for callers that accidentally pass an
			// unsorted map: hide map is disabled until a sorted list is provided.
			filtered = nil
			break
		}
		filtered = append(filtered, line)
		last = line
	}

	if len(filtered) == 0 {
		if t.visibleLines == nil {
			return
		}
		t.visibleLines = nil
		t.invalidateLineMeta()
		t.services.Invalidate()
		return
	}

	unchanged := len(filtered) == len(t.visibleLines)
	if unchanged {
		for i := range filtered {
			if filtered[i] != t.visibleLines[i] {
				unchanged = false
				break
			}
		}
	}
	if unchanged {
		return
	}

	t.visibleLines = filtered
	t.invalidateLineMeta()
	t.services.Invalidate()
}

// VisibleLines returns the active logical line filter. Nil means all lines are
// visible.
func (t *TextArea) VisibleLines() []int {
	if t == nil || len(t.visibleLines) == 0 {
		return nil
	}
	out := make([]int, len(t.visibleLines))
	copy(out, t.visibleLines)
	return out
}

// WordWrap reports whether soft wrapping is enabled.
func (t *TextArea) WordWrap() bool {
	if t == nil {
		return false
	}
	return t.wordWrap
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

func (t *TextArea) invalidateLineMeta() {
	if t == nil {
		return
	}
	t.lineMetaDirty = true
	t.wrappedLineMetaDirty = true
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

// OnChange registers a callback for text changes.
//
// Deprecated: Use SetOnChange instead. This method will be removed in v1.0.
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
	t.selection = Selection{Start: s.selectionStart, End: s.selectionEnd}
	t.invalidateLineMeta()
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
	t.selection = Selection{Start: s.selectionStart, End: s.selectionEnd}
	t.invalidateLineMeta()
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
	s := textAreaState{
		text:           append([]rune(nil), t.text...),
		cursor:         t.cursor,
		selectionStart: t.selection.Start,
		selectionEnd:   t.selection.End,
	}
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

// SetHighlights replaces the highlight ranges. Ranges must be sorted by Start.
// The TextArea does not sort them — the caller is responsible for ordering.
func (t *TextArea) SetHighlights(highlights []TextAreaHighlight) {
	if t == nil {
		return
	}
	t.highlights = highlights
}

// ClearHighlights removes all highlight ranges.
func (t *TextArea) ClearHighlights() {
	if t == nil {
		return
	}
	t.highlights = nil
}

// SetShowLineNumbers enables or disables the line number gutter.
func (t *TextArea) SetShowLineNumbers(show bool) {
	if t == nil {
		return
	}
	t.showLineNumbers = show
}

// SetGutterStyle sets the style for the line number gutter.
func (t *TextArea) SetGutterStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.gutterStyle = style
	t.gutterStyleSet = true
}

// SetTabSize sets the number of spaces for a tab. Default is 4.
func (t *TextArea) SetTabSize(n int) {
	if t == nil {
		return
	}
	if n < 1 {
		n = 1
	}
	t.tabSize = n
}

// SetTabMode controls whether Tab inserts a literal '\t' (true) or spaces (false).
func (t *TextArea) SetTabMode(useTabs bool) {
	if t == nil {
		return
	}
	t.useTabs = useTabs
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

// mergeHighlightStyle merges a highlight's foreground and attributes onto a base style.
func mergeHighlightStyle(base, highlight backend.Style) backend.Style {
	fg, _, hAttrs := highlight.Decompose()
	if fg != backend.ColorDefault {
		base = base.Foreground(fg)
	}
	if hAttrs&backend.AttrBold != 0 {
		base = base.Bold(true)
	}
	if hAttrs&backend.AttrItalic != 0 {
		base = base.Italic(true)
	}
	return base
}

// gutterWidth returns the width of the line number gutter (0 if disabled).
func (t *TextArea) gutterWidth(totalLines int) int {
	if !t.showLineNumbers {
		return 0
	}
	return digits(totalLines) + 1 // +1 for separator space
}

func digits(n int) int {
	if n <= 0 {
		return 1
	}
	d := 0
	for n > 0 {
		d++
		n /= 10
	}
	return d
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

	logicalStarts, logicalLengths := t.lineMeta()
	lineStarts, lineLengths := t.displayLineMeta()
	line, col := t.cursorLineCol(lineStarts, lineLengths)
	t.scrollY = min(max(t.scrollY, 0), max(0, len(lineStarts)-1))
	if line < t.scrollY {
		t.scrollY = line
	} else if line >= t.scrollY+content.Height {
		t.scrollY = line - content.Height + 1
	}

	// Line number gutter
	gw := t.gutterWidth(len(logicalStarts))
	textBounds := content
	if gw > 0 && gw < content.Width {
		textBounds.X += gw
		textBounds.Width -= gw
	}

	// Persistent horizontal scroll
	if col >= textBounds.Width {
		t.scrollX = col - textBounds.Width + 1
	} else if col < t.scrollX {
		t.scrollX = col
	}

	// Gutter style: use explicit or dimmed base style
	gs := t.gutterStyle
	if !t.gutterStyleSet {
		gs = style.Dim(true)
	}

	// Highlight cursor: index into t.highlights
	hIdx := 0

	// Precompute normalized selection range for rendering
	sel := t.selection.Normalize()
	hasSel := !t.selection.IsEmpty()

	for row := 0; row < content.Height; row++ {
		lineIndex := t.scrollY + row
		if lineIndex >= len(lineStarts) {
			break
		}

		// Draw line number
		if gw > 0 {
			lineStart := lineStarts[lineIndex]
			logicalLine, logicalCol := t.cursorLineColForOffset(logicalStarts, logicalLengths, lineStart)

			if logicalCol == 0 {
				numStr := fmt.Sprintf("%*d", gw-1, logicalLine+1)
				for i, ch := range numStr {
					ctx.Buffer.Set(content.X+i, content.Y+row, ch, gs)
				}
				// Separator space
				ctx.Buffer.Set(content.X+gw-1, content.Y+row, ' ', gs)
			} else {
				for i := 0; i < gw; i++ {
					ctx.Buffer.Set(content.X+i, content.Y+row, ' ', gs)
				}
			}
		}

		lineStart := lineStarts[lineIndex]
		lineLen := lineLengths[lineIndex]
		lineText := t.lineSlice(lineStart, lineLen)
		dir := i18n.DetectDirection(lineText)
		if dir == i18n.DirectionRTL {
			lineText = i18n.BidiReorder(lineText, dir)
		}

		// Per-character rendering with highlights
		visibleStart := t.scrollX
		visibleEnd := t.scrollX + textBounds.Width
		if visibleEnd > lineLen {
			visibleEnd = lineLen
		}

		// Advance highlight cursor to first potentially overlapping range
		for hIdx > 0 && hIdx < len(t.highlights) && t.highlights[hIdx].Start > lineStart {
			hIdx--
		}
		for hIdx < len(t.highlights) && t.highlights[hIdx].End <= lineStart+visibleStart {
			hIdx++
		}

		xPos := textBounds.X
		for c := visibleStart; c < visibleEnd && xPos < textBounds.X+textBounds.Width; c++ {
			runeOffset := lineStart + c
			ch := rune(lineText[c])

			charStyle := style
			// Check highlights
			hi := hIdx
			for hi < len(t.highlights) && t.highlights[hi].Start <= runeOffset {
				if t.highlights[hi].End > runeOffset {
					charStyle = mergeHighlightStyle(charStyle, t.highlights[hi].Style)
					break
				}
				hi++
			}

			// Apply selection highlighting (reverse video) on top of syntax highlights
			if hasSel && runeOffset >= sel.Start && runeOffset < sel.End {
				charStyle = charStyle.Reverse(true)
			}

			if dir == i18n.DirectionRTL && visibleEnd-visibleStart < textBounds.Width {
				offset := textBounds.Width - (visibleEnd - visibleStart)
				ctx.Buffer.Set(xPos+offset, content.Y+row, ch, charStyle)
			} else {
				ctx.Buffer.Set(xPos, content.Y+row, ch, charStyle)
			}
			xPos++
		}
	}

	if t.focused {
		cursorRow := line - t.scrollY
		cursorCol := col - t.scrollX
		if cursorRow >= 0 && cursorRow < content.Height && cursorCol >= 0 && cursorCol < textBounds.Width {
			cursorX := textBounds.X + cursorCol
			cursorY := content.Y + cursorRow
			ch := ' '
			lineText := t.lineSlice(lineStarts[line], lineLengths[line])
			if col < len(lineText) {
				ch = rune(lineText[col])
			}
			ctx.Buffer.Set(cursorX, cursorY, ch, style.Reverse(true))
		}
	}
}

// HandleMessage processes keyboard and mouse input.
func (t *TextArea) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}

	// Handle mouse events.
	if mouse, ok := msg.(runtime.MouseMsg); ok {
		return t.handleMouse(mouse)
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlA:
		t.SelectAll()
		return runtime.Handled()
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
		if t.HasSelection() {
			t.deleteSelection()
		}
		t.insertRune('\n')
		return runtime.Handled()
	case terminal.KeyTab:
		t.pushHistory(true)
		if t.HasSelection() {
			t.deleteSelection()
		}
		if t.useTabs {
			t.insertRune('\t')
		} else {
			for i := 0; i < t.tabSize; i++ {
				t.insertRune(' ')
			}
		}
		return runtime.Handled()
	case terminal.KeyBackspace:
		if t.HasSelection() {
			t.pushHistory(true)
			t.deleteSelection()
		} else if t.cursor > 0 {
			t.pushHistory(true)
			t.deleteRune(t.cursor - 1)
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if t.HasSelection() {
			t.pushHistory(true)
			t.deleteSelection()
		} else if t.cursor < len(t.text) {
			t.pushHistory(true)
			t.deleteRune(t.cursor)
		}
		return runtime.Handled()
	case terminal.KeyLeft:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			if key.Ctrl {
				t.cursor = textAreaWordBoundaryLeft(t.text, t.cursor)
			} else if t.cursor > 0 {
				t.cursor--
			}
			t.extendSelection(anchor)
		} else {
			if !t.collapseSelection(true) {
				if key.Ctrl {
					t.cursor = textAreaWordBoundaryLeft(t.text, t.cursor)
				} else if t.cursor > 0 {
					t.cursor--
				}
			}
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			if key.Ctrl {
				t.cursor = textAreaWordBoundaryRight(t.text, t.cursor)
			} else if t.cursor < len(t.text) {
				t.cursor++
			}
			t.extendSelection(anchor)
		} else {
			if !t.collapseSelection(false) {
				if key.Ctrl {
					t.cursor = textAreaWordBoundaryRight(t.text, t.cursor)
				} else if t.cursor < len(t.text) {
					t.cursor++
				}
			}
		}
		return runtime.Handled()
	case terminal.KeyUp:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			t.moveVertical(-1)
			t.extendSelection(anchor)
		} else {
			if !t.collapseSelection(true) {
				t.moveVertical(-1)
			}
		}
		return runtime.Handled()
	case terminal.KeyDown:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			t.moveVertical(1)
			t.extendSelection(anchor)
		} else {
			if !t.collapseSelection(false) {
				t.moveVertical(1)
			}
		}
		return runtime.Handled()
	case terminal.KeyPageUp:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			t.movePage(-1)
			t.extendSelection(anchor)
		} else {
			t.selection = Selection{}
			t.movePage(-1)
		}
		return runtime.Handled()
	case terminal.KeyPageDown:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			t.movePage(1)
			t.extendSelection(anchor)
		} else {
			t.selection = Selection{}
			t.movePage(1)
		}
		return runtime.Handled()
	case terminal.KeyHome:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			if key.Ctrl {
				t.cursor = 0
			} else {
				t.moveLineBoundary(true)
			}
			t.extendSelection(anchor)
		} else {
			t.selection = Selection{}
			if key.Ctrl {
				t.cursor = 0
			} else {
				t.moveLineBoundary(true)
			}
		}
		return runtime.Handled()
	case terminal.KeyEnd:
		if key.Shift {
			anchor := t.cursor
			if !t.selection.IsEmpty() {
				anchor = t.selection.Start
			}
			if key.Ctrl {
				t.cursor = len(t.text)
			} else {
				t.moveLineBoundary(false)
			}
			t.extendSelection(anchor)
		} else {
			t.selection = Selection{}
			if key.Ctrl {
				t.cursor = len(t.text)
			} else {
				t.moveLineBoundary(false)
			}
		}
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune != 0 {
			t.pushHistory(true)
			if t.HasSelection() {
				t.deleteSelection()
			}
			t.insertRune(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// movePage moves the cursor up or down by approximately one page.
func (t *TextArea) movePage(direction int) {
	content := t.ContentBounds()
	pageSize := content.Height
	if pageSize <= 0 {
		pageSize = 1
	}
	t.moveVertical(direction * pageSize)
}

// handleMouse processes mouse click events to position the cursor.
func (t *TextArea) handleMouse(mouse runtime.MouseMsg) runtime.HandleResult {
	if mouse.Button != runtime.MouseLeft || mouse.Action != runtime.MousePress {
		return runtime.Unhandled()
	}

	content := t.ContentBounds()
	logicalStarts, _ := t.lineMeta()
	lineStarts, lineLengths := t.displayLineMeta()
	gw := t.gutterWidth(len(logicalStarts))

	textX := content.X + gw
	textW := content.Width - gw

	// Check if click is within text bounds.
	if mouse.X < textX || mouse.X >= textX+textW {
		return runtime.Unhandled()
	}
	if mouse.Y < content.Y || mouse.Y >= content.Y+content.Height {
		return runtime.Unhandled()
	}

	row := mouse.Y - content.Y
	col := mouse.X - textX + t.scrollX
	targetLine := t.scrollY + row
	if targetLine >= len(lineStarts) {
		targetLine = len(lineStarts) - 1
	}
	if targetLine < 0 {
		targetLine = 0
	}
	lineLen := lineLengths[targetLine]
	if col > lineLen {
		col = lineLen
	}
	if col < 0 {
		col = 0
	}
	newPos := lineStarts[targetLine] + col

	if mouse.Shift && !t.selection.IsEmpty() {
		// Shift+click: extend selection from anchor to click position
		t.cursor = newPos
		t.extendSelection(t.selection.Start)
	} else if mouse.Shift {
		// Shift+click with no selection: select from cursor to click
		anchor := t.cursor
		t.cursor = newPos
		t.extendSelection(anchor)
	} else {
		// Normal click: position cursor, clear selection
		t.cursor = newPos
		t.selection = Selection{}
	}
	return runtime.Handled()
}

func (t *TextArea) insertRune(r rune) {
	t.text = append(t.text[:t.cursor], append([]rune{r}, t.text[t.cursor:]...)...)
	t.cursor++
	t.invalidateLineMeta()
	t.syncValue()
}

func (t *TextArea) insertText(text string) {
	if text == "" {
		return
	}
	runes := []rune(text)
	t.text = append(t.text[:t.cursor], append(runes, t.text[t.cursor:]...)...)
	t.cursor += len(runes)
	t.invalidateLineMeta()
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
	t.invalidateLineMeta()
	t.syncValue()
}

func (t *TextArea) moveVertical(delta int) {
	lineStarts, lineLengths := t.displayLineMeta()
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

func (t *TextArea) displayWidth() int {
	content := t.ContentBounds()
	logicalStarts, _ := t.lineMeta()
	gw := t.gutterWidth(len(logicalStarts))
	textWidth := content.Width - gw
	if textWidth <= 0 {
		textWidth = 1
	}
	return min(textWidth, len(t.text)+1) // avoid excessive wrapping in empty/short lines
}

func (t *TextArea) displayLineMeta() ([]int, []int) {
	width := t.displayWidth()
	return t.lineMetaForWidth(width)
}

func (t *TextArea) lineMetaForWidth(width int) ([]int, []int) {
	logicalStarts, logicalLengths := t.lineMeta()

	if len(t.visibleLines) == 0 {
		if !t.wordWrap {
			return logicalStarts, logicalLengths
		}
		if width < 1 {
			width = 1
		}
		if !t.wrappedLineMetaDirty && t.wrappedLineMetaWidth == width && t.wrappedLineStarts != nil {
			return t.wrappedLineStarts, t.wrappedLineLengths
		}

		starts := make([]int, 0, len(t.text)+1)
		lengths := make([]int, 0, len(t.text)+1)
		for i, lineStart := range logicalStarts {
			lineLen := logicalLengths[i]
			if lineLen <= 0 {
				starts = append(starts, lineStart)
				lengths = append(lengths, 0)
				continue
			}
			for lineOffset := 0; lineOffset < lineLen; lineOffset += width {
				segLen := width
				if remaining := lineLen - lineOffset; remaining < segLen {
					segLen = remaining
				}
				starts = append(starts, lineStart+lineOffset)
				lengths = append(lengths, segLen)
			}
		}

		t.wrappedLineStarts = starts
		t.wrappedLineLengths = lengths
		t.wrappedLineMetaWidth = width
		t.wrappedLineMetaDirty = false
		return t.wrappedLineStarts, t.wrappedLineLengths
	}

	visibleLogical := t.visibleLogicalLines(len(logicalStarts))

	if !t.wordWrap {
		starts := make([]int, 0, len(visibleLogical))
		lengths := make([]int, 0, len(visibleLogical))
		for _, lineIdx := range visibleLogical {
			starts = append(starts, logicalStarts[lineIdx])
			lengths = append(lengths, logicalLengths[lineIdx])
		}
		if len(starts) == 0 {
			return []int{0}, []int{0}
		}
		return starts, lengths
	}
	if width < 1 {
		width = 1
	}
	if !t.wrappedLineMetaDirty && t.wrappedLineMetaWidth == width && t.wrappedLineStarts != nil {
		return t.wrappedLineStarts, t.wrappedLineLengths
	}

	starts := make([]int, 0, len(t.text)+1)
	lengths := make([]int, 0, len(t.text)+1)
	for _, lineIdx := range visibleLogical {
		lineStart := logicalStarts[lineIdx]
		lineLen := logicalLengths[lineIdx]
		if lineLen <= 0 {
			starts = append(starts, lineStart)
			lengths = append(lengths, 0)
			continue
		}
		for lineOffset := 0; lineOffset < lineLen; lineOffset += width {
			segLen := width
			if remaining := lineLen - lineOffset; remaining < segLen {
				segLen = remaining
			}
			starts = append(starts, lineStart+lineOffset)
			lengths = append(lengths, segLen)
		}
	}

	t.wrappedLineStarts = starts
	t.wrappedLineLengths = lengths
	t.wrappedLineMetaWidth = width
	t.wrappedLineMetaDirty = false
	return t.wrappedLineStarts, t.wrappedLineLengths
}

func (t *TextArea) visibleLogicalLines(totalLines int) []int {
	if totalLines <= 0 {
		return nil
	}
	if len(t.visibleLines) == 0 {
		lines := make([]int, totalLines)
		for i := 0; i < totalLines; i++ {
			lines[i] = i
		}
		return lines
	}

	lines := make([]int, 0, len(t.visibleLines))
	for _, line := range t.visibleLines {
		if line < 0 || line >= totalLines {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, 0)
	}
	return lines
}

func (t *TextArea) lineSlice(start, length int) string {
	if start < 0 || length < 0 || start > len(t.text) {
		return ""
	}
	end := start + length
	if end > len(t.text) {
		end = len(t.text)
	}
	return string(t.text[start:end])
}

func (t *TextArea) cursorLineCol(starts []int, lengths []int) (int, int) {
	if len(starts) == 0 {
		return 0, 0
	}
	for i, start := range starts {
		if t.cursor < start {
			if i == 0 {
				return 0, 0
			}
			prev := i - 1
			return prev, lengths[prev]
		}
		end := start + lengths[i]
		if t.cursor <= end {
			return i, t.cursor - start
		}
	}
	last := len(starts) - 1
	return last, lengths[last]
}

func (t *TextArea) cursorLineColForOffset(starts []int, lengths []int, offset int) (int, int) {
	if len(starts) == 0 {
		return 0, 0
	}
	for i, start := range starts {
		end := start + lengths[i]
		if offset <= end {
			return i, offset - start
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
	t.Base.State.Multiline = true
	t.Base.Label = label
	t.Base.Value = &accessibility.ValueInfo{Text: t.Text()}
}

// --- Selection support ---

// GetSelection returns the current selection range.
func (t *TextArea) GetSelection() Selection {
	if t == nil {
		return Selection{}
	}
	return t.selection
}

// SetSelection sets the selection range, clamping to valid bounds.
func (t *TextArea) SetSelection(sel Selection) {
	if t == nil {
		return
	}
	textLen := len(t.text)
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
	t.selection = sel
	t.services.Invalidate()
}

// SelectAll selects all text.
func (t *TextArea) SelectAll() {
	if t == nil {
		return
	}
	t.selection = Selection{Start: 0, End: len(t.text)}
	t.cursor = len(t.text)
	t.services.Invalidate()
}

// SelectNone clears the selection.
func (t *TextArea) SelectNone() {
	if t == nil {
		return
	}
	t.selection = Selection{}
	t.services.Invalidate()
}

// SelectWord selects the word at the cursor position.
func (t *TextArea) SelectWord() {
	if t == nil {
		return
	}
	left := textAreaWordBoundaryLeft(t.text, t.cursor)
	right := textAreaWordBoundaryRight(t.text, t.cursor)
	t.selection = Selection{Start: left, End: right}
	t.services.Invalidate()
}

// SelectLine selects the current line.
func (t *TextArea) SelectLine() {
	if t == nil {
		return
	}
	lineStarts, lineLengths := t.lineMeta()
	line, _ := t.cursorLineCol(lineStarts, lineLengths)
	if line < 0 || line >= len(lineStarts) {
		return
	}
	start := lineStarts[line]
	end := start + lineLengths[line]
	t.selection = Selection{Start: start, End: end}
	t.services.Invalidate()
}

// HasSelection returns true if text is selected.
func (t *TextArea) HasSelection() bool {
	if t == nil {
		return false
	}
	return !t.selection.IsEmpty()
}

// GetSelectedText returns the currently selected text.
func (t *TextArea) GetSelectedText() string {
	if t == nil || t.selection.IsEmpty() {
		return ""
	}
	sel := t.selection.Normalize()
	if sel.Start > len(t.text) {
		sel.Start = len(t.text)
	}
	if sel.End > len(t.text) {
		sel.End = len(t.text)
	}
	return string(t.text[sel.Start:sel.End])
}

// deleteSelection removes the selected text, sets cursor to selection start.
func (t *TextArea) deleteSelection() {
	if t == nil || t.selection.IsEmpty() {
		return
	}
	sel := t.selection.Normalize()
	if sel.End > len(t.text) {
		sel.End = len(t.text)
	}
	if sel.Start > len(t.text) {
		sel.Start = len(t.text)
	}
	t.text = append(t.text[:sel.Start], t.text[sel.End:]...)
	t.cursor = sel.Start
	t.selection = Selection{}
	t.invalidateLineMeta()
	t.syncValue()
}

// collapseSelection moves cursor to selection start or end, clears selection.
// Returns true if there was a selection to collapse.
func (t *TextArea) collapseSelection(toStart bool) bool {
	if t == nil || t.selection.IsEmpty() {
		return false
	}
	sel := t.selection.Normalize()
	if toStart {
		t.cursor = sel.Start
	} else {
		t.cursor = sel.End
	}
	t.selection = Selection{}
	return true
}

// extendSelection sets up or extends selection from anchor to current cursor.
// Call this after moving the cursor with shift held.
func (t *TextArea) extendSelection(anchor int) {
	t.selection = Selection{Start: anchor, End: t.cursor}
}

var _ Selectable = (*TextArea)(nil)

// ClipboardCopy returns selected text, or all text if no selection.
func (t *TextArea) ClipboardCopy() (string, bool) {
	if t == nil {
		return "", false
	}
	if t.HasSelection() {
		return t.GetSelectedText(), true
	}
	return t.Text(), true
}

// ClipboardCut cuts selected text, or all text if no selection.
func (t *TextArea) ClipboardCut() (string, bool) {
	if t == nil {
		return "", false
	}
	t.pushHistory(false)
	if t.HasSelection() {
		text := t.GetSelectedText()
		t.deleteSelection()
		return text, true
	}
	text := t.Text()
	t.text = nil
	t.cursor = 0
	t.scrollY = 0
	t.selection = Selection{}
	t.invalidateLineMeta()
	t.syncValue()
	if text != "" {
		if announcer := t.services.Announcer(); announcer != nil {
			announcer.Announce("Text cleared", accessibility.PriorityPolite)
		}
	}
	return text, true
}

// ClipboardPaste inserts text at the cursor, replacing any selection.
func (t *TextArea) ClipboardPaste(text string) bool {
	if t == nil || text == "" {
		return false
	}
	// Large pastes (> 100 chars) are not grouped
	grouped := len(text) <= 100
	t.pushHistory(grouped)
	if t.HasSelection() {
		t.deleteSelection()
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
var _ runtime.Bindable = (*TextArea)(nil)
var _ runtime.Unbindable = (*TextArea)(nil)
var _ Validatable = (*TextArea)(nil)
