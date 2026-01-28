//go:build js

package gpu

import (
	"image/color"
)

// Stub implementations for GPU effects on WASM builds.
// These functions return false to indicate GPU acceleration is not available,
// causing the caller to fall back to software rendering.

func ensureEffectTexture(driver Driver, tex Texture) (Texture, bool, bool) {
	return tex, false, false
}

func copyTextureGPU(driver Driver, src Texture, dst Framebuffer) bool {
	return false
}

func blurGPU(driver Driver, src Texture, dst Framebuffer, radius float32) bool {
	return false
}

func glowGPU(driver Driver, src Texture, dst Framebuffer, radius float32, intensity float32, col color.RGBA) bool {
	return false
}

func shadowGPU(driver Driver, src Texture, dst Framebuffer, radius, offsetX, offsetY float32, col color.RGBA) bool {
	return false
}

func chromaticGPU(driver Driver, src Texture, dst Framebuffer, amount float32) bool {
	return false
}

func vignetteGPU(driver Driver, src Texture, dst Framebuffer, radius, softness float32) bool {
	return false
}

func pixelateGPU(driver Driver, src Texture, dst Framebuffer, size float32) bool {
	return false
}

func colorGradeGPU(driver Driver, src Texture, dst Framebuffer, brightness, contrast, saturation, hue float32) bool {
	return false
}

func customEffectGPU(driver Driver, src Texture, dst Framebuffer, shader Shader, uniforms map[string]any) bool {
	return false
}

func ensureGPUFallback(driver Driver, src Texture, dst Framebuffer) (Texture, bool, bool, error) {
	return src, false, false, nil
}
