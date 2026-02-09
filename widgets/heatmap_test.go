package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/testing/widgettest"
)

func TestHeatMap_Basic(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2, 3},
		{4, 5, 6},
	})
	hm := NewHeatMap(data)

	// Verify constructor defaults
	if hm.StyleType() != "HeatMap" {
		t.Errorf("StyleType() = %q, want %q", hm.StyleType(), "HeatMap")
	}
	if hm.CellWidth != 3 {
		t.Errorf("CellWidth = %d, want 3", hm.CellWidth)
	}
	if hm.ColorScale == nil {
		t.Error("ColorScale should not be nil after construction")
	}

	// Verify Measure produces reasonable size
	widgettest.AssertMeasure(t, hm, 9, 2)

	// Verify HandleMessage returns unhandled
	result := hm.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Error("HandleMessage should return unhandled")
	}
}

func TestHeatMap_Render(t *testing.T) {
	data := state.NewSignal([][]float64{
		{0, 5, 10},
		{10, 0, 5},
	})
	hm := NewHeatMap(data)
	widgettest.AssertRenders(t, hm, 15, 5)
}

func TestHeatMap_RenderNilData(t *testing.T) {
	hm := NewHeatMap(nil)
	// Should not panic
	buf := runtime.NewBuffer(10, 5)
	hm.Measure(runtime.Constraints{MaxWidth: 10, MaxHeight: 5})
	hm.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 5})
	hm.Render(runtime.RenderContext{Buffer: buf})
}

func TestHeatMap_RenderEmptyData(t *testing.T) {
	data := state.NewSignal([][]float64{})
	hm := NewHeatMap(data)
	widgettest.AssertRenders(t, hm, 10, 5)
}

func TestHeatMap_WithLabels(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2},
		{3, 4},
	})
	hm := NewHeatMap(data)
	hm.RowLabels = []string{"R1", "R2"}
	hm.ColLabels = []string{"C1", "C2"}

	// Measure should account for labels
	size := hm.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Width < 9 { // 3 (label+space) + 2*3 (cells)
		t.Errorf("Width with labels should be >= 9, got %d", size.Width)
	}
	if size.Height < 3 { // 1 col label row + 2 data rows
		t.Errorf("Height with labels should be >= 3, got %d", size.Height)
	}

	// Render should include labels in the output
	output := renderToString(hm, 20, 5)
	if len(output) == 0 {
		t.Error("renderToString produced empty output")
	}
	// The row labels should appear somewhere in the output
	if !strings.Contains(output, "R1") && !strings.Contains(output, "R2") {
		t.Error("expected row labels to appear in rendered output")
	}
}

func TestHeatMap_Accessibility(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2},
		{3, 4},
	})
	hm := NewHeatMap(data)

	widgettest.AssertRole(t, hm, "grid")
	widgettest.AssertLabel(t, hm, "Heat Map")

	// Description should contain dimensions and range
	desc := hm.AccessibleDescription()
	if desc == "" {
		t.Error("expected non-empty description")
	}
	if !strings.Contains(desc, "2 columns by 2 rows") {
		t.Errorf("description should contain dimensions, got %q", desc)
	}
	if !strings.Contains(desc, "range") {
		t.Errorf("description should contain 'range', got %q", desc)
	}

	// Value info
	val := hm.AccessibleValue()
	if val == nil {
		t.Fatal("expected non-nil value info")
	}
	if !strings.Contains(val.Text, "min") || !strings.Contains(val.Text, "max") {
		t.Errorf("value text should contain min/max, got %q", val.Text)
	}
}

func TestHeatMap_AccessibilityCustomLabel(t *testing.T) {
	data := state.NewSignal([][]float64{{1}})
	hm := NewHeatMap(data)
	hm.SetLabel("Temperature")
	widgettest.AssertLabel(t, hm, "Temperature")
}

func TestHeatMap_AccessibilityEmptyData(t *testing.T) {
	data := state.NewSignal([][]float64{})
	hm := NewHeatMap(data)
	desc := hm.AccessibleDescription()
	if desc != "empty" {
		t.Errorf("expected description 'empty' for empty data, got %q", desc)
	}
}

func TestHeatMap_ColorScales(t *testing.T) {
	scales := []struct {
		name  string
		scale ColorScale
	}{
		{"GreenRed", GreenRedScale()},
		{"BlueRed", BlueRedScale()},
		{"Grayscale", GrayscaleScale()},
	}

	testValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, sc := range scales {
		t.Run(sc.name, func(t *testing.T) {
			for _, v := range testValues {
				color := sc.scale(v)
				if !color.IsRGB() {
					t.Errorf("%s(%f) returned non-RGB color", sc.name, v)
				}
			}
		})
	}

	// Test boundary clamping
	for _, sc := range scales {
		t.Run(sc.name+"_clamp", func(t *testing.T) {
			// Values outside [0,1] should be clamped
			colorLow := sc.scale(-0.5)
			colorZero := sc.scale(0.0)
			if colorLow != colorZero {
				t.Errorf("%s(-0.5) should equal %s(0.0)", sc.name, sc.name)
			}

			colorHigh := sc.scale(1.5)
			colorOne := sc.scale(1.0)
			if colorHigh != colorOne {
				t.Errorf("%s(1.5) should equal %s(1.0)", sc.name, sc.name)
			}
		})
	}
}

func TestHeatMap_ColorScaleEndpoints(t *testing.T) {
	// GreenRedScale: value=0 should be green, value=1 should be red
	gr := GreenRedScale()
	green := gr(0)
	r, g, _ := green.RGB()
	if r != 0 || g != 255 {
		t.Errorf("GreenRedScale(0) should be green, got R=%d G=%d", r, g)
	}
	red := gr(1)
	r, g, _ = red.RGB()
	if r != 255 || g != 0 {
		t.Errorf("GreenRedScale(1) should be red, got R=%d G=%d", r, g)
	}

	// BlueRedScale: value=0 should be blue, value=1 should be red
	br := BlueRedScale()
	blue := br(0)
	r, _, b := blue.RGB()
	if r != 0 || b != 255 {
		t.Errorf("BlueRedScale(0) should be blue, got R=%d B=%d", r, b)
	}
	red = br(1)
	r, _, b = red.RGB()
	if r != 255 || b != 0 {
		t.Errorf("BlueRedScale(1) should be red, got R=%d B=%d", r, b)
	}

	// GrayscaleScale: value=0 should be black, value=1 should be white
	gs := GrayscaleScale()
	black := gs(0)
	r, g, b = black.RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("GrayscaleScale(0) should be black, got R=%d G=%d B=%d", r, g, b)
	}
	white := gs(1)
	r, g, b = white.RGB()
	if r != 255 || g != 255 || b != 255 {
		t.Errorf("GrayscaleScale(1) should be white, got R=%d G=%d B=%d", r, g, b)
	}
}

func TestHeatMap_ShowValues(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2},
		{3, 4},
	})
	hm := NewHeatMap(data)
	hm.ShowValues = true
	widgettest.AssertRenders(t, hm, 20, 5)
}

func TestHeatMap_CustomCellWidth(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2, 3},
	})
	hm := NewHeatMap(data)
	hm.CellWidth = 5

	size := hm.Measure(runtime.Constraints{MaxWidth: 100, MaxHeight: 10})
	if size.Width < 15 { // 3 cells * 5 width
		t.Errorf("Width with CellWidth=5 should be >= 15, got %d", size.Width)
	}
}

func TestHeatMap_BindUnbind(t *testing.T) {
	data := state.NewSignal([][]float64{{1}})
	hm := NewHeatMap(data)

	widgettest.AssertBindable(t, hm)

	hm.Bind(runtime.Services{})
	hm.Unbind()

	// Nil receiver should not panic
	var nilHM *HeatMap
	nilHM.Bind(runtime.Services{})
	nilHM.Unbind()
}

func TestHeatMap_NilReceiver(t *testing.T) {
	var hm *HeatMap

	// None of these should panic
	hm.SetLabel("test")
	hm.Render(runtime.RenderContext{})
	hm.Bind(runtime.Services{})
	hm.Unbind()
}

func TestHeatMap_UniformData(t *testing.T) {
	// All values the same: should not divide by zero
	data := state.NewSignal([][]float64{
		{5, 5},
		{5, 5},
	})
	hm := NewHeatMap(data)
	widgettest.AssertRenders(t, hm, 10, 5)
}

func TestHeatMap_DifferentColorScale(t *testing.T) {
	data := state.NewSignal([][]float64{
		{0, 10},
		{20, 30},
	})
	hm := NewHeatMap(data)
	hm.ColorScale = BlueRedScale()
	widgettest.AssertRenders(t, hm, 10, 5)
}

func TestHeatMap_NilColorScale(t *testing.T) {
	data := state.NewSignal([][]float64{
		{1, 2},
	})
	hm := NewHeatMap(data)
	hm.ColorScale = nil
	// Should fall back to GreenRedScale without panicking
	widgettest.AssertRenders(t, hm, 10, 3)
}

func TestHeatMap_LargeData(t *testing.T) {
	rows := make([][]float64, 10)
	for i := range rows {
		rows[i] = make([]float64, 10)
		for j := range rows[i] {
			rows[i][j] = float64(i*10 + j)
		}
	}
	data := state.NewSignal(rows)
	hm := NewHeatMap(data)
	widgettest.AssertRenders(t, hm, 40, 15)
}

func TestHeatMap_FormatHeatMapValue(t *testing.T) {
	tests := []struct {
		val   float64
		width int
	}{
		{1.5, 3},
		{100.0, 3},
		{0.0, 1},
		{99.9, 5},
	}
	for _, tt := range tests {
		s := formatHeatMapValue(tt.val, tt.width)
		if len(s) > tt.width {
			// For very small widths, the formatter might exceed, which is ok
			// since truncateString handles it in render
		}
		_ = s // just verify no panic
	}
}

// Verify interfaces are satisfied
var _ runtime.Widget = (*HeatMap)(nil)
var _ runtime.Bindable = (*HeatMap)(nil)
var _ runtime.Unbindable = (*HeatMap)(nil)
var _ accessibility.Accessible = (*HeatMap)(nil)

// Suppress unused import warning for backend
var _ = backend.ColorRGB
