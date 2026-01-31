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
	templateName := fs.String("template", "minimal", "template: minimal, full, game, dashboard, form, data-viewer")
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
	case "dashboard":
		return dashboardTemplate(), nil
	case "form":
		return formTemplate(), nil
	case "data-viewer":
		return dataViewerTemplate(), nil
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

func dashboardTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
			"themes",
		},
		files: map[string]string{
			"go.mod":              goModTemplate,
			"main.go":             dashboardMainTemplate,
			"fluffy.toml":         fluffyTomlTemplate,
			"themes/default.yaml": defaultThemeTemplate,
		},
	}
}

func formTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
			"themes",
		},
		files: map[string]string{
			"go.mod":              goModTemplate,
			"main.go":             formMainTemplate,
			"fluffy.toml":         fluffyTomlTemplate,
			"themes/default.yaml": defaultThemeTemplate,
		},
	}
}

func dataViewerTemplate() projectTemplate {
	return projectTemplate{
		dirs: []string{
			"widgets",
			"themes",
		},
		files: map[string]string{
			"go.mod":              goModTemplate,
			"main.go":             dataViewerMainTemplate,
			"fluffy.toml":         fluffyTomlTemplate,
			"themes/default.yaml": defaultThemeTemplate,
		},
	}
}

const goModTemplate = `module {{.ModulePath}}

go 1.24
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

	"github.com/odvcencio/fluffyui/fluffy"
)

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
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

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	ui "github.com/odvcencio/fluffyui/widgets"
)

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
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
	d.title = ui.NewLabel("{{.AppTitle}} Dashboard", ui.WithLabelStyle(backend.DefaultStyle().Bold(true)))
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

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	ui "github.com/odvcencio/fluffyui/widgets"
)

func main() {
	app, err := fluffy.NewApp(fluffy.WithTickRate(time.Second / 60))
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
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

const dashboardMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	ui "github.com/odvcencio/fluffyui/widgets"
)

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	app.SetRoot(buildDashboard())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func buildDashboard() runtime.Widget {
	sparkData := state.NewSignal([]float64{12, 18, 14, 22, 16, 24, 19})
	spark := ui.NewSparkline(sparkData)

	progress := ui.NewProgress()
	progress.Label = "Capacity"
	progress.Value = 72

	alert := ui.NewAlert("All systems nominal", ui.AlertSuccess)

	table := ui.NewTable(
		ui.TableColumn{Title: "Service"},
		ui.TableColumn{Title: "Status"},
		ui.TableColumn{Title: "Latency"},
	)
	table.SetRows([][]string{
		{"Auth", "OK", "32ms"},
		{"Billing", "OK", "45ms"},
		{"Search", "OK", "57ms"},
	})

	left := ui.NewPanel(table, ui.WithPanelBorder(backend.DefaultStyle()))
	left.SetTitle("Services")

	rightColumn := ui.VBox(
		ui.FlexFixed(alert),
		ui.FlexFixed(progress),
		ui.FlexFixed(spark),
	)
	rightColumn.Gap = 1

	right := ui.NewPanel(rightColumn, ui.WithPanelBorder(backend.DefaultStyle()))
	right.SetTitle("Signals")

	split := ui.NewSplitter(left, right)
	split.Ratio = 0.6
	return split
}
`

const formMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	ui "github.com/odvcencio/fluffyui/widgets"
)

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	app.SetRoot(buildForm())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func buildForm() runtime.Widget {
	builder := forms.NewBuilder().
		Text("name", "Name", "", forms.Required("Name required")).
		Email("email", "Email", "", forms.Email("Invalid email")).
		Checkbox("tos", "Accept terms", false)

	form, _ := builder.Build()

	status := ui.NewLabel("Ready")

	nameInput := ui.NewInput()
	nameInput.SetPlaceholder("Name")
	nameInput.SetOnChange(func(text string) {
		form.Set("name", text)
	})

	emailInput := ui.NewInput()
	emailInput.SetPlaceholder("Email")
	emailInput.SetOnChange(func(text string) {
		form.Set("email", text)
	})

	terms := ui.NewCheckbox("Accept terms")
	terms.SetOnChange(func(value *bool) {
		if value != nil {
			form.Set("tos", *value)
		}
	})

	submit := ui.NewButton("Submit", ui.WithVariant(ui.VariantPrimary), ui.WithOnClick(func() {
		form.Submit()
	}))

	form.OnSubmit(func(values forms.Values) {
		status.SetText(fmt.Sprintf("Submitted: %v", values))
	})

	layout := ui.VBox(
		ui.FlexFixed(ui.NewLabel("Sign up", ui.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		ui.FlexFixed(nameInput),
		ui.FlexFixed(emailInput),
		ui.FlexFixed(terms),
		ui.FlexFixed(submit),
		ui.FlexFixed(status),
	)
	layout.Gap = 1

	panel := ui.NewPanel(layout, ui.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle("Form")
	return panel
}
`

const dataViewerMainTemplate = `package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	ui "github.com/odvcencio/fluffyui/widgets"
)

type tableSource struct {
	rows int
}

func (t tableSource) RowCount() int {
	return t.rows
}

func (t tableSource) Cell(row, col int) string {
	switch col {
	case 0:
		return fmt.Sprintf("Row %d", row+1)
	case 1:
		return fmt.Sprintf("Value %d", (row+1)*10)
	default:
		return ""
	}
}

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	app.SetRoot(buildViewer())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func buildViewer() runtime.Widget {
	grid := ui.NewDataGrid(
		ui.TableColumn{Title: "Name"},
		ui.TableColumn{Title: "Value"},
	)
	grid.SetDataSource(tableSource{rows: 10000})

	panel := ui.NewPanel(grid, ui.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle("Data Viewer")
	return panel
}
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
