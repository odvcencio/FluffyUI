package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestInputCursorAndClipboard(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	in := NewInput()
	in.Bind(app.Services())

	in.SetText("hello world")
	_ = in.StyleType()
	in.SetCursorOffset(5)
	if in.CursorOffset() != 5 {
		t.Fatalf("expected cursor offset 5")
	}
	_, _ = in.CursorPosition()
	in.SetCursorPosition(3, 0)
	in.CursorWordLeft()
	in.CursorWordRight()

	in.SetValidators(forms.Required("required"))
	_ = in.Errors()
	_ = in.Valid()
	in.OnSubmit(func(text string) {})
	in.OnChange(func(text string) {})

	in.SetSelection(Selection{Start: 0, End: 5})
	if !in.HasSelection() {
		t.Fatalf("expected selection")
	}
	_ = in.GetSelectedText()
	_, _ = in.ClipboardCopy()
	_, _ = in.ClipboardCut()
	_ = in.ClipboardPaste("hi")

	_ = in.copyToClipboard()
	_ = in.cutToClipboard()
	_ = mem.Write("paste")
	_ = in.pasteFromClipboard()
	in.insertText("!")

	in.SelectAll()
	in.SelectNone()
	in.SelectWord()
	in.SelectLine()
	in.Clear()

	_ = flufftest.RenderToString(in, 20, 1)
}

func TestMultilineInputInteractions(t *testing.T) {
	mem := &clipboard.MemoryClipboard{}
	app := runtime.NewApp(runtime.AppConfig{Clipboard: mem})
	m := NewMultilineInput()
	m.Bind(app.Services())

	m.SetText("hello\nworld")
	m.SetCursorPosition(0, 0)
	m.SetCursorOffset(3)
	m.CursorWordRight()
	m.CursorWordLeft()
	m.SetOnSubmit(func(text string) {})
	m.SetOnChange(func(text string) {})
	m.SetValidators(forms.Required("required"))
	_ = m.Errors()
	_ = m.Valid()
	m.SetLabel("Notes")
	m.SetStyle(backend.DefaultStyle())
	m.SetFocusStyle(backend.DefaultStyle().Bold(true))

	m.Focus()
	m.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 4})

	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'x'})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlC})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlX})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlV})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter, Ctrl: true})
	m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	m.SelectAll()
	m.SelectWord()
	m.SelectLine()
	m.SelectNone()
	_, _ = m.ClipboardCopy()
	_, _ = m.ClipboardCut()
	_ = m.ClipboardPaste("paste")
	_ = mem.Write("paste")
	_ = m.pasteFromClipboard()

	_ = flufftest.RenderToString(m, 20, 4)
	m.Unbind()
}
