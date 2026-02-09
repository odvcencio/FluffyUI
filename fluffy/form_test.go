package fluffy

import (
	"testing"
)

func TestFormConvenience(t *testing.T) {
	form := Form(
		FormField("Name", Input("Enter name")),
		FormField("Email", Input("Enter email")),
	)
	if form == nil {
		t.Fatal("expected form")
	}
	if form.FieldCount() != 2 {
		t.Fatalf("expected 2 fields, got %d", form.FieldCount())
	}
}
