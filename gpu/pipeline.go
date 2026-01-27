package gpu

type gpuPipeline struct {
	solid      Shader
	texture    Shader
	projection Matrix3
}

func newGPUPipeline(driver Driver, width, height int) (*gpuPipeline, error) {
	if driver == nil {
		return nil, ErrUnsupported
	}
	solidSrc, err := LoadShaderSource("solid")
	if err != nil {
		return nil, err
	}
	solidShader, err := driver.NewShader(solidSrc)
	if err != nil {
		return nil, err
	}
	textureSrc, err := LoadShaderSource("texture")
	if err != nil {
		solidShader.Dispose()
		return nil, err
	}
	textureShader, err := driver.NewShader(textureSrc)
	if err != nil {
		solidShader.Dispose()
		return nil, err
	}
	p := &gpuPipeline{solid: solidShader, texture: textureShader}
	p.SetProjection(width, height)
	if p.texture != nil {
		p.texture.SetUniform("uTexture", 0)
	}
	return p, nil
}

func (p *gpuPipeline) SetProjection(width, height int) {
	if p == nil {
		return
	}
	proj := projectionMatrix(width, height)
	p.projection = proj
	if p.solid != nil {
		p.solid.SetUniform("uTransform", proj)
	}
	if p.texture != nil {
		p.texture.SetUniform("uTransform", proj)
	}
}

func (p *gpuPipeline) Dispose() {
	if p == nil {
		return
	}
	if p.solid != nil {
		p.solid.Dispose()
		p.solid = nil
	}
	if p.texture != nil {
		p.texture.Dispose()
		p.texture = nil
	}
}

func projectionMatrix(width, height int) Matrix3 {
	if width <= 0 || height <= 0 {
		return Identity()
	}
	sx := float32(2) / float32(width)
	sy := float32(-2) / float32(height)
	return Matrix3{m: [9]float32{
		sx, 0, -1,
		0, sy, 1,
		0, 0, 1,
	}}
}
