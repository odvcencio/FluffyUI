package accessibility

import (
	"strings"
	"testing"
)

func TestStateSetStrings(t *testing.T) {
	checked := true
	expanded := false
	state := StateSet{
		Checked:  &checked,
		Expanded: &expanded,
		Selected: true,
		Disabled: true,
		ReadOnly: true,
		Required: true,
		Invalid:  true,
	}
	got := strings.Join(state.Strings(), ",")
	want := "checked,collapsed,selected,disabled,read-only,required,invalid"
	if got != want {
		t.Fatalf("state strings = %q, want %q", got, want)
	}
}

func TestSimpleAnnouncerCallbacks(t *testing.T) {
	announcer := &SimpleAnnouncer{}
	called := 0
	announcer.SetOnMessage(func(a Announcement) {
		called++
		if a.Message != "hello" {
			t.Fatalf("message = %q, want hello", a.Message)
		}
		if a.Priority != PriorityAssertive {
			t.Fatalf("priority = %v, want assertive", a.Priority)
		}
	})

	announcer.Announce("   ", PriorityPolite)
	if called != 0 {
		t.Fatalf("expected no calls for empty announcement")
	}

	announcer.Announce("hello", PriorityAssertive)
	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}

	history := announcer.History()
	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}
	history[0].Message = "mutated"
	if announcer.History()[0].Message != "hello" {
		t.Fatalf("history should be copied")
	}
}

func TestFormatChangeIncludesValue(t *testing.T) {
	state := StateSet{Checked: BoolPtr(true), Disabled: true}
	value := &ValueInfo{Text: "10%"}
	base := &Base{
		Role:        RoleCheckbox,
		Label:       "Enable",
		Description: "desc",
		State:       state,
		Value:       value,
	}
	msg := FormatChange(base)
	want := "Enable, checkbox, desc, checked disabled, 10%"
	if msg != want {
		t.Fatalf("message = %q, want %q", msg, want)
	}
	if FormatChange(nil) != "" {
		t.Fatalf("expected empty message for nil")
	}
}
