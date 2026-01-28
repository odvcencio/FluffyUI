package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	backendtcell "github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/docs/site/content"
	"github.com/odvcencio/fluffyui/docs/site/viewer"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/runtime"
)

func main() {
	rootDir, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find project root: %v\n", err)
		os.Exit(1)
	}

	docsPath := filepath.Join(rootDir, "docs")
	site, err := content.LoadDir(docsPath, content.LoadOptions{Source: "docs"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load docs: %v\n", err)
		os.Exit(1)
	}

	view := viewer.NewDocsView(site)
	app, err := newDocsApp(view)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func newDocsApp(root runtime.Widget) (*runtime.App, error) {
	be, err := backendtcell.New()
	if err != nil {
		return nil, err
	}
	registry := keybind.NewRegistry()
	keybind.RegisterStandardCommands(registry)
	keybind.RegisterScrollCommands(registry)
	keybind.RegisterClipboardCommands(registry)

	keymap := keybind.DefaultKeymap()
	stack := &keybind.KeymapStack{}
	stack.Push(keymap)
	router := keybind.NewKeyRouter(registry, nil, stack)
	keyHandler := &keybind.RuntimeHandler{Router: router}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		TickRate:          time.Second / 30,
		KeyHandler:        keyHandler,
		Announcer:         &accessibility.SimpleAnnouncer{},
		Clipboard:         &clipboard.MemoryClipboard{},
		FocusRegistration: runtime.FocusRegistrationAuto,
		FocusStyle: &accessibility.FocusStyle{
			Indicator: "> ",
			Style:     backend.DefaultStyle().Bold(true),
		},
	})
	return app, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod")
		}
		dir = parent
	}
}
