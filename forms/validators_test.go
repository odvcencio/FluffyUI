package forms

import "testing"

func TestOneOfValidator(t *testing.T) {
	validator := OneOf("red", "green")
	if err := validator.Validate("red"); err != nil {
		t.Fatalf("expected red to be valid")
	}
	if err := validator.Validate("blue"); err == nil {
		t.Fatalf("expected blue to be invalid")
	}
	if err := validator.Validate(""); err != nil {
		t.Fatalf("expected empty string to be valid")
	}
}

func TestUUIDValidator(t *testing.T) {
	validator := UUID()
	if err := validator.Validate("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatalf("expected valid UUID")
	}
	if err := validator.Validate("not-a-uuid"); err == nil {
		t.Fatalf("expected invalid UUID")
	}
	if err := validator.Validate(""); err != nil {
		t.Fatalf("expected empty string to be valid")
	}
}

func TestJSONValidator(t *testing.T) {
	validator := JSON()
	if err := validator.Validate("{\"a\":1}"); err != nil {
		t.Fatalf("expected valid JSON")
	}
	if err := validator.Validate("{\"a\":"); err == nil {
		t.Fatalf("expected invalid JSON")
	}
	if err := validator.Validate(" \n\t "); err != nil {
		t.Fatalf("expected whitespace to be valid")
	}
}
