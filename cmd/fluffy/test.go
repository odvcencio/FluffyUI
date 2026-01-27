package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
)

func runTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	visual := fs.Bool("visual", false, "enable visual testing")
	race := fs.Bool("race", false, "run with race detector")
	pkg := fs.String("pkg", "./...", "packages to test")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	pkgs := fs.Args()
	if len(pkgs) == 0 {
		pkgs = []string{*pkg}
	}
	cmdArgs := []string{"go", "test"}
	if *race {
		cmdArgs = append(cmdArgs, "-race")
	}
	cmdArgs = append(cmdArgs, pkgs...)

	env := os.Environ()
	if *visual {
		env = append(env, "FLUFFYUI_VISUAL=1")
	}
	return runCommand(cmdArgs, env)
}

func runCommand(args []string, env []string) error {
	if len(args) == 0 {
		return errors.New("missing command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if env != nil {
		cmd.Env = env
	}
	return cmd.Run()
}
