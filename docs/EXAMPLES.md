# FluffyUI Examples

A comprehensive guide to all examples in the FluffyUI framework, organized by complexity and feature area.

## Quick Start Examples

### Hello World
```bash
go run ./examples/quickstart
```
The simplest FluffyUI application - displays "Hello from FluffyUI!" with animated effects.

### Counter
```bash
go run ./examples/counter
```
Demonstrates reactive state management with signals. Shows a counter that can be incremented/decremented with history tracking.

## Widget Showcase

### Widget Gallery
```bash
go run ./examples/widgets/gallery
```
Complete catalog of all 35+ widgets organized by category:
- Layout widgets (Grid, Stack, Panel, etc.)
- Input widgets (Button, Input, Select, etc.)
- Data widgets (Table, List, Tree)
- Navigation widgets (Tabs, Menu, Palette)
- Feedback widgets (Dialog, Toast, Progress)

### Showcase Tabs
```bash
go run ./examples/showcase
```
Curated widget showcase with multiple tabs (overview, inputs, data).

### Widget Categories
```bash
go run ./examples/widgets/layout     # Layout demonstrations
go run ./examples/widgets/input      # Input form demonstrations
go run ./examples/widgets/data       # Data display demonstrations
go run ./examples/widgets/navigation # Navigation demonstrations
go run ./examples/widgets/feedback   # Feedback widget demonstrations
```

## Graphics & Animation

### Canvas Graphics
```bash
go run ./examples/graphics-demo
```
Demonstrates the sub-cell graphics system:
- Drawing shapes (circles, rectangles, triangles)
- Bezier curves and paths
- Transforms (translate, rotate, scale)
- Multiple blitters (Braille, Sextant, Quadrant)

### Fireworks 3D
```bash
go run ./examples/fireworks-demo
```
3D particle effects with perspective projection:
- Physics simulation (gravity, air resistance)
- Color gradients and blending
- Particle emitters
- Real-time animation

### GPU Canvas
```bash
go run ./examples/gpu-canvas-demo
```
Hardware-accelerated canvas rendering:
- Software, OpenGL, and Metal backends
- High-performance graphics
- Large canvas support

### Water Simulation
```bash
go run ./examples/water-demo
```
Interactive water ripple simulation using the canvas API.

### Animation
```bash
go run ./examples/animation-demo
```
Animation system demonstrations:
- Tweening with various easing functions
- Spring physics
- Animation chaining

## Complete Applications

### Candy Wars (Showcase Game)
```bash
go run ./examples/candy-wars
```
A complete trading game demonstrating all FluffyUI features:
- Multiple screens with tab navigation
- Tables with sorting and selection
- Dialogs and modal interactions
- Sparkline charts for data visualization
- Form validation
- Reactive state management
- Keybindings and shortcuts
- Toast notifications
- Save/load functionality

### Todo App
```bash
go run ./examples/todo-app
```
Full CRUD application:
- Task creation, editing, deletion
- Filtering (all/active/completed)
- Persistence
- Keyboard shortcuts

### File Browser
```bash
go run ./examples/file-browser
```
File system browser:
- Tree navigation
- File operations
- Preview pane
- Keyboard shortcuts

### Dashboard
```bash
go run ./examples/dashboard
```
Data visualization dashboard:
- Multiple chart types
- Real-time data updates
- Layout management
- Sparkline charts

### Settings Form
```bash
go run ./examples/settings-form
```
Complex form with validation:
- Multiple field types
- Form validation
- Error messages
- Submit handling

## Advanced Features

### Command Palette
```bash
go run ./examples/command-palette
```
Demonstrates the keybinding system:
- Command registry
- Fuzzy finder
- Keymap stacking
- Custom commands

### Virtual Scrolling
```bash
go run ./examples/virtual-scrolling
```
Performance demonstration:
- 10,000+ item lists
- Efficient rendering
- Smooth scrolling

### Accessibility Demo
```bash
go run ./examples/accessibility-demo
```
Accessibility features:
- Screen reader support
- Focus management
- ARIA-like roles
- Keyboard navigation

### AI Agent Integration
```bash
go run ./examples/ai-agent-demo
```
Out-of-process agent interaction:
- JSONL socket protocol
- Agent-driven input
- Snapshot-based observation

### Recording Demo
```bash
go run ./examples/recording
```
Session recording and export:
- Asciicast capture
- Video export
- Playback

### Performance Dashboard
```bash
go run ./examples/perf-dashboard
```
Live render performance summary:
- FPS and render/flush timing
- Dirty cell ratios
- Layer count sampling

## Furry Formats (FUR)

### FUR Demo
```bash
go run ./examples/fur-demo
```
Rich text and formatting:
- FUR markup language
- Styled text
- Inline images
- Color support

## Community Examples

Community examples live under `examples/community` and will eventually be
mirrored from a dedicated `fluffyui-examples` repository.

## Custom Implementations

### Custom Loop
```bash
go run ./examples/custom-loop
```
Custom application loop:
- Manual event handling
- Custom update logic
- Frame rate control

## Demo Generation

### Generate All Demos
```bash
go run ./examples/generate-demos --out docs/demos --duration 6
```
Headless demo generation using simulation backend - perfect for CI/CD.

### Regenerate All Demos
```bash
./scripts/regenerate-demos.sh [duration_seconds]
```
Convenience script that generates cast files and converts to GIFs.

## Running with Options

### Backend Selection
```bash
# Real terminal (default)
FLUFFYUI_BACKEND=tcell go run ./examples/quickstart

# Simulation backend
FLUFFYUI_BACKEND=sim go run ./examples/quickstart
```

### Recording
```bash
# Record to asciicast
FLUFFYUI_RECORD=session.cast go run ./examples/quickstart

# Record and export to video
FLUFFYUI_RECORD=session.cast FLUFFYUI_RECORD_EXPORT=output.mp4 go run ./examples/quickstart
```

### Audio
```bash
# Enable audio (requires assets)
FLUFFYUI_AUDIO_ASSETS=./examples/quickstart/assets/audio go run ./examples/quickstart

# Disable audio
FLUFFYUI_AUDIO_ASSETS=off go run ./examples/quickstart
```

## Example Complexity Guide

| Example | Lines of Code | Complexity | Focus Area |
|---------|--------------|------------|------------|
| quickstart | ~30 | ⭐ Beginner | Basics |
| counter | ~100 | ⭐ Beginner | State |
| widgets/* | ~500 | ⭐⭐ Intermediate | Widgets |
| graphics-demo | ~300 | ⭐⭐ Intermediate | Graphics |
| fireworks-demo | ~400 | ⭐⭐⭐ Advanced | Animation |
| candy-wars | ~3000 | ⭐⭐⭐ Advanced | Everything |

## Learning Path

1. **Beginner**: Start with `quickstart` and `counter` to understand the basics
2. **Intermediate**: Explore `widgets/gallery` and `graphics-demo` for feature depth
3. **Advanced**: Study `candy-wars` for a complete application architecture

## Troubleshooting

### Terminal Issues
If you see "terminal not supported" errors:
```bash
export TERM=xterm-256color
```

### Display Issues
For headless environments:
```bash
FLUFFYUI_BACKEND=sim go run ./examples/quickstart
```

### Build Errors
Ensure Go 1.24+ is installed:
```bash
go version
```
