package keybind

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestParseKeySequence(t *testing.T) {
	key, err := ParseKeySequence("ctrl+c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key.Sequence) != 1 {
		t.Fatalf("expected 1 key press, got %d", len(key.Sequence))
	}
	if key.Sequence[0].Key != terminal.KeyCtrlC {
		t.Fatalf("expected ctrl+c to map to KeyCtrlC, got %v", key.Sequence[0].Key)
	}

	key, err = ParseKeySequence("g g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key.Sequence) != 2 {
		t.Fatalf("expected 2 key presses, got %d", len(key.Sequence))
	}
	if key.Sequence[0].Rune != 'g' || key.Sequence[1].Rune != 'g' {
		t.Fatalf("expected g g sequence, got %+v", key.Sequence)
	}

	key, err = ParseKeySequence("space")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.Sequence[0].Key != terminal.KeyRune || key.Sequence[0].Rune != ' ' {
		t.Fatalf("expected space to map to rune, got %+v", key.Sequence[0])
	}
}

func TestKeyRouterSequence(t *testing.T) {
	registry := NewRegistry()
	count := 0
	registry.Register(Command{ID: "go-top", Handler: func(ctx Context) {
		count++
	}})
	km := &Keymap{
		Bindings: []Binding{
			{Key: MustParseKeySequence("g g"), Command: "go-top"},
		},
	}
	stack := &KeymapStack{}
	stack.Push(km)
	router := NewKeyRouter(registry, nil, stack)

	if !router.HandleKey(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'g'}, Context{}) {
		t.Fatalf("expected prefix to be handled")
	}
	if count != 0 {
		t.Fatalf("expected no command yet, got %d", count)
	}
	if !router.HandleKey(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'g'}, Context{}) {
		t.Fatalf("expected sequence to be handled")
	}
	if count != 1 {
		t.Fatalf("expected command to run once, got %d", count)
	}
}

func TestKeyRouterCondition(t *testing.T) {
	registry := NewRegistry()
	count := 0
	registry.Register(Command{ID: "save", Handler: func(ctx Context) {
		count++
	}})
	km := &Keymap{
		Bindings: []Binding{
			{
				Key:     MustParseKeySequence("ctrl+c"),
				Command: "save",
				When:    func(ctx Context) bool { return false },
			},
		},
	}
	stack := &KeymapStack{}
	stack.Push(km)
	router := NewKeyRouter(registry, nil, stack)
	handled := router.HandleKey(runtime.KeyMsg{Key: terminal.KeyCtrlC, Ctrl: true}, Context{})
	if handled {
		t.Fatalf("expected condition to block handling")
	}
	if count != 0 {
		t.Fatalf("expected no command, got %d", count)
	}
}

func TestFormatKeySequence(t *testing.T) {
	if got := FormatKeySequence(MustParseKeySequence("ctrl+c")); got != "Ctrl+C" {
		t.Fatalf("expected Ctrl+C, got %q", got)
	}
	if got := FormatKeySequence(MustParseKeySequence("g g")); got != "G G" {
		t.Fatalf("expected G G, got %q", got)
	}
	if got := FormatKeySequence(MustParseKeySequence("shift+tab")); got != "Shift+Tab" {
		t.Fatalf("expected Shift+Tab, got %q", got)
	}
	if got := FormatKeySequence(MustParseKeySequence("space")); got != "Space" {
		t.Fatalf("expected Space, got %q", got)
	}
}
