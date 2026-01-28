//go:build !js

package gpu

import (
	"errors"
	"image"
	"image/color"
	"reflect"
	"sync"
)

type effectShaderCache struct {
	mu       sync.Mutex
	byDriver map[uintptr]map[string]Shader
}

var effectShaders effectShaderCache

func getEffectShader(driver Driver, name string) (Shader, error) {
	if driver == nil || name == "" {
		return nil, ErrUnsupported
	}
	key := driverKey(driver)
	if key == 0 {
		return nil, ErrUnsupported
	}
	effectShaders.mu.Lock()
	if effectShaders.byDriver == nil {
		effectShaders.byDriver = make(map[uintptr]map[string]Shader)
	}
	shaderMap := effectShaders.byDriver[key]
	if shaderMap == nil {
		shaderMap = make(map[string]Shader)
		effectShaders.byDriver[key] = shaderMap
	}
	if shader := shaderMap[name]; shader != nil {
		effectShaders.mu.Unlock()
		return shader, nil
	}
	effectShaders.mu.Unlock()
	src, err := LoadShaderSource(name)
	if err != nil {
		return nil, err
	}
	shader, err := driver.NewShader(src)
	if err != nil {
		return nil, err
	}
	shader.SetUniform("uTexture", 0)
	if name == "texture" {
		shader.SetUniform("uTransform", Identity())
	}
	effectShaders.mu.Lock()
	shaderMap = effectShaders.byDriver[key]
	if shaderMap == nil {
		shaderMap = make(map[string]Shader)
		effectShaders.byDriver[key] = shaderMap
	}
	shaderMap[name] = shader
	effectShaders.mu.Unlock()
	return shader, nil
}

func driverKey(driver Driver) uintptr {
	if driver == nil {
		return 0
	}
	v := reflect.ValueOf(driver)
	if v.Kind() != reflect.Pointer {
		return 0
	}
	return v.Pointer()
}

func ensureEffectTexture(driver Driver, tex Texture) (Texture, bool, bool) {
	if tex == nil || driver == nil {
		return nil, false, false
	}
	if driver.Backend() == BackendOpenGL {
		if gltex, ok := tex.(*openglTexture); ok {
			return gltex, false, true
		}
	}
	pixels, w, h, ok := texturePixels(tex, driver)
	if !ok {
		return nil, false, false
	}
	gtex, err := driver.NewTexture(w, h)
	if err != nil {
		return nil, false, false
	}
	gtex.Upload(pixels, image.Rectangle{})
	return gtex, true, true
}

func effectQuad(width, height int, offsetX, offsetY float32) ([]float32, []uint16) {
	dx := float32(0)
	dy := float32(0)
	if width > 0 {
		dx = (offsetX / float32(width)) * 2
	}
	if height > 0 {
		dy = -(offsetY / float32(height)) * 2
	}
	verts := []float32{
		-1 + dx, 1 + dy, 0, 1,
		1 + dx, 1 + dy, 1, 1,
		1 + dx, -1 + dy, 1, 0,
		-1 + dx, -1 + dy, 0, 0,
	}
	inds := []uint16{0, 1, 2, 2, 3, 0}
	return verts, inds
}

func drawEffect(driver Driver, shader Shader, tex Texture, target Framebuffer, blend BlendMode, vertices []float32, indices []uint16) {
	if driver == nil || shader == nil || tex == nil || target == nil || len(vertices) == 0 || len(indices) == 0 {
		return
	}
	driver.Draw(DrawCall{
		Shader:   shader,
		Vertices: vertices,
		Indices:  indices,
		Textures: []Texture{tex},
		Target:   target,
		Blend:    blend,
	})
}

func copyTextureGPU(driver Driver, src Texture, dst Framebuffer) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	shader, err := getEffectShader(driver, "texture")
	if err != nil {
		return false
	}
	w, h := dst.Size()
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func blurGPU(driver Driver, src Texture, dst Framebuffer, radius float32) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	if radius <= 0 {
		return copyTextureGPU(driver, src, dst)
	}
	w, h := dst.Size()
	if w <= 0 || h <= 0 {
		return false
	}
	blurShader, err := getEffectShader(driver, "blur")
	if err != nil {
		return false
	}
	temp, err := driver.NewFramebuffer(w, h)
	if err != nil {
		return false
	}
	defer temp.Dispose()
	dirX := [2]float32{1 / float32(w), 0}
	blurShader.SetUniform("uDirection", dirX)
	blurShader.SetUniform("uRadius", radius)
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, blurShader, src, temp, BlendNone, verts, inds)
	dirY := [2]float32{0, 1 / float32(h)}
	blurShader.SetUniform("uDirection", dirY)
	blurShader.SetUniform("uRadius", radius)
	drawEffect(driver, blurShader, temp.Texture(), dst, BlendNone, verts, inds)
	return true
}

func glowGPU(driver Driver, src Texture, dst Framebuffer, radius float32, intensity float32, col color.RGBA) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	w, h := dst.Size()
	if w <= 0 || h <= 0 {
		return false
	}
	if intensity <= 0 {
		return copyTextureGPU(driver, src, dst)
	}
	blurTemp, err := driver.NewFramebuffer(w, h)
	if err != nil {
		return false
	}
	defer blurTemp.Dispose()
	if !blurGPU(driver, src, blurTemp, radius) {
		return false
	}
	if dst != nil {
		dst.Bind()
		driver.Clear(0, 0, 0, 0)
	}
	textureShader, err := getEffectShader(driver, "texture")
	if err != nil {
		return false
	}
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, textureShader, src, dst, BlendAlpha, verts, inds)
	glowShader, err := getEffectShader(driver, "glow")
	if err != nil {
		return false
	}
	alpha := float32(col.A) / 255
	alpha *= intensity
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	glowShader.SetUniform("uGlowColor", [4]float32{float32(col.R) / 255, float32(col.G) / 255, float32(col.B) / 255, alpha})
	drawEffect(driver, glowShader, blurTemp.Texture(), dst, BlendAdditive, verts, inds)
	return true
}

func shadowGPU(driver Driver, src Texture, dst Framebuffer, radius, offsetX, offsetY float32, col color.RGBA) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	w, h := dst.Size()
	if w <= 0 || h <= 0 {
		return false
	}
	blurTemp, err := driver.NewFramebuffer(w, h)
	if err != nil {
		return false
	}
	defer blurTemp.Dispose()
	if !blurGPU(driver, src, blurTemp, radius) {
		return false
	}
	dst.Bind()
	driver.Clear(0, 0, 0, 0)
	glowShader, err := getEffectShader(driver, "glow")
	if err != nil {
		return false
	}
	alpha := float32(col.A) / 255
	glowShader.SetUniform("uGlowColor", [4]float32{float32(col.R) / 255, float32(col.G) / 255, float32(col.B) / 255, alpha})
	verts, inds := effectQuad(w, h, offsetX, offsetY)
	drawEffect(driver, glowShader, blurTemp.Texture(), dst, BlendAlpha, verts, inds)
	textureShader, err := getEffectShader(driver, "texture")
	if err != nil {
		return false
	}
	verts, inds = effectQuad(w, h, 0, 0)
	drawEffect(driver, textureShader, src, dst, BlendAlpha, verts, inds)
	return true
}

func chromaticGPU(driver Driver, src Texture, dst Framebuffer, amount float32) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	w, h := dst.Size()
	if w <= 0 || h <= 0 {
		return false
	}
	shader, err := getEffectShader(driver, "chromatic")
	if err != nil {
		return false
	}
	offset := [2]float32{0, 0}
	if amount != 0 {
		offset = [2]float32{amount / float32(w), amount / float32(h)}
	}
	shader.SetUniform("uOffset", offset)
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func vignetteGPU(driver Driver, src Texture, dst Framebuffer, radius, softness float32) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	shader, err := getEffectShader(driver, "vignette")
	if err != nil {
		return false
	}
	if radius <= 0 {
		radius = 0.7
	}
	if softness <= 0 {
		softness = 0.3
	}
	shader.SetUniform("uRadius", radius)
	shader.SetUniform("uSoftness", softness)
	w, h := dst.Size()
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func pixelateGPU(driver Driver, src Texture, dst Framebuffer, size float32) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	if size <= 1 {
		return copyTextureGPU(driver, src, dst)
	}
	w, h := dst.Size()
	if w <= 0 || h <= 0 {
		return false
	}
	shader, err := getEffectShader(driver, "pixelate")
	if err != nil {
		return false
	}
	shader.SetUniform("uPixelSize", size)
	shader.SetUniform("uResolution", [2]float32{float32(w), float32(h)})
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func colorGradeGPU(driver Driver, src Texture, dst Framebuffer, brightness, contrast, saturation, hue float32) bool {
	if driver == nil || src == nil || dst == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	shader, err := getEffectShader(driver, "colorgrade")
	if err != nil {
		return false
	}
	shader.SetUniform("uBrightness", brightness)
	shader.SetUniform("uContrast", contrast)
	shader.SetUniform("uSaturation", saturation)
	shader.SetUniform("uHue", hue)
	w, h := dst.Size()
	verts, inds := effectQuad(w, h, 0, 0)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func customEffectGPU(driver Driver, src Texture, dst Framebuffer, shader Shader, uniforms map[string]any) bool {
	if driver == nil || src == nil || dst == nil || shader == nil {
		return false
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return false
		}
	}
	shader.SetUniform("uTexture", 0)
	shader.SetUniform("uTransform", Identity())
	for k, v := range uniforms {
		shader.SetUniform(k, v)
	}
	verts, inds := effectQuadSize(dst)
	drawEffect(driver, shader, src, dst, BlendNone, verts, inds)
	return true
}

func effectQuadSize(dst Framebuffer) ([]float32, []uint16) {
	if dst == nil {
		return nil, nil
	}
	w, h := dst.Size()
	return effectQuad(w, h, 0, 0)
}

func ensureGPUFallback(driver Driver, src Texture, dst Framebuffer) (Texture, bool, bool, error) {
	if driver == nil || src == nil || dst == nil {
		return nil, false, false, ErrUnsupported
	}
	if driver.Backend() == BackendOpenGL {
		if _, ok := dst.(*openglFramebuffer); !ok {
			return nil, false, false, ErrUnsupported
		}
	}
	tex, temp, ok := ensureEffectTexture(driver, src)
	if !ok {
		return nil, false, false, errors.New("gpu: texture conversion failed")
	}
	return tex, temp, true, nil
}
