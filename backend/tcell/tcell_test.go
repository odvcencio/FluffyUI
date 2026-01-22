package tcell

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func TestReverseConvertKeyEvent(t *testing.T) {
	tests := []struct {
		name  string
		event terminal.KeyEvent
	}{
		{
			name:  "rune a",
			event: terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'a'},
		},
		{
			name:  "enter",
			event: terminal.KeyEvent{Key: terminal.KeyEnter},
		},
		{
			name:  "escape",
			event: terminal.KeyEvent{Key: terminal.KeyEscape},
		},
		{
			name:  "arrow up",
			event: terminal.KeyEvent{Key: terminal.KeyUp},
		},
		{
			name:  "with alt modifier",
			event: terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'x', Alt: true},
		},
		{
			name:  "ctrl+c as KeyCtrlC",
			event: terminal.KeyEvent{Key: terminal.KeyCtrlC},
		},
		{
			name:  "F1 key",
			event: terminal.KeyEvent{Key: terminal.KeyF1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseConvertEvent(tt.event)
			if result == nil {
				t.Fatal("reverseConvertEvent returned nil for KeyEvent")
			}

			keyEvent, ok := result.(*tcell.EventKey)
			if !ok {
				t.Fatalf("expected *tcell.EventKey, got %T", result)
			}

			// Convert back and verify round-trip
			converted := convertEvent(keyEvent)
			keyConverted, ok := converted.(terminal.KeyEvent)
			if !ok {
				t.Fatalf("round-trip failed: expected terminal.KeyEvent, got %T", converted)
			}

			if keyConverted.Key != tt.event.Key {
				t.Errorf("Key mismatch: got %v, want %v", keyConverted.Key, tt.event.Key)
			}
			if tt.event.Key == terminal.KeyRune && keyConverted.Rune != tt.event.Rune {
				t.Errorf("Rune mismatch: got %v, want %v", keyConverted.Rune, tt.event.Rune)
			}
			if keyConverted.Alt != tt.event.Alt {
				t.Errorf("Alt mismatch: got %v, want %v", keyConverted.Alt, tt.event.Alt)
			}
			if keyConverted.Ctrl != tt.event.Ctrl {
				t.Errorf("Ctrl mismatch: got %v, want %v", keyConverted.Ctrl, tt.event.Ctrl)
			}
		})
	}
}

func TestReverseConvertMouseEvent(t *testing.T) {
	tests := []struct {
		name  string
		event terminal.MouseEvent
	}{
		{
			name:  "left click",
			event: terminal.MouseEvent{X: 10, Y: 5, Button: terminal.MouseLeft, Action: terminal.MousePress},
		},
		{
			name:  "right click",
			event: terminal.MouseEvent{X: 20, Y: 10, Button: terminal.MouseRight, Action: terminal.MousePress},
		},
		{
			name:  "wheel up",
			event: terminal.MouseEvent{X: 0, Y: 0, Button: terminal.MouseWheelUp, Action: terminal.MousePress},
		},
		{
			name:  "with shift",
			event: terminal.MouseEvent{X: 5, Y: 5, Button: terminal.MouseLeft, Action: terminal.MousePress, Shift: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseConvertEvent(tt.event)
			if result == nil {
				t.Fatal("reverseConvertEvent returned nil for MouseEvent")
			}

			mouseEvent, ok := result.(*tcell.EventMouse)
			if !ok {
				t.Fatalf("expected *tcell.EventMouse, got %T", result)
			}

			// Verify position
			x, y := mouseEvent.Position()
			if x != tt.event.X || y != tt.event.Y {
				t.Errorf("Position mismatch: got (%d, %d), want (%d, %d)", x, y, tt.event.X, tt.event.Y)
			}
		})
	}
}

func TestReverseConvertResizeEvent(t *testing.T) {
	event := terminal.ResizeEvent{Width: 80, Height: 24}
	result := reverseConvertEvent(event)
	if result == nil {
		t.Fatal("reverseConvertEvent returned nil for ResizeEvent")
	}

	resizeEvent, ok := result.(*tcell.EventResize)
	if !ok {
		t.Fatalf("expected *tcell.EventResize, got %T", result)
	}

	w, h := resizeEvent.Size()
	if w != 80 || h != 24 {
		t.Errorf("Size mismatch: got (%d, %d), want (80, 24)", w, h)
	}
}

func TestReverseConvertKey(t *testing.T) {
	// Test all key mappings for round-trip consistency
	keys := []terminal.Key{
		terminal.KeyUp, terminal.KeyDown, terminal.KeyLeft, terminal.KeyRight,
		terminal.KeyHome, terminal.KeyEnd, terminal.KeyPageUp, terminal.KeyPageDown,
		terminal.KeyInsert, terminal.KeyDelete, terminal.KeyBackspace,
		terminal.KeyTab, terminal.KeyEnter, terminal.KeyEscape,
		terminal.KeyF1, terminal.KeyF2, terminal.KeyF3, terminal.KeyF4,
		terminal.KeyF5, terminal.KeyF6, terminal.KeyF7, terminal.KeyF8,
		terminal.KeyF9, terminal.KeyF10, terminal.KeyF11, terminal.KeyF12,
	}

	for _, k := range keys {
		tcellKey := reverseConvertKey(k)
		backConverted := convertKey(tcellKey)
		if backConverted != k {
			t.Errorf("Round-trip failed for key %v: got %v", k, backConverted)
		}
	}
}

func TestReverseConvertMouseButton(t *testing.T) {
	buttons := []terminal.MouseButton{
		terminal.MouseLeft, terminal.MouseMiddle, terminal.MouseRight,
		terminal.MouseWheelUp, terminal.MouseWheelDown, terminal.MouseNone,
	}

	for _, b := range buttons {
		tcellButton := reverseConvertMouseButton(b)
		backConverted := convertMouseButton(tcellButton)
		if backConverted != b {
			t.Errorf("Round-trip failed for button %v: got %v", b, backConverted)
		}
	}
}
