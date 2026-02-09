package terminal

import "testing"

func TestParseKittyKey_SimpleChar(t *testing.T) {
	// \x1b[97u → 'a' (no modifiers)
	seq := []byte("\x1b[97u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
	if ev.Shift || ev.Alt || ev.Ctrl {
		t.Error("expected no modifiers")
	}
}

func TestParseKittyKey_CharWithShift(t *testing.T) {
	// \x1b[97;2u → 'a' with Shift (modifiers=2, bitmask=1=shift)
	seq := []byte("\x1b[97;2u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
	if !ev.Shift {
		t.Error("expected Shift=true")
	}
	if ev.Alt || ev.Ctrl {
		t.Error("expected Alt=false, Ctrl=false")
	}
}

func TestParseKittyKey_CharWithCtrlAlt(t *testing.T) {
	// \x1b[97;7u → 'a' with Ctrl+Alt (modifiers=7, bitmask=6: alt=2 + ctrl=4)
	seq := []byte("\x1b[97;7u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
	if ev.Shift {
		t.Error("expected Shift=false")
	}
	if !ev.Alt {
		t.Error("expected Alt=true")
	}
	if !ev.Ctrl {
		t.Error("expected Ctrl=true")
	}
}

func TestParseKittyKey_AllModifiers(t *testing.T) {
	// modifiers=8 → bitmask=7: shift+alt+ctrl
	seq := []byte("\x1b[97;8u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if !ev.Shift || !ev.Alt || !ev.Ctrl {
		t.Errorf("expected all modifiers, got Shift=%t Alt=%t Ctrl=%t", ev.Shift, ev.Alt, ev.Ctrl)
	}
}

func TestParseKittyKey_Enter(t *testing.T) {
	// \x1b[13u → KeyEnter
	seq := []byte("\x1b[13u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyEnter {
		t.Errorf("expected KeyEnter, got %d", ev.Key)
	}
	if ev.Shift || ev.Alt || ev.Ctrl {
		t.Error("expected no modifiers")
	}
}

func TestParseKittyKey_Escape(t *testing.T) {
	// \x1b[27u → KeyEscape
	seq := []byte("\x1b[27u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyEscape {
		t.Errorf("expected KeyEscape, got %d", ev.Key)
	}
}

func TestParseKittyKey_Tab(t *testing.T) {
	// \x1b[9u → KeyTab
	seq := []byte("\x1b[9u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyTab {
		t.Errorf("expected KeyTab, got %d", ev.Key)
	}
}

func TestParseKittyKey_Backspace(t *testing.T) {
	// \x1b[127u → KeyBackspace
	seq := []byte("\x1b[127u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyBackspace {
		t.Errorf("expected KeyBackspace, got %d", ev.Key)
	}
}

func TestParseKittyKey_F1_FunctionalRange(t *testing.T) {
	// \x1b[57344u → KeyF1 (first key in Kitty functional range)
	seq := []byte("\x1b[57344u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyF1 {
		t.Errorf("expected KeyF1, got %d", ev.Key)
	}
}

func TestParseKittyKey_FunctionKeys(t *testing.T) {
	tests := []struct {
		name    string
		seq     []byte
		wantKey Key
	}{
		{"F1", []byte("\x1b[57358u"), KeyF1},
		{"F2", []byte("\x1b[57359u"), KeyF2},
		{"F3", []byte("\x1b[57360u"), KeyF3},
		{"F4", []byte("\x1b[57361u"), KeyF4},
		{"F5", []byte("\x1b[57362u"), KeyF5},
		{"F6", []byte("\x1b[57363u"), KeyF6},
		{"F7", []byte("\x1b[57364u"), KeyF7},
		{"F8", []byte("\x1b[57365u"), KeyF8},
		{"F9", []byte("\x1b[57366u"), KeyF9},
		{"F10", []byte("\x1b[57367u"), KeyF10},
		{"F11", []byte("\x1b[57368u"), KeyF11},
		{"F12", []byte("\x1b[57369u"), KeyF12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, ok := ParseKittyKey(tt.seq)
			if !ok {
				t.Fatal("expected successful parse")
			}
			if ev.Key != tt.wantKey {
				t.Errorf("expected %d, got %d", tt.wantKey, ev.Key)
			}
		})
	}
}

func TestParseKittyKey_NavigationKeys(t *testing.T) {
	tests := []struct {
		name    string
		seq     []byte
		wantKey Key
	}{
		{"Insert", []byte("\x1b[57348u"), KeyInsert},
		{"Delete", []byte("\x1b[57349u"), KeyDelete},
		{"Left", []byte("\x1b[57350u"), KeyLeft},
		{"Right", []byte("\x1b[57351u"), KeyRight},
		{"Up", []byte("\x1b[57352u"), KeyUp},
		{"Down", []byte("\x1b[57353u"), KeyDown},
		{"PageUp", []byte("\x1b[57354u"), KeyPageUp},
		{"PageDown", []byte("\x1b[57355u"), KeyPageDown},
		{"Home", []byte("\x1b[57356u"), KeyHome},
		{"End", []byte("\x1b[57357u"), KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, ok := ParseKittyKey(tt.seq)
			if !ok {
				t.Fatal("expected successful parse")
			}
			if ev.Key != tt.wantKey {
				t.Errorf("expected %d, got %d", tt.wantKey, ev.Key)
			}
		})
	}
}

func TestParseKittyKey_ShiftedKeyWithBase(t *testing.T) {
	// \x1b[97:65;2u → shifted 'A' (base='a', shifted='A', shift modifier)
	seq := []byte("\x1b[97:65;2u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'A' {
		t.Errorf("expected 'A', got %c (%d)", ev.Rune, ev.Rune)
	}
	if !ev.Shift {
		t.Error("expected Shift=true")
	}
}

func TestParseKittyKey_ShiftedKeyNoShiftMod(t *testing.T) {
	// \x1b[97:65u → no modifier field, so shifted keycode is ignored, uses base
	seq := []byte("\x1b[97:65u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	// No modifiers means shift is false, so base keycode 97='a' is used
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
}

func TestParseKittyKey_EventType(t *testing.T) {
	// \x1b[97;1:1u → 'a' press event (modifiers=1 means bitmask=0, event_type=1)
	seq := []byte("\x1b[97;1:1u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 'a' {
		t.Errorf("expected 'a', got %c", ev.Rune)
	}
	// modifiers=1 → bitmask=0, no modifiers
	if ev.Shift || ev.Alt || ev.Ctrl {
		t.Error("expected no modifiers")
	}
}

func TestParseKittyKey_EnterWithShift(t *testing.T) {
	// \x1b[13;2u → Enter with Shift
	seq := []byte("\x1b[13;2u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyEnter {
		t.Errorf("expected KeyEnter, got %d", ev.Key)
	}
	if !ev.Shift {
		t.Error("expected Shift=true")
	}
}

func TestParseKittyKey_SpaceChar(t *testing.T) {
	// \x1b[32u → space character
	seq := []byte("\x1b[32u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != ' ' {
		t.Errorf("expected space, got %c (%d)", ev.Rune, ev.Rune)
	}
}

func TestParseKittyKey_UnicodeChar(t *testing.T) {
	// \x1b[228u → 'ä' (U+00E4)
	seq := []byte("\x1b[228u")
	ev, ok := ParseKittyKey(seq)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if ev.Key != KeyRune {
		t.Errorf("expected KeyRune, got %d", ev.Key)
	}
	if ev.Rune != 0xe4 {
		t.Errorf("expected U+00E4 (ä), got U+%04X", ev.Rune)
	}
}

// --- Invalid sequence tests ---

func TestParseKittyKey_TooShort(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1b[u"))
	if ok {
		t.Error("expected failure for too-short sequence")
	}
}

func TestParseKittyKey_WrongPrefix(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1bO97u"))
	if ok {
		t.Error("expected failure for wrong prefix")
	}
}

func TestParseKittyKey_WrongSuffix(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1b[97~"))
	if ok {
		t.Error("expected failure for wrong suffix")
	}
}

func TestParseKittyKey_EmptyPayload(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1b[u"))
	if ok {
		t.Error("expected failure for empty payload")
	}
}

func TestParseKittyKey_NonNumericKeycode(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1b[abcu"))
	if ok {
		t.Error("expected failure for non-numeric keycode")
	}
}

func TestParseKittyKey_NonNumericModifier(t *testing.T) {
	_, ok := ParseKittyKey([]byte("\x1b[97;xu"))
	if ok {
		t.Error("expected failure for non-numeric modifier")
	}
}

func TestParseKittyKey_NilInput(t *testing.T) {
	_, ok := ParseKittyKey(nil)
	if ok {
		t.Error("expected failure for nil input")
	}
}

func TestParseKittyKey_EmptyInput(t *testing.T) {
	_, ok := ParseKittyKey([]byte{})
	if ok {
		t.Error("expected failure for empty input")
	}
}

// --- Enable/Disable sequence tests ---

func TestKittyKeyboardEnable(t *testing.T) {
	tests := []struct {
		flags int
		want  string
	}{
		{0, "\x1b[>0u"},
		{1, "\x1b[>1u"},
		{KittyDisambiguate, "\x1b[>1u"},
		{KittyReportEvents, "\x1b[>2u"},
		{KittyReportAlternates, "\x1b[>4u"},
		{KittyReportAllKeys, "\x1b[>8u"},
		{KittyDisambiguate | KittyReportEvents, "\x1b[>3u"},
		{KittyDisambiguate | KittyReportEvents | KittyReportAlternates | KittyReportAllKeys, "\x1b[>15u"},
	}

	for _, tt := range tests {
		got := KittyKeyboardEnable(tt.flags)
		if got != tt.want {
			t.Errorf("KittyKeyboardEnable(%d) = %q, want %q", tt.flags, got, tt.want)
		}
	}
}

func TestKittyKeyboardDisable(t *testing.T) {
	got := KittyKeyboardDisable()
	want := "\x1b[<u"
	if got != want {
		t.Errorf("KittyKeyboardDisable() = %q, want %q", got, want)
	}
}

// --- Caps detection tests ---

func TestDetectCaps_KittyKeyboard_ViaTermProgram(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("Kitty should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaXtermKitty(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("xterm-kitty TERM should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaGhostty(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("Ghostty should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaWezTerm(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("WezTerm should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaFoot(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "foot")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("Foot should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaContour(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "contour")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("Contour should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ViaFootTerm(t *testing.T) {
	t.Setenv("TERM", "foot")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("foot TERM should enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_ForcedEnable(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "1")

	caps := DetectCaps()
	if !caps.KittyKeyboard {
		t.Error("FLUFFYUI_KITTY_KEYBOARD=1 should force enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_NotEnabled(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "alacritty")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if caps.KittyKeyboard {
		t.Error("Alacritty should not enable KittyKeyboard")
	}
}

func TestDetectCaps_KittyKeyboard_Default(t *testing.T) {
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("FLUFFYUI_KITTY_KEYBOARD", "")

	caps := DetectCaps()
	if caps.KittyKeyboard {
		t.Error("KittyKeyboard should be false by default")
	}
}

func TestDetectCaps_String_IncludesKittyKbd(t *testing.T) {
	caps := &TerminalCaps{
		KittyKeyboard: true,
		Term:          "xterm-256color",
		TermProgram:   "kitty",
	}

	s := caps.String()
	if !contains(s, "KittyKbd=true") {
		t.Errorf("String() should contain KittyKbd=true, got: %s", s)
	}
}
