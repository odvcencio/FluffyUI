package widgets

import (
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/gpu"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	flufftest "github.com/odvcencio/fluffyui/testing"
	"github.com/odvcencio/fluffyui/toast"
)

func renderSmoke(t *testing.T, name string, w runtime.Widget, width, height int) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", name, r)
		}
	}()
	_ = flufftest.RenderToString(w, width, height)
}

func TestWidgetSmokeRender(t *testing.T) {
	chartData := state.NewSignal([]float64{1, 2, 3, 2, 1})
	barData := state.NewSignal([]BarData{{Label: "A", Value: 1}, {Label: "B", Value: 2}})
	sliderValue := state.NewSignal(0.5)
	minValue := state.NewSignal(0.2)
	maxValue := state.NewSignal(0.8)
	now := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name   string
		widget runtime.Widget
		w      int
		h      int
	}{
		{"Accordion", NewAccordion(NewAccordionSection("Section", NewLabel("Content"))), 24, 5},
		{"Alert", NewAlert("Heads up", AlertInfo), 20, 1},
		{"AnimatedGauge", NewAnimatedGauge(0, 100), 20, 1},
		{"AspectRatio", NewAspectRatio(NewLabel("X"), 2.0), 10, 5},
		{"AsyncImage", NewAsyncImageWithLoader(func() (image.Image, error) {
			img := image.NewRGBA(image.Rect(0, 0, 2, 2))
			img.Set(0, 0, color.RGBA{R: 255, A: 255})
			return img, nil
		}), 10, 4},
		{"AutoComplete", func() runtime.Widget {
			ac := NewAutoComplete()
			ac.SetOptions([]string{"Alpha", "Beta", "Gamma"})
			ac.SetQuery("a")
			return ac
		}(), 20, 6},
		{"Breadcrumb", NewBreadcrumb(BreadcrumbItem{Label: "Home"}, BreadcrumbItem{Label: "Library"}), 30, 1},
		{"Button", NewButton("OK"), 10, 1},
		{"Calendar", func() runtime.Widget {
			cal := NewCalendar(WithNowFunc(func() time.Time { return now }))
			return cal
		}(), 28, 9},
		{"CanvasWidget", NewCanvasWidget(func(c *graphics.Canvas) {
			c.SetPixel(0, 0, 1)
		}), 10, 4},
		{"Checkbox", NewCheckbox("Accept"), 15, 1},
		{"DataGrid", func() runtime.Widget {
			grid := NewDataGrid(TableColumn{Title: "Name"}, TableColumn{Title: "Value"})
			grid.SetRows([][]string{{"Alpha", "1"}, {"Beta", "2"}})
			return grid
		}(), 30, 6},
		{"DatePicker", func() runtime.Widget {
			cal := NewDatePicker()
			cal.calendar.now = func() time.Time { return now }
			return cal
		}(), 20, 8},
		{"DateRangePicker", func() runtime.Widget {
			picker := NewDateRangePicker()
			picker.calendar.now = func() time.Time { return now }
			start := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
			end := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
			picker.SetRange(&start, &end)
			return picker
		}(), 34, 10},
		{"DebugOverlay", NewDebugOverlay(NewLabel("Root")), 30, 6},
		{"Dialog", NewDialog("Title", "Body"), 30, 6},
		{"EnhancedPalette", NewEnhancedPalette(keybind.NewRegistry()).Widget, 30, 6},
		{"PerformanceDashboard", NewPerformanceDashboard(runtime.NewRenderSampler(10)), 30, 10},
		{"Grid", func() runtime.Widget {
			grid := NewGrid(2, 2)
			grid.Add(NewLabel("A"), 0, 0, 1, 1)
			grid.Add(NewLabel("B"), 1, 1, 1, 1)
			return grid
		}(), 12, 4},
		{"Input", NewInput(), 12, 1},
		{"LineChart", func() runtime.Widget {
			chart := NewLineChart()
			chart.SetSeries([]ChartSeries{{Data: []float64{1, 3, 2}, Color: 1}})
			return chart
		}(), 20, 8},
		{"List", func() runtime.Widget {
			adapter := NewSliceAdapter([]string{"One", "Two", "Three"}, func(item string, index int, selected bool, ctx runtime.RenderContext) {
				style := backend.DefaultStyle()
				if selected {
					style = style.Reverse(true)
				}
				writePadded(ctx.Buffer, ctx.Bounds.X, ctx.Bounds.Y, ctx.Bounds.Width, item, style)
			})
			return NewList(adapter)
		}(), 12, 4},
		{"Menu", NewMenu(&MenuItem{ID: "1", Title: "File"}, &MenuItem{ID: "2", Title: "Edit"}), 20, 6},
		{"MultiSelect", NewMultiSelect(MultiSelectOption{Label: "Red"}, MultiSelectOption{Label: "Blue"}), 20, 4},
		{"Panel", NewPanel(NewLabel("Content")), 20, 4},
		{"Popover", NewPopover(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 2}, NewLabel("Tip")), 20, 4},
		{"Progress", func() runtime.Widget {
			p := NewProgress()
			p.Value = 40
			return p
		}(), 20, 1},
		{"Radio", func() runtime.Widget {
			group := NewRadioGroup()
			return NewRadio("Option", group)
		}(), 14, 1},
		{"RangeSlider", func() runtime.Widget {
			return NewRangeSlider(minValue, maxValue)
		}(), 20, 3},
		{"RichText", NewRichText("Hello **World**"), 30, 6},
		{"Search", NewSearchWidget(), 20, 1},
		{"Section", NewSection("Section"), 16, 3},
		{"Select", NewSelect(SelectOption{Label: "One", Value: 1}, SelectOption{Label: "Two", Value: 2}), 18, 1},
		{"SimpleWidget", NewSimpleWidget(), 10, 2},
		{"SignalLabel", func() runtime.Widget {
			sig := state.NewSignal("Signal")
			return NewSignalLabel(sig, nil)
		}(), 20, 1},
		{"Slider", NewSlider(sliderValue), 20, 2},
		{"Sparkline", NewSparkline(chartData), 20, 4},
		{"Spinner", NewSpinner(), 6, 1},
		{"Splitter", NewSplitter(NewLabel("Left"), NewLabel("Right")), 20, 3},
		{"Stack", NewStack(NewLabel("Top"), NewLabel("Bottom")), 20, 3},
		{"Stepper", NewStepper(Step{Title: "One"}, Step{Title: "Two"}), 20, 3},
		{"Table", func() runtime.Widget {
			table := NewTable(TableColumn{Title: "Name"}, TableColumn{Title: "Value"})
			table.SetRows([][]string{{"Alpha", "1"}, {"Beta", "2"}})
			return table
		}(), 30, 6},
		{"Tabs", NewTabs(Tab{Title: "One", Content: NewLabel("First")}, Tab{Title: "Two", Content: NewLabel("Second")}), 20, 4},
		{"Text", NewText("Hello"), 10, 2},
		{"TextArea", func() runtime.Widget {
			area := NewTextArea()
			area.SetText("Line 1\nLine 2")
			return area
		}(), 20, 4},
		{"TimePicker", func() runtime.Widget {
			picker := NewTimePicker()
			picker.SetTime(time.Date(2026, time.January, 2, 9, 30, 15, 0, time.UTC))
			picker.SetShowSeconds(true)
			return picker
		}(), 10, 1},
		{"ToastStack", func() runtime.Widget {
			stack := NewToastStack()
			stack.SetNow(now)
			stack.SetToasts([]*toast.Toast{{
				ID:        "1",
				Message:   "Saved",
				Level:     toast.ToastSuccess,
				CreatedAt: now.Add(-time.Second),
				Duration:  5 * time.Second,
			}})
			return stack
		}(), 40, 8},
		{"Tooltip", NewTooltip(NewLabel("Target"), NewLabel("Tip")), 20, 3},
		{"Tree", NewTree(&TreeNode{Label: "Root", Expanded: true, Children: []*TreeNode{{Label: "Child"}}}), 20, 6},
		{"VirtualList", func() runtime.Widget {
			list := NewVirtualList(testVirtualListAdapter{items: []string{"One", "Two", "Three"}})
			return list
		}(), 20, 5},
		{"ScrollView", NewScrollView(NewLabel("Scrollable")), 20, 3},
		{"GPUCanvas", NewGPUCanvasWidget(func(canvas *gpu.GPUCanvas) {
			canvas.Clear(color.RGBA{R: 20, G: 40, B: 60, A: 255})
		}), 12, 4},
		{"BarChart", NewBarChart(barData), 20, 6},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			renderSmoke(t, tc.name, tc.widget, tc.w, tc.h)
		})
	}
}
