// Command fluffy-speak is a screen reader bridge for FluffyUI applications.
// It connects to a FluffyUI MCP server, reads ARIA-like accessibility semantics
// from the widget tree, and speaks changes through platform TTS engines.
//
// Usage:
//
//	fluffy-speak [flags] [-- command args...]
//
// Flags:
//
//	--addr        MCP server address (unix:/path, tcp:host:port, or sse:url)
//	--transport   Transport type: unix, stdio, sse (auto-detected from addr)
//	--token       Authentication token
//	--tts         TTS backend: auto, espeak, spd, say, sapi, mock
//	--rate        Speech rate multiplier (default: 1.0)
//	--dry-run     Print announcements to stderr instead of speaking
//	--log         Write JSONL event log to path
//	-v            Verbose output
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mcp "m31labs.dev/fluffyui/agent/mcp"
	"m31labs.dev/fluffyui/accessibility/tts"
)

var (
	addr      = flag.String("addr", "", "MCP server address (unix:/path, tcp:host:port, or URL)")
	transport = flag.String("transport", "", "Transport: unix, stdio, sse (auto-detected)")
	token     = flag.String("token", "", "Authentication token")
	ttsName   = flag.String("tts", "auto", "TTS backend: auto, espeak, spd, say, sapi, mock")
	rate      = flag.Float64("rate", 1.0, "Speech rate multiplier")
	dryRun    = flag.Bool("dry-run", false, "Print to stderr instead of speaking")
	logPath   = flag.String("log", "", "JSONL event log path")
	verbose   = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Optionally launch the app as a subprocess
	cmdArgs := flag.Args()
	var cmd *exec.Cmd
	if len(cmdArgs) > 0 {
		cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Env = buildEnv()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fail("start command: %v", err)
		}
		// Give the app time to start its MCP server
		sleepCtx(ctx, 500*time.Millisecond)
	}

	// Connect to MCP server
	client, err := connectMCP()
	if err != nil {
		fail("connect: %v", err)
	}
	defer client.Close()

	// Initialize TTS
	var speaker Speaker
	if *dryRun {
		speaker = tts.NewMock()
	} else {
		speaker, err = detectSpeaker(*ttsName, *rate)
		if err != nil {
			fail("tts: %v", err)
		}
	}
	defer speaker.Close()

	// Initialize logger
	logger, err := newLogWriter(*logPath)
	if err != nil {
		fail("log: %v", err)
	}
	if logger != nil {
		defer logger.Close()
	}

	// Create speech queue and runner
	queue := NewSpeechQueue(speaker, 50*time.Millisecond)
	go queue.Run(ctx)

	runner := &speakRunner{
		client:  client,
		queue:   queue,
		logger:  logger,
		verbose: *verbose,
	}

	if *verbose {
		fmt.Fprintln(os.Stderr, "fluffy-speak connected, listening for changes...")
	}

	if err := runner.run(ctx); err != nil {
		fail("run: %v", err)
	}

	if cmd != nil {
		waitCmd(cmd, 2*time.Second)
	}
}

func connectMCP() (*mcp.Client, error) {
	opts := mcp.ClientOptions{
		Token:   *token,
		Timeout: 10 * time.Second,
	}

	if *transport != "" {
		opts.Transport = *transport
	}

	address := strings.TrimSpace(*addr)
	if address != "" {
		opts.Addr = address
		if opts.Transport == "" {
			if strings.HasPrefix(address, "/") || strings.HasPrefix(address, "unix:") {
				opts.Transport = "unix"
				opts.Addr = strings.TrimPrefix(address, "unix:")
			} else if strings.HasPrefix(address, "http") || strings.HasPrefix(address, "sse:") {
				opts.Transport = "sse"
				opts.Addr = strings.TrimPrefix(address, "sse:")
			} else {
				opts.Transport = "unix"
			}
		}
	}

	// If launching a subprocess, try stdio
	if address == "" && len(flag.Args()) > 0 {
		opts.Transport = "stdio"
		opts.Command = flag.Args()[0]
		opts.Args = flag.Args()[1:]
	}

	if opts.Transport == "" && opts.Addr == "" && opts.Command == "" {
		return nil, fmt.Errorf("no MCP address specified; use --addr or provide a command")
	}

	return mcp.ConnectWithOptions(opts)
}

func buildEnv() []string {
	env := os.Environ()
	addr := strings.TrimSpace(*addr)
	if addr != "" {
		env = append(env, "FLUFFY_MCP=unix://"+addr)
	} else {
		env = append(env, "FLUFFY_MCP=1")
	}
	if *token != "" {
		env = append(env, "FLUFFY_MCP_TOKEN="+*token)
	}
	return env
}

func waitCmd(cmd *exec.Cmd, timeout time.Duration) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
}

func sleepCtx(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fluffy-speak: "+format+"\n", args...)
	os.Exit(1)
}
