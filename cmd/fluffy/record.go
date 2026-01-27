package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

func runRecord(args []string) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	output := fs.String("output", "session.cast", "record output path (.cast or export file)")
	export := fs.String("export", "", "export output path (gif/mp4)")
	title := fs.String("title", "", "recording title")
	fs.SetOutput(os.Stderr)

	split := indexOf(args, "--")
	var cmdArgs []string
	if split == -1 {
		if err := fs.Parse(args); err != nil {
			return err
		}
		cmdArgs = []string{"go", "run", "."}
	} else {
		if err := fs.Parse(args[:split]); err != nil {
			return err
		}
		cmdArgs = args[split+1:]
		if len(cmdArgs) == 0 {
			return errors.New("missing command after --")
		}
	}

	castPath := *output
	exportPath := *export
	if ext := strings.ToLower(filepath.Ext(castPath)); ext != "" && ext != ".cast" {
		if exportPath == "" {
			exportPath = castPath
		}
		castPath = strings.TrimSuffix(castPath, ext) + ".cast"
	}
	env := os.Environ()
	env = append(env, "FLUFFYUI_RECORD="+castPath)
	if exportPath != "" {
		env = append(env, "FLUFFYUI_RECORD_EXPORT="+exportPath)
	}
	if *title != "" {
		env = append(env, "FLUFFYUI_RECORD_TITLE="+*title)
	}

	return runCommand(cmdArgs, env)
}
