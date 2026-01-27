//go:build darwin

package gpu

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	objcOnce sync.Once
	objcErr  error

	objcGetClass             func(name *byte) uintptr
	selRegisterName          func(name *byte) uintptr
	objcRetain               func(obj uintptr) uintptr
	objcRelease              func(obj uintptr)
	objcAutoreleasePoolPush  func() uintptr
	objcAutoreleasePoolPop   func(pool uintptr)
	objcMsgSend              func(obj, sel uintptr) uintptr
	objcMsgSend1             func(obj, sel, a1 uintptr) uintptr
	objcMsgSend2             func(obj, sel, a1, a2 uintptr) uintptr
	objcMsgSend3             func(obj, sel, a1, a2, a3 uintptr) uintptr
	objcMsgSend4             func(obj, sel, a1, a2, a3, a4 uintptr) uintptr
	objcMsgSend5             func(obj, sel, a1, a2, a3, a4, a5 uintptr) uintptr
	objcMsgSendScissor       func(obj, sel uintptr, rect mtlScissorRect)
	objcMsgSendViewport      func(obj, sel uintptr, vp mtlViewport)
	objcMsgSendClearColor    func(obj, sel uintptr, color mtlClearColor)
	objcMsgSendReplaceRegion func(obj, sel uintptr, region mtlRegion, level uintptr, bytes unsafe.Pointer, bytesPerRow uintptr)
	objcMsgSendGetBytes      func(obj, sel uintptr, bytes unsafe.Pointer, bytesPerRow uintptr, region mtlRegion, level uintptr)
)

func objcInit() error {
	objcOnce.Do(func() {
		handle, err := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			objcErr = err
			return
		}
		purego.RegisterLibFunc(&objcGetClass, handle, "objc_getClass")
		purego.RegisterLibFunc(&selRegisterName, handle, "sel_registerName")
		purego.RegisterLibFunc(&objcRetain, handle, "objc_retain")
		purego.RegisterLibFunc(&objcRelease, handle, "objc_release")
		purego.RegisterLibFunc(&objcAutoreleasePoolPush, handle, "objc_autoreleasePoolPush")
		purego.RegisterLibFunc(&objcAutoreleasePoolPop, handle, "objc_autoreleasePoolPop")
		purego.RegisterLibFunc(&objcMsgSend, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSend1, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSend2, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSend3, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSend4, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSend5, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSendScissor, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSendViewport, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSendClearColor, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSendReplaceRegion, handle, "objc_msgSend")
		purego.RegisterLibFunc(&objcMsgSendGetBytes, handle, "objc_msgSend")
	})
	return objcErr
}

func objcClass(name string) uintptr {
	if name == "" {
		return 0
	}
	cname := append([]byte(name), 0)
	return objcGetClass((*byte)(unsafe.Pointer(&cname[0])))
}

func sel(name string) uintptr {
	if name == "" {
		return 0
	}
	cname := append([]byte(name), 0)
	return selRegisterName((*byte)(unsafe.Pointer(&cname[0])))
}

func withAutoreleasePool(fn func()) {
	if fn == nil {
		return
	}
	pool := uintptr(0)
	if objcAutoreleasePoolPush != nil {
		pool = objcAutoreleasePoolPush()
	}
	fn()
	if pool != 0 && objcAutoreleasePoolPop != nil {
		objcAutoreleasePoolPop(pool)
	}
}

func nsString(text string) uintptr {
	if text == "" {
		return 0
	}
	cls := objcClass("NSString")
	if cls == 0 {
		return 0
	}
	cstr := append([]byte(text), 0)
	return objcMsgSend1(cls, sel("stringWithUTF8String:"), uintptr(unsafe.Pointer(&cstr[0])))
}

func objcAllocInit(className string) uintptr {
	cls := objcClass(className)
	if cls == 0 {
		return 0
	}
	obj := objcMsgSend(cls, sel("alloc"))
	if obj == 0 {
		return 0
	}
	return objcMsgSend(obj, sel("init"))
}

func objcReleaseObj(obj uintptr) {
	if obj == 0 || objcRelease == nil {
		return
	}
	objcRelease(obj)
}
