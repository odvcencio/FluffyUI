# FluffyUI Demo Showcase

This directory contains recordings and GIFs demonstrating the full depth of FluffyUI's capabilities.

![FluffyUI hero](hero.gif)

## Available Demos

### Core Widgets

#### hero -- Animated FluffyUI logo with rainbow effects
<div class="fluffy-demo" data-cast="hero.cast" data-cols="80" data-rows="24"></div>

#### quickstart -- Hello world with typing animation and sparkles
<div class="fluffy-demo" data-cast="quickstart.cast" data-cols="80" data-rows="24"></div>

#### buttons -- Button variants (primary, secondary, danger, success, warning)
<div class="fluffy-demo" data-cast="buttons.cast" data-cols="80" data-rows="24"></div>

#### input -- Form input with validation states and password strength
<div class="fluffy-demo" data-cast="input.cast" data-cols="80" data-rows="24"></div>

#### dialog -- Modal dialogs with focus management and animations
<div class="fluffy-demo" data-cast="dialog.cast" data-cols="80" data-rows="24"></div>

#### tabs -- Tabbed navigation with content switching
<div class="fluffy-demo" data-cast="tabs.cast" data-cols="80" data-rows="24"></div>

#### select -- Dropdown selection cycling through options
<div class="fluffy-demo" data-cast="select.cast" data-cols="80" data-rows="24"></div>

#### checkbox -- Checkbox toggles and radio button groups
<div class="fluffy-demo" data-cast="checkbox.cast" data-cols="80" data-rows="24"></div>

#### slider -- Volume, brightness, and range sliders
<div class="fluffy-demo" data-cast="slider.cast" data-cols="80" data-rows="24"></div>

#### textarea -- Multiline code editor with syntax highlighting
<div class="fluffy-demo" data-cast="textarea.cast" data-cols="80" data-rows="24"></div>

### Data Widgets

#### table -- Sortable data table with selection and status indicators
<div class="fluffy-demo" data-cast="table.cast" data-cols="80" data-rows="24"></div>

#### list -- File manager-style list with checkboxes and selection
<div class="fluffy-demo" data-cast="list.cast" data-cols="80" data-rows="24"></div>

#### sparkline -- Live data visualization with multiple metrics
<div class="fluffy-demo" data-cast="sparkline.cast" data-cols="80" data-rows="24"></div>

#### progress -- Progress bars, spinners, and multi-step indicators
<div class="fluffy-demo" data-cast="progress.cast" data-cols="80" data-rows="24"></div>

#### counter -- Reactive counter with history sparkline
<div class="fluffy-demo" data-cast="counter.cast" data-cols="80" data-rows="24"></div>

### Graphics & Animation

#### graphics -- Canvas API with shapes, curves, and transforms
<div class="fluffy-demo" data-cast="graphics.cast" data-cols="80" data-rows="24"></div>

#### easing -- Animation easing functions visualization
<div class="fluffy-demo" data-cast="easing.cast" data-cols="80" data-rows="24"></div>

#### fireworks -- 3D particle effects with perspective projection
<div class="fluffy-demo" data-cast="fireworks.cast" data-cols="80" data-rows="24"></div>

#### video -- Video player widget demo
<div class="fluffy-demo" data-cast="video.cast" data-cols="80" data-rows="24"></div>

### Complete Applications

#### candy-wars -- Full trading game demonstrating all features

Run: `go run ./examples/candy-wars`

## Framework Capabilities Demonstrated

### 1. Widget System (35+ Components)
- **Layout**: Grid, Stack, Splitter, ScrollView, Panel, Box, AspectRatio
- **Input**: Button, Input, TextArea, Checkbox, Radio, Select, Slider
- **Data**: List, Table, Tree, SearchWidget
- **Navigation**: Tabs, Menu, Breadcrumb, Stepper, Palette
- **Feedback**: Dialog, Alert, ToastStack, Spinner, Progress, Sparkline, BarChart

### 2. Sub-Cell Graphics
- Canvas API with pixel-precise drawing
- Multiple blitters: Braille (2x4), Sextant (2x3), Quadrant (2x2)
- Shapes: circles, rectangles, triangles, lines, curves
- Transforms: translate, rotate, scale
- Path operations: Bezier curves, arcs

### 3. Animation System
- Tweens with configurable easing (linear, quad, cubic, elastic, bounce)
- Physics-based spring animations
- Particle systems with gravity and air resistance
- Color gradients and effects

### 4. Reactive State
- Signals with automatic UI updates
- Computed values
- Subscription-based reactivity
- State persistence

### 5. Accessibility
- Screen reader support
- Focus management
- ARIA-like roles
- Keyboard navigation

### 6. Keybinding System
- Command registry
- Keymap stacking for modes
- Standard commands (quit, scroll, clipboard)
- Custom command binding

### 7. Recording & Export
- Asciicast capture
- Video export (MP4)
- GIF generation
- Deterministic simulation backend

## Generating Demos

### All Demos
```bash
go run ./examples/generate-demos --out docs/demos --duration 6
```

### Specific Demo
```bash
go run ./examples/generate-demos --out docs/demos --demo hero --duration 10
```

### Convert to GIF
```bash
# Requires agg: cargo install --git https://github.com/asciinema/agg
agg --theme monokai --font-size 16 --fps-cap 30 \
  --last-frame-duration 0.001 docs/demos/hero.cast docs/demos/hero.gif
```

### Batch Convert All
```bash
cd docs/demos
for cast in *.cast; do
  gif="${cast%.cast}.gif"
  echo "Converting: $cast -> $gif"
  agg --theme monokai --font-size 16 --fps-cap 30 \
    --last-frame-duration 0.001 "$cast" "$gif"
done
```

## Viewing Demos

### Using asciinema
```bash
# Play a recording
asciinema play docs/demos/hero.cast

# Play at 2x speed
asciinema play docs/demos/hero.cast --speed 2
```

### Using Web Browser
Open any `.gif` file directly in your browser or image viewer.

## Demo Statistics

```bash
# Count demos
echo "Cast files: $(ls docs/demos/*.cast 2>/dev/null | wc -l)"
echo "GIF files: $(ls docs/demos/*.gif 2>/dev/null | wc -l)"

# Total size
du -sh docs/demos/

# Individual sizes
ls -lh docs/demos/*.gif
```

## Creating Custom Demos

See [examples/generate-demos/main.go](../examples/generate-demos/main.go) for examples of how to create your own demos using the simulation backend.

Key patterns:
1. Create a widget that implements `runtime.Widget`
2. Use `widgets.Component` for automatic invalidation
3. Animate with `runtime.TickMsg`
4. Record with `recording.NewAsciicastRecorder`
