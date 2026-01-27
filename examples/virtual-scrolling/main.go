package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

type bigAdapter struct {
	count int
}

func (a bigAdapter) Count() int {
	return a.count
}

func (a bigAdapter) Item(index int) int {
	return index
}

func (a bigAdapter) Render(item int, index int, selected bool, ctx runtime.RenderContext) {
	style := backend.DefaultStyle()
	prefix := "  "
	if selected {
		prefix = "> "
	}
	text := fmt.Sprintf("%sItem %06d", prefix, item)
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, text, style)
}

func (a bigAdapter) FixedItemHeight() int {
	return 1
}

func main() {
	count := 100000
	if raw := strings.TrimSpace(os.Getenv("FLUFFYUI_VIRTUAL_COUNT")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			count = parsed
		}
	}

	list := widgets.NewVirtualList[int](bigAdapter{count: count})
	list.SetOverscan(4)
	list.SetLabel("Virtual Items")

	panel := widgets.NewPanel(list).WithBorder(backend.DefaultStyle())
	panel.SetTitle(fmt.Sprintf("Virtual List (%d items)", count))

	bundle, err := demo.NewApp(panel, demo.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}
