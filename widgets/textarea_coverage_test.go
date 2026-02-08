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
