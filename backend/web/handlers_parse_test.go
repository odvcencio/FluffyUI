//go:build !js

package web

import (
	"testing"

	"github.com/odvcencio/fluffyui/terminal"
)

func TestParseCSI_MousePress(t *testing.T) {
	s := &Session{}
	ev, consumed := s.parseCSI([]byte("<0;10;5M"))
	if consumed != 8 {
		t.Fatalf("consumed = %d, want 8", consumed)
	}
	mouse, ok := ev.(terminal.MouseEvent)
	if !ok {
		t.Fatalf("expected mouse event, got %T", ev)
	}
	if mouse.X != 9 || mouse.Y != 4 {
		t.Fatalf("mouse coords = (%d,%d), want (9,4)", mouse.X, mouse.Y)
	}
	if mouse.Action != terminal.MousePress {
		t.Fatalf("action = %v, want MousePress", mouse.Action)
	}
	if mouse.Button != terminal.MouseLeft {
		t.Fatalf("button = %v, want MouseLeft", mouse.Button)
	}
}

func TestParseCSI_MouseRelease(t *testing.T) {
	s := &Session{}
	ev, consumed := s.parseCSI([]byte("<2;3;7m"))
	if consumed != 7 {
		t.Fatalf("consumed = %d, want 7", consumed)
	}
	mouse, ok := ev.(terminal.MouseEvent)
	if !ok {
		t.Fatalf("expected mouse event, got %T", ev)
	}
	if mouse.X != 2 || mouse.Y != 6 {
		t.Fatalf("mouse coords = (%d,%d), want (2,6)", mouse.X, mouse.Y)
	}
	if mouse.Action != terminal.MouseRelease {
		t.Fatalf("action = %v, want MouseRelease", mouse.Action)
	}
	if mouse.Button != terminal.MouseRight {
		t.Fatalf("button = %v, want MouseRight", mouse.Button)
	}
}
