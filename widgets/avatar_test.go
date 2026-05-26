package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	fluffytest "m31labs.dev/fluffyui/testing"
)

func TestAvatar_NewAvatar(t *testing.T) {
	a := NewAvatar("Jane Doe")
	if a == nil {
		t.Fatal("expected non-nil avatar")
	}
	if a.Name() != "Jane Doe" {
		t.Fatalf("expected name 'Jane Doe', got %q", a.Name())
	}
	if a.Initials() != "JD" {
		t.Fatalf("expected initials 'JD', got %q", a.Initials())
	}
	if a.AccessibleRole() != "image" {
		t.Fatalf("expected role image, got %q", a.AccessibleRole())
	}
	if a.CanFocus() {
		t.Fatal("expected avatar to not be focusable")
	}
}

func TestAvatar_SingleName(t *testing.T) {
	a := NewAvatar("Alice")
	if a.Initials() != "A" {
		t.Fatalf("expected initials 'A', got %q", a.Initials())
	}
}

func TestAvatar_EmptyName(t *testing.T) {
	a := NewAvatar("")
	if a.Initials() != "?" {
		t.Fatalf("expected initials '?', got %q", a.Initials())
	}
}

func TestAvatar_WithOptions(t *testing.T) {
	a := NewAvatar("John Smith", WithAvatarSize(AvatarLarge), WithAvatarInitials("JS"))
	if a.Size() != AvatarLarge {
		t.Fatalf("expected size AvatarLarge, got %d", a.Size())
	}
	if a.Initials() != "JS" {
		t.Fatalf("expected initials 'JS', got %q", a.Initials())
	}
}

func TestAvatar_MeasureSmall(t *testing.T) {
	a := NewAvatar("Jane Doe")
	size := a.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 4 {
		t.Fatalf("expected width 4 for small avatar, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1 for small avatar, got %d", size.Height)
	}
}

func TestAvatar_MeasureMedium(t *testing.T) {
	a := NewAvatar("Jane Doe", WithAvatarSize(AvatarMedium))
	size := a.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 5 {
		t.Fatalf("expected width 5 for medium avatar, got %d", size.Width)
	}
	if size.Height != 3 {
		t.Fatalf("expected height 3 for medium avatar, got %d", size.Height)
	}
}

func TestAvatar_MeasureLarge(t *testing.T) {
	a := NewAvatar("Jane Doe", WithAvatarSize(AvatarLarge))
	size := a.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 7 {
		t.Fatalf("expected width 7 for large avatar, got %d", size.Width)
	}
	if size.Height != 3 {
		t.Fatalf("expected height 3 for large avatar, got %d", size.Height)
	}
}

func TestAvatar_RenderSmall(t *testing.T) {
	a := NewAvatar("Jane Doe")
	output := fluffytest.RenderToString(a, 10, 1)
	if !strings.Contains(output, "JD") {
		t.Fatalf("expected output to contain 'JD', got %q", output)
	}
}

func TestAvatar_HandleMessage(t *testing.T) {
	a := NewAvatar("test")
	result := a.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected avatar to return unhandled")
	}
}

func TestAvatar_DeterministicColor(t *testing.T) {
	c1 := avatarColor("Alice")
	c2 := avatarColor("Alice")
	if c1 != c2 {
		t.Fatal("expected same color for same name")
	}
	// Different names may produce different colors (but not guaranteed).
}

func TestAvatar_ComputeInitials(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Jane Doe", "JD"},
		{"Alice", "A"},
		{"", "?"},
		{"  Bob  Smith  ", "BS"},
		{"a b c", "AB"},
	}
	for _, tt := range tests {
		got := computeInitials(tt.name)
		if got != tt.expected {
			t.Errorf("computeInitials(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestAvatar_StyleType(t *testing.T) {
	a := NewAvatar("Test")
	if a.StyleType() != "Avatar" {
		t.Fatalf("expected StyleType 'Avatar', got %q", a.StyleType())
	}
}
