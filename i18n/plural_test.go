package i18n

import "testing"

func TestBundleTranslatePlural(t *testing.T) {
	bundle := NewBundle("en")
	bundle.AddMessages("en", map[string]string{
		"items.one":   "{0} item",
		"items.other": "{0} items",
		"count":       "Count {0}",
	})

	if got := bundle.TranslatePlural("en", "items", 1); got != "1 item" {
		t.Fatalf("unexpected singular: %q", got)
	}
	if got := bundle.TranslatePlural("en", "items", 2); got != "2 items" {
		t.Fatalf("unexpected plural: %q", got)
	}
	if got := bundle.TranslatePlural("en", "count", 5); got != "Count 5" {
		t.Fatalf("unexpected fallback plural: %q", got)
	}
}

func TestRegisterPluralRule(t *testing.T) {
	RegisterPluralRule("xx", func(n int) PluralForm {
		if n == 2 {
			return PluralTwo
		}
		return PluralOther
	})

	bundle := NewBundle("xx")
	bundle.AddMessages("xx", map[string]string{
		"thing.two":   "two things",
		"thing.other": "other things",
	})

	if got := bundle.TranslatePlural("xx", "thing", 2); got != "two things" {
		t.Fatalf("unexpected plural rule: %q", got)
	}
}

func TestLocaleDirection(t *testing.T) {
	if !IsRTL("ar") {
		t.Fatalf("expected ar to be RTL")
	}
	if IsRTL("en") {
		t.Fatalf("expected en to be LTR")
	}
	RegisterDirection("en", DirectionRTL)
	if !IsRTL("en-US") {
		t.Fatalf("expected override to apply")
	}
}
