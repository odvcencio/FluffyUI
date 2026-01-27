package gpu

import (
	"image"
	"image/color"
	"math"
)

// Effect applies a post-processing operation.
type Effect interface {
	Apply(src Texture, dst Framebuffer, driver Driver)
}

// BlurEffect applies a box blur.
type BlurEffect struct {
	Radius float32
}

// GlowEffect applies a colored glow.
type GlowEffect struct {
	Radius    float32
	Intensity float32
	Color     color.RGBA
}

// ShadowEffect applies a drop shadow.
type ShadowEffect struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Color   color.RGBA
}

// ChromaticAberration offsets color channels.
type ChromaticAberration struct {
	Amount float32
}

// VignetteEffect darkens edges.
type VignetteEffect struct {
	Radius   float32
	Softness float32
}

// PixelateEffect reduces resolution.
type PixelateEffect struct {
	PixelSize float32
}

// ColorGradeEffect adjusts color properties.
type ColorGradeEffect struct {
	Brightness float32
	Contrast   float32
	Saturation float32
	Hue        float32
}

// CustomEffect applies a custom shader.
type CustomEffect struct {
	Shader   Shader
	Uniforms map[string]any
}

func (e BlurEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if blurGPU(driver, tex, dst, e.Radius) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	r := int(math.Round(float64(e.Radius)))
	blurRGBA(sp, dp, w, h, r)
}

func (e GlowEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if glowGPU(driver, tex, dst, e.Radius, e.Intensity, e.Color) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	r := int(math.Round(float64(e.Radius)))
	blur := make([]byte, len(sp))
	blurRGBA(sp, blur, w, h, r)
	copy(dp, sp)
	intensity := e.Intensity
	if intensity <= 0 {
		return
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			a := float32(blur[idx+3]) / 255
			if a == 0 {
				continue
			}
			alpha := a * intensity
			col := e.Color
			col.A = floatToByte(alpha * float32(col.A) / 255)
			blendPixel(dp, w, h, x, y, col)
		}
	}
}

func (e ShadowEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if shadowGPU(driver, tex, dst, e.Blur, e.OffsetX, e.OffsetY, e.Color) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	blurRadius := int(math.Round(float64(e.Blur)))
	blur := make([]byte, len(sp))
	blurRGBA(sp, blur, w, h, blurRadius)
	clearPixels(dp, w, h, color.RGBA{})
	offX := int(math.Round(float64(e.OffsetX)))
	offY := int(math.Round(float64(e.OffsetY)))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			a := float32(blur[idx+3]) / 255
			if a == 0 {
				continue
			}
			col := e.Color
			col.A = floatToByte(a * float32(col.A) / 255)
			blendPixel(dp, w, h, x+offX, y+offY, col)
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			col := color.RGBA{R: sp[idx], G: sp[idx+1], B: sp[idx+2], A: sp[idx+3]}
			if col.A == 0 {
				continue
			}
			blendPixel(dp, w, h, x, y, col)
		}
	}
}

func (e ChromaticAberration) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if chromaticGPU(driver, tex, dst, e.Amount) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	offset := int(math.Round(float64(e.Amount)))
	if offset == 0 {
		copy(dp, sp)
		return
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			rIdx := (y*w + clampInt(x-offset, 0, w-1)) * 4
			bIdx := (y*w + clampInt(x+offset, 0, w-1)) * 4
			dp[idx] = sp[rIdx]
			dp[idx+1] = sp[idx+1]
			dp[idx+2] = sp[bIdx+2]
			dp[idx+3] = sp[idx+3]
		}
	}
}

func (e VignetteEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if vignetteGPU(driver, tex, dst, e.Radius, e.Softness) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	copy(dp, sp)
	cx := float32(w-1) / 2
	cy := float32(h-1) / 2
	maxDist := float32(math.Hypot(float64(cx), float64(cy)))
	radius := e.Radius
	if radius <= 0 {
		radius = 0.7
	}
	soft := e.Softness
	if soft <= 0 {
		soft = 0.3
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if dp[idx+3] == 0 {
				continue
			}
			dx := float32(x) - cx
			dy := float32(y) - cy
			d := float32(math.Hypot(float64(dx), float64(dy))) / maxDist
			if d <= radius {
				continue
			}
			t := (d - radius) / soft
			if t > 1 {
				t = 1
			}
			factor := 1 - t
			dp[idx] = floatToByte(float32(dp[idx]) / 255 * factor)
			dp[idx+1] = floatToByte(float32(dp[idx+1]) / 255 * factor)
			dp[idx+2] = floatToByte(float32(dp[idx+2]) / 255 * factor)
		}
	}
}

func (e PixelateEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if pixelateGPU(driver, tex, dst, e.PixelSize) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	block := int(math.Round(float64(e.PixelSize)))
	if block <= 1 {
		copy(dp, sp)
		return
	}
	for by := 0; by < h; by += block {
		for bx := 0; bx < w; bx += block {
			var sr, sg, sb, sa, count int
			for y := 0; y < block && by+y < h; y++ {
				for x := 0; x < block && bx+x < w; x++ {
					idx := ((by+y)*w + (bx + x)) * 4
					sr += int(sp[idx])
					sg += int(sp[idx+1])
					sb += int(sp[idx+2])
					sa += int(sp[idx+3])
					count++
				}
			}
			if count == 0 {
				continue
			}
			avg := color.RGBA{
				R: uint8(sr / count),
				G: uint8(sg / count),
				B: uint8(sb / count),
				A: uint8(sa / count),
			}
			for y := 0; y < block && by+y < h; y++ {
				for x := 0; x < block && bx+x < w; x++ {
					idx := ((by+y)*w + (bx + x)) * 4
					dp[idx] = avg.R
					dp[idx+1] = avg.G
					dp[idx+2] = avg.B
					dp[idx+3] = avg.A
				}
			}
		}
	}
}

func (e ColorGradeEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if tex, temp, ok := ensureEffectTexture(driver, src); ok {
			if colorGradeGPU(driver, tex, dst, e.Brightness, e.Contrast, e.Saturation, e.Hue) {
				if temp {
					tex.Dispose()
				}
				return
			}
			if temp {
				tex.Dispose()
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	copy(dp, sp)
	brightness := e.Brightness
	contrast := e.Contrast
	saturation := e.Saturation
	hue := e.Hue
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if dp[idx+3] == 0 {
				continue
			}
			r := float32(dp[idx]) / 255
			g := float32(dp[idx+1]) / 255
			b := float32(dp[idx+2]) / 255
			if brightness != 0 {
				r = clampFloat(r + brightness)
				g = clampFloat(g + brightness)
				b = clampFloat(b + brightness)
			}
			if contrast != 0 {
				factor := 1 + contrast
				r = clampFloat((r-0.5)*factor + 0.5)
				g = clampFloat((g-0.5)*factor + 0.5)
				b = clampFloat((b-0.5)*factor + 0.5)
			}
			if saturation != 0 {
				luma := 0.2126*r + 0.7152*g + 0.0722*b
				r = clampFloat(luma + (r-luma)*(1+saturation))
				g = clampFloat(luma + (g-luma)*(1+saturation))
				b = clampFloat(luma + (b-luma)*(1+saturation))
			}
			if hue != 0 {
				r, g, b = rotateHue(r, g, b, hue)
			}
			dp[idx] = floatToByte(r)
			dp[idx+1] = floatToByte(g)
			dp[idx+2] = floatToByte(b)
		}
	}
}

func (e CustomEffect) Apply(src Texture, dst Framebuffer, driver Driver) {
	if driver != nil && driver.Backend() != BackendSoftware {
		if e.Shader != nil {
			if _, ok := e.Shader.(*softwareShader); !ok {
				if tex, temp, ok := ensureEffectTexture(driver, src); ok {
					if customEffectGPU(driver, tex, dst, e.Shader, e.Uniforms) {
						if temp {
							tex.Dispose()
						}
						return
					}
					if temp {
						tex.Dispose()
					}
				}
			}
		}
	}
	sp, w, h, ok := texturePixels(src, driver)
	dp, dw, dh, flush, okDst := framebufferPixels(dst, driver)
	if !ok || !okDst {
		return
	}
	if w != dw || h != dh {
		return
	}
	defer flush(dp)
	if shader, ok := e.Shader.(*softwareShader); ok && shader.apply != nil {
		out := shader.apply(sp, w, h, mergeUniforms(shader.uniforms, e.Uniforms))
		if len(out) == len(dp) {
			copy(dp, out)
			return
		}
	}
	copy(dp, sp)
}

type textureReader interface {
	ReadTexturePixels(tex Texture, rect image.Rectangle) ([]byte, int, int, error)
}

func texturePixels(tex Texture, driver Driver) ([]byte, int, int, bool) {
	if tex == nil {
		return nil, 0, 0, false
	}
	if sw, ok := tex.(*softwareTexture); ok && sw != nil {
		return sw.pixels, sw.width, sw.height, true
	}
	if driver == nil {
		return nil, 0, 0, false
	}
	if reader, ok := driver.(textureReader); ok {
		pixels, w, h, err := reader.ReadTexturePixels(tex, image.Rectangle{})
		if err != nil || len(pixels) == 0 {
			return nil, 0, 0, false
		}
		return pixels, w, h, true
	}
	return nil, 0, 0, false
}

func framebufferPixels(fb Framebuffer, _ Driver) ([]byte, int, int, func([]byte), bool) {
	if fb == nil {
		return nil, 0, 0, nil, false
	}
	if sw, ok := fb.(*softwareFramebuffer); ok && sw != nil && sw.tex != nil {
		return sw.tex.pixels, sw.tex.width, sw.tex.height, func([]byte) {}, true
	}
	tex := fb.Texture()
	if tex == nil {
		return nil, 0, 0, nil, false
	}
	w, h := tex.Size()
	if w <= 0 || h <= 0 {
		return nil, 0, 0, nil, false
	}
	buf := make([]byte, w*h*4)
	flush := func(pixels []byte) {
		if len(pixels) == 0 {
			return
		}
		tex.Upload(pixels, image.Rect(0, 0, w, h))
	}
	return buf, w, h, flush, true
}

func blurRGBA(src, dst []byte, w, h, radius int) {
	if radius <= 0 || len(src) == 0 {
		copy(dst, src)
		return
	}
	tmp := make([]byte, len(src))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sr, sg, sb, sa, count int
			for k := -radius; k <= radius; k++ {
				sx := clampInt(x+k, 0, w-1)
				idx := (y*w + sx) * 4
				sr += int(src[idx])
				sg += int(src[idx+1])
				sb += int(src[idx+2])
				sa += int(src[idx+3])
				count++
			}
			idx := (y*w + x) * 4
			tmp[idx] = uint8(sr / count)
			tmp[idx+1] = uint8(sg / count)
			tmp[idx+2] = uint8(sb / count)
			tmp[idx+3] = uint8(sa / count)
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sr, sg, sb, sa, count int
			for k := -radius; k <= radius; k++ {
				sy := clampInt(y+k, 0, h-1)
				idx := (sy*w + x) * 4
				sr += int(tmp[idx])
				sg += int(tmp[idx+1])
				sb += int(tmp[idx+2])
				sa += int(tmp[idx+3])
				count++
			}
			idx := (y*w + x) * 4
			dst[idx] = uint8(sr / count)
			dst[idx+1] = uint8(sg / count)
			dst[idx+2] = uint8(sb / count)
			dst[idx+3] = uint8(sa / count)
		}
	}
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func rotateHue(r, g, b, hue float32) (float32, float32, float32) {
	angle := float64(hue)
	cosA := float32(math.Cos(angle))
	sinA := float32(math.Sin(angle))
	m00 := 0.213 + cosA*0.787 - sinA*0.213
	m01 := 0.715 - cosA*0.715 - sinA*0.715
	m02 := 0.072 - cosA*0.072 + sinA*0.928
	m10 := 0.213 - cosA*0.213 + sinA*0.143
	m11 := 0.715 + cosA*0.285 + sinA*0.140
	m12 := 0.072 - cosA*0.072 - sinA*0.283
	m20 := 0.213 - cosA*0.213 - sinA*0.787
	m21 := 0.715 - cosA*0.715 + sinA*0.715
	m22 := 0.072 + cosA*0.928 + sinA*0.072
	nr := clampFloat(r*m00 + g*m01 + b*m02)
	ng := clampFloat(r*m10 + g*m11 + b*m12)
	nb := clampFloat(r*m20 + g*m21 + b*m22)
	return nr, ng, nb
}

func mergeUniforms(base, extra map[string]any) map[string]any {
	if base == nil && extra == nil {
		return nil
	}
	out := make(map[string]any)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}
