package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
	fluffytest "github.com/odvcencio/fluffy-ui/testing"
)

func TestAccordionRenderExpanded(t *testing.T) {
	section1 := NewAccordionSection("First", NewLabel("Alpha"), WithSectionExpanded(true))
	section2 := NewAccordionSection("Second", NewLabel("Beta"))
	acc := NewAccordion(section1, section2)

	out := fluffytest.RenderToString(acc, 20, 4)
	if !strings.Contains(out, "v First") {
		t.Fatalf("expected expanded header, got:\n%s", out)
	}
	if !strings.Contains(out, "Alpha") {
		t.Fatalf("expected content, got:\n%s", out)
	}
	if strings.Contains(out, "Beta") {
		t.Fatalf("expected collapsed content to be hidden, got:\n%s", out)
	}
}

func TestAccordionToggleAndSelection(t *testing.T) {
	section1 := NewAccordionSection("First", NewLabel("Alpha"))
	section2 := NewAccordionSection("Second", NewLabel("Beta"))
	acc := NewAccordion(section1, section2)
	acc.Focus()

	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !section1.Expanded() {
		t.Fatalf("expected first section to expand")
	}

	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !section2.Expanded() {
		t.Fatalf("expected second section to expand")
	}
	if section1.Expanded() {
		t.Fatalf("expected first section to collapse when another expands")
	}
}

func TestAccordionAllowMultiple(t *testing.T) {
	section1 := NewAccordionSection("First", NewLabel("Alpha"))
	section2 := NewAccordionSection("Second", NewLabel("Beta"))
	acc := NewAccordion(section1, section2)
	acc.SetAllowMultiple(true)
	section1.SetExpanded(true)
	section2.SetExpanded(true)
	if !section1.Expanded() || !section2.Expanded() {
		t.Fatalf("expected both sections expanded when allowMultiple is true")
	}

	acc.SetAllowMultiple(false)
	if section1.Expanded() && section2.Expanded() {
		t.Fatalf("expected only one section expanded when allowMultiple is false")
	}
}

func TestAccordionDisabledSelectionSkips(t *testing.T) {
	section1 := NewAccordionSection("First", NewLabel("Alpha"), WithSectionDisabled(true))
	section2 := NewAccordionSection("Second", NewLabel("Beta"))
	acc := NewAccordion(section1, section2)
	acc.Focus()

	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !section2.Expanded() {
		t.Fatalf("expected selection to skip disabled section")
	}
	if section1.Expanded() {
		t.Fatalf("disabled section should not expand")
	}
}
