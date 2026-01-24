package animation

import "math"

// ForceField applies a force based on a position in space.
type ForceField interface {
	Apply(position Vector2) Vector2
}

// GravityField applies a constant force everywhere.
type GravityField struct {
	Gravity Vector2
}

// Apply returns the gravity vector.
func (g *GravityField) Apply(position Vector2) Vector2 {
	if g == nil {
		return Vector2{}
	}
	return g.Gravity
}

// RadialField attracts or repels particles around a center point.
type RadialField struct {
	Center   Vector2
	Strength float64
	Radius   float64
}

// Apply returns the radial force for the position.
func (r *RadialField) Apply(position Vector2) Vector2 {
	if r == nil {
		return Vector2{}
	}
	dx := r.Center.X - position.X
	dy := r.Center.Y - position.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist == 0 || (r.Radius > 0 && dist > r.Radius) {
		return Vector2{}
	}
	strength := r.Strength / (dist * dist)
	return Vector2{X: dx / dist * strength, Y: dy / dist * strength}
}

// VortexField spins particles around a center point.
type VortexField struct {
	Center   Vector2
	Strength float64
	Radius   float64
}

// Apply returns the vortex force for the position.
func (v *VortexField) Apply(position Vector2) Vector2 {
	if v == nil {
		return Vector2{}
	}
	dx := position.X - v.Center.X
	dy := position.Y - v.Center.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist == 0 || (v.Radius > 0 && dist > v.Radius) {
		return Vector2{}
	}
	strength := v.Strength / dist
	return Vector2{X: -dy / dist * strength, Y: dx / dist * strength}
}

// TurbulenceField applies a simple oscillating noise field.
type TurbulenceField struct {
	Strength float64
	Scale    float64
}

// Apply returns the turbulence force for the position.
func (t *TurbulenceField) Apply(position Vector2) Vector2 {
	if t == nil {
		return Vector2{}
	}
	scale := t.Scale
	if scale == 0 {
		scale = 1
	}
	// Use a combination of frequencies for more natural turbulence
	x1 := math.Sin(position.X*scale) * t.Strength
	y1 := math.Cos(position.Y*scale) * t.Strength
	x2 := math.Sin(position.Y*scale*0.7+position.X*scale*0.3) * t.Strength * 0.5
	y2 := math.Cos(position.X*scale*0.7+position.Y*scale*0.3) * t.Strength * 0.5
	return Vector2{X: x1 + x2, Y: y1 + y2}
}

// CompositeField combines multiple force fields.
type CompositeField struct {
	Fields []ForceField
}

// Apply returns the sum of all field forces.
func (c *CompositeField) Apply(position Vector2) Vector2 {
	if c == nil || len(c.Fields) == 0 {
		return Vector2{}
	}
	result := Vector2{}
	for _, field := range c.Fields {
		if field == nil {
			continue
		}
		force := field.Apply(position)
		result.X += force.X
		result.Y += force.Y
	}
	return result
}

// Add appends a force field to the composite.
func (c *CompositeField) Add(field ForceField) *CompositeField {
	if c == nil {
		return nil
	}
	if field != nil {
		c.Fields = append(c.Fields, field)
	}
	return c
}

// DampingField reduces velocity over distance from a center point.
type DampingField struct {
	Center   Vector2
	Strength float64 // 0-1, higher = more damping
	Radius   float64 // 0 = infinite range
}

// Apply returns a damping force opposing current velocity direction.
func (d *DampingField) Apply(position Vector2) Vector2 {
	if d == nil || d.Strength <= 0 {
		return Vector2{}
	}
	dx := position.X - d.Center.X
	dy := position.Y - d.Center.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if d.Radius > 0 && dist > d.Radius {
		return Vector2{}
	}
	// Return position-based damping (caller multiplies by velocity)
	factor := d.Strength
	if d.Radius > 0 && dist > 0 {
		factor *= 1 - (dist / d.Radius)
	}
	return Vector2{X: -factor, Y: -factor}
}
