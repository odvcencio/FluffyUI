package fur

import (
	"strings"
	"testing"
)

func TestInspectBasic(t *testing.T) {
	type Sample struct {
		Name string
		Age  int
	}
	r := Inspect(Sample{Name: "test", Age: 42})
	lines := r.Render(60)

	text := extractText(lines)
	if !strings.Contains(text, "Sample") {
		t.Error("expected type name in output")
	}
	if !strings.Contains(text, "Name") {
		t.Error("expected field Name in output")
	}
	if !strings.Contains(text, "Age") {
		t.Error("expected field Age in output")
	}
}

func TestInspectWithMethods(t *testing.T) {
	r := InspectWith("hello", InspectOpts{Methods: true, Fields: false})
	lines := r.Render(60)

	text := extractText(lines)
	if !strings.Contains(text, "Methods") {
		t.Error("expected Methods section")
	}
}

func TestInspectWithFields(t *testing.T) {
	type Data struct {
		Value int
	}
	r := InspectWith(Data{Value: 10}, InspectOpts{Fields: true, Methods: false})
	lines := r.Render(60)

	text := extractText(lines)
	if !strings.Contains(text, "Fields") {
		t.Error("expected Fields section")
	}
	if !strings.Contains(text, "Value") {
		t.Error("expected Value field")
	}
}

func TestInspectNil(t *testing.T) {
	r := Inspect(nil)
	lines := r.Render(40)

	text := extractText(lines)
	if !strings.Contains(text, "nil") {
		t.Error("expected nil in output")
	}
}

func TestInspectPointer(t *testing.T) {
	type Inner struct {
		X int
	}
	val := &Inner{X: 5}
	r := Inspect(val)
	lines := r.Render(60)

	text := extractText(lines)
	if !strings.Contains(text, "Inner") {
		t.Error("expected type name in output")
	}
}
