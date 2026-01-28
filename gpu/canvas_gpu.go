package gpu

import (
	"image"
	"image/color"
	"math"
)

const maxUint16 = 65535

func (c *GPUCanvas) clearGPU(col color.RGBA) {
	if c == nil || c.driver == nil || c.gpuFb == nil {
		return
	}
	c.gpuFb.Bind()
	c.driver.Clear(float32(col.R)/255, float32(col.G)/255, float32(col.B)/255, float32(col.A)/255)
}

func (c *GPUCanvas) fillRectGPU(x, y, w, h float32) {
	p0 := c.applyTransform(vec2{x: x, y: y})
	p1 := c.applyTransform(vec2{x: x + w, y: y})
	p2 := c.applyTransform(vec2{x: x + w, y: y + h})
	p3 := c.applyTransform(vec2{x: x, y: y + h})
	verts := make([]float32, 0, 24)
	inds := make([]uint16, 0, 6)
	appendSolidQuad(&verts, &inds, p0, p1, p2, p3, c.fillColor)
	c.drawSolid(verts, inds, blendForColor(c.fillColor), c.gpuFb)
}

func (c *GPUCanvas) strokeRectGPU(x, y, w, h float32) {
	points := []vec2{{x: x, y: y}, {x: x + w, y: y}, {x: x + w, y: y + h}, {x: x, y: y + h}}
	c.strokePointsGPU(points, c.strokeColor, c.strokeWidth, true)
}

func (c *GPUCanvas) fillPolygonGPU(points []vec2, col color.RGBA) {
	if len(points) < 3 {
		return
	}
	transformed := make([]vec2, len(points))
	for i, p := range points {
		transformed[i] = c.applyTransform(p)
	}
	vertsPoints, indices := triangulatePolygon(transformed)
	if len(indices) == 0 || len(vertsPoints) == 0 {
		return
	}
	if len(vertsPoints) > maxUint16 {
		return
	}
	verts := make([]float32, 0, len(vertsPoints)*6)
	r, g, b, a := colorToFloat(col)
	for _, p := range vertsPoints {
		verts = append(verts, p.x, p.y, r, g, b, a)
	}
	c.drawSolid(verts, indices, blendForColor(col), c.gpuFb)
}

func (c *GPUCanvas) strokePointsGPU(points []vec2, col color.RGBA, width float32, closed bool) {
	if len(points) < 2 {
		return
	}
	transformed := make([]vec2, len(points))
	for i, p := range points {
		transformed[i] = c.applyTransform(p)
	}
	last := len(transformed) - 1
	for i := 0; i < last; i++ {
		c.drawLineGPU(transformed[i], transformed[i+1], col, width)
	}
	if closed && len(transformed) > 2 {
		c.drawLineGPU(transformed[last], transformed[0], col, width)
	}
}

func (c *GPUCanvas) drawLineGPU(p1, p2 vec2, col color.RGBA, width float32) {
	dx := p2.x - p1.x
	dy := p2.y - p1.y
	len := float32(math.Hypot(float64(dx), float64(dy)))
	if len == 0 {
		return
	}
	if width <= 0 {
		width = 1
	}
	half := width / 2
	nx := -dy / len
	ny := dx / len
	v1 := vec2{x: p1.x + nx*half, y: p1.y + ny*half}
	v2 := vec2{x: p1.x - nx*half, y: p1.y - ny*half}
	v3 := vec2{x: p2.x - nx*half, y: p2.y - ny*half}
	v4 := vec2{x: p2.x + nx*half, y: p2.y + ny*half}
	verts := make([]float32, 0, 24)
	inds := make([]uint16, 0, 6)
	appendSolidQuad(&verts, &inds, v1, v2, v3, v4, col)
	c.drawSolid(verts, inds, blendForColor(col), c.gpuFb)
}

func (c *GPUCanvas) drawTextureGPU(img Texture, x, y, w, h float32) {
	if c == nil || img == nil || c.driver == nil {
		return
	}
	tex, temp, ok := c.ensureGPUTexture(img)
	if !ok {
		return
	}
	if temp {
		defer tex.Dispose()
	}
	c.drawTexturedQuad(tex, x, y, w, h, c.gpuFb, c.transform, BlendAlpha)
}

func (c *GPUCanvas) drawTextGPU(text string, x, y float32, font *Font) {
	if text == "" {
		return
	}
	if font == nil {
		font = DefaultFont()
	}
	pxFont := font.Face
	if pxFont == nil {
		return
	}
	verts := make([]float32, 0, 256)
	inds := make([]uint16, 0, 384)
	cursorX := x
	for _, r := range text {
		glyph := pxFont.Glyph(r)
		if glyph != nil {
			for gy, row := range glyph {
				for gx := 0; gx < len(row); gx++ {
					if row[gx] != '#' {
						continue
					}
					p0 := c.applyTransform(vec2{x: cursorX + float32(gx), y: y + float32(gy)})
					p1 := c.applyTransform(vec2{x: cursorX + float32(gx+1), y: y + float32(gy)})
					p2 := c.applyTransform(vec2{x: cursorX + float32(gx+1), y: y + float32(gy+1)})
					p3 := c.applyTransform(vec2{x: cursorX + float32(gx), y: y + float32(gy+1)})
					if !appendSolidQuad(&verts, &inds, p0, p1, p2, p3, c.fillColor) {
						c.drawSolid(verts, inds, blendForColor(c.fillColor), c.gpuFb)
						verts = verts[:0]
						inds = inds[:0]
						appendSolidQuad(&verts, &inds, p0, p1, p2, p3, c.fillColor)
					}
				}
			}
		}
		cursorX += float32(pxFont.Width + pxFont.Spacing)
	}
	if len(inds) > 0 {
		c.drawSolid(verts, inds, blendForColor(c.fillColor), c.gpuFb)
	}
}

func (c *GPUCanvas) compositeGPU(src, dst Framebuffer) {
	if c == nil || src == nil || dst == nil {
		return
	}
	tex := src.Texture()
	if tex == nil {
		return
	}
	w, h := dst.Size()
	c.drawTexturedQuad(tex, 0, 0, float32(w), float32(h), dst, Identity(), BlendAlpha)
}

func (c *GPUCanvas) drawSolid(vertices []float32, indices []uint16, blend BlendMode, target Framebuffer) {
	if c == nil || c.driver == nil || c.pipeline == nil || len(vertices) == 0 || len(indices) == 0 {
		return
	}
	if target == nil {
		target = c.gpuFb
	}
	c.driver.Draw(DrawCall{
		Shader:   c.pipeline.solid,
		Vertices: vertices,
		Indices:  indices,
		Target:   target,
		Blend:    blend,
	})
}

func (c *GPUCanvas) drawTextured(vertices []float32, indices []uint16, tex Texture, blend BlendMode, target Framebuffer) {
	if c == nil || c.driver == nil || c.pipeline == nil || len(vertices) == 0 || len(indices) == 0 || tex == nil {
		return
	}
	if target == nil {
		target = c.gpuFb
	}
	c.driver.Draw(DrawCall{
		Shader:   c.pipeline.texture,
		Vertices: vertices,
		Indices:  indices,
		Textures: []Texture{tex},
		Target:   target,
		Blend:    blend,
	})
}

func (c *GPUCanvas) drawTexturedQuad(tex Texture, x, y, w, h float32, target Framebuffer, transform Matrix3, blend BlendMode) {
	p0 := applyTransformPoint(transform, vec2{x: x, y: y})
	p1 := applyTransformPoint(transform, vec2{x: x + w, y: y})
	p2 := applyTransformPoint(transform, vec2{x: x + w, y: y + h})
	p3 := applyTransformPoint(transform, vec2{x: x, y: y + h})
	verts := make([]float32, 0, 16)
	inds := make([]uint16, 0, 6)
	appendTexturedQuad(&verts, &inds, p0, p1, p2, p3)
	c.drawTextured(verts, inds, tex, blend, target)
}

func (c *GPUCanvas) ensureGPUTexture(img Texture) (Texture, bool, bool) {
	if img == nil || c.driver == nil {
		return nil, false, false
	}
	// If the texture is already a GPU-native texture for this backend, use it directly
	if isNativeTexture(img, c.driver.Backend()) {
		return img, false, true
	}
	// Otherwise, upload the pixel data to a new texture
	if pixels, w, h, ok := texturePixels(img, c.driver); ok {
		tex, err := c.driver.NewTexture(w, h)
		if err != nil {
			return nil, false, false
		}
		tex.Upload(pixels, image.Rectangle{})
		return tex, true, true
	}
	return nil, false, false
}

func blendForColor(col color.RGBA) BlendMode {
	if col.A < 255 {
		return BlendAlpha
	}
	return BlendNone
}

// isNativeTexture checks if a texture is native to the given backend.
// This function has platform-specific implementations.
func isNativeTexture(img Texture, backend Backend) bool {
	return isNativeTextureForPlatform(img, backend)
}

// isNativeFramebuffer checks if a framebuffer is native to the given backend.
// This function has platform-specific implementations.
func isNativeFramebuffer(fb Framebuffer, backend Backend) bool {
	return isNativeFramebufferForPlatform(fb, backend)
}

func colorToFloat(col color.RGBA) (float32, float32, float32, float32) {
	return float32(col.R) / 255, float32(col.G) / 255, float32(col.B) / 255, float32(col.A) / 255
}

func appendSolidQuad(vertices *[]float32, indices *[]uint16, p0, p1, p2, p3 vec2, col color.RGBA) bool {
	if vertices == nil || indices == nil {
		return false
	}
	base := len(*vertices) / 6
	if base+4 > maxUint16 {
		return false
	}
	r, g, b, a := colorToFloat(col)
	*vertices = append(*vertices,
		p0.x, p0.y, r, g, b, a,
		p1.x, p1.y, r, g, b, a,
		p2.x, p2.y, r, g, b, a,
		p3.x, p3.y, r, g, b, a,
	)
	*indices = append(*indices,
		uint16(base), uint16(base+1), uint16(base+2),
		uint16(base+2), uint16(base+3), uint16(base),
	)
	return true
}

func appendTexturedQuad(vertices *[]float32, indices *[]uint16, p0, p1, p2, p3 vec2) bool {
	if vertices == nil || indices == nil {
		return false
	}
	base := len(*vertices) / 4
	if base+4 > maxUint16 {
		return false
	}
	*vertices = append(*vertices,
		p0.x, p0.y, 0, 1,
		p1.x, p1.y, 1, 1,
		p2.x, p2.y, 1, 0,
		p3.x, p3.y, 0, 0,
	)
	*indices = append(*indices,
		uint16(base), uint16(base+1), uint16(base+2),
		uint16(base+2), uint16(base+3), uint16(base),
	)
	return true
}

func applyTransformPoint(m Matrix3, p vec2) vec2 {
	x, y := m.Apply(p.x, p.y)
	return vec2{x: x, y: y}
}

func triangulatePolygon(points []vec2) ([]vec2, []uint16) {
	cleaned := dedupePolygon(points)
	if len(cleaned) < 3 {
		return nil, nil
	}
	cleaned = stripCollinear(cleaned)
	if len(cleaned) < 3 {
		return nil, nil
	}
	if len(cleaned) > maxUint16 {
		return nil, nil
	}
	ccw := polygonArea(cleaned) > 0
	indices := make([]int, len(cleaned))
	for i := range indices {
		indices[i] = i
	}
	result := make([]uint16, 0, (len(cleaned)-2)*3)
	guard := 0
	for len(indices) > 2 && guard < 10000 {
		earFound := false
		for i := 0; i < len(indices); i++ {
			prev := indices[(i-1+len(indices))%len(indices)]
			curr := indices[i]
			next := indices[(i+1)%len(indices)]
			if !isConvex(cleaned[prev], cleaned[curr], cleaned[next], ccw) {
				continue
			}
			if containsPoint(cleaned, indices, prev, curr, next) {
				continue
			}
			result = append(result, uint16(prev), uint16(curr), uint16(next))
			indices = append(indices[:i], indices[i+1:]...)
			earFound = true
			break
		}
		if !earFound {
			break
		}
		guard++
	}
	return cleaned, result
}

func dedupePolygon(points []vec2) []vec2 {
	if len(points) == 0 {
		return nil
	}
	out := make([]vec2, 0, len(points))
	for _, p := range points {
		if len(out) == 0 || !almostEqualVec(out[len(out)-1], p) {
			out = append(out, p)
		}
	}
	if len(out) > 1 && almostEqualVec(out[0], out[len(out)-1]) {
		out = out[:len(out)-1]
	}
	return out
}

func stripCollinear(points []vec2) []vec2 {
	if len(points) < 3 {
		return points
	}
	out := make([]vec2, 0, len(points))
	for i := 0; i < len(points); i++ {
		prev := points[(i-1+len(points))%len(points)]
		curr := points[i]
		next := points[(i+1)%len(points)]
		if isCollinear(prev, curr, next) {
			continue
		}
		out = append(out, curr)
	}
	return out
}

func isCollinear(a, b, c vec2) bool {
	cross := (b.x-a.x)*(c.y-a.y) - (b.y-a.y)*(c.x-a.x)
	return math.Abs(float64(cross)) < 1e-5
}

func polygonArea(points []vec2) float32 {
	var area float32
	for i := 0; i < len(points); i++ {
		j := (i + 1) % len(points)
		area += points[i].x*points[j].y - points[j].x*points[i].y
	}
	return area * 0.5
}

func isConvex(a, b, c vec2, ccw bool) bool {
	cross := (b.x-a.x)*(c.y-a.y) - (b.y-a.y)*(c.x-a.x)
	if ccw {
		return cross > 0
	}
	return cross < 0
}

func containsPoint(points []vec2, indices []int, i0, i1, i2 int) bool {
	a := points[i0]
	b := points[i1]
	c := points[i2]
	for _, idx := range indices {
		if idx == i0 || idx == i1 || idx == i2 {
			continue
		}
		if pointInTriangle(points[idx], a, b, c) {
			return true
		}
	}
	return false
}

func pointInTriangle(p, a, b, c vec2) bool {
	b1 := sign(p, a, b) < 0
	b2 := sign(p, b, c) < 0
	b3 := sign(p, c, a) < 0
	return (b1 == b2) && (b2 == b3)
}

func sign(p, a, b vec2) float32 {
	return (p.x-a.x)*(b.y-a.y) - (b.x-a.x)*(p.y-a.y)
}

func almostEqualVec(a, b vec2) bool {
	return math.Abs(float64(a.x-b.x)) < 1e-4 && math.Abs(float64(a.y-b.y)) < 1e-4
}

// layoutInfo describes the structure of vertex data.
type layoutInfo struct {
	stride    int
	posSize   int
	uvSize    int
	colorSize int
}

// inferLayout determines vertex layout from vertex data structure.
func inferLayout(vertices []float32) layoutInfo {
	if len(vertices)%8 == 0 {
		return layoutInfo{stride: 8, posSize: 2, uvSize: 2, colorSize: 4}
	}
	if len(vertices)%6 == 0 {
		return layoutInfo{stride: 6, posSize: 2, colorSize: 4}
	}
	if len(vertices)%4 == 0 {
		return layoutInfo{stride: 4, posSize: 2, uvSize: 2}
	}
	return layoutInfo{stride: 2, posSize: 2}
}
