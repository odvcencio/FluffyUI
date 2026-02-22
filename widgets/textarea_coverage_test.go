package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestTextAreaCoverage(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	area := NewTextArea()
	area.Bind(app.Services())

	area.SetText("hello\nworld")
	_ = area.CursorOffset()
	_, _ = area.CursorPosition()
	area.SetCursorOffset(1)
	area.SetCursorPosition(1, 0)
	area.CursorWordLeft()
	area.CursorWordRight()

	area.SetOnChange(func(text string) {})
	area.OnChange(func(text string) {})
	area.SetValidators(forms.Required("required"))
	_ = area.Errors()
	_ = area.Valid()
	area.SetLabel("Notes")
	area.SetStyle(backend.DefaultStyle())
	area.SetFocusStyle(backend.DefaultStyle().Bold(true))
	_ = area.StyleType()

	area.Focus()
	area.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 4})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'x'})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlC})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlX})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlV})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	_, _ = area.ClipboardCopy()
	_, _ = area.ClipboardCut()
	_ = area.ClipboardPaste("paste")
	_ = mem.Write("paste")
	_ = area.pasteFromClipboard()

	// Undo/redo via key messages
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlZ})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlZ, Shift: true})
	area.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlY})

	_ = flufftest.RenderToString(area, 20, 4)
	area.Unbind()
}

func TestTextArea_UndoRedo(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello")
	ta.SetText("world")

	if !ta.CanUndo() {
		t.Fatal("expected CanUndo after two SetText calls")
	}

	ta.Undo()
	if string(ta.text) != "hello" {
		t.Fatalf("after undo expected 'hello', got %q", string(ta.text))
	}

	if !ta.CanRedo() {
		t.Fatal("expected CanRedo after undo")
	}

	ta.Redo()
	if string(ta.text) != "world" {
		t.Fatalf("after redo expected 'world', got %q", string(ta.text))
	}
}

func TestTextArea_UndoAtRoot(t *testing.T) {
	ta := NewTextArea()
	if ta.CanUndo() {
		t.Fatal("expected no undo at start")
	}
	if ta.Undo() {
		t.Fatal("expected Undo to return false at root")
	}
}

func TestTextArea_ClearHistory(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("a")
	ta.SetText("b")
	ta.ClearHistory()
	if ta.CanUndo() {
		t.Fatal("expected no undo after ClearHistory")
	}
}

func TestTextArea_UndoRedoNilSafety(t *testing.T) {
	var ta *TextArea
	if ta.Undo() {
		t.Fatal("expected Undo on nil to return false")
	}
	if ta.Redo() {
		t.Fatal("expected Redo on nil to return false")
	}
	if ta.CanUndo() {
		t.Fatal("expected CanUndo on nil to return false")
	}
	if ta.CanRedo() {
		t.Fatal("expected CanRedo on nil to return false")
	}
	ta.ClearHistory() // should not panic
}

func TestTextArea_Highlights(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("func main()")
	ta.Focus()

	kwStyle := backend.DefaultStyle().Foreground(backend.ColorRed)
	fnStyle := backend.DefaultStyle().Foreground(backend.ColorGreen)

	ta.SetHighlights([]TextAreaHighlight{
		{Start: 0, End: 4, Style: kwStyle}, // "func"
		{Start: 5, End: 9, Style: fnStyle}, // "main"
	})

	buf := flufftest.LayoutAndRender(ta, 20, 3)

	// "func" (positions 0-3) should have red foreground
	for i := 0; i < 4; i++ {
		cell := buf.Get(i, 0)
		fg, _, _ := cell.Style.Decompose()
		if fg != backend.ColorRed {
			t.Errorf("char %d: expected red foreground, got %v", i, fg)
		}
	}

	// "main" (positions 5-8) should have green foreground
	for i := 5; i < 9; i++ {
		cell := buf.Get(i, 0)
		fg, _, _ := cell.Style.Decompose()
		if fg != backend.ColorGreen {
			t.Errorf("char %d: expected green foreground, got %v", i, fg)
		}
	}

	// Non-highlighted chars should have default foreground
	cell := buf.Get(4, 0) // space between "func" and "main"
	fg, _, _ := cell.Style.Decompose()
	if fg == backend.ColorRed || fg == backend.ColorGreen {
		t.Errorf("space at pos 4 should not be highlighted, got fg=%v", fg)
	}

	// ClearHighlights should remove all highlights
	ta.ClearHighlights()
	buf = flufftest.LayoutAndRender(ta, 20, 3)
	cell = buf.Get(0, 0)
	fg, _, _ = cell.Style.Decompose()
	if fg == backend.ColorRed {
		t.Error("after ClearHighlights, char 0 should not be red")
	}
}

func TestTextArea_HighlightMergesAttributes(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("bold")
	ta.Focus()

	hlStyle := backend.DefaultStyle().Bold(true).Foreground(backend.ColorBlue)
	ta.SetHighlights([]TextAreaHighlight{
		{Start: 0, End: 4, Style: hlStyle},
	})

	buf := flufftest.LayoutAndRender(ta, 20, 3)
	cell := buf.Get(0, 0)
	_, _, attrs := cell.Style.Decompose()
	if attrs&backend.AttrBold == 0 {
		t.Error("expected bold attribute on highlighted cell")
	}
	fg, _, _ := cell.Style.Decompose()
	if fg != backend.ColorBlue {
		t.Errorf("expected blue foreground, got %v", fg)
	}
}

func TestTextArea_LineNumbers(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("aaa\nbbb\nccc")
	ta.SetShowLineNumbers(true)
	ta.Focus()

	// 3 lines → gutter width = digits(3) + 1 = 2
	buf := flufftest.LayoutAndRender(ta, 20, 5)

	// Check line numbers at left edge
	// Line 1: "1 " at columns 0-1
	c := buf.Get(0, 0)
	if c.Rune != '1' {
		t.Errorf("expected '1' at (0,0), got %q", c.Rune)
	}
	c = buf.Get(1, 0)
	if c.Rune != ' ' {
		t.Errorf("expected space separator at (1,0), got %q", c.Rune)
	}

	// Text starts at column 2
	c = buf.Get(2, 0)
	if c.Rune != 'a' {
		t.Errorf("expected 'a' at (2,0), got %q", c.Rune)
	}

	// Line 2
	c = buf.Get(0, 1)
	if c.Rune != '2' {
		t.Errorf("expected '2' at (0,1), got %q", c.Rune)
	}
	c = buf.Get(2, 1)
	if c.Rune != 'b' {
		t.Errorf("expected 'b' at (2,1), got %q", c.Rune)
	}

	// Line 3
	c = buf.Get(0, 2)
	if c.Rune != '3' {
		t.Errorf("expected '3' at (0,2), got %q", c.Rune)
	}
}

func TestTextArea_LineNumbersGutterStyle(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("a\nb")
	ta.SetShowLineNumbers(true)
	gs := backend.DefaultStyle().Foreground(backend.ColorYellow)
	ta.SetGutterStyle(gs)
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 20, 5)
	cell := buf.Get(0, 0)
	fg, _, _ := cell.Style.Decompose()
	if fg != backend.ColorYellow {
		t.Errorf("gutter foreground: expected yellow, got %v", fg)
	}
}

func TestTextArea_TabSpaces(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("x")
	ta.SetCursorOffset(1)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	got := ta.Text()
	if got != "x    " {
		t.Fatalf("expected 4 spaces after tab, got %q", got)
	}
}

func TestTextArea_TabLiteral(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("x")
	ta.SetCursorOffset(1)
	ta.SetTabMode(true) // use literal tabs
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	got := ta.Text()
	if got != "x\t" {
		t.Fatalf("expected literal tab, got %q", got)
	}
}

func TestTextArea_TabSize(t *testing.T) {
	ta := NewTextArea()
	ta.SetTabSize(2)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	got := ta.Text()
	if got != "  " {
		t.Fatalf("expected 2 spaces, got %q", got)
	}
}

func TestTextArea_PageUpDown(t *testing.T) {
	// Build 20 lines
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	text := strings.Join(lines, "\n")

	ta := NewTextArea()
	ta.SetText(text)
	ta.SetCursorOffset(0) // start at line 0
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	// PageDown should move cursor down by page height (5)
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	_, line := ta.CursorPosition()
	if line != 5 {
		t.Fatalf("after PageDown: expected line 5, got %d", line)
	}

	// Another PageDown
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	_, line = ta.CursorPosition()
	if line != 10 {
		t.Fatalf("after 2nd PageDown: expected line 10, got %d", line)
	}

	// PageUp
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	_, line = ta.CursorPosition()
	if line != 5 {
		t.Fatalf("after PageUp: expected line 5, got %d", line)
	}
}

func TestTextArea_MouseClick(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello\nworld\nfoo")
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	// Click at (3, 1) should position cursor at line 1, col 3 = "rld" -> 'r' is at col 2
	// Actually (3,1) = col 3 of "world" = 'l'
	ta.HandleMessage(runtime.MouseMsg{
		X:      3,
		Y:      1,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})

	col, line := ta.CursorPosition()
	if line != 1 {
		t.Fatalf("expected line 1, got %d", line)
	}
	if col != 3 {
		t.Fatalf("expected col 3, got %d", col)
	}
}

func TestTextArea_MouseClickWithLineNumbers(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("ab\ncd")
	ta.SetShowLineNumbers(true)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	// Gutter width for 2 lines = digits(2) + 1 = 2
	// Click at (3, 0): text starts at x=2, so text col = 3 - 2 = 1
	ta.HandleMessage(runtime.MouseMsg{
		X:      3,
		Y:      0,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})

	col, line := ta.CursorPosition()
	if line != 0 {
		t.Fatalf("expected line 0, got %d", line)
	}
	if col != 1 {
		t.Fatalf("expected col 1, got %d", col)
	}
}

func TestTextArea_WordWrapRendering(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("abcdefghij")
	ta.SetWordWrap(true)
	ta.SetShowLineNumbers(false)
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 4, 4)
	if buf.Get(0, 0).Rune != 'a' || buf.Get(3, 0).Rune != 'd' {
		t.Fatalf("expected first wrapped row to be 'abcd', got %q%q", buf.Get(0, 0).Rune, buf.Get(3, 0).Rune)
	}
	if buf.Get(0, 1).Rune != 'e' || buf.Get(3, 1).Rune != 'h' {
		t.Fatalf("expected second wrapped row to be 'efgh', got %q%q", buf.Get(0, 1).Rune, buf.Get(3, 1).Rune)
	}
	if buf.Get(0, 2).Rune != 'i' || buf.Get(1, 2).Rune != 'j' {
		t.Fatalf("expected third wrapped row to be 'ij', got %q%q", buf.Get(0, 2).Rune, buf.Get(1, 2).Rune)
	}
}

func TestTextArea_WordWrapVerticalMovement(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("abcdefghij")
	ta.SetWordWrap(true)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 4, Height: 4})

	ta.SetCursorOffset(2)
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if ta.CursorOffset() != 6 {
		t.Fatalf("expected cursor offset 6 after Down in wrapped row, got %d", ta.CursorOffset())
	}

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if ta.CursorOffset() != 2 {
		t.Fatalf("expected cursor offset 2 after Up in wrapped row, got %d", ta.CursorOffset())
	}
}

func TestTextArea_WordWrapMouse(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("abcdefghij")
	ta.SetWordWrap(true)
	ta.SetShowLineNumbers(false)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 4, Height: 4})

	// Click on row 1, col 1 => expected row1 segment "efgh", position 'f' (offset 5).
	ta.HandleMessage(runtime.MouseMsg{
		X:      1,
		Y:      1,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})
	if ta.CursorOffset() != 5 {
		t.Fatalf("expected cursor at offset 5 after wrapped-line click, got %d", ta.CursorOffset())
	}
}

func TestTextArea_WordWrapLineNumbers(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("abcdefghij")
	ta.SetWordWrap(true)
	ta.SetShowLineNumbers(true)
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 6, 4)
	// Width 6 => gutter width = 2, text width = 4.
	if buf.Get(0, 0).Rune != '1' {
		t.Fatalf("expected first logical line number on wrapped line 0, got %q", buf.Get(0, 0).Rune)
	}
	if buf.Get(0, 1).Rune != ' ' {
		t.Fatalf("expected continuation line to have blank gutter, got %q", buf.Get(0, 1).Rune)
	}
	if buf.Get(0, 1).Rune == '1' {
		t.Fatalf("did not expect line number on wrapped continuation row")
	}
}

func TestTextArea_VisibleLinesFilter(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.SetVisibleLines([]int{0, 2})
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 10, 4)
	if got := string([]rune{buf.Get(0, 0).Rune, buf.Get(1, 0).Rune, buf.Get(2, 0).Rune, buf.Get(3, 0).Rune, buf.Get(4, 0).Rune}); got != "line1" {
		t.Fatalf("row 0 = %q, want %q", got, "line1")
	}
	if got := string([]rune{buf.Get(0, 1).Rune, buf.Get(1, 1).Rune, buf.Get(2, 1).Rune, buf.Get(3, 1).Rune, buf.Get(4, 1).Rune}); got != "line3" {
		t.Fatalf("row 1 = %q, want %q", got, "line3")
	}
}

func TestTextArea_VisibleLinesAffectVerticalMovement(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("a\nb\nc\nd")
	ta.SetVisibleLines([]int{0, 2, 3})
	ta.SetCursorPosition(0, 0)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 4})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	_, row := ta.CursorPosition()
	if row != 2 {
		t.Fatalf("cursor row after Down = %d, want 2", row)
	}

	ta.SetVisibleLines(nil)
	if got := ta.VisibleLines(); got != nil {
		t.Fatalf("VisibleLines() after reset = %v, want nil", got)
	}
}

func TestTextArea_CtrlHome(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("aaa\nbbb\nccc")
	ta.SetCursorOffset(len([]rune("aaa\nbbb\nccc"))) // end of text
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome, Ctrl: true})
	if ta.CursorOffset() != 0 {
		t.Fatalf("expected cursor at 0, got %d", ta.CursorOffset())
	}
}

func TestTextArea_CtrlEnd(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("aaa\nbbb\nccc")
	ta.SetCursorOffset(0)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd, Ctrl: true})
	expected := len([]rune("aaa\nbbb\nccc"))
	if ta.CursorOffset() != expected {
		t.Fatalf("expected cursor at %d, got %d", expected, ta.CursorOffset())
	}
}

func TestTextArea_NilSafetyNewMethods(t *testing.T) {
	var ta *TextArea
	// Should not panic
	ta.SetHighlights(nil)
	ta.ClearHighlights()
	ta.SetShowLineNumbers(true)
	ta.SetGutterStyle(backend.DefaultStyle())
	ta.SetTabSize(8)
	ta.SetTabMode(true)
}

func TestTextArea_ScrollXPersists(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("abcdefghijklmnopqrstuvwxyz")
	ta.SetCursorOffset(20) // well past a 10-col viewport
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 10, 3)
	_ = buf

	// After render, scrollX should be set to show cursor
	if ta.scrollX == 0 {
		t.Fatal("expected scrollX > 0 for cursor at offset 20 in 10-col viewport")
	}
}

func TestTextArea_SelectAll(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	ta := NewTextArea()
	ta.Bind(app.Services())
	ta.SetText("hello\nworld")
	ta.SetCursorOffset(0)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})

	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlA})
	if !ta.HasSelection() {
		t.Fatal("expected selection after Ctrl+A")
	}
	sel := ta.GetSelection()
	if sel.Start != 0 || sel.End != len([]rune("hello\nworld")) {
		t.Fatalf("expected full selection, got %d-%d", sel.Start, sel.End)
	}
	if ta.GetSelectedText() != "hello\nworld" {
		t.Fatalf("expected selected text 'hello\\nworld', got %q", ta.GetSelectedText())
	}
}

func TestTextArea_ShiftArrowSelection(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello")
	ta.SetCursorOffset(0)
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	// Shift+Right three times to select "hel"
	for i := 0; i < 3; i++ {
		ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight, Shift: true})
	}
	if !ta.HasSelection() {
		t.Fatal("expected selection after Shift+Right")
	}
	if ta.GetSelectedText() != "hel" {
		t.Fatalf("expected 'hel', got %q", ta.GetSelectedText())
	}

	// Right without shift should collapse selection
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if ta.HasSelection() {
		t.Fatal("expected no selection after Right without Shift")
	}
}

func TestTextArea_SelectionDelete(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello world")
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	// Select "hello" (0-5)
	ta.SetSelection(Selection{Start: 0, End: 5})
	ta.SetCursorOffset(5)

	// Backspace should delete selection
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if ta.Text() != " world" {
		t.Fatalf("expected ' world', got %q", ta.Text())
	}
	if ta.HasSelection() {
		t.Fatal("expected no selection after delete")
	}
}

func TestTextArea_SelectionReplace(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello world")
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	// Select "hello" and type "hi"
	ta.SetSelection(Selection{Start: 0, End: 5})
	ta.SetCursorOffset(5)
	ta.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'H'})
	if ta.Text() != "H world" {
		t.Fatalf("expected 'H world', got %q", ta.Text())
	}
}

func TestTextArea_ClipboardCopySelection(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	ta := NewTextArea()
	ta.Bind(app.Services())
	ta.SetText("hello world")
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	// Select "world" (6-11)
	ta.SetSelection(Selection{Start: 6, End: 11})

	text, ok := ta.ClipboardCopy()
	if !ok {
		t.Fatal("expected ClipboardCopy to succeed")
	}
	if text != "world" {
		t.Fatalf("expected 'world', got %q", text)
	}
}

func TestTextArea_ClipboardCutSelection(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	ta := NewTextArea()
	ta.Bind(app.Services())
	ta.SetText("hello world")
	ta.Focus()
	ta.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})

	// Select "hello" (0-5)
	ta.SetSelection(Selection{Start: 0, End: 5})

	text, ok := ta.ClipboardCut()
	if !ok {
		t.Fatal("expected ClipboardCut to succeed")
	}
	if text != "hello" {
		t.Fatalf("expected 'hello', got %q", text)
	}
	if ta.Text() != " world" {
		t.Fatalf("expected ' world', got %q", ta.Text())
	}
}

func TestTextArea_SelectionRendering(t *testing.T) {
	ta := NewTextArea()
	// Use non-reverse styles so selection reverse is distinguishable
	plainStyle := backend.DefaultStyle().Foreground(backend.ColorWhite)
	ta.SetStyle(plainStyle)
	ta.SetFocusStyle(plainStyle)
	ta.SetText("abcdef")
	ta.SetCursorOffset(0)
	ta.SetSelection(Selection{Start: 1, End: 4}) // select "bcd"
	ta.Focus()

	buf := flufftest.LayoutAndRender(ta, 20, 3)

	// Characters in selection should have reverse style
	for i := 1; i < 4; i++ {
		cell := buf.Get(i, 0)
		_, _, attrs := cell.Style.Decompose()
		if attrs&backend.AttrReverse == 0 {
			t.Errorf("char %d: expected reverse attribute for selection", i)
		}
	}

	// Character before selection should not be reversed (cursor is at 0, which IS reversed for cursor)
	// Check character after selection instead
	cell := buf.Get(4, 0) // 'e' (outside selection)
	_, _, attrs := cell.Style.Decompose()
	if attrs&backend.AttrReverse != 0 {
		t.Error("char 4 should not have reverse attribute")
	}
}

func TestTextArea_SelectNone(t *testing.T) {
	ta := NewTextArea()
	app := runtime.NewApp(runtime.AppConfig{})
	ta.Bind(app.Services())
	ta.SetText("hello")
	ta.SelectAll()
	if !ta.HasSelection() {
		t.Fatal("expected selection after SelectAll")
	}
	ta.SelectNone()
	if ta.HasSelection() {
		t.Fatal("expected no selection after SelectNone")
	}
}

func TestTextArea_SelectWord(t *testing.T) {
	ta := NewTextArea()
	app := runtime.NewApp(runtime.AppConfig{})
	ta.Bind(app.Services())
	ta.SetText("hello world")
	ta.SetCursorOffset(7) // inside "world"
	ta.SelectWord()
	got := ta.GetSelectedText()
	if got != "world" {
		t.Fatalf("expected 'world', got %q", got)
	}
}

func TestTextArea_SelectLine(t *testing.T) {
	ta := NewTextArea()
	app := runtime.NewApp(runtime.AppConfig{})
	ta.Bind(app.Services())
	ta.SetText("hello\nworld\nfoo")
	ta.SetCursorOffset(7) // inside "world" (line 1)
	ta.SelectLine()
	got := ta.GetSelectedText()
	if got != "world" {
		t.Fatalf("expected 'world', got %q", got)
	}
}

func TestTextArea_SelectableInterface(t *testing.T) {
	// Compile-time check is done with var _ Selectable = (*TextArea)(nil)
	// This test verifies nil safety
	var ta *TextArea
	if ta.HasSelection() {
		t.Fatal("nil HasSelection should return false")
	}
	if ta.GetSelectedText() != "" {
		t.Fatal("nil GetSelectedText should return empty")
	}
	_ = ta.GetSelection()
	ta.SetSelection(Selection{Start: 0, End: 5})
	ta.SelectAll()
	ta.SelectNone()
	ta.SelectWord()
	ta.SelectLine()
}
