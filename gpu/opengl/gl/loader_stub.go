//go:build js

package gl

import (
	"errors"
	"unsafe"
)

// Stub implementation for WASM builds where purego is not available.

var (
	initErr  = errors.New("OpenGL not supported on WASM")
	lib      = uintptr(0)
	procAddr = func(name string) uintptr { return 0 }
)

// Init returns an error on WASM since OpenGL is not supported.
func Init() error {
	return initErr
}

func registerFunctions() error {
	return initErr
}

func register(target interface{}, name string) error {
	return initErr
}

func openLibrary(name string) (uintptr, error) {
	return 0, initErr
}

func closeLibrary(handle uintptr) error {
	return nil
}

func initProcAddress(handle uintptr) error {
	return initErr
}

func libGLName() string {
	return ""
}

// Stub function pointers - these are no-ops on WASM
var (
	ClearColor               = func(r, g, b, a float32) {}
	Clear                    = func(mask uint32) {}
	Viewport                 = func(x, y, width, height int32) {}
	BindTexture              = func(target uint32, texture uint32) {}
	DeleteTextures           = func(n int32, textures *uint32) {}
	GenTextures              = func(n int32, textures *uint32) {}
	TexImage2D               = func(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels *uint8) {}
	TexSubImage2D            = func(target uint32, level int32, xoffset int32, yoffset int32, width int32, height int32, format uint32, xtype uint32, pixels *uint8) {}
	TexParameteri            = func(target uint32, pname uint32, param int32) {}
	ActiveTexture            = func(texture uint32) {}
	BindBuffer               = func(target uint32, buffer uint32) {}
	DeleteBuffers            = func(n int32, buffers *uint32) {}
	GenBuffers               = func(n int32, buffers *uint32) {}
	BufferData               = func(target uint32, size int, data unsafe.Pointer, usage uint32) {}
	BufferSubData            = func(target uint32, offset int, size int, data unsafe.Pointer) {}
	BindFramebuffer          = func(target uint32, framebuffer uint32) {}
	DeleteFramebuffers       = func(n int32, framebuffers *uint32) {}
	GenFramebuffers          = func(n int32, framebuffers *uint32) {}
	FramebufferTexture2D     = func(target uint32, attachment uint32, textarget uint32, texture uint32, level int32) {}
	CheckFramebufferStatus   = func(target uint32) uint32 { return 0 }
	CreateShader             = func(xtype uint32) uint32 { return 0 }
	DeleteShader             = func(shader uint32) {}
	ShaderSource             = func(shader uint32, count int32, xstring **uint8, length *int32) {}
	CompileShader            = func(shader uint32) {}
	GetShaderiv              = func(shader uint32, pname uint32, params *int32) {}
	GetShaderInfoLog         = func(shader uint32, bufSize int32, length *int32, infoLog *uint8) {}
	CreateProgram            = func() uint32 { return 0 }
	DeleteProgram            = func(program uint32) {}
	AttachShader             = func(program uint32, shader uint32) {}
	LinkProgram              = func(program uint32) {}
	GetProgramiv             = func(program uint32, pname uint32, params *int32) {}
	GetProgramInfoLog        = func(program uint32, bufSize int32, length *int32, infoLog *uint8) {}
	UseProgram               = func(program uint32) {}
	GetUniformLocation       = func(program uint32, name *uint8) int32 { return -1 }
	Uniform1i                = func(location int32, v0 int32) {}
	Uniform1f                = func(location int32, v0 float32) {}
	Uniform2f                = func(location int32, v0 float32, v1 float32) {}
	Uniform3f                = func(location int32, v0 float32, v1 float32, v2 float32) {}
	Uniform4f                = func(location int32, v0 float32, v1 float32, v2 float32, v3 float32) {}
	UniformMatrix3fv         = func(location int32, count int32, transpose uint8, value *float32) {}
	UniformMatrix4fv         = func(location int32, count int32, transpose uint8, value *float32) {}
	GetAttribLocation        = func(program uint32, name *uint8) int32 { return -1 }
	EnableVertexAttribArray  = func(index uint32) {}
	VertexAttribPointer      = func(index uint32, size int32, xtype uint32, normalized uint8, stride int32, pointer unsafe.Pointer) {}
	DisableVertexAttribArray = func(index uint32) {}
	DrawElements             = func(mode uint32, count int32, xtype uint32, indices unsafe.Pointer) {}
	Enable                   = func(cap uint32) {}
	Disable                  = func(cap uint32) {}
	BlendFunc                = func(sfactor uint32, dfactor uint32) {}
	BlendFuncSeparate        = func(srcRGB uint32, dstRGB uint32, srcAlpha uint32, dstAlpha uint32) {}
	BlendEquationSeparate    = func(modeRGB uint32, modeAlpha uint32) {}
	ReadPixels               = func(x int32, y int32, width int32, height int32, format uint32, xtype uint32, pixels *uint8) {}
	GetIntegerv              = func(pname uint32, data *int32) {}
	GetString                = func(name uint32) *uint8 { return nil }
	GetError                 = func() uint32 { return 0 }
)
