package fur

import (
	"bytes"
	"strings"
	"testing"
)

func TestConsolePrint(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor())

	c.Println("Hello, World!")

	got := buf.String()
	if got != "Hello, World!\n" {
		t.Errorf("got %q, want %q", got, "Hello, World!\n")
	}
}

func TestConsoleMarkup(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor())

	c.Println("[bold]Bold[/] text")

	got := buf.String()
	if got != "Bold text\n" {
		t.Errorf("got %q, want %q", got, "Bold text\n")
	}
}

func TestConsoleRule(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor(), WithWidth(20))

	c.Rule("Test")

	got := strings.TrimSpace(buf.String())
	if !strings.Contains(got, "Test") {
		t.Errorf("rule should contain title, got %q", got)
	}
	// Rule uses Unicode box characters which are wider than 1 byte
	// Just verify it contains the title and line chars
	if !strings.Contains(got, "─") {
		t.Errorf("rule should contain line character, got %q", got)
	}
}

func TestConsoleWidth(t *testing.T) {
	c := New(WithWidth(120))
	if c.Width() != 120 {
		t.Errorf("got width %d, want 120", c.Width())
	}
}

func TestConsoleDefault(t *testing.T) {
	c := Default()
	if c == nil {
		t.Error("Default() returned nil")
	}
}

func TestConsoleRender(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor(), WithWidth(40))

	c.Render(Text("Hello from renderable"))

	got := buf.String()
	if !strings.Contains(got, "Hello from renderable") {
		t.Errorf("output should contain text, got %q", got)
	}
}
