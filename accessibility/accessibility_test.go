package accessibility

import (
	"strings"
	"testing"
)

func TestFormatChange(t *testing.T) {
	checked := true
	base := &Base{
		Role:  RoleButton,
		Label: "Save",
		State: StateSet{
			Checked:  &checked,
			Disabled: true,
		},
	}
	msg := FormatChange(base)
	if msg == "" {
		t.Fatalf("expected message, got empty")
	}
	if !containsAll(msg, []string{"Save", "button"}) {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestSimpleAnnouncer(t *testing.T) {
	a := &SimpleAnnouncer{}
	a.Announce("hello", PriorityPolite)
	history := a.History()
	if len(history) != 1 {
		t.Fatalf("expected 1 announcement, got %d", len(history))
	}
	if history[0].Message != "hello" {
		t.Fatalf("unexpected message %q", history[0].Message)
	}
}

func containsAll(message string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(message, part) {
			return false
		}
	}
	return true
}
