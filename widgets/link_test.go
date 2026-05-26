package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestLink_NewLink(t *testing.T) {
	l := NewLink("FluffyUI", "https://example.com")
	if l.label != "FluffyUI" {
		t.Errorf("expected label FluffyUI, got %s", l.label)
	}
	if l.url != "https://example.com" {
		t.Errorf("expected url https://example.com, got %s", l.url)
	}
	if l.Base.Role != "link" {
		t.Errorf("expected role link, got %s", string(l.Base.Role))
	}
}

func TestLink_Measure(t *testing.T) {
	l := NewLink("Hello", "https://example.com")
	size := l.Measure(runtime.Loose(80, 24))
	if size.Width != 5 {
		t.Errorf("expected width 5, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Errorf("expected height 1, got %d", size.Height)
	}
}

func TestLink_Activate(t *testing.T) {
	l := NewLink("Click", "https://example.com")
	l.Focus()
	var activated string
	l.SetOnActivate(func(url string) {
		activated = url
	})

	result := l.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("expected Enter to be handled")
	}
	if activated != "https://example.com" {
		t.Errorf("expected activation with URL, got %s", activated)
	}
}

func TestLink_IgnoresOtherKeys(t *testing.T) {
	l := NewLink("Click", "https://example.com")
	result := l.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if result.Handled {
		t.Error("expected Escape to not be handled")
	}
}

func TestLink_URLAccessors(t *testing.T) {
	l := NewLink("Click", "https://example.com")
	if l.URL() != "https://example.com" {
		t.Errorf("URL() = %s, want https://example.com", l.URL())
	}
	l.SetURL("https://other.com")
	if l.URL() != "https://other.com" {
		t.Errorf("URL() = %s, want https://other.com", l.URL())
	}
	if l.Base.Description != "https://other.com" {
		t.Errorf("Description = %s, want https://other.com", l.Base.Description)
	}
}

func TestLink_SetLabel(t *testing.T) {
	l := NewLink("Click", "https://example.com")
	l.SetLabel("New Label")
	if l.label != "New Label" {
		t.Errorf("label = %s, want New Label", l.label)
	}
	if l.Base.Label != "New Label" {
		t.Errorf("Base.Label = %s, want New Label", l.Base.Label)
	}
}

func TestLink_CanFocus(t *testing.T) {
	l := NewLink("Click", "https://example.com")
	if !l.CanFocus() {
		t.Error("Link should be focusable")
	}
}
