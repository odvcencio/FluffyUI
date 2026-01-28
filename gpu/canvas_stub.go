//go:build js

package gpu

import (
	"image/color"
)

// Stub implementations for WASM builds where GPU acceleration is not available.
// These methods are no-ops and rely on the software fallback in canvas.go.

func (c *GPUCanvas) clearGPU(col color.RGBA) {}

func (c *GPUCanvas) fillRectGPU(x, y, w, h float32) {}

func (c *GPUCanvas) strokeRectGPU(x, y, w, h float32) {}

func (c *GPUCanvas) fillPolygonGPU(points []vec2, col color.RGBA) {}

func (c *GPUCanvas) strokePointsGPU(points []vec2, col color.RGBA, width float32, closed bool) {}

func (c *GPUCanvas) drawLineGPU(p1, p2 vec2, col color.RGBA, width float32) {}

func (c *GPUCanvas) drawTextureGPU(img Texture, x, y, w, h float32) {}

func (c *GPUCanvas) drawTextGPU(text string, x, y float32, font *Font) {}

func (c *GPUCanvas) compositeGPU(src, dst Framebuffer) {}

func (c *GPUCanvas) drawSolid(vertices []float32, indices []uint16, blend BlendMode, target Framebuffer) {}

func (c *GPUCanvas) drawTextured(vertices []float32, indices []uint16, tex Texture, blend BlendMode, target Framebuffer) {}

func (c *GPUCanvas) drawTexturedQuad(tex Texture, x, y, w, h float32, target Framebuffer, transform Matrix3, blend BlendMode) {}

func (c *GPUCanvas) ensureGPUTexture(img Texture) (Texture, bool, bool) {
	return img, false, false
}
