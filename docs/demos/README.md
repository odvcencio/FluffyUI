# FluffyUI Demo Showcase

This directory contains recordings and GIFs demonstrating the full depth of FluffyUI's capabilities.

## Available Demos

### Core Widgets

| Demo | Description | File |
|------|-------------|------|
| **hero** | Animated FluffyUI logo with rainbow effects | [hero.gif](hero.gif) |
| **quickstart** | Hello world with typing animation and sparkles | [quickstart.gif](quickstart.gif) |
| **buttons** | Button variants (primary, secondary, danger, success, warning) | [buttons.gif](buttons.gif) |
| **input** | Form input with validation states and password strength | [input.gif](input.gif) |
| **dialog** | Modal dialogs with focus management and animations | [dialog.gif](dialog.gif) |
| **tabs** | Tabbed navigation with content switching | [tabs.gif](tabs.gif) |

### Data Widgets

| Demo | Description | File |
|------|-------------|------|
| **table** | Sortable data table with selection and status indicators | [table.gif](table.gif) |
| **list** | File manager-style list with checkboxes and selection | [list.gif](list.gif) |
| **sparkline** | Live data visualization with multiple metrics | [sparkline.gif](sparkline.gif) |
| **progress** | Progress bars, spinners, and multi-step indicators | [progress.gif](progress.gif) |
| **counter** | Reactive counter with history sparkline | [counter.gif](counter.gif) |

### Graphics & Animation

| Demo | Description | File |
|------|-------------|------|
| **graphics** | Canvas API with shapes, curves, and transforms | [graphics.gif](graphics.gif) |
| **easing** | Animation easing functions visualization | [easing.gif](easing.gif) |
| **fireworks** | 3D particle effects with perspective projection | [fireworks.gif](fireworks.gif) |
| **video** | Video player widget demo | [video.gif](video.gif) |

### Complete Applications

| Demo | Description | File |
|------|-------------|------|
| **candy-wars** | Full trading game demonstrating all features | [candy-wars-playback.gif](candy-wars-playback.gif) |

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
