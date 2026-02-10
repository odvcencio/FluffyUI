package prompts

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/terminal"
)

// simBackend creates a sim backend and returns it along with a fluffy.AppOption
// that injects it. The caller can use be.InjectKey* to simulate user input.
func simBackend() (*sim.Backend, fluffy.AppOption) {
	be := sim.New(80, 24)
	return be, fluffy.WithBackend(be)
}

// injectAfter schedules key injection after a short delay to allow the app
// event loop to start.
func injectAfter(be *sim.Backend, delay time.Duration, fn func(*sim.Backend)) {
	go func() {
		time.Sleep(delay)
		fn(be)
	}()
}

const keyDelay = 50 * time.Millisecond

// --- Confirm tests ---

func TestConfirm_Yes(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKeyRune('y')
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !result {
		t.Fatal("expected true for 'y' key, got false")
	}
}

func TestConfirm_YesUppercase(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKeyRune('Y')
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !result {
		t.Fatal("expected true for 'Y' key, got false")
	}
}

func TestConfirm_No(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKeyRune('n')
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if result {
		t.Fatal("expected false for 'n' key, got true")
	}
}

func TestConfirm_EnterDefault(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKey(terminal.KeyEnter, 0)
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !result {
		t.Fatal("expected true for Enter (default yes), got false")
	}
}

func TestConfirm_Escape(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKey(terminal.KeyEscape, 0)
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if result {
		t.Fatal("expected false for Escape, got true")
	}
}

func TestConfirm_CtrlC(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKey(terminal.KeyCtrlC, 0)
	})

	result, err := Confirm("Continue?", opt)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if result {
		t.Fatal("expected false for Ctrl+C, got true")
	}
}

// --- Choose tests ---

func TestChoose_FirstOption(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		// Select first option (default) and confirm.
		be.InjectKey(terminal.KeyEnter, 0)
	})

	idx, err := Choose("Pick one:", []string{"Alpha", "Beta", "Gamma"}, opt)
	if err != nil {
		t.Fatalf("Choose returned error: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected index 0, got %d", idx)
	}
}

func TestChoose_Navigate(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		// Move down twice to select "Gamma" (index 2).
		be.InjectKey(terminal.KeyDown, 0)
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyDown, 0)
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyEnter, 0)
	})

	idx, err := Choose("Pick one:", []string{"Alpha", "Beta", "Gamma"}, opt)
	if err != nil {
		t.Fatalf("Choose returned error: %v", err)
	}
	if idx != 2 {
		t.Fatalf("expected index 2, got %d", idx)
	}
}

func TestChoose_NavigateVim(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		// Move down with j, then up with k.
		be.InjectKeyRune('j')
		time.Sleep(10 * time.Millisecond)
		be.InjectKeyRune('j')
		time.Sleep(10 * time.Millisecond)
		be.InjectKeyRune('k')
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyEnter, 0)
	})

	idx, err := Choose("Pick one:", []string{"Alpha", "Beta", "Gamma"}, opt)
	if err != nil {
		t.Fatalf("Choose returned error: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected index 1 (down, down, up), got %d", idx)
	}
}

func TestChoose_Cancel(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKey(terminal.KeyEscape, 0)
	})

	idx, err := Choose("Pick one:", []string{"Alpha", "Beta"}, opt)
	if err != nil {
		t.Fatalf("Choose returned error: %v", err)
	}
	if idx != -1 {
		t.Fatalf("expected -1 on cancel, got %d", idx)
	}
}

func TestChoose_EmptyOptions(t *testing.T) {
	_, err := Choose("Pick one:", nil)
	if err == nil {
		t.Fatal("expected error for empty options")
	}
}

func TestChoose_BoundaryNavigation(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		// Try to move up past the first item (should stay at 0).
		be.InjectKey(terminal.KeyUp, 0)
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyEnter, 0)
	})

	idx, err := Choose("Pick:", []string{"Only"}, opt)
	if err != nil {
		t.Fatalf("Choose returned error: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected index 0 after up-boundary, got %d", idx)
	}
}

// --- Input tests ---

func TestInput_Submit(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKeyString("hello")
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyEnter, 0)
	})

	result, err := Input("Name?", opt)
	if err != nil {
		t.Fatalf("Input returned error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result)
	}
}

func TestInput_Cancel(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKeyString("partial")
		time.Sleep(10 * time.Millisecond)
		be.InjectKey(terminal.KeyEscape, 0)
	})

	result, err := Input("Name?", opt)
	if err != nil {
		t.Fatalf("Input returned error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string on cancel, got %q", result)
	}
}

func TestInput_EmptySubmit(t *testing.T) {
	be, opt := simBackend()
	injectAfter(be, keyDelay, func(be *sim.Backend) {
		be.InjectKey(terminal.KeyEnter, 0)
	})

	result, err := Input("Name?", opt)
	if err != nil {
		t.Fatalf("Input returned error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string for empty submit, got %q", result)
	}
}
