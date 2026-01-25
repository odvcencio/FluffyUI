package fur

import (
	"strings"
	"testing"
)

func TestLookupEmoji(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		found    bool
	}{
		{"wave", "\U0001F44B", true},
		{"rocket", "\U0001F680", true},
		{"check", "\u2705", true},
		{"heart", "\u2764\uFE0F", true},
		{"fire", "\U0001F525", true},
		{"nonexistent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji, found := LookupEmoji(tt.name)
			if found != tt.found {
				t.Errorf("LookupEmoji(%q) found = %v, want %v", tt.name, found, tt.found)
			}
			if found && emoji != tt.expected {
				t.Errorf("LookupEmoji(%q) = %q, want %q", tt.name, emoji, tt.expected)
			}
		})
	}
}

func TestLookupEmojiCaseInsensitive(t *testing.T) {
	emoji1, found1 := LookupEmoji("WAVE")
	emoji2, found2 := LookupEmoji("wave")

	if !found1 || !found2 {
		t.Error("emoji lookup should be case insensitive")
	}
	if emoji1 != emoji2 {
		t.Error("case should not affect emoji result")
	}
}

func TestEmojiNames(t *testing.T) {
	names := EmojiNames()

	if len(names) < 100 {
		t.Errorf("expected at least 100 emojis, got %d", len(names))
	}

	// Check some expected names exist
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	expected := []string{"wave", "rocket", "check", "heart", "fire", "star"}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected emoji %q in names list", name)
		}
	}
}

func TestRegisterEmoji(t *testing.T) {
	// Register a custom emoji
	RegisterEmoji("custom_test", "\U0001F9EA")

	emoji, found := LookupEmoji("custom_test")
	if !found {
		t.Error("custom emoji should be found")
	}
	if emoji != "\U0001F9EA" {
		t.Errorf("custom emoji = %q, want test tube", emoji)
	}

	// Clean up
	delete(emojiMap, "custom_test")
}

func TestRegisterEmojiOverwrite(t *testing.T) {
	original, _ := LookupEmoji("star")

	RegisterEmoji("star", "\U0001F31F")
	modified, _ := LookupEmoji("star")

	if modified == original {
		t.Error("RegisterEmoji should allow overwriting")
	}

	// Restore original
	RegisterEmoji("star", original)
}

func TestEnableDisableEmoji(t *testing.T) {
	parser := DefaultMarkupParser()
	originalState := parser.EnableEmoji

	EnableEmoji()
	if !parser.EnableEmoji {
		t.Error("EnableEmoji should set EnableEmoji to true")
	}

	DisableEmoji()
	if parser.EnableEmoji {
		t.Error("DisableEmoji should set EnableEmoji to false")
	}

	// Restore original state
	parser.EnableEmoji = originalState
}

func TestEmojiInMarkup(t *testing.T) {
	parser := NewMarkupParser()
	parser.EnableEmoji = true

	lines := parser.Parse(":wave: Hello :rocket:")

	text := extractText(lines)
	if text == ":wave: Hello :rocket:" {
		t.Error("emojis should be replaced when enabled")
	}
}

func TestEmojiDisabledInMarkup(t *testing.T) {
	parser := NewMarkupParser()
	parser.EnableEmoji = false

	lines := parser.Parse(":wave: Hello")

	text := strings.TrimSpace(extractText(lines))
	if text != ":wave: Hello" {
		t.Errorf("emojis should not be replaced when disabled, got %q", text)
	}
}

func TestEmojiCategories(t *testing.T) {
	// Test emojis from different categories exist
	categories := map[string][]string{
		"smileys":  {"smile", "grin", "joy"},
		"gestures": {"wave", "thumbsup", "clap"},
		"animals":  {"cat", "dog", "bug"},
		"nature":   {"rose", "tree", "mushroom"},
		"food":     {"apple", "pizza", "coffee"},
		"travel":   {"car", "airplane", "rocket"},
		"objects":  {"phone", "laptop", "key"},
		"symbols":  {"check", "x", "warning"},
		"hearts":   {"heart", "broken_heart"},
		"weather":  {"sun", "cloud", "rainbow"},
	}

	for category, emojis := range categories {
		for _, name := range emojis {
			if _, found := LookupEmoji(name); !found {
				t.Errorf("category %s: emoji %q not found", category, name)
			}
		}
	}
}

func TestComputerMouseEmoji(t *testing.T) {
	// Verify the fix for duplicate mouse key
	mouse, foundMouse := LookupEmoji("mouse")
	computerMouse, foundComputerMouse := LookupEmoji("computer_mouse")

	if !foundMouse {
		t.Error("mouse emoji should exist")
	}
	if !foundComputerMouse {
		t.Error("computer_mouse emoji should exist")
	}
	if mouse == computerMouse {
		t.Error("mouse and computer_mouse should be different emojis")
	}
}
