package animation

import (
	"math"
	"testing"
)

func TestSpringSettles(t *testing.T) {
	spring := NewSpring(0, SpringDefault)
	spring.SetTarget(1)

	for i := 0; i < 2000 && !spring.AtRest(); i++ {
		spring.Update(0.016)
	}

	if !spring.AtRest() {
		t.Fatalf("expected spring to settle")
	}
	if math.Abs(spring.Value-1) > 0.01 {
		t.Fatalf("value = %v, want ~1", spring.Value)
	}
}
