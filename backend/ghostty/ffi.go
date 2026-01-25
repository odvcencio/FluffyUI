//go:build linux || darwin || windows

package ghostty

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
)

type ghosttyConfig struct{}
type ghosttyApp struct{}
type ghosttySurface struct{}

type ghosttyLib struct {
	handle uintptr

	init func(uintptr, **byte) int32

	configNew      func() *ghosttyConfig
	configFree     func(*ghosttyConfig)
	configSet      func(*ghosttyConfig, string, string)
	configFinalize func(*ghosttyConfig)

	appNewHeadless func(*ghosttyConfig) *ghosttyApp
	appFree        func(*ghosttyApp)
	appTick        func(*ghosttyApp)

	surfaceNewHeadless func(*ghosttyApp) *ghosttySurface
	surfaceFree        func(*ghosttySurface)

	surfaceSize         func(*ghosttySurface) ghosttySurfaceSize
	surfaceGetSize      func(*ghosttySurface, *int32, *int32)
	surfaceSetCell      func(*ghosttySurface, uint32, uint32, uint32, uint32, uint32, uint8)
	surfaceClear        func(*ghosttySurface)
	surfaceShow         func(*ghosttySurface)
	surfacePoll         func(*ghosttySurface, *ghosttyEvent, int32) int32
	surfaceSetCursorPos func(*ghosttySurface, int32, int32)
	surfaceShowCursor   func(*ghosttySurface)
	surfaceHideCursor   func(*ghosttySurface)

	surfaceKeyTranslationMods func(*ghosttySurface, int32) int32
	surfaceKey                func(*ghosttySurface, ghosttyInputKey) bool
	surfaceKeyIsBinding       func(*ghosttySurface, ghosttyInputKey, *int32) bool
	surfaceText               func(*ghosttySurface, *byte, uintptr)
	surfacePreedit            func(*ghosttySurface, *byte, uintptr)
	surfaceMouseCaptured      func(*ghosttySurface) bool
	surfaceMouseButton        func(*ghosttySurface, int32, int32, int32) bool
	surfaceMousePos           func(*ghosttySurface, float64, float64, int32)
	surfaceMouseScroll        func(*ghosttySurface, float64, float64, int32)
}

var (
	ghosttyOnce sync.Once
	ghosttyLibs *ghosttyLib
	ghosttyErr  error
)

func loadGhosttyLib() (*ghosttyLib, error) {
	ghosttyOnce.Do(func() {
		ghosttyLibs, ghosttyErr = openGhosttyLib()
	})
	return ghosttyLibs, ghosttyErr
}

func openGhosttyLib() (*ghosttyLib, error) {
	libName := defaultLibName()
	if libName == "" {
		return nil, fmt.Errorf("ghostty backend unsupported on %s", runtime.GOOS)
	}
	candidates, err := ghosttyLibCandidates(libName)
	if err != nil {
		return nil, err
	}
	var errs []string
	for _, path := range candidates {
		handle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		lib := &ghosttyLib{handle: handle}
		if err := registerGhosttyFuncs(lib); err != nil {
			_ = purego.Dlclose(handle)
			return nil, err
		}
		if lib.init != nil {
			if rc := lib.init(0, nil); rc != 0 {
				_ = purego.Dlclose(handle)
				return nil, fmt.Errorf("ghostty_init failed with code %d", rc)
			}
		}
		return lib, nil
	}
	return nil, errors.New("libghostty not found; tried: " + strings.Join(errs, "; "))
}

func ghosttyLibCandidates(libName string) ([]string, error) {
	if path := strings.TrimSpace(os.Getenv("FLUFFYUI_GHOSTTY_LIB")); path != "" {
		return []string{path}, nil
	}
	if path := strings.TrimSpace(os.Getenv("GHOSTTY_LIB_PATH")); path != "" {
		return []string{path}, nil
	}
	subdir := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	candidates := []string{
		filepath.Join("backend", "ghostty", "libs", subdir, libName),
		filepath.Join("libs", subdir, libName),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "libs", subdir, libName))
	}
	if srcDir := ghosttySourceDir(); srcDir != "" {
		candidates = append(candidates, filepath.Join(srcDir, "libs", subdir, libName))
	}
	candidates = append(candidates, libName)
	return candidates, nil
}

func ghosttySourceDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Dir(file)
}

func registerGhosttyFuncs(lib *ghosttyLib) error {
	if lib == nil {
		return errors.New("ghostty library handle is nil")
	}
	if err := registerGhosttyFunc(lib, &lib.init, "ghostty_init"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.configNew, "ghostty_config_new"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.configFree, "ghostty_config_free"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.configSet, "ghostty_config_set"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.configFinalize, "ghostty_config_finalize"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.appNewHeadless, "ghostty_app_new_headless"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.appFree, "ghostty_app_free"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.appTick, "ghostty_app_tick"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfaceNewHeadless, "ghostty_surface_new_headless"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfaceFree, "ghostty_surface_free"); err != nil {
		return err
	}
	registerGhosttyFuncOptional(lib, &lib.surfaceSize, "ghostty_surface_size")
	if err := registerGhosttyFunc(lib, &lib.surfaceGetSize, "ghostty_surface_get_size"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfaceSetCell, "ghostty_surface_set_cell"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfaceClear, "ghostty_surface_clear"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfaceShow, "ghostty_surface_show"); err != nil {
		return err
	}
	if err := registerGhosttyFunc(lib, &lib.surfacePoll, "ghostty_surface_poll"); err != nil {
		return err
	}
	registerGhosttyFuncOptional(lib, &lib.surfaceSetCursorPos, "ghostty_surface_set_cursor_pos")
	registerGhosttyFuncOptional(lib, &lib.surfaceShowCursor, "ghostty_surface_show_cursor")
	registerGhosttyFuncOptional(lib, &lib.surfaceHideCursor, "ghostty_surface_hide_cursor")
	registerGhosttyFuncOptional(lib, &lib.surfaceKeyTranslationMods, "ghostty_surface_key_translation_mods")
	registerGhosttyFuncOptional(lib, &lib.surfaceKey, "ghostty_surface_key")
	registerGhosttyFuncOptional(lib, &lib.surfaceKeyIsBinding, "ghostty_surface_key_is_binding")
	registerGhosttyFuncOptional(lib, &lib.surfaceText, "ghostty_surface_text")
	registerGhosttyFuncOptional(lib, &lib.surfacePreedit, "ghostty_surface_preedit")
	registerGhosttyFuncOptional(lib, &lib.surfaceMouseCaptured, "ghostty_surface_mouse_captured")
	registerGhosttyFuncOptional(lib, &lib.surfaceMouseButton, "ghostty_surface_mouse_button")
	registerGhosttyFuncOptional(lib, &lib.surfaceMousePos, "ghostty_surface_mouse_pos")
	registerGhosttyFuncOptional(lib, &lib.surfaceMouseScroll, "ghostty_surface_mouse_scroll")
	return nil
}

func registerGhosttyFunc(lib *ghosttyLib, target any, name string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s: %v", name, r)
		}
	}()
	purego.RegisterLibFunc(target, lib.handle, name)
	return err
}

func registerGhosttyFuncOptional(lib *ghosttyLib, target any, name string) {
	_ = registerGhosttyFunc(lib, target, name)
}
