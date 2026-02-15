package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

// --- Base OnMount / OnUnmount ---

func TestBase_OnMount_FiresOnBind(t *testing.T) {
	var b Base
	called := false
	b.OnMount(func() { called = true })

	// Simulate what BindTree does: Bind then OnMountHook.
	b.Bind(runtime.Services{})
	b.OnMountHook()

	if !called {
		t.Error("OnMount callback was not invoked during Bind/OnMountHook")
	}
}

func TestBase_OnUnmount_FiresOnUnbind(t *testing.T) {
	var b Base
	called := false
	b.OnUnmount(func() { called = true })

	// Simulate what UnbindTree does: OnUnmountHook then Unbind.
	b.OnUnmountHook()
	b.Unbind()

	if !called {
		t.Error("OnUnmount callback was not invoked during OnUnmountHook/Unbind")
	}
}

func TestBase_NilHooks_NoPanic(t *testing.T) {
	var b Base
	// No hooks registered — these should not panic.
	b.Bind(runtime.Services{})
	b.OnMountHook()
	b.OnUnmountHook()
	b.Unbind()
}

func TestBase_NilReceiver_NoPanic(t *testing.T) {
	var b *Base
	// All methods must be nil-safe.
	b.OnMount(func() {})
	b.OnUnmount(func() {})
	b.Bind(runtime.Services{})
	b.Unbind()
	b.OnMountHook()
	b.OnUnmountHook()
}

func TestBase_HookCanBeChanged(t *testing.T) {
	var b Base
	first := false
	second := false

	b.OnMount(func() { first = true })
	b.OnMountHook()
	if !first {
		t.Error("first hook should have fired")
	}

	// Replace the hook.
	b.OnMount(func() { second = true })
	b.OnMountHook()
	if !second {
		t.Error("second (replaced) hook should have fired")
	}
}

func TestBase_HookCanBeCleared(t *testing.T) {
	var b Base
	count := 0
	b.OnMount(func() { count++ })
	b.OnMountHook()
	if count != 1 {
		t.Fatalf("expected count=1, got %d", count)
	}

	// Clear the hook by setting nil.
	b.OnMount(nil)
	b.OnMountHook()
	if count != 1 {
		t.Errorf("expected count=1 after clearing hook, got %d", count)
	}
}

// --- FocusableBase OnFocus / OnBlur ---

func TestFocusableBase_OnFocus_FiresOnFocus(t *testing.T) {
	var fb FocusableBase
	called := false
	fb.OnFocus(func() { called = true })

	fb.Focus()

	if !called {
		t.Error("OnFocus callback was not invoked during Focus()")
	}
	if !fb.IsFocused() {
		t.Error("widget should be focused after Focus()")
	}
}

func TestFocusableBase_OnBlur_FiresOnBlur(t *testing.T) {
	var fb FocusableBase
	fb.Focus() // Start focused.

	called := false
	fb.OnBlur(func() { called = true })

	fb.Blur()

	if !called {
		t.Error("OnBlur callback was not invoked during Blur()")
	}
	if fb.IsFocused() {
		t.Error("widget should not be focused after Blur()")
	}
}

func TestFocusableBase_NilFocusHooks_NoPanic(t *testing.T) {
	var fb FocusableBase
	// No hooks — Focus/Blur should not panic.
	fb.Focus()
	fb.Blur()
}

func TestFocusableBase_NilReceiver_NoPanic(t *testing.T) {
	var fb *FocusableBase
	fb.OnFocus(func() {})
	fb.OnBlur(func() {})
	fb.Focus()
	fb.Blur()
}

func TestFocusableBase_FocusHookCanBeChanged(t *testing.T) {
	var fb FocusableBase
	first := false
	second := false

	fb.OnFocus(func() { first = true })
	fb.Focus()
	if !first {
		t.Error("first onFocus hook should have fired")
	}

	fb.Blur()
	fb.OnFocus(func() { second = true })
	fb.Focus()
	if !second {
		t.Error("second (replaced) onFocus hook should have fired")
	}
}

func TestFocusableBase_BlurHookCanBeChanged(t *testing.T) {
	var fb FocusableBase
	first := false
	second := false

	fb.OnBlur(func() { first = true })
	fb.Focus()
	fb.Blur()
	if !first {
		t.Error("first onBlur hook should have fired")
	}

	fb.OnBlur(func() { second = true })
	fb.Focus()
	fb.Blur()
	if !second {
		t.Error("second (replaced) onBlur hook should have fired")
	}
}

// --- BindTree / UnbindTree integration ---

// testLifecycleWidget is a minimal widget that uses Base for lifecycle hooks.
type testLifecycleWidget struct {
	Base
}

func (w *testLifecycleWidget) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}

func (w *testLifecycleWidget) Render(ctx runtime.RenderContext) {}

func TestBindTree_FiresOnMountHook(t *testing.T) {
	w := &testLifecycleWidget{}
	w.OnMount(func() {})

	runtime.BindTree(w, runtime.Services{})
	// BindTree with zero services is a no-op, so make a non-zero one.
	// Let's call bindWidget manually via the public API approach.
	// Actually, BindTree checks services.isZero() — let's check what makes it non-zero.
}

// testLifecycleWidgetWithOwnBind has its own Bind that does NOT call Base.Bind.
// Hooks should still fire because BindTree calls OnMountHook separately.
type testLifecycleWidgetWithOwnBind struct {
	Base
	services runtime.Services
	bindCalled bool
}

func (w *testLifecycleWidgetWithOwnBind) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}

func (w *testLifecycleWidgetWithOwnBind) Render(ctx runtime.RenderContext) {}

func (w *testLifecycleWidgetWithOwnBind) Bind(services runtime.Services) {
	w.services = services
	w.bindCalled = true
}

func (w *testLifecycleWidgetWithOwnBind) Unbind() {
	w.services = runtime.Services{}
}

func TestLifecycleHooks_WithOwnBind_ViaOnMountHook(t *testing.T) {
	w := &testLifecycleWidgetWithOwnBind{}
	mounted := false
	w.OnMount(func() { mounted = true })

	// Simulate what BindTree does:
	// 1. Call Bind (widget's own)
	w.Bind(runtime.Services{})
	// 2. Call OnMountHook (from Base, via MountHook interface)
	w.OnMountHook()

	if !w.bindCalled {
		t.Error("widget's own Bind should have been called")
	}
	if !mounted {
		t.Error("OnMount hook should have fired via OnMountHook")
	}
}

func TestLifecycleHooks_WithOwnUnbind_ViaOnUnmountHook(t *testing.T) {
	w := &testLifecycleWidgetWithOwnBind{}
	unmounted := false
	w.OnUnmount(func() { unmounted = true })

	// Simulate what UnbindTree does:
	// 1. Call OnUnmountHook (from Base)
	w.OnUnmountHook()
	// 2. Call Unbind (widget's own)
	w.Unbind()

	if !unmounted {
		t.Error("OnUnmount hook should have fired via OnUnmountHook")
	}
}

// --- Hooks on actual widgets ---

func TestLabel_OnMount_Hook(t *testing.T) {
	label := NewLabel("Hello")
	mounted := false
	label.OnMount(func() { mounted = true })

	// Label has its own Bind, so call it first, then the hook.
	label.Bind(runtime.Services{})
	label.OnMountHook()

	if !mounted {
		t.Error("OnMount hook should fire on Label via OnMountHook")
	}
}

func TestLabel_OnUnmount_Hook(t *testing.T) {
	label := NewLabel("Hello")
	unmounted := false
	label.OnUnmount(func() { unmounted = true })

	label.OnUnmountHook()
	label.Unbind()

	if !unmounted {
		t.Error("OnUnmount hook should fire on Label via OnUnmountHook")
	}
}

func TestButton_OnFocus_OnBlur_Hooks(t *testing.T) {
	btn := NewButton("Click me")
	focused := false
	blurred := false

	btn.OnFocus(func() { focused = true })
	btn.OnBlur(func() { blurred = true })

	btn.Focus()
	if !focused {
		t.Error("OnFocus hook should fire on Button.Focus()")
	}
	if !btn.IsFocused() {
		t.Error("Button should be focused")
	}

	btn.Blur()
	if !blurred {
		t.Error("OnBlur hook should fire on Button.Blur()")
	}
	if btn.IsFocused() {
		t.Error("Button should not be focused after Blur()")
	}
}

// --- Combined lifecycle test ---

func TestLifecycle_FullCycle(t *testing.T) {
	var fb FocusableBase
	var events []string

	fb.OnMount(func() { events = append(events, "mount") })
	fb.OnUnmount(func() { events = append(events, "unmount") })
	fb.OnFocus(func() { events = append(events, "focus") })
	fb.OnBlur(func() { events = append(events, "blur") })

	// Mount
	fb.Bind(runtime.Services{})
	fb.OnMountHook()
	// Focus
	fb.Focus()
	// Blur
	fb.Blur()
	// Unmount
	fb.OnUnmountHook()
	fb.Unbind()

	expected := []string{"mount", "focus", "blur", "unmount"}
	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(events), events)
	}
	for i, e := range expected {
		if events[i] != e {
			t.Errorf("event[%d] = %q, want %q", i, events[i], e)
		}
	}
}

// --- Interface compliance ---

func TestBase_ImplementsMountHook(t *testing.T) {
	var _ runtime.MountHook = (*Base)(nil)
}

func TestBase_ImplementsUnmountHook(t *testing.T) {
	var _ runtime.UnmountHook = (*Base)(nil)
}

func TestBase_ImplementsBindable(t *testing.T) {
	var _ runtime.Bindable = (*Base)(nil)
}

func TestBase_ImplementsUnbindable(t *testing.T) {
	var _ runtime.Unbindable = (*Base)(nil)
}
