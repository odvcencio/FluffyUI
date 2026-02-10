package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func runSnapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	output := fs.String("output", "", "output file (default: stdout)")
	format := fs.String("format", "ansi", "output format: ansi, text")
	width := fs.Int("width", 80, "terminal width")
	height := fs.Int("height", 24, "terminal height")
	delay := fs.Duration("delay", 500*time.Millisecond, "delay before capture")
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: fluffy snapshot [flags] -- <command> [args...]")
		fmt.Fprintln(os.Stderr, "\nCaptures a single frame from a FluffyUI application.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		return fmt.Errorf("no command specified")
	}

	// Set environment for the subprocess
	env := os.Environ()
	env = append(env,
		"FLUFFYUI_SNAPSHOT=1",
		fmt.Sprintf("FLUFFYUI_SNAPSHOT_FORMAT=%s", *format),
		fmt.Sprintf("FLUFFYUI_SNAPSHOT_WIDTH=%d", *width),
		fmt.Sprintf("FLUFFYUI_SNAPSHOT_HEIGHT=%d", *height),
		fmt.Sprintf("FLUFFYUI_SNAPSHOT_DELAY=%s", delay.String()),
	)
	if *output != "" {
		env = append(env, fmt.Sprintf("FLUFFYUI_SNAPSHOT_OUTPUT=%s", *output))
	}

	return runCommand(remaining, env)
}
