package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestTagInput_AddTag(t *testing.T) {
	ti := NewTagInput()
	ti.Focus()

	// Type "hello" then Enter.
	for _, r := range "hello" {
		ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}
	if ti.InputBuffer() != "hello" {
		t.Fatalf("expected input buffer 'hello', got %q", ti.InputBuffer())
	}

	ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	tags := ti.Tags()
	if len(tags) != 1 || tags[0] != "hello" {
		t.Fatalf("expected tags [hello], got %v", tags)
	}
	if ti.InputBuffer() != "" {
		t.Fatalf("expected empty input buffer after enter, got %q", ti.InputBuffer())
	}

	// Add via Tab.
	for _, r := range "world" {
		ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}
	ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	tags = ti.Tags()
	if len(tags) != 2 || tags[1] != "world" {
		t.Fatalf("expected tags [hello world], got %v", tags)
	}

	// Duplicate tag should not be added.
	ok := ti.AddTag("hello")
	if ok {
		t.Fatal("expected duplicate tag to not be added")
	}
	if len(ti.Tags()) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(ti.Tags()))
	}
}

func TestTagInput_RemoveTag(t *testing.T) {
	ti := NewTagInput()
	ti.Focus()
	ti.SetTags([]string{"one", "two", "three"})

	// Backspace on empty input removes last tag.
	ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	tags := ti.Tags()
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags after backspace, got %v", tags)
	}
	if tags[1] != "two" {
		t.Fatalf("expected last tag to be 'two', got %q", tags[1])
	}

	// Backspace with input text removes character, not tag.
	for _, r := range "ab" {
		ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}
	ti.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if ti.InputBuffer() != "a" {
		t.Fatalf("expected input buffer 'a', got %q", ti.InputBuffer())
	}
	if len(ti.Tags()) != 2 {
		t.Fatalf("expected tags unchanged at 2, got %d", len(ti.Tags()))
	}
}

func TestTagInput_MaxTags(t *testing.T) {
	ti := NewTagInput()
	ti.SetMaxTags(2)

	ok := ti.AddTag("one")
	if !ok {
		t.Fatal("expected first tag to be added")
	}
	ok = ti.AddTag("two")
	if !ok {
		t.Fatal("expected second tag to be added")
	}
	ok = ti.AddTag("three")
	if ok {
		t.Fatal("expected third tag to be rejected due to max limit")
	}
	if len(ti.Tags()) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(ti.Tags()))
	}
}

func TestTagInput_Measure(t *testing.T) {
	ti := NewTagInput()
	size := ti.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 3 {
		t.Fatalf("expected width >= 3, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestTagInput_OnChange(t *testing.T) {
	ti := NewTagInput()
	var changed []string
	ti.SetOnChange(func(tags []string) {
		changed = tags
	})
	ti.AddTag("foo")
	if len(changed) != 1 || changed[0] != "foo" {
		t.Fatalf("expected onChange with [foo], got %v", changed)
	}
}

func TestTagInput_RemoveByValue(t *testing.T) {
	ti := NewTagInput()
	ti.SetTags([]string{"a", "b", "c"})

	ok := ti.RemoveTag("b")
	if !ok {
		t.Fatal("expected RemoveTag to return true")
	}
	tags := ti.Tags()
	if len(tags) != 2 || tags[0] != "a" || tags[1] != "c" {
		t.Fatalf("expected [a c], got %v", tags)
	}

	ok = ti.RemoveTag("nonexistent")
	if ok {
		t.Fatal("expected RemoveTag to return false for nonexistent tag")
	}
}
