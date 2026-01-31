//go:build !js

package widgettest

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/widgets"
)

func TestHarnessCapture(t *testing.T) {
	h := New(t, widgets.NewLabel("Hello"), 20, 1)

	if !strings.Contains(h.Capture(), "Hello") {
		t.Fatalf("expected capture to contain label text")
	}
}
