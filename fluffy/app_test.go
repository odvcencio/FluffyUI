package fluffy

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

type quitOnTickWidget struct{}

func (w *quitOnTickWidget) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}

func (w *quitOnTickWidget) Layout(bounds runtime.Rect) {}

func (w *quitOnTickWidget) Render(ctx runtime.RenderContext) {}

func (w *quitOnTickWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		return runtime.WithCommand(runtime.Quit{})
	}
	return runtime.Unhandled()
}

type inlineTrackingBackend struct {
	backend.Backend

	inline       bool
	inlineHeight int
}

func (b *inlineTrackingBackend) SetInlineMode(enabled bool) {
	if b == nil {
		return
	}
	b.inline = enabled
}

func (b *inlineTrackingBackend) SetInlineHeight(lines int) {
	if b == nil {
		return
	}
	b.inlineHeight = lines
}

func TestRunContext(t *testing.T) {
	be := sim.New(20, 5)
	root := &quitOnTickWidget{}

	err := RunContext(context.Background(), root, WithBackend(be), WithTickRate(time.Millisecond))
	if err != nil {
		t.Fatalf("RunContext returned error: %v", err)
	}
}

func TestRun(t *testing.T) {
	be := sim.New(20, 5)
	root := &quitOnTickWidget{}

	err := Run(root, WithBackend(be), WithTickRate(time.Millisecond))
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestRunInline(t *testing.T) {
	be := &inlineTrackingBackend{Backend: sim.New(20, 5)}
	root := &quitOnTickWidget{}

	err := RunInline(root, 7, WithBackend(be), WithTickRate(time.Millisecond))
	if err != nil {
		t.Fatalf("RunInline returned error: %v", err)
	}

	if !be.inline {
		t.Fatal("expected inline mode to be enabled")
	}
	if be.inlineHeight != 7 {
		t.Fatalf("inline height = %d, want 7", be.inlineHeight)
	}
}

func TestRunInlineContext(t *testing.T) {
	be := &inlineTrackingBackend{Backend: sim.New(20, 5)}
	root := &quitOnTickWidget{}

	err := RunInlineContext(context.Background(), root, 9, WithBackend(be), WithTickRate(time.Millisecond))
	if err != nil {
		t.Fatalf("RunInlineContext returned error: %v", err)
	}

	if !be.inline {
		t.Fatal("expected inline mode to be enabled")
	}
	if be.inlineHeight != 9 {
		t.Fatalf("inline height = %d, want 9", be.inlineHeight)
	}
}

func TestLabelAndTextConvenience(t *testing.T) {
	labelOutput := fluffytest.RenderToString(Label("hello"), 10, 1)
	if !strings.Contains(labelOutput, "hello") {
		t.Fatalf("Label convenience did not render text: %q", labelOutput)
	}

	textOutput := fluffytest.RenderToString(Text("world"), 10, 1)
	if !strings.Contains(textOutput, "world") {
		t.Fatalf("Text convenience did not render text: %q", textOutput)
	}
}

var _ runtime.Widget = (*quitOnTickWidget)(nil)
