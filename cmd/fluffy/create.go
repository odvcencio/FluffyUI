package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type projectData struct {
	AppName    string
	AppTitle   string
	ModulePath string
}

type projectTemplate struct {
	dirs  []string
	files map[string]string
}

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	templateName := fs.String("template", "minimal", "template: minimal, full, game")
	modulePath := fs.String("module", "", "go module path")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("missing app name")
	}

	appName := fs.Arg(0)
	targetDir := appName
	if empty, err := dirEmpty(targetDir); err != nil {
		return err
	} else if !empty && !*force {
		return fmt.Errorf("target directory is not empty: %s", targetDir)
	}

	if *modulePath == "" {
		*modulePath = fmt.Sprintf("example.com/%s", filepath.Base(appName))
	}

	data := projectData{
		AppName:    appName,
		AppTitle:   titleFromName(appName),
		ModulePath: *modulePath,
	}

	tmpl, err := selectTemplate(*templateName)
	if err != nil {
		return err
	}

	return createProject(targetDir, data, tmpl, *force)
}

func selectTemplate(name string) (projectTemplate, error) {
	switch name {
	case "minimal":
		return minimalTemplate(), nil
	case "full":
		return fullTemplate(), nil
	case "game":
		return gameTemplate(), nil
	default:
		return projectTemplate{}, fmt.Errorf("unknown template: %s", name)
	}
}

func createProject(root string, data projectData, tmpl projectTemplate, force bool) error {
	if err := ensureDir(root); err != nil {
		return err
	}
	for _, dir := range tmpl.dirs {
		if err := ensureDir(filepath.Join(root, dir)); err != nil {
			return err
		}
	}
	for relPath, content := range tmpl.files {
		rendered, err := renderTemplate(content, data)
		if err != nil {
			return fmt.Errorf("render %s: %w", relPath, err)
		}
		if err := writeFile(filepath.Join(root, relPath), []byte(rendered), 0o644, force); err != nil {
			return err
		}
	}
	return nil
}

func minimalTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
		},
		files: map[string]string{
			"go.mod":      goModTemplate,
			"main.go":     minimalMainTemplate,
			"fluffy.toml": fluffyTomlTemplate,
		},
	}
}

func fullTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
			"themes",
			"assets/audio",
			"assets/images",
			"tests/visual",
			"tests/e2e",
		},
		files: map[string]string{
			"go.mod":              goModTemplate,
			"main.go":             fullMainTemplate,
			"fluffy.toml":         fluffyTomlTemplate,
			"themes/default.yaml": defaultThemeTemplate,
		},
	}
}

func gameTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
			"themes",
			"assets/audio",
			"assets/images",
			"tests/visual",
		},
		files: map[string]string{
			"go.mod":              goModTemplate,
			"main.go":             gameMainTemplate,
			"fluffy.toml":         fluffyTomlTemplate,
			"themes/default.yaml": defaultThemeTemplate,
		},
	}
}

const goModTemplate = `module {{.ModulePath}}

go 1.22
`

const fluffyTomlTemplate = `name = "{{.AppName}}"
module = "{{.ModulePath}}"
theme = "themes/default.yaml"
`

const minimalMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffy-ui/fluffy"
)

func main() {
	app := fluffy.NewApp()
	app.SetRoot(fluffy.NewLabel("{{.AppTitle}}"))

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}
`

const fullMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/fluffy"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	ui "github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	app := fluffy.NewApp()
	app.SetRoot(NewDashboard())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type Dashboard struct {
	ui.Component
	count      *state.Signal[int]
	title      *ui.Label
	countLabel *ui.Label
	incBtn     *ui.Button
	decBtn     *ui.Button
	grid       *ui.Grid
}

func NewDashboard() *Dashboard {
	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	d := &Dashboard{count: count}
	d.title = ui.NewLabel("{{.AppTitle}} Dashboard").WithStyle(backend.DefaultStyle().Bold(true))
	d.countLabel = ui.NewLabel("Count: 0")
	d.incBtn = ui.NewButton("Increment", ui.WithVariant(ui.VariantPrimary), ui.WithOnClick(func() {
		d.update(1)
	}))
	d.decBtn = ui.NewButton("Decrement", ui.WithVariant(ui.VariantSecondary), ui.WithOnClick(func() {
		d.update(-1)
	}))

	grid := ui.NewGrid(3, 2)
	grid.Gap = 1
	grid.Add(d.title, 0, 0, 1, 2)
	grid.Add(d.countLabel, 1, 0, 1, 2)
	grid.Add(d.decBtn, 2, 0, 1, 1)
	grid.Add(d.incBtn, 2, 1, 1, 1)
	d.grid = grid

	d.refresh()
	return d
}

func (d *Dashboard) Mount() {
	d.Observe(d.count, d.refresh)
	d.refresh()
}

func (d *Dashboard) Unmount() {
	d.Subs.Clear()
}

func (d *Dashboard) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *Dashboard) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	if d.grid != nil {
		d.grid.Layout(bounds)
	}
}

func (d *Dashboard) Render(ctx runtime.RenderContext) {
	if d.grid != nil {
		d.grid.Render(ctx)
	}
}

func (d *Dashboard) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d.grid != nil {
		if result := d.grid.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (d *Dashboard) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if d.grid != nil {
		children = append(children, d.grid)
	}
	return children
}

func (d *Dashboard) refresh() {
	if d.countLabel != nil && d.count != nil {
		d.countLabel.SetText(fmt.Sprintf("Count: %d", d.count.Get()))
	}
	d.Invalidate()
}

func (d *Dashboard) update(delta int) {
	if d.count == nil {
		return
	}
	d.count.Update(func(v int) int { return v + delta })
}

var _ runtime.Widget = (*Dashboard)(nil)
var _ runtime.ChildProvider = (*Dashboard)(nil)
`

const gameMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/fluffy"
	"github.com/odvcencio/fluffy-ui/runtime"
	ui "github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	app := fluffy.NewApp(fluffy.WithTickRate(time.Second / 60))
	app.SetRoot(NewBouncer())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type Bouncer struct {
	ui.Base
	x     int
	dir   int
	width int
}

func NewBouncer() *Bouncer {
	return &Bouncer{dir: 1}
}

func (b *Bouncer) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (b *Bouncer) Layout(bounds runtime.Rect) {
	b.Base.Layout(bounds)
	b.width = bounds.Width
	if b.x >= b.width {
		b.x = b.width - 1
	}
	if b.x < 0 {
		b.x = 0
	}
}

func (b *Bouncer) Render(ctx runtime.RenderContext) {
	bounds := b.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.Fill(bounds, ' ', backend.DefaultStyle())
	x := bounds.X + b.x
	y := bounds.Y + bounds.Height/2
	if x < bounds.X {
		x = bounds.X
	}
	if x >= bounds.X+bounds.Width {
		x = bounds.X + bounds.Width - 1
	}
	ctx.Buffer.Set(x, y, '@', backend.DefaultStyle().Bold(true))
}

func (b *Bouncer) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); !ok {
		return runtime.Unhandled()
	}
	if b.width <= 1 {
		return runtime.Unhandled()
	}
	b.x += b.dir
	if b.x <= 0 {
		b.x = 0
		b.dir = 1
	}
	if b.x >= b.width-1 {
		b.x = b.width - 1
		b.dir = -1
	}
	b.Invalidate()
	return runtime.Handled()
}

var _ runtime.Widget = (*Bouncer)(nil)
`

const defaultThemeTemplate = `name: "Default Theme"
colors:
  background: "#0c0c10"
  surface: "#16161c"
  text: "#f0eee8"
  accent: "#ffb74d"
styles:
  app:
    foreground: "text"
    background: "background"
  panel:
    foreground: "text"
    background: "surface"
  button.primary:
    foreground: "background"
    background: "accent"
`
