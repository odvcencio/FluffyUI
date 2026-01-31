package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

// PerformanceDashboard renders render-loop performance metrics.
type PerformanceDashboard struct {
	Component

	sampler        *runtime.RenderSampler
	label          string
	refreshEvery   time.Duration
	style          backend.Style
	headerStyle    backend.Style
	mutedStyle     backend.Style
	styleSet       bool
	headerStyleSet bool
	mutedStyleSet  bool
	scheduled      bool
}

// PerformanceDashboardOption configures the dashboard.
type PerformanceDashboardOption = Option[PerformanceDashboard]

// NewPerformanceDashboard creates a dashboard wired to a render sampler.
func NewPerformanceDashboard(sampler *runtime.RenderSampler, opts ...PerformanceDashboardOption) *PerformanceDashboard {
	d := &PerformanceDashboard{
		sampler:      sampler,
		label:        "Performance",
		refreshEvery: 500 * time.Millisecond,
		style:        backend.DefaultStyle(),
		headerStyle:  backend.DefaultStyle().Bold(true),
		mutedStyle:   backend.DefaultStyle().Dim(true),
	}
	d.Base.Role = accessibility.RoleStatus
	d.syncA11y()
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}
	return d
}

// WithPerformanceLabel sets the accessibility label.
func WithPerformanceLabel(label string) PerformanceDashboardOption {
	return func(d *PerformanceDashboard) {
		if d == nil {
			return
		}
		d.label = strings.TrimSpace(label)
		d.syncA11y()
	}
}

// WithPerformanceRefresh sets the refresh interval (0 disables auto refresh).
func WithPerformanceRefresh(interval time.Duration) PerformanceDashboardOption {
	return func(d *PerformanceDashboard) {
		if d == nil {
			return
		}
		d.refreshEvery = interval
	}
}

// WithPerformanceStyle overrides the base style.
func WithPerformanceStyle(style backend.Style) PerformanceDashboardOption {
	return func(d *PerformanceDashboard) {
		if d == nil {
			return
		}
		d.style = style
		d.styleSet = true
	}
}

// WithPerformanceHeaderStyle overrides the header style.
func WithPerformanceHeaderStyle(style backend.Style) PerformanceDashboardOption {
	return func(d *PerformanceDashboard) {
		if d == nil {
			return
		}
		d.headerStyle = style
		d.headerStyleSet = true
	}
}

// WithPerformanceMutedStyle overrides the muted style.
func WithPerformanceMutedStyle(style backend.Style) PerformanceDashboardOption {
	return func(d *PerformanceDashboard) {
		if d == nil {
			return
		}
		d.mutedStyle = style
		d.mutedStyleSet = true
	}
}

// SetSampler updates the render sampler.
func (d *PerformanceDashboard) SetSampler(sampler *runtime.RenderSampler) {
	if d == nil {
		return
	}
	d.sampler = sampler
	if d.Services != (runtime.Services{}) {
		d.Services.Invalidate()
	}
}

// SetRefreshInterval updates the auto-refresh interval.
func (d *PerformanceDashboard) SetRefreshInterval(interval time.Duration) {
	if d == nil {
		return
	}
	d.refreshEvery = interval
}

// StyleType returns the selector type name.
func (d *PerformanceDashboard) StyleType() string {
	return "PerformanceDashboard"
}

// Bind attaches app services and schedules refreshes.
func (d *PerformanceDashboard) Bind(services runtime.Services) {
	d.Component.Bind(services)
	if d.refreshEvery <= 0 || d.scheduled {
		return
	}
	d.scheduled = true
	services.Every(d.refreshEvery, func(time.Time) runtime.Message {
		return runtime.InvalidateMsg{}
	})
}

// Unbind releases subscriptions.
func (d *PerformanceDashboard) Unbind() {
	d.Component.Unbind()
}

// Measure returns the desired size.
func (d *PerformanceDashboard) Measure(constraints runtime.Constraints) runtime.Size {
	lines := d.summaryLines()
	width := 0
	for _, line := range lines {
		if w := textWidth(line.text); w > width {
			width = w
		}
	}
	if width == 0 {
		width = 1
	}
	height := len(lines)
	if height == 0 {
		height = 1
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

// Render draws the dashboard summary.
func (d *PerformanceDashboard) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	bounds := d.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	baseStyle := resolveBaseStyle(ctx, d, backend.DefaultStyle(), false)
	if d.styleSet {
		baseStyle = mergeBackendStyles(baseStyle, d.style)
	}
	headerStyle := baseStyle
	if d.headerStyleSet {
		headerStyle = mergeBackendStyles(baseStyle, d.headerStyle)
	} else {
		headerStyle = baseStyle.Bold(true)
	}
	mutedStyle := baseStyle
	if d.mutedStyleSet {
		mutedStyle = mergeBackendStyles(baseStyle, d.mutedStyle)
	} else {
		mutedStyle = baseStyle.Dim(true)
	}

	ctx.Buffer.Fill(bounds, ' ', baseStyle)
	lines := d.summaryLines()
	for i, line := range lines {
		if i >= bounds.Height {
			break
		}
		style := baseStyle
		switch line.kind {
		case perfLineHeader:
			style = headerStyle
		case perfLineMuted:
			style = mutedStyle
		}
		writePadded(ctx.Buffer, bounds.X, bounds.Y+i, bounds.Width, truncateString(line.text, bounds.Width), style)
	}
}

func (d *PerformanceDashboard) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleStatus
	}
	label := strings.TrimSpace(d.label)
	if label == "" {
		label = "Performance"
	}
	d.Base.Label = label
}

type perfLineKind int

const (
	perfLineBase perfLineKind = iota
	perfLineHeader
	perfLineMuted
)

type perfLine struct {
	text string
	kind perfLineKind
}

func (d *PerformanceDashboard) summaryLines() []perfLine {
	if d == nil {
		return []perfLine{{text: "Performance", kind: perfLineHeader}}
	}
	if d.sampler == nil {
		return []perfLine{{text: "No render sampler attached", kind: perfLineMuted}}
	}
	summary := d.sampler.Summary()
	if summary.Samples == 0 {
		return []perfLine{{text: "No render samples yet", kind: perfLineMuted}}
	}
	fps := 0.0
	if summary.AvgTotal > 0 {
		fps = 1.0 / summary.AvgTotal.Seconds()
	}
	lines := []perfLine{{text: fmt.Sprintf("Render Summary (frames %d)", summary.Frames), kind: perfLineHeader}}
	lines = append(lines,
		perfLine{text: fmt.Sprintf("Window: %d samples", summary.Samples), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Avg total: %s (%.1f fps)", formatDuration(summary.AvgTotal), fps), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Avg render: %s", formatDuration(summary.AvgRender)), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Avg flush: %s", formatDuration(summary.AvgFlush)), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Dirty avg: %.1f%%", summary.AvgDirtyRatio*100), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Max total: %s", formatDuration(summary.MaxTotal)), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Max render: %s", formatDuration(summary.MaxRender)), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Max flush: %s", formatDuration(summary.MaxFlush)), kind: perfLineBase},
	)
	lastDirty := "n/a"
	if summary.Last.TotalCells > 0 {
		lastDirty = fmt.Sprintf("%d/%d", summary.Last.DirtyCells, summary.Last.TotalCells)
	}
	lines = append(lines,
		perfLine{text: fmt.Sprintf("Last dirty: %s", lastDirty), kind: perfLineBase},
		perfLine{text: fmt.Sprintf("Layers: %d", summary.Last.LayerCount), kind: perfLineBase},
	)
	return lines
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0ms"
	}
	ms := float64(d) / float64(time.Millisecond)
	if ms < 1 {
		return fmt.Sprintf("%.2fms", ms)
	}
	if ms < 100 {
		return fmt.Sprintf("%.1fms", ms)
	}
	return fmt.Sprintf("%.0fms", ms)
}

var _ runtime.Widget = (*PerformanceDashboard)(nil)
