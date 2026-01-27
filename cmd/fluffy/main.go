package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	if value == "" {
		return nil
	}
	*s = append(*s, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "-h", "--help", "help":
		usage()
		return
	}
	switch os.Args[1] {
	case "create":
		if err := runCreate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "add":
		if err := runAdd(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "theme":
		if err := runTheme(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "test":
		if err := runTest(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "record":
		if err := runRecord(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "dev":
		if err := runDev(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `fluffy - FluffyUI developer tools

usage:
  fluffy dev [--watch path] [--ext .go,.fss] [--debounce 200ms] -- <cmd> [args...]
  fluffy create <name> [--template minimal|full|game] [--module path] [--force]
  fluffy add widget|page <Name> [--dir path] [--force]
  fluffy theme init|check|export [--path theme.yaml] [--output theme.css] [--force]
  fluffy test [--visual] [--race] [--pkg ./...]
  fluffy record [--output file.cast|file.gif] [--export file] [--title title] [-- <cmd>]
`)
}

func runDev(args []string) error {
	fs := flag.NewFlagSet("dev", flag.ContinueOnError)
	var watches stringSlice
	var exts string
	var debounce time.Duration
	fs.Var(&watches, "watch", "watch path (repeatable)")
	fs.StringVar(&exts, "ext", ".go,.fss,.yaml,.json", "comma-separated extensions")
	fs.DurationVar(&debounce, "debounce", 200*time.Millisecond, "restart debounce window")
	fs.SetOutput(os.Stderr)

	split := indexOf(args, "--")
	if split == -1 {
		return errors.New("missing -- separator before command")
	}
	if err := fs.Parse(args[:split]); err != nil {
		return err
	}
	cmdArgs := args[split+1:]
	if len(cmdArgs) == 0 {
		return errors.New("missing command after --")
	}
	if len(watches) == 0 {
		watches = append(watches, ".")
	}

	extSet := parseExts(exts)
	if len(extSet) == 0 {
		return errors.New("no extensions to watch")
	}

	restarts := make(chan struct{}, 1)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchLoop(watches, extSet, 500*time.Millisecond, debounce, restarts, stop)
	}()

	cmd, err := startCmd(cmdArgs)
	if err != nil {
		close(stop)
		wg.Wait()
		return err
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigc)

	for {
		select {
		case <-restarts:
			_ = stopCmd(cmd)
			cmd, err = startCmd(cmdArgs)
			if err != nil {
				close(stop)
				wg.Wait()
				return err
			}
		case <-sigc:
			close(stop)
			wg.Wait()
			_ = stopCmd(cmd)
			return nil
		}
	}
}

func startCmd(args []string) (*exec.Cmd, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopCmd(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		return nil
	}
}

func watchLoop(paths []string, exts map[string]struct{}, interval, debounce time.Duration, restart chan<- struct{}, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	last := map[string]time.Time{}
	_ = scanPaths(paths, exts, last)
	var lastChange time.Time
	var pending bool
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			changed := scanPaths(paths, exts, last)
			if changed {
				lastChange = time.Now()
				pending = true
			}
			if pending && time.Since(lastChange) >= debounce {
				pending = false
				select {
				case restart <- struct{}{}:
				default:
				}
			}
		}
	}
}

func scanPaths(paths []string, exts map[string]struct{}, last map[string]time.Time) bool {
	changed := false
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "vendor" || name == "node_modules" || name == "dist" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if _, ok := exts[ext]; !ok {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			mod := info.ModTime()
			if prev, ok := last[path]; !ok || mod.After(prev) {
				last[path] = mod
				changed = true
			}
			return nil
		})
	}
	return changed
}

func parseExts(value string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(value, ",") {
		ext := strings.TrimSpace(part)
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		out[strings.ToLower(ext)] = struct{}{}
	}
	return out
}

func indexOf(args []string, value string) int {
	for i, arg := range args {
		if arg == value {
			return i
		}
	}
	return -1
}
