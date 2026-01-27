package gl

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	initOnce sync.Once
	initErr  error
	lib      uintptr
	procAddr func(name string) uintptr
)

// Init loads OpenGL symbols.
func Init() error {
	initOnce.Do(func() {
		libName := libGLName()
		h, err := openLibrary(libName)
		if err != nil {
			initErr = err
			return
		}
		lib = h
		if err := initProcAddress(lib); err != nil {
			initErr = err
			return
		}
		if err := registerFunctions(); err != nil {
			initErr = err
			return
		}
	})
	return initErr
}

func registerFunctions() error {
	if lib == 0 {
		return errors.New("gl library not loaded")
	}
	if err := register(&ClearColor, "glClearColor"); err != nil {
		return err
	}
	if err := register(&Clear, "glClear"); err != nil {
		return err
	}
	if err := register(&Viewport, "glViewport"); err != nil {
		return err
	}
	if err := register(&GetString, "glGetString"); err != nil {
		return err
	}
	if err := register(&GenTextures, "glGenTextures"); err != nil {
		return err
	}
	if err := register(&BindTexture, "glBindTexture"); err != nil {
		return err
	}
	if err := register(&TexParameteri, "glTexParameteri"); err != nil {
		return err
	}
	if err := register(&TexImage2D, "glTexImage2D"); err != nil {
		return err
	}
	if err := register(&TexSubImage2D, "glTexSubImage2D"); err != nil {
		return err
	}
	if err := register(&DeleteTextures, "glDeleteTextures"); err != nil {
		return err
	}
	if err := register(&ActiveTexture, "glActiveTexture"); err != nil {
		return err
	}
	if err := register(&GenFramebuffers, "glGenFramebuffers"); err != nil {
		return err
	}
	if err := register(&BindFramebuffer, "glBindFramebuffer"); err != nil {
		return err
	}
	if err := register(&FramebufferTexture2D, "glFramebufferTexture2D"); err != nil {
		return err
	}
	if err := register(&CheckFramebufferStatus, "glCheckFramebufferStatus"); err != nil {
		return err
	}
	if err := register(&DeleteFramebuffers, "glDeleteFramebuffers"); err != nil {
		return err
	}
	if err := register(&ReadPixels, "glReadPixels"); err != nil {
		return err
	}
	if err := register(&CreateShader, "glCreateShader"); err != nil {
		return err
	}
	if err := register(&ShaderSource, "glShaderSource"); err != nil {
		return err
	}
	if err := register(&CompileShader, "glCompileShader"); err != nil {
		return err
	}
	if err := register(&GetShaderiv, "glGetShaderiv"); err != nil {
		return err
	}
	if err := register(&GetShaderInfoLog, "glGetShaderInfoLog"); err != nil {
		return err
	}
	if err := register(&DeleteShader, "glDeleteShader"); err != nil {
		return err
	}
	if err := register(&CreateProgram, "glCreateProgram"); err != nil {
		return err
	}
	if err := register(&AttachShader, "glAttachShader"); err != nil {
		return err
	}
	if err := register(&LinkProgram, "glLinkProgram"); err != nil {
		return err
	}
	if err := register(&GetProgramiv, "glGetProgramiv"); err != nil {
		return err
	}
	if err := register(&GetProgramInfoLog, "glGetProgramInfoLog"); err != nil {
		return err
	}
	if err := register(&UseProgram, "glUseProgram"); err != nil {
		return err
	}
	if err := register(&DeleteProgram, "glDeleteProgram"); err != nil {
		return err
	}
	if err := register(&GetUniformLocation, "glGetUniformLocation"); err != nil {
		return err
	}
	if err := register(&Uniform1i, "glUniform1i"); err != nil {
		return err
	}
	if err := register(&Uniform1f, "glUniform1f"); err != nil {
		return err
	}
	if err := register(&Uniform2f, "glUniform2f"); err != nil {
		return err
	}
	if err := register(&Uniform4f, "glUniform4f"); err != nil {
		return err
	}
	if err := register(&UniformMatrix3fv, "glUniformMatrix3fv"); err != nil {
		return err
	}
	if err := register(&GenBuffers, "glGenBuffers"); err != nil {
		return err
	}
	if err := register(&BindBuffer, "glBindBuffer"); err != nil {
		return err
	}
	if err := register(&BufferData, "glBufferData"); err != nil {
		return err
	}
	if err := register(&BufferSubData, "glBufferSubData"); err != nil {
		return err
	}
	if err := register(&DeleteBuffers, "glDeleteBuffers"); err != nil {
		return err
	}
	if err := register(&GenVertexArrays, "glGenVertexArrays"); err != nil {
		return err
	}
	if err := register(&BindVertexArray, "glBindVertexArray"); err != nil {
		return err
	}
	if err := register(&DeleteVertexArrays, "glDeleteVertexArrays"); err != nil {
		return err
	}
	if err := register(&EnableVertexAttribArray, "glEnableVertexAttribArray"); err != nil {
		return err
	}
	if err := register(&VertexAttribPointer, "glVertexAttribPointer"); err != nil {
		return err
	}
	if err := register(&DrawElements, "glDrawElements"); err != nil {
		return err
	}
	if err := register(&Enable, "glEnable"); err != nil {
		return err
	}
	if err := register(&Disable, "glDisable"); err != nil {
		return err
	}
	if err := register(&BlendFunc, "glBlendFunc"); err != nil {
		return err
	}
	if err := register(&Scissor, "glScissor"); err != nil {
		return err
	}
	return nil
}

func register(target any, name string) error {
	if procAddr == nil {
		return errors.New("gl proc loader missing")
	}
	addr := procAddr(name)
	if addr == 0 {
		return errors.New("gl: missing symbol " + name)
	}
	purego.RegisterFunc(target, addr)
	return nil
}

// OpenGL function pointers.
var (
	ClearColor              func(r, g, b, a float32)
	Clear                   func(mask uint32)
	Viewport                func(x, y int32, width, height int32)
	GetString               func(name uint32) unsafe.Pointer
	GenTextures             func(n int32, textures *uint32)
	BindTexture             func(target uint32, texture uint32)
	TexParameteri           func(target uint32, pname uint32, param int32)
	TexImage2D              func(target uint32, level int32, internalFormat int32, width, height int32, border int32, format uint32, typ uint32, pixels unsafe.Pointer)
	TexSubImage2D           func(target uint32, level int32, xoffset, yoffset, width, height int32, format uint32, typ uint32, pixels unsafe.Pointer)
	DeleteTextures          func(n int32, textures *uint32)
	ActiveTexture           func(texture uint32)
	GenFramebuffers         func(n int32, framebuffers *uint32)
	BindFramebuffer         func(target uint32, framebuffer uint32)
	FramebufferTexture2D    func(target uint32, attachment uint32, textarget uint32, texture uint32, level int32)
	CheckFramebufferStatus  func(target uint32) uint32
	DeleteFramebuffers      func(n int32, framebuffers *uint32)
	ReadPixels              func(x, y int32, width, height int32, format uint32, typ uint32, pixels unsafe.Pointer)
	CreateShader            func(shaderType uint32) uint32
	ShaderSource            func(shader uint32, count int32, sources **uint8, lengths *int32)
	CompileShader           func(shader uint32)
	GetShaderiv             func(shader uint32, pname uint32, params *int32)
	GetShaderInfoLog        func(shader uint32, maxLength int32, length *int32, infoLog *uint8)
	DeleteShader            func(shader uint32)
	CreateProgram           func() uint32
	AttachShader            func(program uint32, shader uint32)
	LinkProgram             func(program uint32)
	GetProgramiv            func(program uint32, pname uint32, params *int32)
	GetProgramInfoLog       func(program uint32, maxLength int32, length *int32, infoLog *uint8)
	UseProgram              func(program uint32)
	DeleteProgram           func(program uint32)
	GetUniformLocation      func(program uint32, name *uint8) int32
	Uniform1i               func(location int32, v0 int32)
	Uniform1f               func(location int32, v0 float32)
	Uniform2f               func(location int32, v0, v1 float32)
	Uniform4f               func(location int32, v0, v1, v2, v3 float32)
	UniformMatrix3fv        func(location int32, count int32, transpose uint8, value *float32)
	GenBuffers              func(n int32, buffers *uint32)
	BindBuffer              func(target uint32, buffer uint32)
	BufferData              func(target uint32, size uintptr, data unsafe.Pointer, usage uint32)
	BufferSubData           func(target uint32, offset uintptr, size uintptr, data unsafe.Pointer)
	DeleteBuffers           func(n int32, buffers *uint32)
	GenVertexArrays         func(n int32, arrays *uint32)
	BindVertexArray         func(array uint32)
	DeleteVertexArrays      func(n int32, arrays *uint32)
	EnableVertexAttribArray func(index uint32)
	VertexAttribPointer     func(index uint32, size int32, typ uint32, normalized uint8, stride int32, pointer unsafe.Pointer)
	DrawElements            func(mode uint32, count int32, typ uint32, indices unsafe.Pointer)
	Enable                  func(cap uint32)
	Disable                 func(cap uint32)
	BlendFunc               func(sfactor uint32, dfactor uint32)
	Scissor                 func(x, y int32, width, height int32)
)

// Ptr converts a slice or pointer to an unsafe pointer.
func Ptr(data any) unsafe.Pointer {
	switch v := data.(type) {
	case unsafe.Pointer:
		return v
	case []byte:
		if len(v) == 0 {
			return unsafe.Pointer(uintptr(0))
		}
		return unsafe.Pointer(&v[0])
	case []float32:
		if len(v) == 0 {
			return unsafe.Pointer(uintptr(0))
		}
		return unsafe.Pointer(&v[0])
	case []uint16:
		if len(v) == 0 {
			return unsafe.Pointer(uintptr(0))
		}
		return unsafe.Pointer(&v[0])
	default:
		return unsafe.Pointer(uintptr(0))
	}
}
