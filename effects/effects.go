package effects

import (
	"math"
	"math/rand"

	"github.com/odvcencio/fluffy-ui/animation"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/graphics"
)

// Shimmer creates a loading skeleton effect.
func Shimmer(canvas *graphics.Canvas, x, y, w, h int, phase float64, color backend.Color) {
	if canvas == nil || w <= 0 || h <= 0 {
		return
	}
	if phase < 0 {
		phase = 0
	}
	if phase > 1 {
		phase = 1
	}
	shimmerWidth := max(1, w/4)
	shimmerPos := int(phase*float64(w+shimmerWidth)) - shimmerWidth

	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			relX := px - x
			dist := float64(relX - shimmerPos)
			if dist >= 0 && dist < float64(shimmerWidth) {
				intensity := 1.0 - math.Abs(dist-float64(shimmerWidth)/2)/(float64(shimmerWidth)/2)
				if intensity < 0 {
					intensity = 0
				}
				canvas.Blend(px, py, color, float32(intensity*0.5))
			}
		}
	}
}

// Glow creates a glow effect around a point.
func Glow(canvas *graphics.Canvas, cx, cy, radius int, color backend.Color, intensity float64) {
	if canvas == nil || radius <= 0 {
		return
	}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				alpha := (1 - dist/float64(radius)) * intensity
				if alpha > 0 {
					canvas.Blend(cx+dx, cy+dy, color, float32(alpha))
				}
			}
		}
	}
}

// Ripple creates an expanding ring effect.
func Ripple(canvas *graphics.Canvas, cx, cy int, radius, maxRadius int, color backend.Color, thickness float64) {
	if canvas == nil || maxRadius <= 0 {
		return
	}
	for dy := -maxRadius; dy <= maxRadius; dy++ {
		for dx := -maxRadius; dx <= maxRadius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if math.Abs(dist-float64(radius)) < thickness {
				alpha := 1 - float64(radius)/float64(maxRadius)
				if alpha > 0 {
					canvas.Blend(cx+dx, cy+dy, color, float32(alpha))
				}
			}
		}
	}
}

// LinearGradient fills a region with a linear gradient.
func LinearGradient(canvas *graphics.Canvas, x, y, w, h int, startColor, endColor backend.Color, angle float64) {
	if canvas == nil || w <= 0 || h <= 0 {
		return
	}
	cos, sin := math.Cos(angle), math.Sin(angle)
	maxDist := math.Abs(float64(w)*cos) + math.Abs(float64(h)*sin)
	if maxDist == 0 {
		return
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			dist := float64(px)*cos + float64(py)*sin
			t := dist / maxDist
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			canvas.SetPixel(x+px, y+py, lerpColor(startColor, endColor, t))
		}
	}
}

// RadialGradient fills a region with a radial gradient.
func RadialGradient(canvas *graphics.Canvas, cx, cy, radius int, centerColor, edgeColor backend.Color) {
	if canvas == nil || radius <= 0 {
		return
	}
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= float64(radius) {
				t := dist / float64(radius)
				canvas.SetPixel(cx+dx, cy+dy, lerpColor(centerColor, edgeColor, t))
			}
		}
	}
}

// Shadow draws a drop shadow.
func Shadow(canvas *graphics.Canvas, x, y, w, h, offsetX, offsetY, blur int, color backend.Color) {
	if canvas == nil || w <= 0 || h <= 0 {
		return
	}
	for py := 0; py < h+blur*2; py++ {
		for px := 0; px < w+blur*2; px++ {
			sx := x + px + offsetX - blur
			sy := y + py + offsetY - blur

			inX := px >= blur && px < w+blur
			inY := py >= blur && py < h+blur

			if inX && inY {
				canvas.Blend(sx, sy, color, 0.3)
				continue
			}

			dx := 0
			if px < blur {
				dx = blur - px
			} else if px >= w+blur {
				dx = px - w - blur + 1
			}
			dy := 0
			if py < blur {
				dy = blur - py
			} else if py >= h+blur {
				dy = py - h - blur + 1
			}

			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist < float64(blur) {
				alpha := 0.3 * (1 - dist/float64(blur))
				canvas.Blend(sx, sy, color, float32(alpha))
			}
		}
	}
}

// Confetti emits a burst of confetti particles.
func Confetti(ps *animation.ParticleSystem, x, y int, count int) {
	if ps == nil || count <= 0 {
		return
	}
	colors := []backend.Color{
		backend.ColorRGB(255, 107, 107),
		backend.ColorRGB(255, 193, 7),
		backend.ColorRGB(76, 175, 80),
		backend.ColorRGB(33, 150, 243),
		backend.ColorRGB(156, 39, 176),
	}
	start := colors[rand.Intn(len(colors))]
	end := colors[rand.Intn(len(colors))]

	ps.Burst(animation.Vector2{X: float64(x), Y: float64(y)}, count, animation.ParticleConfig{
		Speed:     animation.Range{Min: 50, Max: 150},
		Life:      animation.Range{Min: 1.0, Max: 2.0},
		Size:      animation.Range{Min: 1, Max: 3},
		Spread:    math.Pi * 2,
		Direction: -math.Pi / 2,
		Color:     animation.ColorRange{Start: start, End: end},
		Gravity:   animation.Vector2{X: 0, Y: 80},
	})
}

// Sparkle emits small sparkles in a region.
func Sparkle(ps *animation.ParticleSystem, x, y, w, h int, density float64) {
	if ps == nil || w <= 0 || h <= 0 {
		return
	}
	count := int(float64(w*h) * density)
	for i := 0; i < count; i++ {
		px := x + rand.Intn(w)
		py := y + rand.Intn(h)
		ps.Emit(animation.Particle{
			Position: animation.Vector2{X: float64(px), Y: float64(py)},
			Velocity: animation.Vector2{X: 0, Y: 0},
			Color:    backend.ColorRGB(255, 255, 200),
			Size:     1,
			Life:     0.3 + rand.Float64()*0.5,
			MaxLife:  0.8,
			Alpha:    1.0,
		})
	}
}

func lerpColor(a, b backend.Color, t float64) backend.Color {
	if !a.IsRGB() || !b.IsRGB() {
		if t >= 0.5 {
			return b
		}
		return a
	}
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	r := uint8(float64(ar) + (float64(br)-float64(ar))*t)
	g := uint8(float64(ag) + (float64(bg)-float64(ag))*t)
	bval := uint8(float64(ab) + (float64(bb)-float64(ab))*t)
	return backend.ColorRGB(r, g, bval)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
