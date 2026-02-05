package main

import "testing"

func TestDetectTrueColor(t *testing.T) {
	tests := []struct {
		name      string
		term      string
		colorterm string
		want      bool
	}{
		{name: "colorterm_truecolor", term: "xterm-256color", colorterm: "truecolor", want: true},
		{name: "term_direct", term: "xterm-direct", colorterm: "", want: true},
		{name: "term_truecolor", term: "screen-truecolor", colorterm: "", want: true},
		{name: "not_detected", term: "xterm-256color", colorterm: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTrueColor(tt.term, tt.colorterm)
			if got != tt.want {
				t.Fatalf("detectTrueColor(%q, %q) = %v, want %v", tt.term, tt.colorterm, got, tt.want)
			}
		})
	}
}

func TestDetectMouseSupport(t *testing.T) {
	if !detectMouseSupport("xterm-256color") {
		t.Fatalf("xterm-256color should be detected as mouse capable")
	}
	if !detectMouseSupport("tmux-256color") {
		t.Fatalf("tmux-256color should be detected as mouse capable")
	}
	if detectMouseSupport("dumb") {
		t.Fatalf("dumb terminal should not be detected as mouse capable")
	}
}

func TestDetectKittySupport(t *testing.T) {
	t.Setenv("KITTY_WINDOW_ID", "")
	if detectKittySupport("xterm-kitty") != true {
		t.Fatalf("xterm-kitty should be detected as kitty-capable")
	}

	t.Setenv("KITTY_WINDOW_ID", "12")
	if detectKittySupport("xterm-256color") != true {
		t.Fatalf("KITTY_WINDOW_ID should indicate kitty support")
	}
}

func TestDoctorSummary(t *testing.T) {
	checks := []doctorCheck{
		{Name: "a", Status: doctorPass},
		{Name: "b", Status: doctorPass},
		{Name: "c", Status: doctorWarn},
		{Name: "d", Status: doctorFail},
		{Name: "e", Status: doctorInfo},
	}

	pass, warn, fail, info := doctorSummary(checks)
	if pass != 2 || warn != 1 || fail != 1 || info != 1 {
		t.Fatalf("summary = (%d,%d,%d,%d), want (2,1,1,1)", pass, warn, fail, info)
	}
}
