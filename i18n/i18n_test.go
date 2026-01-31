package i18n

import "testing"

func TestBundleTranslate(t *testing.T) {
	bundle := NewBundle("en")
	bundle.AddMessages("en", map[string]string{"hello": "Hello {0}"})
	bundle.AddMessages("es", map[string]string{"hello": "Hola {0}"})

	loc := bundle.Localizer("es")
	if got := loc.Tf("hello", "World"); got != "Hola World" {
		t.Fatalf("expected Spanish translation, got %q", got)
	}
	loc = bundle.Localizer("fr")
	if got := loc.Tf("hello", "World"); got != "Hello World" {
		t.Fatalf("expected fallback translation, got %q", got)
	}
	if got := loc.T("missing"); got != "missing" {
		t.Fatalf("expected fallback to key, got %q", got)
	}
}

func TestMapLocalizer(t *testing.T) {
	loc := MapLocalizer{
		LocaleCode: "en",
		Messages:   map[string]string{"greet": "Hi"},
		Fallback:   map[string]string{"bye": "Bye"},
	}
	if got := loc.T("greet"); got != "Hi" {
		t.Fatalf("expected direct message, got %q", got)
	}
	if got := loc.T("bye"); got != "Bye" {
		t.Fatalf("expected fallback message, got %q", got)
	}
	if got := loc.T("missing"); got != "missing" {
		t.Fatalf("expected fallback key, got %q", got)
	}
}
