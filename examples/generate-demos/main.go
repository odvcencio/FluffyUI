// Demo Generator - Creates asciicast recordings of FluffyUI widgets
//
// This tool generates demo recordings using the simulation backend,
// which doesn't require a real terminal. Perfect for CI/CD pipelines.
//
// Usage:
//
//	go run ./examples/generate-demos --out docs/demos
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/graphics"
	"github.com/odvcencio/fluffy-ui/recording"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

var (
	outDir     = flag.String("out", "docs/demos", "output directory for recordings")
	width      = flag.Int("width", 80, "recording width")
	height     = flag.Int("height", 24, "recording height")
	demoFilter = flag.String("demo", "", "comma-separated list of demos to record")
	fps        = flag.Int("fps", 30, "frames per second")
	duration   = flag.Float64("duration", 5.0, "recording duration in seconds")
)

func main() {
	flag.Parse()

	selected := parseDemoFilter(*demoFilter)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	demos := []struct {
		name   string
		fn     func() runtime.Widget
		width  int
		height int
		frames int
	}{
		{"quickstart", demoQuickstart, 60, 16, 0},
		{"buttons", demoButtons, 80, 24, 0},
		{"counter", demoCounter, 80, 24, 0},
		{"table", demoTable, 80, 24, 0},
		{"progress", demoProgress, 80, 24, 0},
		{"list", demoList, 80, 24, 0},
		{"dialog", demoDialog, 60, 16, 0},
		{"sparkline", demoSparkline, 80, 24, 0},
		{"tabs", demoTabs, 80, 24, 0},
		{"input", demoInput, 80, 24, 0},
		{"graphics", demoGraphics, 100, 32, 0},
		{"easing", demoEasing, 120, 40, 0},
		{"hero", demoHero, 80, 24, 0},
		{"fireworks", demoFireworks, 100, 36, 0},
	}

	for _, demo := range demos {
		if len(selected) > 0 && !selected[demo.name] {
			continue
		}
		outPath := filepath.Join(*outDir, demo.name+".cast")
		fmt.Printf("Recording: %s -> %s\n", demo.name, outPath)

		w, h := demo.width, demo.height
		if w == 0 {
			w = *width
		}
		if h == 0 {
			h = *height
		}
		frames := demo.frames
		if frames == 0 {
			frames = int(*duration * float64(*fps))
		}

		if err := recordDemo(outPath, demo.fn(), w, h, frames, *fps); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
			continue
		}
		fmt.Println("  OK")
	}

	fmt.Println("\nDemos recorded successfully!")
	fmt.Println("\nTo view recordings:")
	fmt.Printf("  asciinema play %s/hero.cast\n", *outDir)
	fmt.Println("\nTo convert to GIF (requires agg):")
	fmt.Printf("  agg --theme monokai --last-frame-duration 0.001 %s/hero.cast %s/hero.gif\n", *outDir, *outDir)
}

func parseDemoFilter(input string) map[string]bool {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	selected := make(map[string]bool)
	for _, name := range strings.Split(input, ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		selected[trimmed] = true
	}
	return selected
}

func recordDemo(path string, root runtime.Widget, width, height, maxFrames, fps int) error {
	recorder, err := recording.NewAsciicastRecorder(path, recording.AsciicastOptions{
		Title: "FluffyUI Demo",
	})
	if err != nil {
		return err
	}

	frameCount := 0

	update := func(app *runtime.App, msg runtime.Message) bool {
		switch msg.(type) {
		case runtime.TickMsg:
			frameCount++
			if frameCount >= maxFrames {
				app.ExecuteCommand(runtime.Quit{})
				return false
			}
			// Forward tick to widgets so animations work
			runtime.DefaultUpdate(app, msg)
			return true
		}
		return runtime.DefaultUpdate(app, msg)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  sim.New(width, height),
		Root:     root,
		Update:   update,
		TickRate: time.Second / time.Duration(fps),
		Recorder: recorder,
	})

	return app.Run(context.Background())
}

// =============================================================================
// Demo Widgets
// =============================================================================

// =============================================================================
// Quickstart Demo - Simple hello world with animation
// =============================================================================

func demoQuickstart() runtime.Widget {
	return &quickstartDemo{}
}

type quickstartDemo struct {
	widgets.Component
	frame int
}

func (q *quickstartDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (q *quickstartDemo) Layout(bounds runtime.Rect) {
	q.Component.Layout(bounds)
}

func (q *quickstartDemo) Render(ctx runtime.RenderContext) {
	bounds := q.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Animated rainbow colors
	colors := []backend.Color{
		backend.ColorBrightRed, backend.ColorBrightYellow, backend.ColorBrightGreen,
		backend.ColorBrightCyan, backend.ColorBrightBlue, backend.ColorBrightMagenta,
	}

	// Title with typing effect
	title := "Hello from FluffyUI!"
	visibleChars := q.frame / 2
	if visibleChars > len(title) {
		visibleChars = len(title)
	}

	titleX := (bounds.Width - len(title)) / 2
	titleY := bounds.Height / 2 - 1

	// Draw visible characters with rainbow effect
	for i := 0; i < visibleChars; i++ {
		color := colors[(i+q.frame/4)%len(colors)]
		style := backend.DefaultStyle().Foreground(color).Bold(true)
		ctx.Buffer.Set(bounds.X+titleX+i, bounds.Y+titleY, rune(title[i]), style)
	}

	// Blinking cursor at end
	if visibleChars < len(title) && (q.frame/8)%2 == 0 {
		ctx.Buffer.SetString(bounds.X+titleX+visibleChars, bounds.Y+titleY, "▌", backend.DefaultStyle().Foreground(backend.ColorBrightWhite))
	}

	// Subtitle appears after title is complete
	if visibleChars >= len(title) {
		subtitle := "Press 'q' to quit"
		subX := (bounds.Width - len(subtitle)) / 2
		subY := titleY + 2

		// Fade in effect
		fadeFrame := q.frame - len(title)*2
		if fadeFrame > 0 {
			style := backend.DefaultStyle().Dim(true)
			if fadeFrame > 15 {
				style = backend.DefaultStyle()
			}
			ctx.Buffer.SetString(bounds.X+subX, bounds.Y+subY, subtitle, style)
		}
	}

	// Animated border sparkles - flows clockwise around the perimeter
	borderChars := []rune{'·', '•', '◦', '○', '◌'}
	perimeter := 2*(bounds.Width+bounds.Height-2)
	if perimeter <= 0 {
		perimeter = 1
	}

	drawBorderCell := func(x, y, perimPos int) {
		// Subtract frame to make it flow clockwise
		idx := (perimPos - q.frame/2 + perimeter*100) % perimeter
		char := borderChars[idx%len(borderChars)]
		color := colors[(idx/2)%len(colors)]
		style := backend.DefaultStyle().Foreground(color)
		ctx.Buffer.Set(x, y, char, style)
	}

	pos := 0
	// Top edge (left to right)
	for i := 0; i < bounds.Width; i++ {
		drawBorderCell(bounds.X+i, bounds.Y, pos)
		pos++
	}
	// Right edge (top to bottom)
	for i := 1; i < bounds.Height-1; i++ {
		drawBorderCell(bounds.X+bounds.Width-1, bounds.Y+i, pos)
		pos++
	}
	// Bottom edge (right to left)
	for i := bounds.Width - 1; i >= 0; i-- {
		drawBorderCell(bounds.X+i, bounds.Y+bounds.Height-1, pos)
		pos++
	}
	// Left edge (bottom to top)
	for i := bounds.Height - 2; i >= 1; i-- {
		drawBorderCell(bounds.X, bounds.Y+i, pos)
		pos++
	}
}

func (q *quickstartDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		q.frame++
		q.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// =============================================================================
// Buttons Demo
// =============================================================================

func demoButtons() runtime.Widget {
	return &buttonsDemo{}
}

type buttonsDemo struct {
	widgets.Component
	frame      int
	focused    int
	clicked    int
	clickFlash int
}

func (b *buttonsDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (b *buttonsDemo) Layout(bounds runtime.Rect) {
	b.Component.Layout(bounds)
}

func (b *buttonsDemo) Render(ctx runtime.RenderContext) {
	bounds := b.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Animated title with color
	titleColors := []backend.Color{
		backend.ColorBrightCyan, backend.ColorBrightGreen, backend.ColorBrightYellow,
		backend.ColorBrightMagenta, backend.ColorBrightBlue,
	}
	titleColor := titleColors[(b.frame/10)%len(titleColors)]
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "FluffyUI Button Variants", backend.DefaultStyle().Bold(true).Foreground(titleColor))

	// Button definitions with colors
	buttons := []struct {
		label string
		style backend.Style
	}{
		{"Primary", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorCyan).Bold(true)},
		{"Secondary", backend.DefaultStyle().Foreground(backend.ColorWhite).Background(backend.ColorBlue)},
		{"Danger", backend.DefaultStyle().Foreground(backend.ColorWhite).Background(backend.ColorRed).Bold(true)},
		{"Success", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorGreen).Bold(true)},
		{"Warning", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorYellow)},
	}

	y := bounds.Y + 3
	x := bounds.X + 2
	for i, btn := range buttons {
		style := btn.style
		// Highlight focused button with pulsing effect
		if i == b.focused {
			pulse := (b.frame / 4) % 2
			if pulse == 0 {
				style = style.Reverse(true)
			}
		}
		// Flash effect on click
		if i == b.clicked && b.clickFlash > 0 {
			style = backend.DefaultStyle().Background(backend.ColorBrightWhite).Foreground(backend.ColorBlack).Bold(true)
		}
		label := fmt.Sprintf(" %s ", btn.label)
		ctx.Buffer.SetString(x, y, label, style)
		x += len(label) + 2
	}

	// Second row with icon buttons
	y += 2
	x = bounds.X + 2

	iconButtons := []struct {
		icon  string
		label string
		style backend.Style
	}{
		{"[+]", "Add", backend.DefaultStyle().Foreground(backend.ColorGreen).Bold(true)},
		{"[-]", "Remove", backend.DefaultStyle().Foreground(backend.ColorRed).Bold(true)},
		{"[*]", "Star", backend.DefaultStyle().Foreground(backend.ColorYellow).Bold(true)},
		{"[>]", "Play", backend.DefaultStyle().Foreground(backend.ColorCyan).Bold(true)},
	}

	for i, btn := range iconButtons {
		style := btn.style
		if i+5 == b.focused {
			style = style.Reverse(true)
		}
		ctx.Buffer.SetString(x, y, btn.icon, style)
		ctx.Buffer.SetString(x+4, y, btn.label, backend.DefaultStyle().Dim(true))
		x += len(btn.icon) + len(btn.label) + 4
	}

	// Instructions with animated cursor
	cursor := " "
	if (b.frame/15)%2 == 0 {
		cursor = "_"
	}
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+7, fmt.Sprintf("Tab cycles focus, Enter activates%s", cursor), backend.DefaultStyle().Dim(true))

	// Focus indicator with progress bar
	focusText := fmt.Sprintf("Focus: %d/9", b.focused+1)
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+9, focusText, backend.DefaultStyle())

	// Mini progress bar showing focus position
	barWidth := 20
	progress := float64(b.focused) / 8.0
	filledWidth := int(progress * float64(barWidth))
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	ctx.Buffer.SetString(bounds.X+15, bounds.Y+9, bar, backend.DefaultStyle().Foreground(backend.ColorCyan))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (b *buttonsDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		b.frame++
		// Cycle focus every ~400ms (12 frames at 30fps)
		if b.frame%12 == 0 {
			b.focused = (b.focused + 1) % 9
			// Simulate click every few focuses
			if b.focused%3 == 0 {
				b.clicked = b.focused
				b.clickFlash = 6
			}
			b.Invalidate()
		}
		// Decay click flash
		if b.clickFlash > 0 {
			b.clickFlash--
			b.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoCounter() runtime.Widget {
	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	return &counterDemo{count: count}
}

type counterDemo struct {
	widgets.Component
	count       *state.Signal[int]
	frame       int
	focusedBtn  int // 0=decrement, 1=increment
	flashEffect int
	history     []int
}

func (c *counterDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *counterDemo) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
}

func (c *counterDemo) Render(ctx runtime.RenderContext) {
	bounds := c.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Title
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Reactive Counter", backend.DefaultStyle().Bold(true).Foreground(backend.ColorBrightCyan))

	// Big count display
	countVal := c.count.Get()

	// Large number display with flash effect
	bigNumStyle := backend.DefaultStyle().Bold(true).Foreground(backend.ColorBrightWhite)
	if c.flashEffect > 0 {
		if c.focusedBtn == 1 {
			bigNumStyle = bigNumStyle.Foreground(backend.ColorBrightGreen)
		} else {
			bigNumStyle = bigNumStyle.Foreground(backend.ColorBrightRed)
		}
	}

	centerX := bounds.X + bounds.Width/2
	centerY := bounds.Y + 5

	// Draw count with padding
	displayStr := fmt.Sprintf("[ %3d ]", countVal)
	ctx.Buffer.SetString(centerX-len(displayStr)/2, centerY, displayStr, bigNumStyle)

	// Buttons
	y := centerY + 2
	decStyle := backend.DefaultStyle()
	incStyle := backend.DefaultStyle()

	if c.focusedBtn == 0 {
		decStyle = backend.DefaultStyle().Background(backend.ColorBrightRed).Foreground(backend.ColorBlack).Bold(true)
	} else {
		incStyle = backend.DefaultStyle().Background(backend.ColorBrightGreen).Foreground(backend.ColorBlack).Bold(true)
	}

	ctx.Buffer.SetString(centerX-12, y, "  -  ", decStyle)
	ctx.Buffer.SetString(centerX+7, y, "  +  ", incStyle)

	// History sparkline
	if len(c.history) > 1 {
		y += 3
		ctx.Buffer.SetString(bounds.X+2, y, "History:", backend.DefaultStyle().Dim(true))

		// Draw mini sparkline
		maxVal := 1
		for _, v := range c.history {
			if v > maxVal {
				maxVal = v
			}
		}

		sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
		sparkX := bounds.X + 12
		for i, v := range c.history {
			if i >= 30 {
				break
			}
			idx := v * (len(sparkChars) - 1) / maxVal
			if idx < 0 {
				idx = 0
			}
			if idx >= len(sparkChars) {
				idx = len(sparkChars) - 1
			}
			ctx.Buffer.Set(sparkX+i, y, sparkChars[idx], backend.DefaultStyle().Foreground(backend.ColorBrightCyan))
		}
	}

	// Keyboard hints
	y = bounds.Y + bounds.Height - 3
	ctx.Buffer.SetString(bounds.X+2, y, "←→", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+5, y, "Switch", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+14, y, "Space", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+20, y, "Press", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+28, y, "R", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+30, y, "Reset", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (c *counterDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		c.frame++

		// Decay flash effect
		if c.flashEffect > 0 {
			c.flashEffect--
		}

		// Switch focus periodically
		if c.frame%20 == 0 {
			c.focusedBtn = (c.focusedBtn + 1) % 2
		}

		// Press button periodically
		if c.frame%12 == 0 {
			if c.focusedBtn == 1 {
				c.count.Update(func(v int) int { return v + 1 })
			} else {
				c.count.Update(func(v int) int {
					if v > 0 {
						return v - 1
					}
					return 0
				})
			}
			c.flashEffect = 4
			c.history = append(c.history, c.count.Get())
			if len(c.history) > 30 {
				c.history = c.history[1:]
			}
		}

		c.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoTable() runtime.Widget {
	return &tableDemo{}
}

type tableDemo struct {
	widgets.Component
	frame     int
	selected  int
	sortCol   int
	sortAsc   bool
	hoverCol  int
}

func (t *tableDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (t *tableDemo) Layout(bounds runtime.Rect) {
	t.Component.Layout(bounds)
}

func (t *tableDemo) Render(ctx runtime.RenderContext) {
	bounds := t.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Column definitions
	columns := []struct {
		title string
		width int
	}{
		{"Product", 16},
		{"Category", 10},
		{"Price", 8},
		{"Stock", 8},
		{"Status", 10},
	}

	rows := [][]string{
		{"Gummy Bears", "Candy", "$2.99", "150", "●"},
		{"Chocolate Bar", "Candy", "$4.50", "89", "●"},
		{"Sour Straws", "Candy", "$1.99", "234", "●"},
		{"Lollipops", "Candy", "$0.99", "500", "●"},
		{"Jawbreakers", "Candy", "$3.25", "12", "○"},
		{"Energy Drink", "Beverage", "$5.99", "0", "○"},
	}

	// Title with record count
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Inventory", backend.DefaultStyle().Bold(true).Foreground(backend.ColorBrightCyan))
	ctx.Buffer.SetString(bounds.X+14, bounds.Y+1, fmt.Sprintf("%d items", len(rows)), backend.DefaultStyle().Dim(true))

	// Sort indicator in title area
	sortInfo := fmt.Sprintf("Sorted by: %s", columns[t.sortCol].title)
	if t.sortAsc {
		sortInfo += " ▲"
	} else {
		sortInfo += " ▼"
	}
	ctx.Buffer.SetString(bounds.X+bounds.Width-25, bounds.Y+1, sortInfo, backend.DefaultStyle().Foreground(backend.ColorBrightYellow))

	// Header row
	y := bounds.Y + 3
	x := bounds.X + 4
	for i, col := range columns {
		headerStyle := backend.DefaultStyle().Bold(true)
		if i == t.sortCol {
			headerStyle = headerStyle.Foreground(backend.ColorBrightYellow).Underline(true)
		} else if i == t.hoverCol {
			headerStyle = headerStyle.Foreground(backend.ColorBrightCyan)
		}
		title := col.title
		if i == t.sortCol {
			if t.sortAsc {
				title += "▲"
			} else {
				title += "▼"
			}
		}
		for len(title) < col.width {
			title += " "
		}
		ctx.Buffer.SetString(x, y, title, headerStyle)
		x += col.width + 1
	}

	// Separator line
	y++
	ctx.Buffer.SetString(bounds.X+2, y, strings.Repeat("─", bounds.Width-4), backend.DefaultStyle().Dim(true))

	// Data rows
	for rowIdx, row := range rows {
		y++
		x = bounds.X + 2

		isSelected := rowIdx == t.selected
		rowStyle := backend.DefaultStyle()

		// Row number
		numStyle := backend.DefaultStyle().Dim(true)
		if isSelected {
			numStyle = backend.DefaultStyle().Foreground(backend.ColorBrightCyan)
		}
		ctx.Buffer.SetString(x, y, fmt.Sprintf("%d", rowIdx+1), numStyle)
		x += 2

		// Selection highlight
		if isSelected {
			for dx := x; dx < bounds.X+bounds.Width-2; dx++ {
				ctx.Buffer.Set(dx, y, ' ', backend.DefaultStyle().Background(backend.ColorBlue))
			}
			rowStyle = rowStyle.Background(backend.ColorBlue).Foreground(backend.ColorBrightWhite)
		}

		for colIdx, col := range columns {
			cell := ""
			if colIdx < len(row) {
				cell = row[colIdx]
			}

			cellStyle := rowStyle

			// Special styling for status column
			if colIdx == 4 {
				if cell == "●" {
					cellStyle = cellStyle.Foreground(backend.ColorBrightGreen)
					cell = "● In Stock"
				} else {
					cellStyle = cellStyle.Foreground(backend.ColorBrightRed)
					cell = "○ Low"
				}
			}

			// Price column styling
			if colIdx == 2 && !isSelected {
				cellStyle = cellStyle.Foreground(backend.ColorBrightYellow)
			}

			for len(cell) < col.width {
				cell += " "
			}
			ctx.Buffer.SetString(x, y, cell, cellStyle)
			x += col.width + 1
		}
	}

	// Footer
	y = bounds.Y + bounds.Height - 3
	ctx.Buffer.SetString(bounds.X+2, y, "↑↓", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+5, y, "Select", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+14, y, "←→", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+17, y, "Sort", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+24, y, "Enter", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+30, y, "Edit", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (t *tableDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		t.frame++

		// Navigate rows
		if t.frame%10 == 0 {
			t.selected = (t.selected + 1) % 6
		}

		// Change sort column periodically
		if t.frame%45 == 0 {
			t.hoverCol = (t.sortCol + 1) % 5
		}
		if t.frame%50 == 0 {
			t.sortCol = t.hoverCol
			t.sortAsc = !t.sortAsc
			t.hoverCol = -1
		}

		t.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoProgress() runtime.Widget {
	return &progressDemo{}
}

type progressDemo struct {
	widgets.Component
	frame   int
	values  []float64
	current int
}

func (p *progressDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (p *progressDemo) Layout(bounds runtime.Rect) {
	p.Component.Layout(bounds)
}

func (p *progressDemo) Render(ctx runtime.RenderContext) {
	bounds := p.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Animated title
	titleColors := []backend.Color{backend.ColorBrightCyan, backend.ColorBrightGreen, backend.ColorBrightYellow}
	titleColor := titleColors[(p.frame/15)%len(titleColors)]
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Progress & Gauges Demo", backend.DefaultStyle().Bold(true).Foreground(titleColor))

	// Multiple progress bars with different styles
	progress1 := float64(p.frame%100) / 100.0
	progress2 := float64((p.frame+33)%100) / 100.0
	progress3 := float64((p.frame+66)%100) / 100.0

	y := bounds.Y + 3
	// Download with animated icon
	dlIcon := []string{"↓", "⬇", "▼"}[(p.frame/5)%3]
	ctx.Buffer.SetString(bounds.X+2, y, fmt.Sprintf("%s Download:", dlIcon), backend.DefaultStyle().Foreground(backend.ColorCyan))
	drawColoredGauge(ctx.Buffer, bounds.X+14, y, 40, progress1, backend.ColorCyan)
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress1*100), backend.DefaultStyle().Foreground(backend.ColorCyan))

	y += 2
	// Upload with animated icon
	ulIcon := []string{"↑", "⬆", "▲"}[(p.frame/5)%3]
	ctx.Buffer.SetString(bounds.X+2, y, fmt.Sprintf("%s Upload:  ", ulIcon), backend.DefaultStyle().Foreground(backend.ColorGreen))
	drawColoredGauge(ctx.Buffer, bounds.X+14, y, 40, progress2, backend.ColorGreen)
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress2*100), backend.DefaultStyle().Foreground(backend.ColorGreen))

	y += 2
	// Process with spinner
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinChars[p.frame%len(spinChars)]
	ctx.Buffer.SetString(bounds.X+2, y, fmt.Sprintf("%s Process: ", spinner), backend.DefaultStyle().Foreground(backend.ColorYellow))
	drawColoredGauge(ctx.Buffer, bounds.X+14, y, 40, progress3, backend.ColorYellow)
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress3*100), backend.DefaultStyle().Foreground(backend.ColorYellow))

	// Multi-step progress indicator
	y += 3
	steps := []string{"Fetch", "Parse", "Build", "Deploy"}
	currentStep := (p.frame / 25) % (len(steps) + 1)
	ctx.Buffer.SetString(bounds.X+2, y, "Pipeline:", backend.DefaultStyle().Bold(true))
	x := bounds.X + 14
	for i, step := range steps {
		style := backend.DefaultStyle().Dim(true)
		icon := "○"
		if i < currentStep {
			style = backend.DefaultStyle().Foreground(backend.ColorGreen)
			icon = "●"
		} else if i == currentStep {
			style = backend.DefaultStyle().Foreground(backend.ColorYellow).Bold(true)
			icon = spinChars[p.frame%len(spinChars)]
		}
		ctx.Buffer.SetString(x, y, fmt.Sprintf("%s %s", icon, step), style)
		x += len(step) + 4
		if i < len(steps)-1 {
			connStyle := backend.DefaultStyle().Dim(true)
			if i < currentStep {
				connStyle = backend.DefaultStyle().Foreground(backend.ColorGreen)
			}
			ctx.Buffer.SetString(x-2, y, "→", connStyle)
		}
	}

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func drawColoredGauge(buf backend.RenderTarget, x, y, width int, progress float64, color backend.Color) {
	filled := int(progress * float64(width))
	for i := 0; i < width; i++ {
		char := "░"
		style := backend.DefaultStyle().Dim(true)
		if i < filled {
			char = "█"
			style = backend.DefaultStyle().Foreground(color)
		}
		buf.SetContent(x+i, y, rune(char[0]), nil, style)
	}
}

func (p *progressDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		p.frame++
		p.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoList() runtime.Widget {
	return &listDemo{}
}

type listDemo struct {
	widgets.Component
	frame    int
	selected int
	checked  map[int]bool
	phase    int // 0=navigating, 1=selecting items, 2=show selected count
}

func (l *listDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (l *listDemo) Layout(bounds runtime.Rect) {
	l.Component.Layout(bounds)
	if l.checked == nil {
		l.checked = make(map[int]bool)
	}
}

func (l *listDemo) Render(ctx runtime.RenderContext) {
	bounds := l.Bounds()
	ctx.Clear(backend.DefaultStyle())

	items := []struct {
		icon  string
		label string
		desc  string
		size  string
	}{
		{"📁", "Documents", "Personal files", "2.4 GB"},
		{"📥", "Downloads", "Recent downloads", "856 MB"},
		{"🎵", "Music", "Audio collection", "12.3 GB"},
		{"🖼", "Pictures", "Image gallery", "4.7 GB"},
		{"🎬", "Videos", "Video collection", "28.1 GB"},
		{"💻", "Projects", "Code repositories", "1.2 GB"},
	}

	// Title with item count
	selectedCount := 0
	for _, v := range l.checked {
		if v {
			selectedCount++
		}
	}
	titleStyle := backend.DefaultStyle().Bold(true).Foreground(backend.ColorBrightCyan)
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "File Manager", titleStyle)
	if selectedCount > 0 {
		ctx.Buffer.SetString(bounds.X+16, bounds.Y+1, fmt.Sprintf("(%d selected)", selectedCount), backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	}

	// Column headers
	y := bounds.Y + 3
	headerStyle := backend.DefaultStyle().Bold(true).Underline(true)
	ctx.Buffer.SetString(bounds.X+5, y, "Name", headerStyle)
	ctx.Buffer.SetString(bounds.X+25, y, "Description", headerStyle)
	ctx.Buffer.SetString(bounds.X+50, y, "Size", headerStyle)
	y++

	// Items
	for i, item := range items {
		y++
		isSelected := i == l.selected
		isChecked := l.checked[i]

		// Checkbox
		checkStyle := backend.DefaultStyle()
		checkBox := "○"
		if isChecked {
			checkBox = "●"
			checkStyle = checkStyle.Foreground(backend.ColorBrightGreen)
		}
		ctx.Buffer.SetString(bounds.X+2, y, checkBox, checkStyle)

		// Row content
		rowStyle := backend.DefaultStyle()
		if isSelected {
			rowStyle = rowStyle.Background(backend.ColorBlue).Foreground(backend.ColorBrightWhite)
			// Draw selection background
			for dx := 4; dx < bounds.Width-3; dx++ {
				ctx.Buffer.Set(bounds.X+dx, y, ' ', rowStyle)
			}
		}

		ctx.Buffer.SetString(bounds.X+5, y, item.icon+" "+item.label, rowStyle)
		ctx.Buffer.SetString(bounds.X+25, y, item.desc, rowStyle.Dim(!isSelected))
		ctx.Buffer.SetString(bounds.X+50, y, item.size, rowStyle.Foreground(backend.ColorBrightYellow))

		// Selection indicator
		if isSelected {
			ctx.Buffer.SetString(bounds.X+bounds.Width-4, y, "◀", rowStyle.Foreground(backend.ColorBrightCyan))
		}
	}

	// Footer with keyboard hints
	y = bounds.Y + bounds.Height - 3
	ctx.Buffer.SetString(bounds.X+2, y, "↑↓", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+5, y, "Navigate", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+16, y, "Space", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+22, y, "Select", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+31, y, "Enter", backend.DefaultStyle().Foreground(backend.ColorBrightYellow))
	ctx.Buffer.SetString(bounds.X+37, y, "Open", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (l *listDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		l.frame++

		// Navigate through list
		if l.frame%10 == 0 {
			l.selected = (l.selected + 1) % 6
		}

		// Toggle selection periodically
		if l.frame%15 == 0 && l.frame > 30 {
			l.checked[l.selected] = !l.checked[l.selected]
		}

		l.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoDialog() runtime.Widget {
	return &dialogDemo{}
}

type dialogDemo struct {
	widgets.Component
	frame       int
	phase       int // 0=appearing, 1=focus cancel, 2=focus confirm, 3=confirming, 4=success, 5=fade out
	dialogScale float64
	focusedBtn  int
}

func (d *dialogDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *dialogDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
}

func (d *dialogDemo) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Draw dimmed background with some content
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Main Application", backend.DefaultStyle().Dim(true))
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+2, "───────────────", backend.DefaultStyle().Dim(true))
	for i := 3; i < bounds.Height-1; i++ {
		ctx.Buffer.SetString(bounds.X+2, bounds.Y+i, "Lorem ipsum content...", backend.DefaultStyle().Dim(true))
	}

	// Dialog box (centered)
	if d.phase >= 0 && d.phase <= 4 {
		dialogW, dialogH := 40, 9
		dialogX := bounds.X + (bounds.Width-dialogW)/2
		dialogY := bounds.Y + (bounds.Height-dialogH)/2

		// Draw dialog shadow
		shadowStyle := backend.DefaultStyle().Dim(true)
		for dy := 1; dy <= dialogH; dy++ {
			ctx.Buffer.SetString(dialogX+dialogW, dialogY+dy, "░", shadowStyle)
		}
		for dx := 1; dx <= dialogW; dx++ {
			ctx.Buffer.SetString(dialogX+dx, dialogY+dialogH, "░", shadowStyle)
		}

		// Draw dialog box
		boxStyle := backend.DefaultStyle()
		if d.phase == 4 {
			boxStyle = boxStyle.Foreground(backend.ColorBrightGreen)
		}
		for dy := 0; dy < dialogH; dy++ {
			for dx := 0; dx < dialogW; dx++ {
				ctx.Buffer.Set(dialogX+dx, dialogY+dy, ' ', boxStyle.Background(backend.ColorBlack))
			}
		}

		// Dialog border
		borderColor := backend.ColorBrightBlue
		if d.phase == 4 {
			borderColor = backend.ColorBrightGreen
		}
		borderStyle := backend.DefaultStyle().Foreground(borderColor).Background(backend.ColorBlack)
		ctx.Buffer.SetString(dialogX, dialogY, "╔"+strings.Repeat("═", dialogW-2)+"╗", borderStyle)
		ctx.Buffer.SetString(dialogX, dialogY+dialogH-1, "╚"+strings.Repeat("═", dialogW-2)+"╝", borderStyle)
		for dy := 1; dy < dialogH-1; dy++ {
			ctx.Buffer.Set(dialogX, dialogY+dy, '║', borderStyle)
			ctx.Buffer.Set(dialogX+dialogW-1, dialogY+dy, '║', borderStyle)
		}

		contentStyle := backend.DefaultStyle().Background(backend.ColorBlack).Foreground(backend.ColorWhite)

		if d.phase < 4 {
			// Title
			title := " Confirm Action "
			ctx.Buffer.SetString(dialogX+(dialogW-len(title))/2, dialogY, title, borderStyle.Bold(true))

			// Message
			ctx.Buffer.SetString(dialogX+3, dialogY+2, "Are you sure you want to proceed?", contentStyle)
			ctx.Buffer.SetString(dialogX+3, dialogY+3, "This action cannot be undone.", contentStyle.Dim(true))

			// Buttons
			cancelStyle := contentStyle
			confirmStyle := contentStyle
			if d.focusedBtn == 0 {
				cancelStyle = backend.DefaultStyle().Background(backend.ColorBrightWhite).Foreground(backend.ColorBlack).Bold(true)
			} else {
				confirmStyle = backend.DefaultStyle().Background(backend.ColorBrightGreen).Foreground(backend.ColorBlack).Bold(true)
			}

			// Loading state for confirm
			if d.phase == 3 {
				spinChars := []string{"◐", "◓", "◑", "◒"}
				spinner := spinChars[(d.frame/4)%len(spinChars)]
				confirmStyle = backend.DefaultStyle().Background(backend.ColorBrightYellow).Foreground(backend.ColorBlack)
				ctx.Buffer.SetString(dialogX+23, dialogY+6, fmt.Sprintf(" %s Processing ", spinner), confirmStyle)
			} else {
				ctx.Buffer.SetString(dialogX+6, dialogY+6, " Cancel ", cancelStyle)
				ctx.Buffer.SetString(dialogX+23, dialogY+6, " Confirm ", confirmStyle)
			}

			// Focus hint
			ctx.Buffer.SetString(dialogX+3, dialogY+dialogH-2, "Tab: switch  Enter: select", contentStyle.Dim(true))
		} else {
			// Success state
			title := " Success! "
			ctx.Buffer.SetString(dialogX+(dialogW-len(title))/2, dialogY, title, borderStyle.Bold(true))
			ctx.Buffer.SetString(dialogX+10, dialogY+3, "✓ Action completed!", contentStyle.Foreground(backend.ColorBrightGreen).Bold(true))
		}
	}
}

func (d *dialogDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		d.frame++

		switch d.phase {
		case 0: // Appearing
			if d.frame > 15 {
				d.phase = 1
			}
		case 1: // Focus on cancel
			if d.frame%30 == 0 {
				d.focusedBtn = 1
				d.phase = 2
			}
		case 2: // Focus on confirm
			if d.frame%25 == 0 {
				d.phase = 3
			}
		case 3: // Processing
			if d.frame%40 == 0 {
				d.phase = 4
			}
		case 4: // Success
			if d.frame%50 == 0 {
				d.phase = 5
			}
		case 5: // Reset
			if d.frame%20 == 0 {
				d.phase = 0
				d.focusedBtn = 0
			}
		}

		d.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoSparkline() runtime.Widget {
	data1 := state.NewSignal(make([]float64, 40))
	data2 := state.NewSignal(make([]float64, 40))
	data3 := state.NewSignal(make([]float64, 40))

	// Initialize with some data
	initData := func(sig *state.Signal[[]float64], offset float64) {
		sig.Update(func(d []float64) []float64 {
			for i := range d {
				d[i] = 50 + 30*math.Sin(float64(i)*0.2+offset)
			}
			return d
		})
	}
	initData(data1, 0)
	initData(data2, math.Pi/3)
	initData(data3, 2*math.Pi/3)

	return &sparklineDemo{data1: data1, data2: data2, data3: data3}
}

type sparklineDemo struct {
	widgets.Component
	data1 *state.Signal[[]float64]
	data2 *state.Signal[[]float64]
	data3 *state.Signal[[]float64]
	frame int
}

func (s *sparklineDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (s *sparklineDemo) Layout(bounds runtime.Rect) {
	s.Component.Layout(bounds)
}

func (s *sparklineDemo) Render(ctx runtime.RenderContext) {
	bounds := s.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Animated title
	titleColor := []backend.Color{backend.ColorBrightCyan, backend.ColorBrightGreen, backend.ColorBrightMagenta}[(s.frame/10)%3]
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Live Data Visualization", backend.DefaultStyle().Bold(true).Foreground(titleColor))

	// CPU sparkline
	y := bounds.Y + 3
	ctx.Buffer.SetString(bounds.X+2, y, "CPU  ", backend.DefaultStyle().Foreground(backend.ColorCyan))
	sparkline1 := widgets.NewSparkline(s.data1)
	sparkline1.Layout(runtime.Rect{X: bounds.X + 8, Y: y, Width: bounds.Width - 20, Height: 1})
	sparkline1.Render(ctx)
	// Current value
	data1 := s.data1.Get()
	ctx.Buffer.SetString(bounds.X+bounds.Width-10, y, fmt.Sprintf("%5.1f%%", data1[len(data1)-1]), backend.DefaultStyle().Foreground(backend.ColorCyan))

	// Memory sparkline
	y += 3
	ctx.Buffer.SetString(bounds.X+2, y, "MEM  ", backend.DefaultStyle().Foreground(backend.ColorGreen))
	sparkline2 := widgets.NewSparkline(s.data2)
	sparkline2.Layout(runtime.Rect{X: bounds.X + 8, Y: y, Width: bounds.Width - 20, Height: 1})
	sparkline2.Render(ctx)
	data2 := s.data2.Get()
	ctx.Buffer.SetString(bounds.X+bounds.Width-10, y, fmt.Sprintf("%5.1f%%", data2[len(data2)-1]), backend.DefaultStyle().Foreground(backend.ColorGreen))

	// Network sparkline
	y += 3
	ctx.Buffer.SetString(bounds.X+2, y, "NET  ", backend.DefaultStyle().Foreground(backend.ColorMagenta))
	sparkline3 := widgets.NewSparkline(s.data3)
	sparkline3.Layout(runtime.Rect{X: bounds.X + 8, Y: y, Width: bounds.Width - 20, Height: 1})
	sparkline3.Render(ctx)
	data3 := s.data3.Get()
	ctx.Buffer.SetString(bounds.X+bounds.Width-10, y, fmt.Sprintf("%5.1f%%", data3[len(data3)-1]), backend.DefaultStyle().Foreground(backend.ColorMagenta))

	// Stats summary
	y += 3
	ctx.Buffer.SetString(bounds.X+2, y, "Statistics:", backend.DefaultStyle().Bold(true))
	y++

	// Calculate stats
	calcStats := func(data []float64) (min, max, avg float64) {
		min, max = data[0], data[0]
		sum := 0.0
		for _, v := range data {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
			sum += v
		}
		avg = sum / float64(len(data))
		return
	}

	min1, max1, avg1 := calcStats(data1)
	ctx.Buffer.SetString(bounds.X+4, y, fmt.Sprintf("CPU  - Min: %5.1f  Max: %5.1f  Avg: %5.1f", min1, max1, avg1), backend.DefaultStyle().Dim(true))
	y++
	min2, max2, avg2 := calcStats(data2)
	ctx.Buffer.SetString(bounds.X+4, y, fmt.Sprintf("MEM  - Min: %5.1f  Max: %5.1f  Avg: %5.1f", min2, max2, avg2), backend.DefaultStyle().Dim(true))
	y++
	min3, max3, avg3 := calcStats(data3)
	ctx.Buffer.SetString(bounds.X+4, y, fmt.Sprintf("NET  - Min: %5.1f  Max: %5.1f  Avg: %5.1f", min3, max3, avg3), backend.DefaultStyle().Dim(true))

	// Update indicator
	indicator := []string{"◐", "◓", "◑", "◒"}[s.frame%4]
	ctx.Buffer.SetString(bounds.X+bounds.Width-4, bounds.Y+1, indicator, backend.DefaultStyle().Foreground(backend.ColorYellow))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (s *sparklineDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		s.frame++
		if s.frame%3 == 0 {
			// Update data with smooth wave patterns
			updateData := func(sig *state.Signal[[]float64], offset float64) {
				sig.Update(func(d []float64) []float64 {
					// Shift left
					copy(d, d[1:])
					// Add new value with some noise
					newVal := 50 + 30*math.Sin(float64(s.frame)*0.1+offset) + float64(rand.Intn(10)-5)
					if newVal < 0 {
						newVal = 0
					}
					if newVal > 100 {
						newVal = 100
					}
					d[len(d)-1] = newVal
					return d
				})
			}
			updateData(s.data1, 0)
			updateData(s.data2, math.Pi/3)
			updateData(s.data3, 2*math.Pi/3)
			s.Invalidate()
		}
		s.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoTabs() runtime.Widget {
	return &tabsDemo{}
}

type tabsDemo struct {
	widgets.Component
	frame    int
	activeTab int
}

func (t *tabsDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (t *tabsDemo) Layout(bounds runtime.Rect) {
	t.Component.Layout(bounds)
}

func (t *tabsDemo) Render(ctx runtime.RenderContext) {
	bounds := t.Bounds()
	ctx.Clear(backend.DefaultStyle())

	tabs := []struct {
		title   string
		content string
		icon    string
	}{
		{"Overview", "Welcome to FluffyUI! A batteries-included TUI framework.", "◉"},
		{"Features", "35+ widgets, reactive state, accessibility built-in.", "★"},
		{"Install", "go get github.com/odvcencio/fluffy-ui", "↓"},
		{"Docs", "Comprehensive guides and examples included.", "📖"},
	}

	// Draw tab headers
	x := bounds.X + 2
	y := bounds.Y + 1
	for i, tab := range tabs {
		style := backend.DefaultStyle()
		if i == t.activeTab {
			style = style.Bold(true).Foreground(backend.ColorCyan).Underline(true)
		} else {
			style = style.Dim(true)
		}
		label := fmt.Sprintf(" %s %s ", tab.icon, tab.title)
		ctx.Buffer.SetString(x, y, label, style)
		x += len(label) + 1
	}

	// Draw separator
	y += 2
	for i := bounds.X + 1; i < bounds.X+bounds.Width-1; i++ {
		ctx.Buffer.SetString(i, y, "─", backend.DefaultStyle().Dim(true))
	}

	// Draw active tab content with animation
	y += 2
	activeContent := tabs[t.activeTab].content
	// Typing animation
	visibleChars := (t.frame % 60) * 2
	if visibleChars > len(activeContent) {
		visibleChars = len(activeContent)
	}
	ctx.Buffer.SetString(bounds.X+4, y, activeContent[:visibleChars], backend.DefaultStyle())

	// Draw cursor if still typing
	if visibleChars < len(activeContent) {
		ctx.Buffer.SetString(bounds.X+4+visibleChars, y, "▌", backend.DefaultStyle().Foreground(backend.ColorCyan))
	}

	// Navigation hint
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+bounds.Height-2, "← → to switch tabs", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (t *tabsDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		t.frame++
		// Switch tabs every ~2 seconds
		if t.frame%60 == 0 {
			t.activeTab = (t.activeTab + 1) % 4
			t.Invalidate()
		}
		t.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// =============================================================================
// Input Demo - Text input with validation states
// =============================================================================

func demoInput() runtime.Widget {
	return &inputDemo{}
}

type inputDemo struct {
	widgets.Component
	frame      int
	typedText  string
	passLen    int
	rememberMe bool
	focusField int // 0=email, 1=password, 2=checkbox, 3=submit
	phase      int // 0=email, 1=password, 2=checkbox, 3=submit, 4=loading, 5=success, 6=reset
}

var inputDemoText = "user@example.com"

func (d *inputDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *inputDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
}

func (d *inputDemo) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()
	ctx.Clear(backend.DefaultStyle())

	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Sign Up", backend.DefaultStyle().Bold(true).Foreground(backend.ColorBrightCyan))

	inputWidth := 28

	// Email field
	y := bounds.Y + 3
	labelStyle := backend.DefaultStyle()
	if d.focusField == 0 {
		labelStyle = labelStyle.Foreground(backend.ColorBrightCyan).Bold(true)
	}
	ctx.Buffer.SetString(bounds.X+2, y, "Email:", labelStyle)

	// Email styling
	emailStyle := backend.DefaultStyle().Background(backend.ColorBrightWhite).Foreground(backend.ColorBlack)
	emailBorder := backend.ColorWhite
	if d.focusField == 0 {
		emailBorder = backend.ColorBrightCyan
	}
	if d.phase >= 1 {
		emailStyle = backend.DefaultStyle().Background(backend.ColorBrightGreen).Foreground(backend.ColorBlack)
		emailBorder = backend.ColorBrightGreen
	}

	ctx.Buffer.SetString(bounds.X+12, y, "[", backend.DefaultStyle().Foreground(emailBorder).Bold(true))
	ctx.Buffer.SetString(bounds.X+13+inputWidth, y, "]", backend.DefaultStyle().Foreground(emailBorder).Bold(true))
	emailText := d.typedText
	for len(emailText) < inputWidth {
		emailText += " "
	}
	ctx.Buffer.SetString(bounds.X+13, y, emailText, emailStyle)

	if d.focusField == 0 && d.phase == 0 && (d.frame/10)%2 == 0 {
		ctx.Buffer.SetString(bounds.X+13+len(d.typedText), y, "▌", emailStyle.Foreground(backend.ColorBrightYellow))
	}

	// Email validation
	if d.phase >= 1 {
		ctx.Buffer.SetString(bounds.X+44, y, "✓", backend.DefaultStyle().Foreground(backend.ColorBrightGreen).Bold(true))
	}

	// Password field
	y += 2
	labelStyle = backend.DefaultStyle()
	if d.focusField == 1 {
		labelStyle = labelStyle.Foreground(backend.ColorBrightCyan).Bold(true)
	}
	ctx.Buffer.SetString(bounds.X+2, y, "Password:", labelStyle)

	passBorder := backend.ColorWhite
	if d.focusField == 1 {
		passBorder = backend.ColorBrightCyan
	}
	if d.phase >= 2 && d.passLen >= 8 {
		passBorder = backend.ColorBrightGreen
	}

	ctx.Buffer.SetString(bounds.X+12, y, "[", backend.DefaultStyle().Foreground(passBorder).Bold(true))
	ctx.Buffer.SetString(bounds.X+13+inputWidth, y, "]", backend.DefaultStyle().Foreground(passBorder).Bold(true))

	passStyle := backend.DefaultStyle().Background(backend.ColorBrightWhite).Foreground(backend.ColorBlack)
	for i := 0; i < inputWidth; i++ {
		ctx.Buffer.Set(bounds.X+13+i, y, ' ', passStyle)
	}
	for i := 0; i < d.passLen; i++ {
		ctx.Buffer.Set(bounds.X+13+i, y, '●', passStyle)
	}

	if d.focusField == 1 && d.phase == 1 && (d.frame/10)%2 == 0 && d.passLen < inputWidth {
		ctx.Buffer.SetString(bounds.X+13+d.passLen, y, "▌", passStyle.Foreground(backend.ColorBrightYellow))
	}

	// Password strength
	if d.passLen > 0 {
		strengthColors := []backend.Color{backend.ColorBrightRed, backend.ColorBrightYellow, backend.ColorBrightGreen}
		strengthIdx := (d.passLen - 1) / 4
		if strengthIdx > 2 {
			strengthIdx = 2
		}
		for i := 0; i < 3; i++ {
			char := "○"
			style := backend.DefaultStyle().Dim(true)
			if i <= strengthIdx {
				char = "●"
				style = backend.DefaultStyle().Foreground(strengthColors[strengthIdx])
			}
			ctx.Buffer.SetString(bounds.X+44+i*2, y, char, style)
		}
	}

	// Remember me checkbox
	y += 2
	checkStyle := backend.DefaultStyle()
	if d.focusField == 2 {
		checkStyle = checkStyle.Foreground(backend.ColorBrightCyan).Bold(true)
	}
	checkBox := "☐"
	if d.rememberMe {
		checkBox = "☑"
		checkStyle = checkStyle.Foreground(backend.ColorBrightGreen)
	}
	ctx.Buffer.SetString(bounds.X+12, y, checkBox, checkStyle)
	ctx.Buffer.SetString(bounds.X+14, y, "Remember me", backend.DefaultStyle())

	// Submit button
	y += 2
	btnStyle := backend.DefaultStyle().Background(backend.ColorBlue).Foreground(backend.ColorWhite)
	btnText := "  Sign Up  "
	if d.focusField == 3 {
		btnStyle = backend.DefaultStyle().Background(backend.ColorBrightCyan).Foreground(backend.ColorBlack).Bold(true)
	}
	if d.phase == 4 {
		spinChars := []string{"◐", "◓", "◑", "◒"}
		btnText = fmt.Sprintf(" %s Signing up... ", spinChars[(d.frame/4)%len(spinChars)])
		btnStyle = backend.DefaultStyle().Background(backend.ColorBrightYellow).Foreground(backend.ColorBlack)
	} else if d.phase == 5 {
		btnText = " ✓ Success! "
		btnStyle = backend.DefaultStyle().Background(backend.ColorBrightGreen).Foreground(backend.ColorBlack).Bold(true)
	}
	ctx.Buffer.SetString(bounds.X+12, y, btnText, btnStyle)

	// Focus indicator
	focusY := bounds.Y + 3 + d.focusField*2
	if d.phase < 4 {
		ctx.Buffer.SetString(bounds.X+10, focusY, "▶", backend.DefaultStyle().Foreground(backend.ColorBrightCyan))
	}

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (d *inputDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		d.frame++

		switch d.phase {
		case 0: // Typing email (focus on email field)
			d.focusField = 0
			if d.frame%3 == 0 && len(d.typedText) < len(inputDemoText) {
				d.typedText = inputDemoText[:len(d.typedText)+1]
			}
			if len(d.typedText) == len(inputDemoText) && d.frame%15 == 0 {
				d.phase = 1
				d.focusField = 1
			}

		case 1: // Typing password (focus on password field)
			if d.frame%4 == 0 && d.passLen < 10 {
				d.passLen++
			}
			if d.passLen >= 10 && d.frame%15 == 0 {
				d.phase = 2
				d.focusField = 2
			}

		case 2: // Toggle checkbox
			if d.frame%20 == 0 {
				d.rememberMe = true
				d.phase = 3
				d.focusField = 3
			}

		case 3: // Focus on submit, then click
			if d.frame%25 == 0 {
				d.phase = 4
			}

		case 4: // Loading/submitting
			if d.frame%40 == 0 {
				d.phase = 5
			}

		case 5: // Success
			if d.frame%50 == 0 {
				d.phase = 6
			}

		case 6: // Reset
			d.typedText = ""
			d.passLen = 0
			d.rememberMe = false
			d.focusField = 0
			d.phase = 0
		}

		d.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// =============================================================================
// Graphics Demo - Canvas shapes, curves, and effects
// =============================================================================

func demoGraphics() runtime.Widget {
	return &graphicsDemo{
		blitter: &graphics.BrailleBlitter{},
	}
}

type graphicsDemo struct {
	widgets.Component
	canvas  *graphics.Canvas
	blitter graphics.Blitter
	frame   int
}

func (d *graphicsDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *graphicsDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	d.ensureCanvas(bounds)
}

func (d *graphicsDemo) ensureCanvas(bounds runtime.Rect) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	if d.canvas == nil {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
		return
	}
	cellW, cellH := d.canvas.CellSize()
	if cellW != bounds.Width || cellH != bounds.Height {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
	}
}

func (d *graphicsDemo) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()
	ctx.Clear(backend.DefaultStyle())
	d.ensureCanvas(bounds)
	canvas := d.canvas
	if canvas == nil {
		return
	}
	canvas.Clear()

	w, h := canvas.Size()
	if w == 0 || h == 0 {
		return
	}

	phase := float64(d.frame) * 0.05

	// Draw rotating circles
	centerX, centerY := w/2, h/2
	for i := 0; i < 6; i++ {
		angle := phase + float64(i)*math.Pi/3
		radius := 40 + int(20*math.Sin(phase*0.5))
		x := centerX + int(float64(radius)*math.Cos(angle))
		y := centerY + int(float64(radius)*math.Sin(angle))

		colors := []backend.Color{
			backend.ColorRGB(255, 100, 100),
			backend.ColorRGB(255, 200, 100),
			backend.ColorRGB(100, 255, 100),
			backend.ColorRGB(100, 255, 255),
			backend.ColorRGB(100, 100, 255),
			backend.ColorRGB(255, 100, 255),
		}
		canvas.SetFillColor(colors[i])
		canvas.FillCircle(x, y, 15+int(5*math.Sin(phase+float64(i))))
	}

	// Draw bezier curves
	canvas.SetStrokeColor(backend.ColorRGB(255, 255, 100))
	for i := 0; i < 3; i++ {
		offset := float64(i) * 30
		p0 := graphics.Point{X: 20, Y: int(50 + offset)}
		p1 := graphics.Point{X: 60, Y: int(20 + offset + 30*math.Sin(phase))}
		p2 := graphics.Point{X: 100, Y: int(80 + offset + 30*math.Cos(phase))}
		p3 := graphics.Point{X: 140, Y: int(50 + offset)}
		canvas.DrawBezier(p0, p1, p2, p3)
	}

	// Draw spinning rectangle
	canvas.Save()
	canvas.Translate(w-60, h/2)
	canvas.Rotate(phase)
	canvas.SetStrokeColor(backend.ColorRGB(100, 200, 255))
	canvas.DrawRect(-20, -20, 40, 40)
	canvas.Restore()

	// Draw wave at bottom
	canvas.SetStrokeColor(backend.ColorRGB(100, 255, 200))
	prevX, prevY := 0, 0
	for x := 0; x < w; x += 2 {
		y := h - 30 + int(10*math.Sin(float64(x)*0.05+phase))
		if x > 0 {
			canvas.DrawLine(prevX, prevY, x, y)
		}
		prevX, prevY = x, y
	}

	// Draw triangle
	canvas.SetFillColor(backend.ColorRGB(255, 150, 100))
	triOffset := int(20 * math.Sin(phase))
	canvas.FillTriangle(
		graphics.Point{X: w - 40, Y: 30 + triOffset},
		graphics.Point{X: w - 60, Y: 60 + triOffset},
		graphics.Point{X: w - 20, Y: 60 + triOffset},
	)

	canvas.Render(ctx.Buffer, bounds.X, bounds.Y)

	// Render title as crisp terminal text
	title := "✦ CANVAS GRAPHICS DEMO ✦"
	titleStyle := backend.DefaultStyle().Bold(true).Foreground(backend.ColorRGB(255, 220, 100))
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, title, titleStyle)

	subtitle := "Shapes, Curves, Transforms"
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+2, subtitle, backend.DefaultStyle().Dim(true))
}

func (d *graphicsDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		d.frame++
		d.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// =============================================================================
// Easing Demo - Animation easing functions visualization
// =============================================================================

func demoEasing() runtime.Widget {
	return &easingDemo{
		blitter: &graphics.BrailleBlitter{},
	}
}

type easingDemo struct {
	widgets.Component
	canvas  *graphics.Canvas
	blitter graphics.Blitter
	frame   int
}

func (d *easingDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *easingDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	d.ensureCanvas(bounds)
}

func (d *easingDemo) ensureCanvas(bounds runtime.Rect) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	if d.canvas == nil {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
		return
	}
	cellW, cellH := d.canvas.CellSize()
	if cellW != bounds.Width || cellH != bounds.Height {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
	}
}

type easingEntry struct {
	name   string
	fn     func(float64) float64
	color  backend.Color
}

var easingFuncs = []easingEntry{
	{"Linear", func(t float64) float64 { return t }, backend.ColorRGB(255, 255, 255)},
	{"OutQuad", func(t float64) float64 { return t * (2 - t) }, backend.ColorRGB(255, 100, 100)},
	{"OutCubic", func(t float64) float64 { t--; return t*t*t + 1 }, backend.ColorRGB(100, 255, 100)},
	{"OutElastic", func(t float64) float64 {
		if t == 0 || t == 1 {
			return t
		}
		return math.Pow(2, -10*t)*math.Sin((t-0.1)*5*math.Pi) + 1
	}, backend.ColorRGB(100, 100, 255)},
	{"OutBounce", func(t float64) float64 {
		if t < 1/2.75 {
			return 7.5625 * t * t
		} else if t < 2/2.75 {
			t -= 1.5 / 2.75
			return 7.5625*t*t + 0.75
		} else if t < 2.5/2.75 {
			t -= 2.25 / 2.75
			return 7.5625*t*t + 0.9375
		}
		t -= 2.625 / 2.75
		return 7.5625*t*t + 0.984375
	}, backend.ColorRGB(255, 200, 100)},
}

func (d *easingDemo) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()
	ctx.Clear(backend.DefaultStyle())
	d.ensureCanvas(bounds)
	canvas := d.canvas
	if canvas == nil {
		return
	}
	canvas.Clear()

	w, h := canvas.Size()
	if w == 0 || h == 0 {
		return
	}

	// Calculate animation progress (loop every 90 frames = 3 seconds)
	loopFrames := 90
	progress := float64(d.frame%loopFrames) / float64(loopFrames)

	graphWidth := w - 100
	graphHeight := h - 60
	graphX := 80
	graphY := 30

	// Draw graph background grid
	canvas.SetStrokeColor(backend.ColorRGB(50, 50, 50))
	for i := 0; i <= 10; i++ {
		y := graphY + i*graphHeight/10
		canvas.DrawLine(graphX, y, graphX+graphWidth, y)
	}
	for i := 0; i <= 10; i++ {
		x := graphX + i*graphWidth/10
		canvas.DrawLine(x, graphY, x, graphY+graphHeight)
	}

	// Draw easing curves
	for _, entry := range easingFuncs {
		canvas.SetStrokeColor(entry.color)
		prevX, prevY := 0, 0
		for i := 0; i <= graphWidth; i += 2 {
			t := float64(i) / float64(graphWidth)
			value := entry.fn(t)
			x := graphX + i
			y := graphY + graphHeight - int(value*float64(graphHeight))
			if i > 0 {
				canvas.DrawLine(prevX, prevY, x, y)
			}
			prevX, prevY = x, y
		}
	}

	// Draw animated balls showing current position
	ballX := graphX + int(progress*float64(graphWidth))
	for i, entry := range easingFuncs {
		value := entry.fn(progress)
		ballY := graphY + graphHeight - int(value*float64(graphHeight))

		// Draw ball
		canvas.SetFillColor(entry.color)
		canvas.FillCircle(ballX, ballY, 4)

		// Draw horizontal position indicator
		indicatorX := graphX + graphWidth + 20 + i*30
		indicatorY := graphY + graphHeight - int(value*float64(graphHeight))
		canvas.FillCircle(indicatorX, indicatorY, 6)
	}

	canvas.Render(ctx.Buffer, bounds.X, bounds.Y)

	// Render text labels
	title := "✦ EASING FUNCTIONS ✦"
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, title, backend.DefaultStyle().Bold(true).Foreground(backend.ColorRGB(255, 220, 100)))

	// Legend
	legendY := bounds.Y + 3
	for i, entry := range easingFuncs {
		style := backend.DefaultStyle().Foreground(entry.color)
		ctx.Buffer.SetString(bounds.X+2, legendY+i, fmt.Sprintf("● %s", entry.name), style)
	}

	// Progress indicator
	progressBar := ""
	barWidth := 30
	filled := int(progress * float64(barWidth))
	for i := 0; i < barWidth; i++ {
		if i < filled {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+bounds.Height-2, progressBar, backend.DefaultStyle().Foreground(backend.ColorCyan))
	ctx.Buffer.SetString(bounds.X+34, bounds.Y+bounds.Height-2, fmt.Sprintf("%3.0f%%", progress*100), backend.DefaultStyle())
}

func (d *easingDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		d.frame++
		d.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoHero() runtime.Widget {
	return &heroDemo{}
}

type heroDemo struct {
	widgets.Component
	frame int
}

func (h *heroDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (h *heroDemo) Layout(bounds runtime.Rect) {
	h.Component.Layout(bounds)
}

// Rainbow colors for the rotating border
var rainbowColors = []backend.Color{
	backend.ColorBrightRed,
	backend.ColorBrightYellow,
	backend.ColorBrightGreen,
	backend.ColorBrightCyan,
	backend.ColorBrightBlue,
	backend.ColorBrightMagenta,
}

// Border characters - stars, sparkles, diamonds
var borderChars = []rune{'★', '✦', '◆', '✧', '❖', '✶', '◇', '✴', '❋', '✸'}

func (h *heroDemo) Render(ctx runtime.RenderContext) {
	bounds := h.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Draw rotating rainbow border
	h.drawRainbowBorder(ctx, bounds)

	// ASCII art title with color
	title := []string{
		" _____ _       __  __       _   _ ___ ",
		"|  ___| |_   _ / _|/ _|_   _| | | |_ _|",
		"| |_  | | | | | |_| |_| | | | | | || | ",
		"|  _| | | |_| |  _|  _| |_| | |_| || | ",
		"|_|   |_|\\__,_|_| |_|  \\__, |\\___/|___|",
		"                      |___/           ",
	}

	startY := bounds.Y + 3
	titleColor := rainbowColors[(h.frame/8)%len(rainbowColors)]
	for i, line := range title {
		x := (bounds.Width - len(line)) / 2
		style := backend.DefaultStyle().Bold(true).Foreground(titleColor)
		ctx.Buffer.SetString(x, startY+i, line, style)
	}

	// Subtitle
	subtitle := "A batteries-included TUI framework for Go"
	x := (bounds.Width - len(subtitle)) / 2
	ctx.Buffer.SetString(x, startY+7, subtitle, backend.DefaultStyle().Dim(true))

	// Features (animated) with colorful bullets
	features := []string{
		"35+ Ready-to-Use Widgets",
		"Reactive State Management",
		"Accessibility Built-In",
		"Deterministic Testing",
	}

	featureY := startY + 10
	visibleFeatures := h.frame / 10
	if visibleFeatures > len(features) {
		visibleFeatures = len(features)
	}

	for i := 0; i < visibleFeatures; i++ {
		fx := (bounds.Width - len(features[i]) - 4) / 2
		bulletColor := rainbowColors[(i+h.frame/10)%len(rainbowColors)]
		bulletStyle := backend.DefaultStyle().Foreground(bulletColor).Bold(true)
		textStyle := backend.DefaultStyle()
		ctx.Buffer.SetString(fx, featureY+i, "★ ", bulletStyle)
		ctx.Buffer.SetString(fx+2, featureY+i, features[i], textStyle)
	}

	// Install command with pulsing highlight
	installY := bounds.Y + bounds.Height - 3
	install := " go get github.com/odvcencio/fluffy-ui "
	ix := (bounds.Width - len(install)) / 2
	installColor := rainbowColors[(h.frame/5)%len(rainbowColors)]
	ctx.Buffer.SetString(ix, installY, install, backend.DefaultStyle().Background(installColor).Foreground(backend.ColorBlack).Bold(true))
}

func (h *heroDemo) drawRainbowBorder(ctx runtime.RenderContext, bounds runtime.Rect) {
	width := bounds.Width
	height := bounds.Height
	if width <= 0 || height <= 0 {
		return
	}
	patternLen := len(borderChars) * len(rainbowColors)
	if patternLen == 0 {
		return
	}

	draw := func(x, y, offset int) {
		pos := (offset + h.frame) % patternLen
		char := borderChars[pos%len(borderChars)]
		color := rainbowColors[(pos/2)%len(rainbowColors)]
		style := backend.DefaultStyle().Foreground(color).Bold(true)
		ctx.Buffer.Set(x, y, char, style)
	}

	// Top edge (left to right)
	for i := 0; i < width; i++ {
		draw(bounds.X+i, bounds.Y, i)
	}

	sideLen := height - 2
	if sideLen < 0 {
		sideLen = 0
	}

	// Right edge (top to bottom)
	for i := 0; i < sideLen; i++ {
		draw(bounds.X+width-1, bounds.Y+1+i, width+i)
	}

	// Bottom edge (right to left)
	if height > 1 {
		base := width + sideLen
		for i := 0; i < width; i++ {
			x := bounds.X + width - 1 - i
			draw(x, bounds.Y+height-1, base+i)
		}
	}

	// Left edge (bottom to top)
	if width > 1 {
		base := width + sideLen + width
		for i := 0; i < sideLen; i++ {
			y := bounds.Y + height - 2 - i
			draw(bounds.X, y, base+i)
		}
	}
}

func (h *heroDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		h.frame++
		h.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =============================================================================
// Fireworks Demo - 3D particle effects with perspective projection
// =============================================================================

func demoFireworks() runtime.Widget {
	return &fireworksDemo{
		blitter:    &graphics.BrailleBlitter{},
		spawnDelay: 500 * time.Millisecond,
	}
}

type fireworksDemo struct {
	widgets.Component
	canvas     *graphics.Canvas
	blitter    graphics.Blitter
	fireworks  []*firework3D
	cameraZ    float64
	frame      int
	lastSpawn  time.Time
	spawnDelay time.Duration
}

type particle3D struct {
	x, y, z       float64
	vx, vy, vz    float64
	r, g, b       uint8
	life, maxLife float64
}

func (p *particle3D) update(dt, gravity, airResistance float64) {
	p.vy += gravity * dt
	p.vx *= airResistance
	p.vy *= airResistance
	p.vz *= airResistance
	p.x += p.vx * dt
	p.y += p.vy * dt
	p.z += p.vz * dt
	p.life += dt
}

func (p *particle3D) isAlive() bool { return p.life < p.maxLife }

func (p *particle3D) alpha() float32 {
	return float32(1.0 - p.life/p.maxLife)
}

func (p *particle3D) project(cameraZ, cameraDist, centerX, centerY float64) (int, int, bool) {
	zRel := p.z - cameraZ
	zOffset := zRel + cameraDist
	if zOffset <= 0 {
		return 0, 0, false
	}
	scale := cameraDist / zOffset
	return int(centerX + (p.x-centerX)*scale), int(centerY + (p.y-centerY)*scale), true
}

type firework3D struct {
	x, y, z       float64
	vx, vy, vz    float64
	r, g, b       uint8
	exploded      bool
	particles     []particle3D
	trail         [][2]float64
	apexTime      float64
}

func newFirework3D(canvasW, canvasH int, cameraZ float64) *firework3D {
	colors := [][3]uint8{
		{255, 50, 50}, {255, 140, 0}, {255, 215, 0}, {50, 255, 50},
		{100, 150, 255}, {200, 100, 255}, {255, 192, 203}, {0, 255, 255}, {255, 255, 255},
	}
	c := colors[rand.Intn(len(colors))]
	x := float64(canvasW)*0.2 + rand.Float64()*float64(canvasW)*0.6
	y := float64(canvasH) - 1
	z := cameraZ + 50 + rand.Float64()*250
	targetY := float64(canvasH)*0.1 + rand.Float64()*float64(canvasH)*0.23
	gravity := 100.0
	dist := targetY - y
	requiredV := math.Sqrt(-2 * gravity * dist)
	return &firework3D{
		x: x, y: y, z: z,
		vx: rand.Float64()*40 - 20, vy: -requiredV, vz: 0,
		r: c[0], g: c[1], b: c[2],
	}
}

func (f *firework3D) update(dt float64) {
	if !f.exploded {
		gravity := 100.0
		f.vy += gravity * dt
		f.x += f.vx * dt
		f.y += f.vy * dt
		f.trail = append(f.trail, [2]float64{f.x, f.y})
		if len(f.trail) > 15 {
			f.trail = f.trail[1:]
		}
		if f.vy > 0 {
			f.apexTime += dt
			if f.apexTime >= 0.3 {
				f.explode()
			}
		}
	} else {
		alive := f.particles[:0]
		for i := range f.particles {
			f.particles[i].update(dt, 50.0, 0.97)
			if f.particles[i].isAlive() {
				alive = append(alive, f.particles[i])
			}
		}
		f.particles = alive
	}
}

func (f *firework3D) explode() {
	f.exploded = true
	numParticles := 300 + rand.Intn(200)
	speed := 100.0 + rand.Float64()*80
	for i := 0; i < numParticles; i++ {
		theta := rand.Float64() * 2 * math.Pi
		phi := rand.Float64() * math.Pi
		vx := speed * math.Sin(phi) * math.Cos(theta)
		vy := speed * math.Cos(phi)
		vz := speed * math.Sin(phi) * math.Sin(theta)
		life := 1.5 + rand.Float64()*1.0
		f.particles = append(f.particles, particle3D{
			x: f.x, y: f.y, z: f.z,
			vx: vx, vy: vy, vz: vz,
			r: f.r, g: f.g, b: f.b,
			maxLife: life,
		})
	}
}

func (f *firework3D) isFinished() bool {
	return f.exploded && len(f.particles) == 0
}

func (d *fireworksDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *fireworksDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	d.ensureCanvas(bounds)
}

func (d *fireworksDemo) ensureCanvas(bounds runtime.Rect) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	if d.blitter == nil {
		d.blitter = &graphics.BrailleBlitter{}
	}
	if d.canvas == nil {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
		return
	}
	cellW, cellH := d.canvas.CellSize()
	if cellW != bounds.Width || cellH != bounds.Height {
		d.canvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, d.blitter)
	}
}

func (d *fireworksDemo) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()
	ctx.Clear(backend.DefaultStyle())
	d.ensureCanvas(bounds)
	canvas := d.canvas
	if canvas == nil {
		return
	}
	canvas.Clear()

	w, h := canvas.Size()
	if w == 0 || h == 0 {
		return
	}

	cameraDist := 200.0
	centerX := float64(w) / 2
	centerY := float64(h) / 2

	for _, fw := range d.fireworks {
		color := backend.ColorRGB(fw.r, fw.g, fw.b)
		if !fw.exploded {
			for _, pt := range fw.trail {
				zRel := fw.z - d.cameraZ
				zOffset := zRel + cameraDist
				if zOffset > 0 {
					scale := cameraDist / zOffset
					sx := int(centerX + (pt[0]-centerX)*scale)
					sy := int(centerY + (pt[1]-centerY)*scale)
					if sx >= 0 && sx < w && sy >= 0 && sy < h {
						canvas.Blend(sx, sy, color, 0.8)
					}
				}
			}
		} else {
			for i := range fw.particles {
				p := &fw.particles[i]
				sx, sy, visible := p.project(d.cameraZ, cameraDist, centerX, centerY)
				if visible && sx >= 0 && sx < w && sy >= 0 && sy < h {
					pcolor := backend.ColorRGB(p.r, p.g, p.b)
					canvas.Blend(sx, sy, pcolor, p.alpha())
				}
			}
		}
	}

	canvas.Render(ctx.Buffer, bounds.X, bounds.Y)

	// Render title as crisp terminal text (not pixel font)
	title := "✦ FIREWORKS ✦"
	titleStyle := backend.DefaultStyle().Bold(true).Foreground(backend.ColorRGB(255, 220, 100))
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, title, titleStyle)
}

func (d *fireworksDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if tick, ok := msg.(runtime.TickMsg); ok {
		d.frame++
		dt := 1.0 / 30.0

		bounds := d.Bounds()
		px, py := d.blitter.PixelsPerCell()
		canvasW := bounds.Width * px
		canvasH := bounds.Height * py

		d.cameraZ += 15.0 * dt

		if tick.Time.Sub(d.lastSpawn) > d.spawnDelay && canvasW > 0 && canvasH > 0 {
			d.fireworks = append(d.fireworks, newFirework3D(canvasW, canvasH, d.cameraZ))
			d.lastSpawn = tick.Time
			d.spawnDelay = time.Duration(300+rand.Intn(400)) * time.Millisecond
		}

		alive := d.fireworks[:0]
		for _, fw := range d.fireworks {
			fw.update(dt)
			if !fw.isFinished() && fw.z-d.cameraZ > -50 {
				alive = append(alive, fw)
			}
		}
		d.fireworks = alive

		d.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}
