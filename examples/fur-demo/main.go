// Package main demonstrates the fur console output package.
//
// Run with: go run ./examples/fur-demo
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/odvcencio/fluffy-ui/fur"
)

func main() {
	c := fur.Default()

	// Banner
	c.Println()
	c.Rule("Fur Demo - Beautiful Console Output")
	c.Println()

	// Markup examples
	c.Println("[bold cyan]1. Markup Syntax[/]")
	c.Println()
	c.Println("  [bold]Bold[/], [italic]Italic[/], [underline]Underline[/], [dim]Dim[/]")
	c.Println("  [red]Red[/], [green]Green[/], [blue]Blue[/], [yellow]Yellow[/], [magenta]Magenta[/], [cyan]Cyan[/]")
	c.Println("  [#ff6600]Hex Color[/], [rgb(100,200,50)]RGB Color[/]")
	c.Println("  [bold italic red on white]Combined styles[/]")
	c.Println()

	// Pretty printing
	c.Rule("2. Pretty Printing")
	c.Println()

	type Server struct {
		Host     string
		Port     int
		TLS      bool
		Features []string
	}

	config := Server{
		Host:     "localhost",
		Port:     8080,
		TLS:      true,
		Features: []string{"auth", "metrics", "tracing"},
	}

	c.Render(fur.Pretty(config))
	c.Println()

	// Inspect
	c.Rule("3. Object Inspection")
	c.Println()
	c.Render(fur.Inspect(http.Client{}))
	c.Println()

	// Progress with Live display
	c.Rule("4. Live Progress")
	c.Println()

	progress := fur.NewProgress(100).WithLabel("Downloading")
	live := fur.NewLive(progress).WithRate(50 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		_ = live.Start(ctx)
		close(done)
	}()

	for i := 0; i <= 100; i += 2 {
		progress.Set(i)
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	<-done
	c.Println()
	c.Println()

	// Columns
	c.Rule("5. Multi-Column Layout")
	c.Println()
	c.Render(fur.Columns(
		fur.Text("Column 1\nWith multiple\nlines of text"),
		fur.Text("Column 2\nAlso has\nseveral lines"),
		fur.Text("Column 3\nAnd more\ncontent here"),
	))
	c.Println()

	// Error tracebacks
	c.Rule("6. Error Tracebacks")
	c.Println()

	err := doSomethingThatFails()
	if err != nil {
		c.Render(fur.Traceback(err))
	}
	c.Println()

	// Emoji (when enabled)
	c.Rule("7. Emoji Support")
	c.Println()
	fur.EnableEmoji()
	c.Println("  :wave: Hello! :rocket: Ready for launch! :check: All systems go!")
	fur.DisableEmoji()
	c.Println()

	// Export
	c.Rule("8. Export Formats")
	c.Println()
	sample := fur.Markup("[bold green]Sample[/] output for export")
	c.Println("  Text: " + fur.ExportText(sample, 40))
	c.Println("  HTML available via fur.ExportHTML()")
	c.Println("  SVG available via fur.ExportSVG()")
	c.Println()

	// Final
	c.Rule()
	c.Println("[bold green]✓[/] Demo complete!")
	c.Println()
}

func doSomethingThatFails() error {
	return fur.WrapMsg(fmt.Errorf("connection refused"), "failed to connect to database")
}

func init() {
	// Check if running in a non-TTY environment
	if os.Getenv("CI") != "" {
		fmt.Println("Skipping interactive demo in CI")
		os.Exit(0)
	}
}
