//go:build js

package gpu

// isNativeTextureForPlatform checks if a texture is native to the given backend (WASM).
func isNativeTextureForPlatform(img Texture, backend Backend) bool {
	switch backend {
	case BackendWebGL:
		_, ok := img.(*webglTexture)
		return ok
	default:
		return false
	}
}

// isNativeFramebufferForPlatform checks if a framebuffer is native to the given backend (WASM).
func isNativeFramebufferForPlatform(fb Framebuffer, backend Backend) bool {
	switch backend {
	case BackendWebGL:
		_, ok := fb.(*webglFramebuffer)
		return ok
	default:
		return false
	}
}
