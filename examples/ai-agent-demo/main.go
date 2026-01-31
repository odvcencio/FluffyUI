package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

const defaultAgentAddr = "tcp:127.0.0.1:7777"

func main() {
	addr := os.Getenv("FLUFFYUI_AGENT_ADDR")
	if addr == "" {
		addr = defaultAgentAddr
	}

	view := NewAgentDemoView(addr)
	bundle, err := demo.NewApp(view, demo.Options{TickRate: time.Second / 30})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	srv, err := agent.NewServer(agent.ServerOptions{
		Addr:      addr,
		App:       bundle.App,
		AllowText: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent server init failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer srv.Close()

	go func() {
		if err := srv.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "agent server error: %v\n", err)
		}
	}()

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type AgentDemoView struct {
	widgets.Component
	title  *widgets.Label
	panel  *widgets.Panel
	addr   *widgets.Label
	help   *widgets.Label
	input  *widgets.Input
	status *widgets.Label
}

func NewAgentDemoView(addr string) *AgentDemoView {
	view := &AgentDemoView{}
	view.title = widgets.NewLabel("AI Agent Demo", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	view.addr = widgets.NewLabel("Agent server: " + addr)
	view.help = widgets.NewLabel("Run: python examples/ai-agent-demo/agent.py")

	view.input = widgets.NewInput()
	view.input.SetLabel("Command")
	view.input.Focus()
	view.status = widgets.NewLabel("Submitted: (none)")
	view.input.SetOnSubmit(func(text string) {
		view.status.SetText("Submitted: " + text)
	})

	stack := demo.NewVBox(view.input, view.status)
	stack.Gap = 1
	view.panel = widgets.NewPanel(stack, widgets.WithPanelBorder(backend.DefaultStyle()))
	view.panel.SetTitle("Input")
	return view
}

func (v *AgentDemoView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (v *AgentDemoView) Layout(bounds runtime.Rect) {
	v.Component.Layout(bounds)
	y := bounds.Y
	if v.title != nil {
		v.title.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y += 2
	}

	panelHeight := bounds.Height - (y - bounds.Y) - 2
	if panelHeight > 6 {
		panelHeight = 6
	}
	if panelHeight < 1 {
		panelHeight = 1
	}
	if v.panel != nil {
		v.panel.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: panelHeight})
		y += panelHeight + 1
	}

	if v.addr != nil && y < bounds.Y+bounds.Height {
		v.addr.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	if v.help != nil && y < bounds.Y+bounds.Height {
		v.help.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
	}
}

func (v *AgentDemoView) Render(ctx runtime.RenderContext) {
	if v.title != nil {
		v.title.Render(ctx)
	}
	if v.panel != nil {
		v.panel.Render(ctx)
	}
	if v.addr != nil {
		v.addr.Render(ctx)
	}
	if v.help != nil {
		v.help.Render(ctx)
	}
}

func (v *AgentDemoView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if v.panel != nil {
		if result := v.panel.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (v *AgentDemoView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if v.title != nil {
		children = append(children, v.title)
	}
	if v.panel != nil {
		children = append(children, v.panel)
	}
	if v.addr != nil {
		children = append(children, v.addr)
	}
	if v.help != nil {
		children = append(children, v.help)
	}
	return children
}
