package fluffy

import (
	"testing"

	"m31labs.dev/fluffyui/style"
)

func TestThemeToggle(t *testing.T) {
	dark := style.DarkTheme()
	light := style.LightTheme()
	toggle := NewThemeToggle(dark, light)

	if toggle.Current() != dark {
		t.Fatal("expected dark theme initially")
	}
	if toggle.Len() != 2 {
		t.Fatalf("expected 2 themes, got %d", toggle.Len())
	}
	toggle.Toggle()
	if toggle.Current() != light {
		t.Fatal("expected light after first toggle")
	}
	if toggle.Index() != 1 {
		t.Fatalf("expected index 1, got %d", toggle.Index())
	}
	toggle.Toggle()
	if toggle.Current() != dark {
		t.Fatal("expected dark after second toggle (wrap)")
	}
}

func TestThemeToggle_Default(t *testing.T) {
	toggle := NewThemeToggle()
	if toggle.Len() != 2 {
		t.Fatal("default should have dark+light")
	}
}

func TestThemeToggle_SetIndex(t *testing.T) {
	a := style.DarkTheme()
	b := style.LightTheme()
	c := style.MonokaiTheme()
	toggle := NewThemeToggle(a, b, c)
	toggle.SetIndex(2)
	if toggle.Current() != c {
		t.Fatal("expected monokai")
	}
	toggle.SetIndex(-1) // should be no-op
	if toggle.Current() != c {
		t.Fatal("negative index should be no-op")
	}
	toggle.SetIndex(99) // should be no-op
	if toggle.Current() != c {
		t.Fatal("out of range should be no-op")
	}
}
