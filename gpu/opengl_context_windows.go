//go:build windows

package gpu

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	csOwnDC         = 0x0020
	wsOverlapped    = 0x00000000
	wsCaption       = 0x00C00000
	wsSysMenu       = 0x00080000
	wsThickFrame    = 0x00040000
	wsMinimizeBox   = 0x00020000
	wsMaximizeBox   = 0x00010000
	wsOverlappedWin = wsOverlapped | wsCaption | wsSysMenu | wsThickFrame | wsMinimizeBox | wsMaximizeBox

	pfdDrawToWindow  = 0x00000004
	pfdSupportOpenGL = 0x00000020
	pfdDoubleBuffer  = 0x00000001
	pfdTypeRGBA      = 0
	pfdMainPlane     = 0

	wglContextMajorVersionARB = 0x2091
	wglContextMinorVersionARB = 0x2092
	wglContextProfileMaskARB  = 0x9126
	wglContextCoreProfileBit  = 0x00000001
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type pixelFormatDescriptor struct {
	Size           uint16
	Version        uint16
	Flags          uint32
	PixelType      uint8
	ColorBits      uint8
	RedBits        uint8
	RedShift       uint8
	GreenBits      uint8
	GreenShift     uint8
	BlueBits       uint8
	BlueShift      uint8
	AlphaBits      uint8
	AlphaShift     uint8
	AccumBits      uint8
	AccumRedBits   uint8
	AccumGreenBits uint8
	AccumBlueBits  uint8
	AccumAlphaBits uint8
	DepthBits      uint8
	StencilBits    uint8
	AuxBuffers     uint8
	LayerType      uint8
	Reserved       uint8
	LayerMask      uint32
	VisibleMask    uint32
	DamageMask     uint32
}

type osmesaContext struct {
	hwnd      windows.Handle
	hdc       windows.Handle
	hglrc     windows.Handle
	className *uint16
	instance  windows.Handle
	width     int
	height    int
}

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	gdi32                 = windows.NewLazySystemDLL("gdi32.dll")
	opengl32              = windows.NewLazySystemDLL("opengl32.dll")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procRegisterClassEx   = user32.NewProc("RegisterClassExW")
	procCreateWindowEx    = user32.NewProc("CreateWindowExW")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procGetDC             = user32.NewProc("GetDC")
	procReleaseDC         = user32.NewProc("ReleaseDC")
	procUnregisterClass   = user32.NewProc("UnregisterClassW")
	procGetModuleHandle   = kernel32.NewProc("GetModuleHandleW")
	procChoosePixelFormat = gdi32.NewProc("ChoosePixelFormat")
	procSetPixelFormat    = gdi32.NewProc("SetPixelFormat")
	procWglCreateContext  = opengl32.NewProc("wglCreateContext")
	procWglMakeCurrent    = opengl32.NewProc("wglMakeCurrent")
	procWglDeleteContext  = opengl32.NewProc("wglDeleteContext")
	procWglGetProcAddress = opengl32.NewProc("wglGetProcAddress")
)

func newOSMesaContext(width, height int) (*osmesaContext, error) {
	instance, _, _ := procGetModuleHandle.Call(0)
	if instance == 0 {
		return nil, errors.New("wgl: GetModuleHandle failed")
	}
	className := windows.StringToUTF16Ptr("FluffyUIWGL")
	wc := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		Style:     csOwnDC,
		WndProc:   procDefWindowProcW.Addr(),
		Instance:  windows.Handle(instance),
		ClassName: className,
	}
	if ret, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		if err != nil && err != windows.ERROR_CLASS_ALREADY_EXISTS {
			return nil, errors.New("wgl: RegisterClassEx failed")
		}
	}
	hwnd, _, _ := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("FluffyUI WGL"))),
		wsOverlappedWin,
		0, 0, 1, 1,
		0, 0,
		instance,
		0,
	)
	if hwnd == 0 {
		return nil, errors.New("wgl: CreateWindowEx failed")
	}
	hdc, _, _ := procGetDC.Call(hwnd)
	if hdc == 0 {
		procDestroyWindow.Call(hwnd)
		return nil, errors.New("wgl: GetDC failed")
	}
	pfd := pixelFormatDescriptor{
		Size:        uint16(unsafe.Sizeof(pixelFormatDescriptor{})),
		Version:     1,
		Flags:       pfdDrawToWindow | pfdSupportOpenGL | pfdDoubleBuffer,
		PixelType:   pfdTypeRGBA,
		ColorBits:   32,
		DepthBits:   24,
		StencilBits: 8,
		LayerType:   pfdMainPlane,
	}
	pf, _, _ := procChoosePixelFormat.Call(hdc, uintptr(unsafe.Pointer(&pfd)))
	if pf == 0 {
		procReleaseDC.Call(hwnd, hdc)
		procDestroyWindow.Call(hwnd)
		return nil, errors.New("wgl: ChoosePixelFormat failed")
	}
	if ret, _, _ := procSetPixelFormat.Call(hdc, pf, uintptr(unsafe.Pointer(&pfd))); ret == 0 {
		procReleaseDC.Call(hwnd, hdc)
		procDestroyWindow.Call(hwnd)
		return nil, errors.New("wgl: SetPixelFormat failed")
	}
	hglrc, _, _ := procWglCreateContext.Call(hdc)
	if hglrc == 0 {
		procReleaseDC.Call(hwnd, hdc)
		procDestroyWindow.Call(hwnd)
		return nil, errors.New("wgl: CreateContext failed")
	}
	if ret, _, _ := procWglMakeCurrent.Call(hdc, hglrc); ret == 0 {
		procWglDeleteContext.Call(hglrc)
		procReleaseDC.Call(hwnd, hdc)
		procDestroyWindow.Call(hwnd)
		return nil, errors.New("wgl: MakeCurrent failed")
	}
	if modern, ok := createModernContext(windows.Handle(hdc), windows.Handle(hglrc)); ok {
		if ret, _, _ := procWglMakeCurrent.Call(hdc, uintptr(modern)); ret != 0 {
			procWglDeleteContext.Call(hglrc)
			hglrc = uintptr(modern)
		} else {
			procWglMakeCurrent.Call(hdc, hglrc)
			procWglDeleteContext.Call(uintptr(modern))
		}
	}
	return &osmesaContext{
		hwnd:      windows.Handle(hwnd),
		hdc:       windows.Handle(hdc),
		hglrc:     windows.Handle(hglrc),
		className: className,
		instance:  windows.Handle(instance),
		width:     width,
		height:    height,
	}, nil
}

func (c *osmesaContext) Resize(width, height int) error {
	if c == nil {
		return errors.New("wgl: context missing")
	}
	c.width = width
	c.height = height
	return nil
}

func (c *osmesaContext) Buffer() []byte {
	return nil
}

func (c *osmesaContext) Destroy() {
	if c == nil {
		return
	}
	if c.hglrc != 0 {
		procWglMakeCurrent.Call(0, 0)
		procWglDeleteContext.Call(uintptr(c.hglrc))
		c.hglrc = 0
	}
	if c.hdc != 0 && c.hwnd != 0 {
		procReleaseDC.Call(uintptr(c.hwnd), uintptr(c.hdc))
		c.hdc = 0
	}
	if c.hwnd != 0 {
		procDestroyWindow.Call(uintptr(c.hwnd))
		c.hwnd = 0
	}
	if c.className != nil {
		procUnregisterClass.Call(uintptr(unsafe.Pointer(c.className)), uintptr(c.instance))
		c.className = nil
	}
}

func createModernContext(hdc windows.Handle, share windows.Handle) (windows.Handle, bool) {
	addr := wglGetProcAddress("wglCreateContextAttribsARB")
	if addr == 0 || isInvalidWGLAddress(addr) {
		return 0, false
	}
	attribs := []int32{
		wglContextMajorVersionARB, 3,
		wglContextMinorVersionARB, 3,
		wglContextProfileMaskARB, wglContextCoreProfileBit,
		0,
	}
	hglrc, _, _ := syscall.SyscallN(addr, uintptr(hdc), uintptr(share), uintptr(unsafe.Pointer(&attribs[0])))
	if hglrc == 0 {
		return 0, false
	}
	return windows.Handle(hglrc), true
}

func wglGetProcAddress(name string) uintptr {
	if name == "" || procWglGetProcAddress.Find() != nil {
		return 0
	}
	cname, err := windows.BytePtrFromString(name)
	if err != nil {
		return 0
	}
	addr, _, _ := procWglGetProcAddress.Call(uintptr(unsafe.Pointer(cname)))
	return addr
}

func isInvalidWGLAddress(addr uintptr) bool {
	switch addr {
	case 0, 1, 2, 3, 4, ^uintptr(0):
		return true
	default:
		return false
	}
}
