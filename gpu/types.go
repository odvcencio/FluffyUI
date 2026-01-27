package gpu

import "image"

// BlendMode controls how source pixels blend with the destination.
type BlendMode int

const (
	BlendNone BlendMode = iota
	BlendAlpha
	BlendAdditive
)

// ShaderSource contains shader source for multiple backends.
type ShaderSource struct {
	GLSL  GLSLSource
	Metal string
}

// GLSLSource contains vertex and fragment shader source.
type GLSLSource struct {
	Vertex   string
	Fragment string
}

// Texture represents a GPU texture resource.
type Texture interface {
	ID() uint32
	Size() (w, h int)
	Upload(pixels []byte, region image.Rectangle)
	Dispose()
}

// Framebuffer represents a render target.
type Framebuffer interface {
	ID() uint32
	Size() (w, h int)
	Bind()
	Texture() Texture
	Dispose()
}

// Shader represents a compiled shader program.
type Shader interface {
	ID() uint32
	SetUniform(name string, value any)
	Dispose()
}

// DrawCall describes a batched draw operation.
type DrawCall struct {
	Shader   Shader
	Vertices []float32
	Indices  []uint16
	Textures []Texture
	Target   Framebuffer
	Blend    BlendMode
	Scissor  *image.Rectangle
	Layout   *VertexLayout
}

// VertexLayout describes vertex attribute layout in bytes.
type VertexLayout struct {
	Stride     int
	Attributes []VertexAttrib
}

// VertexAttrib defines a single vertex attribute.
type VertexAttrib struct {
	Index      uint32
	Size       int32
	Type       uint32
	Normalized bool
	Stride     int
	Offset     int
}
