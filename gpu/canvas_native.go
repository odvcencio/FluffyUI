//go:build !js

package gpu

// isNativeTextureForPlatform checks if a texture is native to the given backend (non-WASM).
func isNativeTextureForPlatform(img Texture, backend Backend) bool {
	switch backend {
	case BackendOpenGL:
		_, ok := img.(*openglTexture)
		return ok
	default:
		return false
	}
}

// isNativeFramebufferForPlatform checks if a framebuffer is native to the given backend (non-WASM).
func isNativeFramebufferForPlatform(fb Framebuffer, backend Backend) bool {
	switch backend {
	case BackendOpenGL:
		_, ok := fb.(*openglFramebuffer)
		return ok
	default:
		return false
	}
}
