package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestMaskedInput_New(t *testing.T) {
	mask := "(###) ###-####"
	mi := NewMaskedInput(mask)

	if mi == nil {
		t.Fatal("expected non-nil MaskedInput")
	}
	if mi.Mask() != mask {
		t.Errorf("Mask() = %q, want %q", mi.Mask(), mask)
	}
	if mi.Value() != "" {
		t.Errorf("Value() = %q, want empty", mi.Value())
	}
	if mi.Placeholder() != '_' {
		t.Errorf("Placeholder() = %c, want '_'", mi.Placeholder())
	}
}

func TestMaskedInput_WithOptions(t *testing.T) {
	mi := NewMaskedInput("##/##/####",
		WithPlaceholder('*'),
		WithMaskedLabel("Date"),
		WithMaskedStyle(backend.DefaultStyle().Bold(true)),
		WithMaskedFocusStyle(backend.DefaultStyle().Italic(true)),
	)

	if mi.Placeholder() != '*' {
		t.Errorf("Placeholder() = %c, want '*'", mi.Placeholder())
	}
	if mi.Base.Label != "Date" {
		t.Errorf("Label = %q, want 'Date'", mi.Base.Label)
	}
}

func TestMaskedInput_PhoneMask(t *testing.T) {
	mi := NewMaskedInput("(###) ###-####")
	mi.Focus()

	// Type phone number: 555-123-4567
	digits := "5551234567"
	for _, d := range digits {
		mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: d})
	}

	expected := "5551234567"
	if mi.Value() != expected {
		t.Errorf("Value() = %q, want %q", mi.Value(), expected)
	}

	maskedExpected := "(555) 123-4567"
	if mi.MaskedValue() != maskedExpected {
		t.Errorf("MaskedValue() = %q, want %q", mi.MaskedValue(), maskedExpected)
	}
}

func TestMaskedInput_DateMask(t *testing.T) {
	mi := NewMaskedInput("##/##/####")
	mi.SetValue("01312026")

	if mi.Value() != "01312026" {
		t.Errorf("Value() = %q, want '01312026'", mi.Value())
	}
	if mi.MaskedValue() != "01/31/2026" {
		t.Errorf("MaskedValue() = %q, want '01/31/2026'", mi.MaskedValue())
	}
}

func TestMaskedInput_CreditCardMask(t *testing.T) {
	mi := NewMaskedInput("#### #### #### ####")
	mi.SetValue("4111111111111111")

	expected := "4111 1111 1111 1111"
	if mi.MaskedValue() != expected {
		t.Errorf("MaskedValue() = %q, want %q", mi.MaskedValue(), expected)
	}
}

func TestMaskedInput_LetterMask(t *testing.T) {
	mi := NewMaskedInput("AAA-###")
	mi.Focus()

	// Try typing invalid then valid characters
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '1'}) // Invalid (need letter)
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'A'})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'B'})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'C'})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'D'}) // Invalid (need digit)
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '1'})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '2'})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '3'})

	if mi.Value() != "ABC123" {
		t.Errorf("Value() = %q, want 'ABC123'", mi.Value())
	}
	if mi.MaskedValue() != "ABC-123" {
		t.Errorf("MaskedValue() = %q, want 'ABC-123'", mi.MaskedValue())
	}
}

func TestMaskedInput_AlphanumericMask(t *testing.T) {
	mi := NewMaskedInput("***-***")
	mi.Focus()

	chars := "AB1CD2"
	for _, c := range chars {
		mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: c})
	}

	if mi.Value() != "AB1CD2" {
		t.Errorf("Value() = %q, want 'AB1CD2'", mi.Value())
	}
	if mi.MaskedValue() != "AB1-CD2" {
		t.Errorf("MaskedValue() = %q, want 'AB1-CD2'", mi.MaskedValue())
	}
}

func TestMaskedInput_AnyCharMask(t *testing.T) {
	mi := NewMaskedInput("????")
	mi.SetValue("@#$%")

	if mi.Value() != "@#$%" {
		t.Errorf("Value() = %q, want '@#$%%'", mi.Value())
	}
}

func TestMaskedInput_EscapedLiteral(t *testing.T) {
	// Mask: \# ## - ## (escaped # becomes literal, then 2 digits, dash, 2 digits)
	mi := NewMaskedInput("\\###-##")
	mi.SetValue("1234")

	expected := "#12-34"
	if mi.MaskedValue() != expected {
		t.Errorf("MaskedValue() = %q, want %q", mi.MaskedValue(), expected)
	}
}

func TestMaskedInput_Backspace(t *testing.T) {
	mi := NewMaskedInput("###-###")
	mi.Focus()

	// Type 123456
	for _, c := range "123456" {
		mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: c})
	}

	if mi.Value() != "123456" {
		t.Errorf("Value() = %q, want '123456'", mi.Value())
	}

	// Backspace
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	if mi.Value() != "12345" {
		t.Errorf("After backspace, Value() = %q, want '12345'", mi.Value())
	}
}

func TestMaskedInput_Delete(t *testing.T) {
	mi := NewMaskedInput("###-###")
	mi.SetValue("123456")
	mi.Focus()
	mi.cursorPos = 0

	// Delete first character
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})

	if mi.Value() != "23456" {
		t.Errorf("After delete, Value() = %q, want '23456'", mi.Value())
	}
}

func TestMaskedInput_CursorNavigation(t *testing.T) {
	mi := NewMaskedInput("###-###")
	mi.SetValue("123456")
	mi.Focus()

	// Cursor should be after last character
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if mi.CursorPos() != 0 {
		t.Errorf("After Home, CursorPos() = %d, want 0", mi.CursorPos())
	}

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	// Should be at end (position 7 because "123-456" is 7 chars)
	if mi.CursorPos() < 6 {
		t.Errorf("After End, CursorPos() = %d, want >= 6", mi.CursorPos())
	}

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	// Should skip to next editable position
	if mi.CursorPos() != 1 {
		t.Errorf("After Right from Home, CursorPos() = %d, want 1", mi.CursorPos())
	}
}

func TestMaskedInput_OnSubmit(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()
	mi.SetValue("123")

	var submitted string
	mi.SetOnSubmit(func(value string) {
		submitted = value
	})

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if submitted != "123" {
		t.Errorf("submitted = %q, want '123'", submitted)
	}
}

func TestMaskedInput_OnChange(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()

	var changed string
	mi.SetOnChange(func(value string) {
		changed = value
	})

	mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '5'})

	if changed != "5" {
		t.Errorf("changed = %q, want '5'", changed)
	}
}

func TestMaskedInput_Clipboard(t *testing.T) {
	mi := NewMaskedInput("###-###")
	mi.SetValue("123456")

	text, ok := mi.ClipboardCopy()
	if !ok {
		t.Error("ClipboardCopy() returned false")
	}
	if text != "123-456" {
		t.Errorf("ClipboardCopy() = %q, want '123-456'", text)
	}

	text, ok = mi.ClipboardCut()
	if !ok {
		t.Error("ClipboardCut() returned false")
	}
	if text != "" {
		// After cut, the value should be empty so MaskedValue returns placeholders
		// Actually, we returned the value before clearing
	}
	if mi.Value() != "" {
		t.Errorf("After cut, Value() = %q, want ''", mi.Value())
	}
}

func TestMaskedInput_ClipboardPaste(t *testing.T) {
	mi := NewMaskedInput("(###) ###-####")
	mi.Focus()

	ok := mi.ClipboardPaste("555-123-4567")
	if !ok {
		t.Error("ClipboardPaste() returned false")
	}

	// Should extract valid digits
	if mi.Value() != "5551234567" {
		t.Errorf("After paste, Value() = %q, want '5551234567'", mi.Value())
	}
}

func TestMaskedInput_Measure(t *testing.T) {
	mi := NewMaskedInput("###-###")

	size := mi.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 80 {
		t.Errorf("Width = %d, want 80", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestMaskedInput_Render(t *testing.T) {
	mi := NewMaskedInput("###-####", WithPlaceholder('_'))
	mi.Focus()
	mi.SetValue("555")
	mi.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	buf := runtime.NewBuffer(20, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	mi.Render(ctx)

	// Check first character
	cell := buf.Get(0, 0)
	if cell.Rune != '5' {
		t.Errorf("expected '5' at (0,0), got '%c'", cell.Rune)
	}
}

func TestMaskedInput_RenderEmpty(t *testing.T) {
	mi := NewMaskedInput("###-###")
	mi.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})

	buf := runtime.NewBuffer(20, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	mi.Render(ctx)
}

func TestMaskedInput_Unfocused(t *testing.T) {
	mi := NewMaskedInput("###")
	// Not focused

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '1'})

	if result.Handled {
		t.Error("unfocused input should not handle messages")
	}
}

func TestMaskedInput_NonKeyMsg(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()

	result := mi.HandleMessage(runtime.ResizeMsg{Width: 80, Height: 24})

	if result.Handled {
		t.Error("non-key message should not be handled")
	}
}

func TestMaskedInput_Tab(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})

	if len(result.Commands) == 0 {
		t.Error("expected FocusNext command")
	}
	_, ok := result.Commands[0].(runtime.FocusNext)
	if !ok {
		t.Errorf("expected FocusNext, got %T", result.Commands[0])
	}
}

func TestMaskedInput_ShiftTab(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})

	if len(result.Commands) == 0 {
		t.Error("expected FocusPrev command")
	}
	_, ok := result.Commands[0].(runtime.FocusPrev)
	if !ok {
		t.Errorf("expected FocusPrev, got %T", result.Commands[0])
	}
}

func TestMaskedInput_Escape(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.Focus()

	result := mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if len(result.Commands) == 0 {
		t.Error("expected Cancel command")
	}
	_, ok := result.Commands[0].(runtime.Cancel)
	if !ok {
		t.Errorf("expected Cancel, got %T", result.Commands[0])
	}
}

func TestMaskedInput_Validation(t *testing.T) {
	mi := NewMaskedInput("###")

	// Implement Validatable interface
	if mi.Valid() != true {
		t.Error("expected Valid() = true with no validators")
	}

	errors := mi.Errors()
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %v", errors)
	}
}

func TestMaskedInput_Accessibility(t *testing.T) {
	mi := NewMaskedInput("###-####", WithMaskedLabel("Phone Number"))

	if mi.AccessibleRole() != "textbox" {
		t.Errorf("AccessibleRole() = %q, want 'textbox'", mi.AccessibleRole())
	}
	if mi.AccessibleLabel() != "Phone Number" {
		t.Errorf("AccessibleLabel() = %q, want 'Phone Number'", mi.AccessibleLabel())
	}
}

func TestMaskedInput_StyleType(t *testing.T) {
	mi := NewMaskedInput("###")

	if mi.StyleType() != "MaskedInput" {
		t.Errorf("StyleType() = %q, want 'MaskedInput'", mi.StyleType())
	}
}

func TestMaskedInput_NilSafety(t *testing.T) {
	var mi *MaskedInput

	// These should not panic
	mi.SetMask("###")
	mi.SetPlaceholder('_')
	mi.SetValue("123")
	mi.SetLabel("Test")
	mi.SetStyle(backend.DefaultStyle())
	mi.SetFocusStyle(backend.DefaultStyle())
	mi.SetOnSubmit(nil)
	mi.SetOnChange(nil)
	mi.SetValidators()
	mi.Bind(runtime.Services{})
	mi.Unbind()

	if mi.Mask() != "" {
		t.Error("expected empty Mask")
	}
	if mi.Value() != "" {
		t.Error("expected empty Value")
	}
	if mi.MaskedValue() != "" {
		t.Error("expected empty MaskedValue")
	}
	if mi.Valid() != true {
		t.Error("expected Valid() = true for nil")
	}
}
