package widgets

import (
	"fmt"
	"math"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
)

// ColorScale maps a normalized value in [0,1] to a color.
type ColorScale func(value float64) backend.Color

// GreenRedScale returns a color scale from green (low) to red (high).
func GreenRedScale() ColorScale {
	return func(value float64) backend.Color {
		if value < 0 {
			value = 0
		}
		if value > 1 {
			value = 1
		}
		r := uint8(value * 255)
		g := uint8((1 - value) * 255)
		return backend.ColorRGB(r, g, 0)
	}
}

// BlueRedScale returns a color scale from blue (low) to red (high).
func BlueRedScale() ColorScale {
	return func(value float64) backend.Color {
		if value < 0 {
			value = 0
		}
		if value > 1 {
			value = 1
		}
		r := uint8(value * 255)
		b := uint8((1 - value) * 255)
		return backend.ColorRGB(r, 0, b)
	}
}

// GrayscaleScale returns a color scale from black (low) to white (high).
func GrayscaleScale() ColorScale {
	return func(value float64) backend.Color {
		if value < 0 {
			value = 0
		}
		if value > 1 {
			value = 1
		}
		v := uint8(value * 255)
		return backend.ColorRGB(v, v, v)
	}
}

// HeatMap displays a 2D grid of values as color-coded cells.
type HeatMap struct {
	Base
	Data       *state.Signal[[][]float64]
	RowLabels  []string
	ColLabels  []string
	ColorScale ColorScale
	CellWidth  int  // width per cell (default 3)
	ShowValues bool // show numeric values in cells
	label      string
	services   runtime.Services
	lastDesc   string // track changes for announcements
}

// NewHeatMap creates a heat map widget.
func NewHeatMap(data *state.Signal[[][]float64]) *HeatMap {
	hm := &HeatMap{
		Data:       data,
		ColorScale: GreenRedScale(),
		CellWidth:  3,
		label:      "Heat Map",
	}
	hm.Base.Role = accessibility.RoleGrid
	hm.Base.RoleDescription = "heat map"
	hm.Base.Live = accessibility.LivePolite
	hm.Base.Relevant = accessibility.RelevantText
	hm.Base.Atomic = true
	hm.syncA11y()
	return hm
}

// SetLabel sets the accessible label.
func (hm *HeatMap) SetLabel(label string) {
	if hm == nil {
		return
	}
	hm.label = label
	hm.syncA11y()
}

// StyleType returns the selector type name.
func (hm *HeatMap) StyleType() string {
	return "HeatMap"
}

// Measure returns desired size.
func (hm *HeatMap) Measure(constraints runtime.Constraints) runtime.Size {
	return hm.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		rows, cols := hm.dataDims()
		cellWidth := hm.CellWidth
		if cellWidth <= 0 {
			cellWidth = 3
		}

		labelOffset := 0
		if len(hm.RowLabels) > 0 {
			for _, l := range hm.RowLabels {
				w := textWidth(l)
				if w > labelOffset {
					labelOffset = w
				}
			}
			labelOffset++ // space after label
		}

		colLabelHeight := 0
		if len(hm.ColLabels) > 0 {
			colLabelHeight = 1
		}

		width := labelOffset + cols*cellWidth
		height := colLabelHeight + rows

		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the heat map.
func (hm *HeatMap) Render(ctx runtime.RenderContext) {
	if hm == nil || hm.Data == nil {
		return
	}
	hm.syncA11y()
	bounds := hm.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	data := hm.Data.Get()
	if len(data) == 0 {
		return
	}

	cellWidth := hm.CellWidth
	if cellWidth <= 0 {
		cellWidth = 3
	}

	colorScale := hm.ColorScale
	if colorScale == nil {
		colorScale = GreenRedScale()
	}

	// Compute label offset
	labelOffset := 0
	if len(hm.RowLabels) > 0 {
		for _, l := range hm.RowLabels {
			w := textWidth(l)
			if w > labelOffset {
				labelOffset = w
			}
		}
		labelOffset++ // space after label
	}

	// Find min/max across all data
	minVal, maxVal := math.Inf(1), math.Inf(-1)
	for _, row := range data {
		for _, v := range row {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	style := mergeBackendStyles(resolveBaseStyle(ctx, hm, backend.DefaultStyle(), false), backend.DefaultStyle())

	// Render column labels
	colLabelY := 0
	if len(hm.ColLabels) > 0 {
		for ci, cl := range hm.ColLabels {
			if ci >= hm.maxCols(data) {
				break
			}
			x := bounds.X + labelOffset + ci*cellWidth
			label := truncateString(cl, cellWidth)
			ctx.Buffer.SetString(x, bounds.Y+colLabelY, label, style)
		}
		colLabelY = 1
	}

	// Render rows
	for ri, row := range data {
		y := bounds.Y + colLabelY + ri
		if y >= bounds.Y+bounds.Height {
			break
		}

		// Row label
		if ri < len(hm.RowLabels) {
			label := truncateString(hm.RowLabels[ri], labelOffset-1)
			ctx.Buffer.SetString(bounds.X, y, label, style)
		}

		// Data cells
		for ci, val := range row {
			x := bounds.X + labelOffset + ci*cellWidth
			if x >= bounds.X+bounds.Width {
				break
			}

			// Normalize value
			ratio := (val - minVal) / (maxVal - minVal)
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}

			bgColor := colorScale(ratio)
			cellStyle := backend.DefaultStyle().Background(bgColor)

			if hm.ShowValues {
				// Render the numeric value centered in the cell
				valStr := formatHeatMapValue(val, cellWidth)
				valStr = truncateString(valStr, cellWidth)
				// Pad to cellWidth
				for len(valStr) < cellWidth {
					valStr += " "
				}
				for j, r := range valStr {
					if x+j < bounds.X+bounds.Width {
						ctx.Buffer.Set(x+j, y, r, cellStyle)
					}
				}
			} else {
				// Render colored spaces
				for j := 0; j < cellWidth; j++ {
					if x+j < bounds.X+bounds.Width {
						ctx.Buffer.Set(x+j, y, ' ', cellStyle)
					}
				}
			}
		}
	}
}

// HandleMessage returns unhandled.
func (hm *HeatMap) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// Bind attaches app services.
func (hm *HeatMap) Bind(services runtime.Services) {
	if hm == nil {
		return
	}
	hm.services = services
}

// Unbind releases app services.
func (hm *HeatMap) Unbind() {
	if hm == nil {
		return
	}
	hm.services = runtime.Services{}
}

func (hm *HeatMap) syncA11y() {
	if hm == nil {
		return
	}
	if hm.Base.Role == "" {
		hm.Base.Role = accessibility.RoleGrid
	}
	label := strings.TrimSpace(hm.label)
	if label == "" {
		label = "Heat Map"
	}
	hm.Base.Label = label

	data := [][]float64(nil)
	if hm.Data != nil {
		data = hm.Data.Get()
	}

	rows, cols := len(data), 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}

	if rows == 0 || cols == 0 {
		hm.Base.Description = "empty"
		hm.Base.Value = nil
		return
	}

	minVal, maxVal := math.Inf(1), math.Inf(-1)
	for _, row := range data {
		for _, v := range row {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}

	desc := fmt.Sprintf("%d columns by %d rows, range %s to %s",
		cols, rows, formatFloat(minVal), formatFloat(maxVal))
	hm.Base.Description = desc
	hm.Base.Value = &accessibility.ValueInfo{
		Text: fmt.Sprintf("min %s, max %s", formatFloat(minVal), formatFloat(maxVal)),
	}
	if desc != hm.lastDesc {
		hm.lastDesc = desc
		if announcer := hm.services.Announcer(); announcer != nil {
			announcer.AnnounceChange(hm)
		}
	}
}

func (hm *HeatMap) dataDims() (rows, cols int) {
	if hm == nil || hm.Data == nil {
		return 0, 0
	}
	data := hm.Data.Get()
	rows = len(data)
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return rows, cols
}

func (hm *HeatMap) maxCols(data [][]float64) int {
	cols := 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return cols
}

// formatHeatMapValue formats a float value to fit in a given width.
func formatHeatMapValue(val float64, width int) string {
	if width <= 1 {
		return fmt.Sprintf("%.0f", val)
	}
	s := fmt.Sprintf("%.1f", val)
	if len(s) > width {
		s = fmt.Sprintf("%.0f", val)
	}
	return s
}

var _ runtime.Widget = (*HeatMap)(nil)
var _ runtime.Bindable = (*HeatMap)(nil)
var _ runtime.Unbindable = (*HeatMap)(nil)
