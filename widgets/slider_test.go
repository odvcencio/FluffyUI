package widgets

import (
	"testing"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func TestSliderKeyAdjust(t *testing.T) {
	value := state.NewSignal(0.0)
	slider := NewSlider(value, WithSliderRange(0, 10, 1))
	slider.Focus()
	slider.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if got := value.Get(); got != 1 {
		t.Fatalf("value = %v, want 1", got)
	}
	slider.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if got := value.Get(); got != 0 {
		t.Fatalf("value after home = %v, want 0", got)
	}
}

func TestRangeSliderKeyAdjust(t *testing.T) {
	minVal := state.NewSignal(2.0)
	maxVal := state.NewSignal(8.0)
	slider := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 10, 1))
	slider.Focus()
	slider.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if got := minVal.Get(); got != 3 {
		t.Fatalf("min value = %v, want 3", got)
	}
}
