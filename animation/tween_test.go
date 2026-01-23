package animation

import (
	"testing"
	"time"
)

func TestTweenCompletes(t *testing.T) {
	value := Float64(0)
	tween := NewTween(
		func() Animatable { return value },
		func(v Animatable) { value = v.(Float64) },
		Float64(1),
		TweenConfig{Duration: 50 * time.Millisecond},
	)
	tween.Start()
	// Force completion.
	tween.startTime = time.Now().Add(-100 * time.Millisecond)
	tween.Update(time.Now())

	if !tween.completed {
		t.Fatalf("expected tween to complete")
	}
	if value != Float64(1) {
		t.Fatalf("value = %v, want 1", value)
	}
}
