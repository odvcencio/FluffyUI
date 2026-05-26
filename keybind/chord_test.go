package keybind

import (
	"testing"
	"time"

	"m31labs.dev/fluffyui/terminal"
)

func ctrlK() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'k', Ctrl: true}
}

func ctrlC() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyCtrlC, Ctrl: true}
}

func ctrlX() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyCtrlX, Ctrl: true}
}

func keyA() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'a'}
}

func keyB() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'b'}
}

func keyG() terminal.KeyEvent {
	return terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'g'}
}

func TestChordHandler_TwoKeyChord(t *testing.T) {
	h := NewChordHandler()
	called := false
	h.Bind("ctrl+k ctrl+c", func() { called = true })

	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending after first key, got %d", state)
	}
	if called {
		t.Fatalf("action should not fire after first key")
	}

	state = h.HandleKey(ctrlC())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted after second key, got %d", state)
	}
	if !called {
		t.Fatalf("action should have fired")
	}
}

func TestKeyPressEqualHandlesCtrlAliases(t *testing.T) {
	a := KeyPress{Key: terminal.KeyCtrlK, Ctrl: true}
	b := terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'k', Ctrl: true}

	press := KeyPressFromTerminal(b)

	if !a.Equal(press) {
		t.Fatalf("expected KeyPress alias match for ctrl+k rune/constant forms")
	}
	if !press.Equal(a) {
		t.Fatalf("expected reverse alias match for ctrl+k constant/rune forms")
	}
}

func TestChordHandler_ThreeKeyChord(t *testing.T) {
	h := NewChordHandler()
	callCount := 0
	h.Bind("ctrl+k a b", func() { callCount++ })

	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending after first key, got %d", state)
	}

	state = h.HandleKey(keyA())
	if state != ChordPending {
		t.Fatalf("expected ChordPending after second key, got %d", state)
	}

	state = h.HandleKey(keyB())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted after third key, got %d", state)
	}
	if callCount != 1 {
		t.Fatalf("expected action to fire once, got %d", callCount)
	}
}

func TestChordHandler_TimeoutResetsPending(t *testing.T) {
	h := NewChordHandler()
	h.SetTimeout(50 * time.Millisecond)
	called := false
	h.Bind("ctrl+k ctrl+c", func() { called = true })

	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	// Wait for timeout to expire.
	time.Sleep(100 * time.Millisecond)

	// The pending state should have been cleared by the timeout.
	// Now pressing ctrl+c alone should not complete the chord.
	state = h.HandleKey(ctrlC())
	if state == ChordCompleted {
		t.Fatalf("expected timeout to reset pending state, but chord completed")
	}
	if called {
		t.Fatalf("action should not have fired after timeout")
	}
}

func TestChordHandler_ResetClearsPending(t *testing.T) {
	h := NewChordHandler()
	called := false
	h.Bind("ctrl+k ctrl+c", func() { called = true })

	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	pending := h.PendingKeys()
	if pending == "" {
		t.Fatalf("expected non-empty pending keys")
	}

	h.Reset()

	pending = h.PendingKeys()
	if pending != "" {
		t.Fatalf("expected empty pending keys after reset, got %q", pending)
	}

	// After reset, the second key alone should not complete the chord.
	state = h.HandleKey(ctrlC())
	if state == ChordCompleted {
		t.Fatalf("expected chord not to complete after reset")
	}
	if called {
		t.Fatalf("action should not fire after reset")
	}
}

func TestChordHandler_SingleKeyBinding(t *testing.T) {
	h := NewChordHandler()
	called := false
	h.Bind("ctrl+c", func() { called = true })

	state := h.HandleKey(ctrlC())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted for single-key binding, got %d", state)
	}
	if !called {
		t.Fatalf("action should have fired")
	}
}

func TestChordHandler_OverlappingPrefixes(t *testing.T) {
	h := NewChordHandler()
	singleCalled := false
	chordCalled := false

	// Bind ctrl+k as a single-key action.
	h.Bind("ctrl+k", func() { singleCalled = true })
	// Bind ctrl+k ctrl+c as a two-key chord.
	h.Bind("ctrl+k ctrl+c", func() { chordCalled = true })

	// Pressing ctrl+k should be pending (because it's a prefix of the chord).
	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending for overlapping prefix, got %d", state)
	}

	// Pressing ctrl+c should complete the two-key chord.
	state = h.HandleKey(ctrlC())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted, got %d", state)
	}
	if !chordCalled {
		t.Fatalf("two-key chord action should have fired")
	}
	if singleCalled {
		t.Fatalf("single-key action should not have fired")
	}
}

func TestChordHandler_OverlappingPrefixes_SingleKeyFallback(t *testing.T) {
	h := NewChordHandler()
	singleCalled := false

	// Bind only the single-key action (no longer chord registered).
	h.Bind("ctrl+k", func() { singleCalled = true })

	state := h.HandleKey(ctrlK())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted for single key, got %d", state)
	}
	if !singleCalled {
		t.Fatalf("single-key action should have fired")
	}
}

func TestChordHandler_Unbind(t *testing.T) {
	h := NewChordHandler()
	called := false
	h.Bind("ctrl+k ctrl+c", func() { called = true })
	h.Unbind("ctrl+k ctrl+c")

	state := h.HandleKey(ctrlK())
	if state != ChordIdle {
		t.Fatalf("expected ChordIdle after unbind, got %d", state)
	}

	state = h.HandleKey(ctrlC())
	if state == ChordCompleted {
		t.Fatalf("expected chord not to complete after unbind")
	}
	if called {
		t.Fatalf("action should not fire after unbind")
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		desc    string
		wantKey terminal.Key
		wantR   rune
		ctrl    bool
		alt     bool
		shift   bool
	}{
		{"ctrl+k", terminal.KeyRune, 'k', true, false, false},
		{"alt+a", terminal.KeyRune, 'a', false, true, false},
		{"shift+enter", terminal.KeyEnter, 0, false, false, true},
		{"F5", terminal.KeyF5, 0, false, false, false},
		{"enter", terminal.KeyEnter, 0, false, false, false},
		{"escape", terminal.KeyEscape, 0, false, false, false},
		{"esc", terminal.KeyEscape, 0, false, false, false},
		{"tab", terminal.KeyTab, 0, false, false, false},
		{"space", terminal.KeyRune, ' ', false, false, false},
		{"ctrl+c", terminal.KeyCtrlC, 0, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ev, err := ParseKey(tt.desc)
			if err != nil {
				t.Fatalf("ParseKey(%q) error: %v", tt.desc, err)
			}
			if ev.Key != tt.wantKey {
				t.Fatalf("ParseKey(%q).Key = %d, want %d", tt.desc, ev.Key, tt.wantKey)
			}
			if ev.Rune != tt.wantR {
				t.Fatalf("ParseKey(%q).Rune = %c, want %c", tt.desc, ev.Rune, tt.wantR)
			}
			if ev.Ctrl != tt.ctrl {
				t.Fatalf("ParseKey(%q).Ctrl = %v, want %v", tt.desc, ev.Ctrl, tt.ctrl)
			}
			if ev.Alt != tt.alt {
				t.Fatalf("ParseKey(%q).Alt = %v, want %v", tt.desc, ev.Alt, tt.alt)
			}
			if ev.Shift != tt.shift {
				t.Fatalf("ParseKey(%q).Shift = %v, want %v", tt.desc, ev.Shift, tt.shift)
			}
		})
	}

	// Error cases.
	if _, err := ParseKey(""); err == nil {
		t.Fatalf("expected error for empty key description")
	}
	if _, err := ParseKey("ctrl+unknown"); err == nil {
		t.Fatalf("expected error for unknown key")
	}
}

func TestFormatKey(t *testing.T) {
	tests := []struct {
		event terminal.KeyEvent
		want  string
	}{
		{terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'k', Ctrl: true}, "Ctrl+K"},
		{terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'a', Alt: true}, "Alt+A"},
		{terminal.KeyEvent{Key: terminal.KeyEnter}, "Enter"},
		{terminal.KeyEvent{Key: terminal.KeyF5}, "F5"},
		{terminal.KeyEvent{Key: terminal.KeyCtrlC}, "Ctrl+C"},
		{terminal.KeyEvent{Key: terminal.KeyEscape}, "Esc"},
		{terminal.KeyEvent{Key: terminal.KeyRune, Rune: ' '}, "Space"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatKey(tt.event)
			if got != tt.want {
				t.Fatalf("FormatKey(%+v) = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}

func TestFormatChord(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ctrl+k ctrl+c", "Ctrl+K Ctrl+C"},
		{"g g", "G G"},
		{"ctrl+c", "Ctrl+C"},
		{"shift+tab", "Shift+Tab"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatChord(tt.input)
			if got != tt.want {
				t.Fatalf("FormatChord(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestChordHandler_HandleKeyReturnsCorrectStates(t *testing.T) {
	h := NewChordHandler()
	h.Bind("ctrl+k ctrl+c", func() {})

	// Unrelated key returns Idle.
	state := h.HandleKey(keyA())
	if state != ChordIdle {
		t.Fatalf("expected ChordIdle for unrelated key, got %d", state)
	}

	// First chord key returns Pending.
	state = h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	// Completing the chord returns Completed.
	state = h.HandleKey(ctrlC())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted, got %d", state)
	}

	// After completion, unrelated key returns Idle.
	state = h.HandleKey(keyA())
	if state != ChordIdle {
		t.Fatalf("expected ChordIdle after completion, got %d", state)
	}
}

func TestChordHandler_WrongSecondKey(t *testing.T) {
	h := NewChordHandler()
	callCount := 0
	h.Bind("ctrl+k ctrl+c", func() { callCount++ })

	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	// Wrong second key should reset.
	state = h.HandleKey(keyA())
	if state != ChordIdle {
		t.Fatalf("expected ChordIdle for wrong second key, got %d", state)
	}
	if callCount != 0 {
		t.Fatalf("action should not have fired")
	}
}

func TestChordHandler_NilAndEmptyEdgeCases(t *testing.T) {
	var h *ChordHandler

	// All methods should be safe on nil.
	if state := h.HandleKey(ctrlK()); state != ChordIdle {
		t.Fatalf("expected ChordIdle on nil handler, got %d", state)
	}
	h.Reset()
	h.Bind("ctrl+k", func() {})
	h.Unbind("ctrl+k")
	h.SetTimeout(time.Second)
	if s := h.PendingKeys(); s != "" {
		t.Fatalf("expected empty from nil handler, got %q", s)
	}

	// Empty chord string should be ignored.
	h2 := NewChordHandler()
	h2.Bind("", func() {})
	h2.Bind("ctrl+k", nil)
	h2.Unbind("")
}

func TestChordHandler_MultipleChords(t *testing.T) {
	h := NewChordHandler()
	commentCalled := false
	formatCalled := false

	h.Bind("ctrl+k ctrl+c", func() { commentCalled = true })
	h.Bind("ctrl+k ctrl+x", func() { formatCalled = true })

	// Trigger the first chord.
	h.HandleKey(ctrlK())
	h.HandleKey(ctrlC())
	if !commentCalled {
		t.Fatalf("ctrl+k ctrl+c action should have fired")
	}

	// Trigger the second chord.
	h.HandleKey(ctrlK())
	h.HandleKey(ctrlX())
	if !formatCalled {
		t.Fatalf("ctrl+k ctrl+x action should have fired")
	}
}

func TestChordHandler_PendingKeysDisplay(t *testing.T) {
	h := NewChordHandler()
	h.Bind("ctrl+k ctrl+c", func() {})

	if s := h.PendingKeys(); s != "" {
		t.Fatalf("expected empty pending before any keys, got %q", s)
	}

	h.HandleKey(ctrlK())
	pending := h.PendingKeys()
	if pending == "" {
		t.Fatalf("expected non-empty pending keys after first key")
	}
	if pending != "Ctrl+K" {
		t.Fatalf("expected pending to show Ctrl+K, got %q", pending)
	}

	h.HandleKey(ctrlC())
	if s := h.PendingKeys(); s != "" {
		t.Fatalf("expected empty pending after chord completes, got %q", s)
	}
}

func TestChordHandler_VimStyleGG(t *testing.T) {
	h := NewChordHandler()
	called := false
	h.Bind("g g", func() { called = true })

	state := h.HandleKey(keyG())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	state = h.HandleKey(keyG())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted, got %d", state)
	}
	if !called {
		t.Fatalf("g g action should have fired")
	}
}

func TestChordHandler_DeadEndFallsBackToSolo(t *testing.T) {
	h := NewChordHandler()
	chordCalled := false
	soloCalled := false

	h.Bind("ctrl+k ctrl+c", func() { chordCalled = true })
	h.Bind("a", func() { soloCalled = true })

	// Start a chord prefix.
	state := h.HandleKey(ctrlK())
	if state != ChordPending {
		t.Fatalf("expected ChordPending, got %d", state)
	}

	// Press 'a' which doesn't continue the chord, but is itself a single-key binding.
	state = h.HandleKey(keyA())
	if state != ChordCompleted {
		t.Fatalf("expected ChordCompleted for solo fallback, got %d", state)
	}
	if !soloCalled {
		t.Fatalf("solo binding should have fired as fallback")
	}
	if chordCalled {
		t.Fatalf("chord binding should not have fired")
	}
}
