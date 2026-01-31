package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type addData struct {
	TypeName string
}

func runAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	targetDir := fs.String("dir", ".", "target project directory")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return errors.New("usage: fluffy add widget|page <Name>")
	}
	kind := fs.Arg(0)
	name := fs.Arg(1)

	switch kind {
	case "widget":
		return addWidget(*targetDir, name, *force)
	case "page":
		return addPage(*targetDir, name, *force)
	default:
		return fmt.Errorf("unknown add target: %s", kind)
	}
}

func addWidget(root, name string, force bool) error {
	typeName := toPascal(name)
	if typeName == "" {
		return errors.New("invalid widget name")
	}
	data := addData{TypeName: typeName}
	rendered, err := renderTemplate(widgetTemplate, data)
	if err != nil {
		return err
	}
	filename := filepath.Join(root, "widgets", toSnake(name)+".go")
	return writeFile(filename, []byte(rendered), 0o644, force)
}

func addPage(root, name string, force bool) error {
	typeName := toPascal(name)
	if typeName == "" {
		return errors.New("invalid page name")
	}
	data := addData{TypeName: typeName}
	rendered, err := renderTemplate(pageTemplate, data)
	if err != nil {
		return err
	}
	filename := filepath.Join(root, "pages", toSnake(name)+".go")
	return writeFile(filename, []byte(rendered), 0o644, force)
}

const widgetTemplate = `package widgets

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	ui "github.com/odvcencio/fluffyui/widgets"
)

type {{.TypeName}} struct {
	ui.Base
}

func New{{.TypeName}}() *{{.TypeName}} {
	return &{{.TypeName}}{}
}

func (w *{{.TypeName}}) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 10, Height: 1}
}

func (w *{{.TypeName}}) Layout(bounds runtime.Rect) {
	w.Base.Layout(bounds)
}

func (w *{{.TypeName}}) Render(ctx runtime.RenderContext) {
	bounds := w.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.SetString(bounds.X, bounds.Y, "{{.TypeName}}", backend.DefaultStyle())
}

func (w *{{.TypeName}}) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

var _ runtime.Widget = (*{{.TypeName}})(nil)
`

const pageTemplate = `package pages

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	ui "github.com/odvcencio/fluffyui/widgets"
)

type {{.TypeName}} struct {
	ui.Component
	title *ui.Label
}

func New{{.TypeName}}() *{{.TypeName}} {
	p := &{{.TypeName}}{}
	p.title = ui.NewLabel("{{.TypeName}}", ui.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	return p
}

func (p *{{.TypeName}}) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (p *{{.TypeName}}) Layout(bounds runtime.Rect) {
	p.Component.Layout(bounds)
	if p.title != nil {
		p.title.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: 1})
	}
}

func (p *{{.TypeName}}) Render(ctx runtime.RenderContext) {
	if p.title != nil {
		p.title.Render(ctx)
	}
}

func (p *{{.TypeName}}) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if p.title != nil {
		if result := p.title.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (p *{{.TypeName}}) ChildWidgets() []runtime.Widget {
	if p.title == nil {
		return nil
	}
	return []runtime.Widget{p.title}
}

var _ runtime.Widget = (*{{.TypeName}})(nil)
var _ runtime.ChildProvider = (*{{.TypeName}})(nil)
`
