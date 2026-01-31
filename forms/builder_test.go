package forms

import "testing"

func TestBuilderBuild(t *testing.T) {
	builder := NewBuilder().
		Text("name", "Name", "", Required("Name required")).
		Checkbox("tos", "Terms", false).
		Validator(FieldsMatch("password", "confirm", "Passwords must match"))

	form, specs := builder.Build()
	if form == nil {
		t.Fatalf("expected form")
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
	if form.Field("name") == nil {
		t.Fatalf("expected name field")
	}
	form.Set("name", "Jane")
	if form.Get("name") != "Jane" {
		t.Fatalf("expected form value set")
	}
}
