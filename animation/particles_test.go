package animation

import (
	"testing"

	"github.com/odvcencio/fluffy-ui/backend"
)

func TestParticleSystemUpdate(t *testing.T) {
	ps := NewParticleSystem(4)
	ps.Emit(Particle{Life: 1, MaxLife: 1, Color: backend.ColorRed})
	ps.Update(0.5)
	if len(ps.particles) != 1 {
		t.Fatalf("expected 1 particle, got %d", len(ps.particles))
	}
	if ps.particles[0].Life >= 1 {
		t.Fatalf("expected life to decrease")
	}
	ps.Update(1)
	if len(ps.particles) != 0 {
		t.Fatalf("expected particles to expire")
	}
}

func TestParticleBurst(t *testing.T) {
	ps := NewParticleSystem(4)
	ps.Burst(Vector2{X: 1, Y: 1}, 2, ParticleConfig{Life: Range{Min: 1, Max: 1}})
	if len(ps.particles) != 2 {
		t.Fatalf("expected 2 particles, got %d", len(ps.particles))
	}
}
