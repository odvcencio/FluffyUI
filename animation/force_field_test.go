package animation

import "testing"

func TestGravityField(t *testing.T) {
	field := &GravityField{Gravity: Vector2{X: 1, Y: -2}}
	got := field.Apply(Vector2{X: 10, Y: 10})
	if got.X != 1 || got.Y != -2 {
		t.Fatalf("force = %#v, want (1,-2)", got)
	}
}

func TestRadialField(t *testing.T) {
	field := &RadialField{Center: Vector2{}, Strength: 10}
	got := field.Apply(Vector2{X: 1, Y: 0})
	if got.X >= 0 || got.Y != 0 {
		t.Fatalf("force = %#v, want negative X toward center", got)
	}
	outside := (&RadialField{Center: Vector2{}, Strength: 10, Radius: 0.5}).Apply(Vector2{X: 1, Y: 0})
	if outside.X != 0 || outside.Y != 0 {
		t.Fatalf("force outside radius = %#v, want zero", outside)
	}
}

func TestVortexField(t *testing.T) {
	field := &VortexField{Center: Vector2{}, Strength: 10}
	got := field.Apply(Vector2{X: 1, Y: 0})
	if got.X != 0 || got.Y == 0 {
		t.Fatalf("force = %#v, want Y rotation", got)
	}
}

func TestTurbulenceField(t *testing.T) {
	field := &TurbulenceField{Strength: 2, Scale: 1}
	got := field.Apply(Vector2{})
	// At origin: sin(0)=0, cos(0)=1, so base force is (0, 2)
	// With multi-frequency: adds sin(0)*0.5=0 and cos(0)*0.5=1
	// Result: (0+0, 2+1) = (0, 3)
	if got.X != 0 || got.Y != 3 {
		t.Fatalf("force = %#v, want (0,3)", got)
	}
}

func TestCompositeField(t *testing.T) {
	composite := &CompositeField{}
	composite.Add(&GravityField{Gravity: Vector2{X: 1, Y: 0}})
	composite.Add(&GravityField{Gravity: Vector2{X: 0, Y: -1}})
	got := composite.Apply(Vector2{X: 5, Y: 5})
	if got.X != 1 || got.Y != -1 {
		t.Fatalf("force = %#v, want (1,-1)", got)
	}
}

func TestDampingField(t *testing.T) {
	field := &DampingField{Center: Vector2{}, Strength: 0.5, Radius: 10}
	got := field.Apply(Vector2{X: 5, Y: 0})
	// At distance 5 from center with radius 10: factor = 0.5 * (1 - 5/10) = 0.25
	if got.X != -0.25 || got.Y != -0.25 {
		t.Fatalf("force = %#v, want (-0.25,-0.25)", got)
	}
	// Outside radius
	outside := field.Apply(Vector2{X: 15, Y: 0})
	if outside.X != 0 || outside.Y != 0 {
		t.Fatalf("outside force = %#v, want (0,0)", outside)
	}
}

func TestParticleSystemForceFields(t *testing.T) {
	ps := NewParticleSystem(4)
	ps.Emit(Particle{Life: 2, MaxLife: 2})
	ps.AddForceField(&GravityField{Gravity: Vector2{X: 1, Y: 0}})
	ps.Update(1)
	if len(ps.particles) != 1 {
		t.Fatalf("particle count = %d, want 1", len(ps.particles))
	}
	if ps.particles[0].Velocity.X != 1 {
		t.Fatalf("velocity.X = %f, want 1", ps.particles[0].Velocity.X)
	}
}
