//go:build !js

package agent

import (
	"context"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// FocusByID focuses a widget by its ID.
func (a *Agent) FocusByID(id string) error {
	return a.focusByID(id)
}

// ClearFocus clears focus on the current focus scope.
func (a *Agent) ClearFocus() error {
	if a == nil {
		return ErrNoApp
	}
	if a.app != nil {
		return a.app.Call(context.Background(), func(app *runtime.App) error {
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.autoAttach && a.screen == nil {
				a.screen = app.Screen()
			}
			if a.screen != nil {
				if scope := a.screen.FocusScope(); scope != nil {
					scope.ClearFocus()
				}
			}
			app.Invalidate()
			return nil
		})
	}
	if screen := a.Screen(); screen != nil {
		if scope := screen.FocusScope(); scope != nil {
			scope.ClearFocus()
		}
	}
	return nil
}

// WithWidgetByID runs fn on the widget with the given ID.
func (a *Agent) WithWidgetByID(ctx context.Context, id string, fn func(runtime.Widget, accessibility.Accessible) error) error {
	if a == nil {
		return ErrNoApp
	}
	if fn == nil {
		return nil
	}
	if strings.TrimSpace(id) == "" {
		return ErrWidgetNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if a.app != nil {
		return a.app.Call(ctx, func(app *runtime.App) error {
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.autoAttach && a.screen == nil {
				a.screen = app.Screen()
			}
			w, _ := a.widgetByIDLocked(id)
			if w == nil {
				return ErrWidgetNotFound
			}
			return fn(w, accessibleFromWidget(w))
		})
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	w, _ := a.widgetByIDLocked(id)
	if w == nil {
		return ErrWidgetNotFound
	}
	return fn(w, accessibleFromWidget(w))
}

// SelectByID focuses the widget by ID and selects the option by label.
func (a *Agent) SelectByID(id, option string) error {
	if a == nil {
		return ErrNoApp
	}
	if strings.TrimSpace(id) == "" {
		return ErrWidgetNotFound
	}
	if strings.TrimSpace(option) == "" {
		return ErrWidgetNotFound
	}

	_, acc, err := a.focusWidgetByID(id)
	if err != nil {
		return err
	}
	if acc == nil {
		return ErrNotInteractive
	}

	current := accessibleChoice(acc)
	if strings.EqualFold(current, option) {
		return nil
	}
	seen := map[string]bool{current: true}
	for i := 0; i < 100; i++ {
		if err := a.sendKey(terminal.KeyDown, 0); err != nil {
			return err
		}
		a.Tick()
		current = accessibleChoice(acc)
		if strings.EqualFold(current, option) {
			return nil
		}
		if seen[current] {
			break
		}
		seen[current] = true
	}
	return ErrWidgetNotFound
}

// CaptureRegion returns raw text for the requested region.
func (a *Agent) CaptureRegion(x, y, width, height int) string {
	if a == nil {
		return ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.captureRegionLocked(x, y, width, height)
}

// CellAt returns the screen cell at the given position.
func (a *Agent) CellAt(x, y int) (backend.Cell, bool) {
	if a == nil {
		return backend.Cell{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureScreenLocked()
	if a.sim != nil {
		mainc, _, style := a.sim.CaptureCell(x, y)
		if mainc == 0 {
			mainc = ' '
		}
		return backend.Cell{Rune: mainc, Style: style}, true
	}
	if a.screen == nil {
		return backend.Cell{}, false
	}
	if buf := a.screen.Buffer(); buf != nil {
		return buf.Get(x, y), true
	}
	return backend.Cell{}, false
}

// Dimensions returns the current screen size.
func (a *Agent) Dimensions() (width, height int) {
	if a == nil {
		return 0, 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureScreenLocked()
	if a.sim != nil {
		return a.sim.Size()
	}
	if a.screen != nil {
		return a.screen.Size()
	}
	return 0, 0
}

// WaitForWidgetGone waits until a widget with the label disappears.
func (a *Agent) WaitForWidgetGone(label string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.FindByLabel(label) == nil {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForTextGone waits until text disappears.
func (a *Agent) WaitForTextGone(text string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !a.ContainsText(text) {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForFocus waits until the widget with label is focused.
func (a *Agent) WaitForFocus(label string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.IsFocused(label) {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForValue waits until the widget's value matches.
func (a *Agent) WaitForValue(label, value string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		current, err := a.GetValue(label)
		if err == nil && current == value {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForEnabled waits until the widget is enabled.
func (a *Agent) WaitForEnabled(label string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.IsEnabled(label) {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForIdle waits until two consecutive snapshots match.
func (a *Agent) WaitForIdle(timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	prev := a.Snapshot()
	for time.Now().Before(deadline) {
		a.Tick()
		next := a.Snapshot()
		if snapshotsEqual(prev, next) {
			return nil
		}
		prev = next
	}
	return ErrTimeout
}

func (a *Agent) captureRegionLocked(x, y, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	a.ensureScreenLocked()
	if a.sim != nil {
		return a.sim.CaptureRegion(x, y, width, height)
	}
	if a.screen == nil {
		return ""
	}
	buf := a.screen.Buffer()
	if buf == nil {
		return ""
	}
	var lines []string
	for row := 0; row < height; row++ {
		var line strings.Builder
		for col := 0; col < width; col++ {
			cell := buf.Get(x+col, y+row)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			line.WriteRune(r)
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

func (a *Agent) widgetByIDLocked(id string) (runtime.Widget, *runtime.Layer) {
	screen := a.ensureScreenLocked()
	if screen == nil {
		return nil, nil
	}
	for i := screen.LayerCount() - 1; i >= 0; i-- {
		layer := screen.Layer(i)
		if layer == nil || layer.Root == nil {
			continue
		}
		explicitCounts := make(map[string]int)
		if found := findWidgetByID(layer.Root, id, i, []int{0}, explicitCounts); found != nil {
			return found, layer
		}
	}
	return nil, nil
}

func snapshotsEqual(aSnap, bSnap Snapshot) bool {
	if aSnap.Width != bSnap.Width || aSnap.Height != bSnap.Height || aSnap.LayerCount != bSnap.LayerCount {
		return false
	}
	if aSnap.FocusedID != bSnap.FocusedID {
		return false
	}
	if aSnap.Text != bSnap.Text {
		return false
	}
	return widgetsEqual(aSnap.Widgets, bSnap.Widgets)
}

func widgetsEqual(aWidgets, bWidgets []WidgetInfo) bool {
	if len(aWidgets) != len(bWidgets) {
		return false
	}
	for i := range aWidgets {
		a := aWidgets[i]
		b := bWidgets[i]
		if a.ID != b.ID ||
			a.Role != b.Role ||
			a.Label != b.Label ||
			a.Description != b.Description ||
			a.Value != b.Value ||
			a.Focusable != b.Focusable ||
			a.Focused != b.Focused {
			return false
		}
		if a.Bounds != b.Bounds {
			return false
		}
		if a.State != b.State {
			return false
		}
		if len(a.Actions) != len(b.Actions) {
			return false
		}
		for j := range a.Actions {
			if a.Actions[j] != b.Actions[j] {
				return false
			}
		}
		if !widgetsEqual(a.Children, b.Children) {
			return false
		}
	}
	return true
}
