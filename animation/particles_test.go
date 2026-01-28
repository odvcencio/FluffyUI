package animation

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
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

func TestParticleCountAndMax(t *testing.T) {
	ps := NewParticleSystem(10)
	if ps.MaxParticles() != 10 {
		t.Fatalf("expected max 10, got %d", ps.MaxParticles())
	}
	if ps.ParticleCount() != 0 {
		t.Fatalf("expected 0 particles, got %d", ps.ParticleCount())
	}
	ps.Emit(Particle{Life: 1, MaxLife: 1})
	ps.Emit(Particle{Life: 1, MaxLife: 1})
	if ps.ParticleCount() != 2 {
		t.Fatalf("expected 2 particles, got %d", ps.ParticleCount())
	}
}

func TestSetMaxParticles(t *testing.T) {
	ps := NewParticleSystem(10)
	for i := 0; i < 5; i++ {
		ps.Emit(Particle{Life: 1, MaxLife: 1})
	}
	if ps.ParticleCount() != 5 {
		t.Fatalf("expected 5 particles, got %d", ps.ParticleCount())
	}
	ps.SetMaxParticles(3)
	if ps.ParticleCount() != 3 {
		t.Fatalf("expected 3 particles after resize, got %d", ps.ParticleCount())
	}
	if ps.MaxParticles() != 3 {
		t.Fatalf("expected max 3, got %d", ps.MaxParticles())
	}
}

func TestRemoveEmitter(t *testing.T) {
	ps := NewParticleSystem(10)
	e1 := &Emitter{Active: true, Rate: 1}
	e2 := &Emitter{Active: true, Rate: 2}
	ps.AddEmitter(e1)
	ps.AddEmitter(e2)
	ps.RemoveEmitter(e1)
	if len(ps.emitters) != 1 {
		t.Fatalf("expected 1 emitter, got %d", len(ps.emitters))
	}
	if ps.emitters[0] != e2 {
		t.Fatal("wrong emitter remaining")
	}
}

func TestClearEmitters(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.AddEmitter(&Emitter{Active: true})
	ps.AddEmitter(&Emitter{Active: true})
	ps.ClearEmitters()
	if len(ps.emitters) != 0 {
		t.Fatalf("expected 0 emitters, got %d", len(ps.emitters))
	}
}

func TestRemoveForceField(t *testing.T) {
	ps := NewParticleSystem(10)
	f1 := &GravityField{Gravity: Vector2{X: 1}}
	f2 := &GravityField{Gravity: Vector2{Y: 1}}
	ps.AddForceField(f1)
	ps.AddForceField(f2)
	ps.RemoveForceField(f1)
	if len(ps.forceFields) != 1 {
		t.Fatalf("expected 1 force field, got %d", len(ps.forceFields))
	}
}

func TestClearForceFields(t *testing.T) {
	ps := NewParticleSystem(10)
	ps.AddForceField(&GravityField{})
	ps.AddForceField(&GravityField{})
	ps.ClearForceFields()
	if len(ps.forceFields) != 0 {
		t.Fatalf("expected 0 force fields, got %d", len(ps.forceFields))
	}
}
