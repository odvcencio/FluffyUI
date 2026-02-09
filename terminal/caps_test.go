package terminal

import "testing"

func TestDetectCaps_Default(t *testing.T) {
	// Clear all relevant env vars to test baseline defaults.
	t.Setenv("TERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	// With no TERM set (empty string, not "dumb"), modern defaults apply.
	if caps.TrueColor {
		t.Error("TrueColor should be false by default")
	}
	if caps.SynchronizedOut {
		t.Error("SynchronizedOut should be false by default")
	}
	if caps.CursorShapes {
		t.Error("CursorShapes should be false by default")
	}
	if caps.OSC8Hyperlinks {
		t.Error("OSC8Hyperlinks should be false by default")
	}
	if caps.Sixel {
		t.Error("Sixel should be false by default")
	}
	if caps.KittyGraphics {
		t.Error("KittyGraphics should be false by default")
	}
	if !caps.BracketedPaste {
		t.Error("BracketedPaste should be true by default")
	}
	if !caps.MouseSGR {
		t.Error("MouseSGR should be true by default")
	}
	if !caps.UnicodeWidth {
		t.Error("UnicodeWidth should be true by default")
	}
	if caps.TermProgram != "" {
		t.Errorf("TermProgram should be empty, got %q", caps.TermProgram)
	}
	if caps.Term != "" {
		t.Errorf("Term should be empty, got %q", caps.Term)
	}
}

func TestDetectCaps_TrueColor(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "truecolor")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("TrueColor should be true when COLORTERM=truecolor")
	}
	// SynchronizedOut etc should remain false — no TERM_PROGRAM set.
	if caps.SynchronizedOut {
		t.Error("SynchronizedOut should be false")
	}

	// Also test the "24bit" variant.
	t.Setenv("COLORTERM", "24bit")
	caps = DetectCaps()
	if !caps.TrueColor {
		t.Error("TrueColor should be true when COLORTERM=24bit")
	}
}

func TestDetectCaps_Kitty(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "kitty")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Kitty should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("Kitty should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("Kitty should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("Kitty should enable OSC8Hyperlinks")
	}
	if !caps.KittyGraphics {
		t.Error("Kitty should enable KittyGraphics")
	}
	if caps.Sixel {
		t.Error("Kitty should not enable Sixel")
	}
	if caps.TermProgram != "kitty" {
		t.Errorf("TermProgram should be 'kitty', got %q", caps.TermProgram)
	}
}

func TestDetectCaps_KittyViaTerm(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("xterm-kitty TERM should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("xterm-kitty TERM should enable SynchronizedOut")
	}
	if !caps.KittyGraphics {
		t.Error("xterm-kitty TERM should enable KittyGraphics")
	}
}

func TestDetectCaps_WezTerm(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("WezTerm should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("WezTerm should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("WezTerm should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("WezTerm should enable OSC8Hyperlinks")
	}
	if !caps.KittyGraphics {
		t.Error("WezTerm should enable KittyGraphics")
	}
	if caps.Sixel {
		t.Error("WezTerm should not enable Sixel via TERM_PROGRAM alone")
	}
	if caps.TermProgram != "WezTerm" {
		t.Errorf("TermProgram should be 'WezTerm', got %q", caps.TermProgram)
	}
}

func TestDetectCaps_Foot(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "foot")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Foot should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("Foot should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("Foot should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("Foot should enable OSC8Hyperlinks")
	}
	if !caps.Sixel {
		t.Error("Foot should enable Sixel")
	}
	if caps.KittyGraphics {
		t.Error("Foot should not enable KittyGraphics")
	}
}

func TestDetectCaps_FootViaTerm(t *testing.T) {
	t.Setenv("TERM", "foot")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("foot TERM should enable TrueColor")
	}
	if !caps.Sixel {
		t.Error("foot TERM should enable Sixel")
	}

	// Also test foot-extra variant.
	t.Setenv("TERM", "foot-extra")
	caps = DetectCaps()
	if !caps.TrueColor {
		t.Error("foot-extra TERM should enable TrueColor")
	}
	if !caps.Sixel {
		t.Error("foot-extra TERM should enable Sixel")
	}
}

func TestDetectCaps_VSCode(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "vscode")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("VSCode should enable TrueColor")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("VSCode should enable OSC8Hyperlinks")
	}
	if !caps.CursorShapes {
		t.Error("VSCode should enable CursorShapes")
	}
	if caps.SynchronizedOut {
		t.Error("VSCode should not enable SynchronizedOut")
	}
	if caps.KittyGraphics {
		t.Error("VSCode should not enable KittyGraphics")
	}
}

func TestDetectCaps_DumbTerminal(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if caps.BracketedPaste {
		t.Error("dumb terminal should not enable BracketedPaste")
	}
	if caps.MouseSGR {
		t.Error("dumb terminal should not enable MouseSGR")
	}
	if caps.UnicodeWidth {
		t.Error("dumb terminal should not enable UnicodeWidth")
	}
	if caps.TrueColor {
		t.Error("dumb terminal should not enable TrueColor")
	}
	if caps.Term != "dumb" {
		t.Errorf("Term should be 'dumb', got %q", caps.Term)
	}
}

func TestDetectCaps_Alacritty(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "alacritty")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Alacritty should enable TrueColor")
	}
	if !caps.CursorShapes {
		t.Error("Alacritty should enable CursorShapes")
	}
	if caps.SynchronizedOut {
		t.Error("Alacritty should not enable SynchronizedOut")
	}
	if caps.OSC8Hyperlinks {
		t.Error("Alacritty should not enable OSC8Hyperlinks")
	}
	if caps.KittyGraphics {
		t.Error("Alacritty should not enable KittyGraphics")
	}
}

func TestDetectCaps_Ghostty(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Ghostty should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("Ghostty should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("Ghostty should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("Ghostty should enable OSC8Hyperlinks")
	}
	if !caps.KittyGraphics {
		t.Error("Ghostty should enable KittyGraphics")
	}
}

func TestDetectCaps_Contour(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "contour")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Contour should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("Contour should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("Contour should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("Contour should enable OSC8Hyperlinks")
	}
	if !caps.Sixel {
		t.Error("Contour should enable Sixel")
	}
	if !caps.KittyGraphics {
		t.Error("Contour should enable KittyGraphics")
	}
}

func TestDetectCaps_Mintty(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "mintty")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Mintty should enable TrueColor")
	}
	if !caps.CursorShapes {
		t.Error("Mintty should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("Mintty should enable OSC8Hyperlinks")
	}
	if caps.SynchronizedOut {
		t.Error("Mintty should not enable SynchronizedOut")
	}
}

func TestDetectCaps_Rio(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "rio")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("Rio should enable TrueColor")
	}
	if !caps.CursorShapes {
		t.Error("Rio should enable CursorShapes")
	}
	if caps.SynchronizedOut {
		t.Error("Rio should not enable SynchronizedOut")
	}
	if caps.OSC8Hyperlinks {
		t.Error("Rio should not enable OSC8Hyperlinks")
	}
}

func TestDetectCaps_iTerm(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()

	if !caps.TrueColor {
		t.Error("iTerm should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("iTerm should enable SynchronizedOut")
	}
	if !caps.CursorShapes {
		t.Error("iTerm should enable CursorShapes")
	}
	if !caps.OSC8Hyperlinks {
		t.Error("iTerm should enable OSC8Hyperlinks")
	}
	if caps.KittyGraphics {
		t.Error("iTerm should not enable KittyGraphics")
	}

	// Also test iTerm2 variant.
	t.Setenv("TERM_PROGRAM", "iTerm2")
	caps = DetectCaps()
	if !caps.TrueColor {
		t.Error("iTerm2 should enable TrueColor")
	}
	if !caps.SynchronizedOut {
		t.Error("iTerm2 should enable SynchronizedOut")
	}
}

func TestDetectCaps_String(t *testing.T) {
	caps := &TerminalCaps{
		TrueColor:       true,
		SynchronizedOut: true,
		CursorShapes:    true,
		OSC8Hyperlinks:  true,
		Sixel:           false,
		KittyGraphics:   true,
		BracketedPaste:  true,
		MouseSGR:        true,
		UnicodeWidth:    true,
		TermProgram:     "kitty",
		Term:            "xterm-256color",
	}

	s := caps.String()
	if s == "" {
		t.Error("String() should not return empty")
	}

	// Verify key substrings are present.
	for _, want := range []string{
		"TrueColor=true",
		"SyncOut=true",
		"CursorShapes=true",
		"OSC8=true",
		"Sixel=false",
		"KittyGfx=true",
		"ITermImg=false",
		"BracketPaste=true",
		"MouseSGR=true",
		"Unicode=true",
		`TERM="xterm-256color"`,
		`TERM_PROGRAM="kitty"`,
	} {
		if !contains(s, want) {
			t.Errorf("String() missing %q, got: %s", want, s)
		}
	}
}

func TestDetectCaps_ColortermOverridesDefault(t *testing.T) {
	// Even without a known TERM_PROGRAM, COLORTERM should set TrueColor.
	t.Setenv("TERM", "xterm")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "24bit")

	caps := DetectCaps()
	if !caps.TrueColor {
		t.Error("COLORTERM=24bit should set TrueColor regardless of TERM_PROGRAM")
	}
	if caps.SynchronizedOut {
		t.Error("Unknown terminal should not have SynchronizedOut")
	}
}

func TestDetectCaps_RawEnvStored(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("COLORTERM", "truecolor")

	caps := DetectCaps()
	if caps.Term != "xterm-256color" {
		t.Errorf("Term should store raw value, got %q", caps.Term)
	}
	if caps.TermProgram != "WezTerm" {
		t.Errorf("TermProgram should store raw value, got %q", caps.TermProgram)
	}
}

func TestCaps_ITermInlineImages(t *testing.T) {
	// iTerm.app should enable ITermInlineImages.
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("COLORTERM", "")

	caps := DetectCaps()
	if !caps.ITermInlineImages {
		t.Error("iTerm.app should enable ITermInlineImages")
	}

	// iTerm2 variant should also enable it.
	t.Setenv("TERM_PROGRAM", "iTerm2")
	caps = DetectCaps()
	if !caps.ITermInlineImages {
		t.Error("iTerm2 should enable ITermInlineImages")
	}

	// Other terminals should not have it.
	t.Setenv("TERM_PROGRAM", "kitty")
	caps = DetectCaps()
	if caps.ITermInlineImages {
		t.Error("kitty should not enable ITermInlineImages")
	}

	// Default (no TERM_PROGRAM) should not have it.
	t.Setenv("TERM_PROGRAM", "")
	caps = DetectCaps()
	if caps.ITermInlineImages {
		t.Error("default should not enable ITermInlineImages")
	}
}

// contains is a test helper to check substring presence.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
