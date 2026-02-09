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

func TestMatchField(t *testing.T) {
	password := NewField("password", "secret123")
	confirm := MatchField(password, "Passwords must match")

	// Matching values
	if err := confirm.Validate("secret123"); err != nil {
		t.Fatalf("expected matching values to pass, got: %s", err.Message)
	}

	// Non-matching values
	if err := confirm.Validate("wrong"); err == nil {
		t.Fatalf("expected non-matching values to fail")
	} else if err.Message != "Passwords must match" {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	// Empty against non-empty
	if err := confirm.Validate(""); err == nil {
		t.Fatalf("expected empty vs non-empty to fail")
	}

	// After changing the other field, validation should reflect new value
	password.SetValue("newpassword")
	if err := confirm.Validate("newpassword"); err != nil {
		t.Fatalf("expected match after field change, got: %s", err.Message)
	}
	if err := confirm.Validate("secret123"); err == nil {
		t.Fatalf("expected old value to no longer match")
	}
}

func TestMatchFieldNilOther(t *testing.T) {
	v := MatchField(nil, "")
	if err := v.Validate("anything"); err != nil {
		t.Fatalf("expected nil other to pass, got: %s", err.Message)
	}
}

func TestMatchFieldDefaultMessage(t *testing.T) {
	f := NewField("a", "x")
	v := MatchField(f, "")
	if err := v.Validate("y"); err == nil {
		t.Fatalf("expected error")
	} else if err.Message != "Fields do not match." {
		t.Fatalf("expected default message, got: %s", err.Message)
	}
}

func TestConditionalRequired(t *testing.T) {
	enabled := true
	cond := func() bool { return enabled }

	v := ConditionalRequired(cond, "Required when enabled")

	// Condition true, empty value -> fail
	if err := v.Validate(""); err == nil {
		t.Fatalf("expected required error when condition true and value empty")
	} else if err.Message != "Required when enabled" {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	// Condition true, non-empty value -> pass
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("expected non-empty to pass, got: %s", err.Message)
	}

	// Condition false, empty value -> pass
	enabled = false
	if err := v.Validate(""); err != nil {
		t.Fatalf("expected empty to pass when condition false, got: %s", err.Message)
	}

	// Condition false, non-empty value -> pass
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("expected value to pass when condition false, got: %s", err.Message)
	}
}

func TestConditionalRequiredNilCondition(t *testing.T) {
	v := ConditionalRequired(nil, "required")
	if err := v.Validate(""); err != nil {
		t.Fatalf("expected nil condition to pass, got: %s", err.Message)
	}
}

func TestConditionalRequiredDefaultMessage(t *testing.T) {
	v := ConditionalRequired(func() bool { return true }, "")
	if err := v.Validate(""); err == nil {
		t.Fatalf("expected error")
	} else if err.Message != "This field is required." {
		t.Fatalf("expected default message, got: %s", err.Message)
	}
}

func TestConditionalRequiredWithNilValue(t *testing.T) {
	v := ConditionalRequired(func() bool { return true }, "required")
	if err := v.Validate(nil); err == nil {
		t.Fatalf("expected nil value to fail when condition true")
	}
}

func TestAndAllPass(t *testing.T) {
	v := And(MinLength(3, ""), MaxLength(10, ""))
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("expected all to pass, got: %s", err.Message)
	}
}

func TestAndFirstFails(t *testing.T) {
	v := And(MinLength(3, "too short"), MaxLength(10, ""))
	if err := v.Validate("ab"); err == nil {
		t.Fatalf("expected first validator to fail")
	} else if err.Message != "too short" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
}

func TestAndSecondFails(t *testing.T) {
	v := And(MinLength(3, ""), MaxLength(5, "too long"))
	if err := v.Validate("toolong"); err == nil {
		t.Fatalf("expected second validator to fail")
	} else if err.Message != "too long" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
}

func TestAndEmpty(t *testing.T) {
	v := And()
	if err := v.Validate("anything"); err != nil {
		t.Fatalf("expected empty And to pass")
	}
}

func TestAndWithNilValidator(t *testing.T) {
	v := And(nil, MinLength(1, ""))
	if err := v.Validate("x"); err != nil {
		t.Fatalf("expected nil to be skipped, got: %s", err.Message)
	}
}

func TestOrFirstPasses(t *testing.T) {
	v := Or(MinLength(3, ""), MinLength(100, ""))
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("expected first to pass, got: %s", err.Message)
	}
}

func TestOrSecondPasses(t *testing.T) {
	v := Or(MinLength(100, ""), MaxLength(10, ""))
	if err := v.Validate("hello"); err != nil {
		t.Fatalf("expected second to pass, got: %s", err.Message)
	}
}

func TestOrAllFail(t *testing.T) {
	v := Or(MinLength(100, "need 100"), MinLength(200, "need 200"))
	if err := v.Validate("short"); err == nil {
		t.Fatalf("expected all to fail")
	} else if err.Message != "need 100" {
		t.Fatalf("expected first error message, got: %s", err.Message)
	}
}

func TestOrEmpty(t *testing.T) {
	v := Or()
	// No validators: nothing can pass, returns nil (vacuously)
	if err := v.Validate("anything"); err != nil {
		t.Fatalf("expected empty Or to return nil, got: %s", err.Message)
	}
}

func TestOrWithNilValidator(t *testing.T) {
	v := Or(nil, MinLength(1, ""))
	if err := v.Validate("x"); err != nil {
		t.Fatalf("expected nil to be skipped and second to pass, got: %s", err.Message)
	}
}

func TestNotInvertsPass(t *testing.T) {
	// MinLength(3) passes for "hello", so Not should fail
	v := Not(MinLength(3, ""), "must be shorter than 3")
	if err := v.Validate("hello"); err == nil {
		t.Fatalf("expected Not to fail when inner passes")
	} else if err.Message != "must be shorter than 3" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
}

func TestNotInvertsFail(t *testing.T) {
	// MinLength(3) fails for "ab", so Not should pass
	v := Not(MinLength(3, ""), "must be shorter than 3")
	if err := v.Validate("ab"); err != nil {
		t.Fatalf("expected Not to pass when inner fails, got: %s", err.Message)
	}
}

func TestNotNilValidator(t *testing.T) {
	v := Not(nil, "msg")
	if err := v.Validate("anything"); err != nil {
		t.Fatalf("expected nil validator in Not to pass, got: %s", err.Message)
	}
}

func TestNotDefaultMessage(t *testing.T) {
	v := Not(MinLength(1, ""), "")
	if err := v.Validate("x"); err == nil {
		t.Fatalf("expected error")
	} else if err.Message != "Validation constraint violated." {
		t.Fatalf("expected default message, got: %s", err.Message)
	}
}

func TestAndOrCombination(t *testing.T) {
	// Value must be either (>=3 chars AND <=5 chars) OR exactly "ab"
	shortExact := Custom(func(v any) *ValidationError {
		if s, ok := v.(string); ok && s == "ab" {
			return nil
		}
		return &ValidationError{Message: "not ab"}
	})
	v := Or(
		And(MinLength(3, ""), MaxLength(5, "")),
		shortExact,
	)

	// "abc" passes via And branch
	if err := v.Validate("abc"); err != nil {
		t.Fatalf("expected abc to pass via And, got: %s", err.Message)
	}

	// "ab" passes via shortExact branch
	if err := v.Validate("ab"); err != nil {
		t.Fatalf("expected ab to pass via exact match, got: %s", err.Message)
	}

	// "toolongvalue" fails both branches
	if err := v.Validate("toolongvalue"); err == nil {
		t.Fatalf("expected toolongvalue to fail both branches")
	}

	// "a" fails both branches
	if err := v.Validate("a"); err == nil {
		t.Fatalf("expected single char to fail both branches")
	}
}

func TestNotWithAnd(t *testing.T) {
	// Not(And(MinLength(3), MaxLength(5))) -- fails when value is 3-5 chars
	v := Not(And(MinLength(3, ""), MaxLength(5, "")), "value must not be 3-5 chars")

	// "ab" (2 chars) -- And fails, so Not passes
	if err := v.Validate("ab"); err != nil {
		t.Fatalf("expected 2 chars to pass Not, got: %s", err.Message)
	}

	// "abc" (3 chars) -- And passes, so Not fails
	if err := v.Validate("abc"); err == nil {
		t.Fatalf("expected 3 chars to fail Not")
	}

	// "abcdef" (6 chars) -- And fails (MaxLength), so Not passes
	if err := v.Validate("abcdef"); err != nil {
		t.Fatalf("expected 6 chars to pass Not, got: %s", err.Message)
	}
}

func TestMatchFieldWithFormDependency(t *testing.T) {
	// Integration-style test: use MatchField within a Form with dependencies
	password := NewField("password", "")
	confirm := NewField("confirm", "", MatchField(password, "Must match password"))

	form := NewForm(password, confirm)
	form.SetDependencies("confirm", "password")

	// Both empty: MatchField passes (both are "")
	errs := form.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no errors with both empty, got: %v", errs)
	}

	// Set password, confirm stays empty
	form.Set("password", "secret")
	errs = form.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected mismatch error")
	}

	// Set confirm to match
	form.Set("confirm", "secret")
	errs = form.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no errors when matching, got: %v", errs)
	}
}
