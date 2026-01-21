package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type request map[string]any

type response struct {
	ID       int       `json:"id,omitempty"`
	OK       bool      `json:"ok,omitempty"`
	Error    string    `json:"error,omitempty"`
	Message  string    `json:"message,omitempty"`
	Snapshot *snapshot `json:"snapshot,omitempty"`
}

type snapshot struct {
	Widgets []widgetInfo `json:"widgets,omitempty"`
}

type widgetInfo struct {
	Label       string       `json:"label,omitempty"`
	Description string       `json:"description,omitempty"`
	Value       string       `json:"value,omitempty"`
	Children    []widgetInfo `json:"children,omitempty"`
}

var (
	addr         = flag.String("addr", "unix:/tmp/fluffyui.sock", "agent server address (unix:/path or tcp:host:port)")
	scriptPath   = flag.String("script", "", "path to JSONL driver script")
	token        = flag.String("token", "", "agent server token (if required)")
	dialTimeout  = flag.Duration("dial-timeout", 5*time.Second, "agent server dial timeout")
	recordPath   = flag.String("record", "", "record output path (sets FLUFFYUI_RECORD for child cmd)")
	exportPath   = flag.String("export", "", "record export output path (sets FLUFFYUI_RECORD_EXPORT for child cmd)")
	recordTitle  = flag.String("record-title", "FluffyUI Demo", "recording title (sets FLUFFYUI_RECORD_TITLE)")
	backend      = flag.String("backend", "", "backend override for child cmd (e.g., sim)")
	width        = flag.Int("width", 0, "backend width for child cmd (sim only)")
	height       = flag.Int("height", 0, "backend height for child cmd (sim only)")
	startupDelay = flag.Duration("startup-delay", 0, "extra delay before dialing the agent")
	verbose      = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()

	if strings.TrimSpace(*scriptPath) == "" {
		fail("script path is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmdArgs := flag.Args()
	var cmd *exec.Cmd
	if len(cmdArgs) > 0 {
		cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Env = append(os.Environ(), buildChildEnv()...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fail("start command: %v", err)
		}
	}

	if *startupDelay > 0 {
		time.Sleep(*startupDelay)
	}

	conn, err := dialAgent(*addr, *dialTimeout)
	if err != nil {
		fail("dial agent: %v", err)
	}
	defer conn.Close()

	client := newClient(conn)
	if _, err := client.send(request{
		"type":  "hello",
		"token": *token,
	}); err != nil && *token != "" {
		fail("hello failed: %v", err)
	}

	if err := runScript(ctx, client, *scriptPath); err != nil {
		fail("script failed: %v", err)
	}

	if cmd != nil {
		waitWithTimeout(cmd, 2*time.Second)
	}
}

func buildChildEnv() []string {
	env := []string{
		"FLUFFYUI_AGENT=" + *addr,
	}
	if *token != "" {
		env = append(env, "FLUFFYUI_AGENT_TOKEN="+*token)
	}
	if *recordPath != "" {
		env = append(env, "FLUFFYUI_RECORD="+*recordPath)
	}
	if *exportPath != "" {
		env = append(env, "FLUFFYUI_RECORD_EXPORT="+*exportPath)
	}
	if *recordTitle != "" && (*recordPath != "" || *exportPath != "") {
		env = append(env, "FLUFFYUI_RECORD_TITLE="+*recordTitle)
	}
	if *backend != "" {
		env = append(env, "FLUFFYUI_BACKEND="+*backend)
	}
	if *width > 0 {
		env = append(env, fmt.Sprintf("FLUFFYUI_WIDTH=%d", *width))
	}
	if *height > 0 {
		env = append(env, fmt.Sprintf("FLUFFYUI_HEIGHT=%d", *height))
	}
	return env
}

func runScript(ctx context.Context, client *agentClient, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		stepType, _ := req["type"].(string)
		switch strings.ToLower(stepType) {
		case "sleep":
			ms := intValue(req, "ms", intValue(req, "delay_ms", 0))
			if ms > 0 {
				sleepWithContext(ctx, time.Duration(ms)*time.Millisecond)
			}
			continue
		case "wait_label":
			label, _ := req["label"].(string)
			if strings.TrimSpace(label) == "" {
				return fmt.Errorf("line %d: wait_label requires label", lineNum)
			}
			timeoutMs := intValue(req, "timeout_ms", 2000)
			if err := waitForLabel(ctx, client, label, time.Duration(timeoutMs)*time.Millisecond); err != nil {
				return fmt.Errorf("line %d: %w", lineNum, err)
			}
			continue
		}

		resp, err := client.send(req)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
		if resp.Error != "" || !resp.OK {
			return fmt.Errorf("line %d: %s %s", lineNum, resp.Error, resp.Message)
		}

		delayMs := intValue(req, "delay_ms", 0)
		if delayMs > 0 {
			sleepWithContext(ctx, time.Duration(delayMs)*time.Millisecond)
		}
	}
	return scanner.Err()
}

func waitForLabel(ctx context.Context, client *agentClient, label string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.send(request{"type": "snapshot"})
		if err != nil {
			return err
		}
		if resp.Snapshot != nil && snapshotHasLabel(resp.Snapshot, label) {
			return nil
		}
		sleepWithContext(ctx, 100*time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for label %q", label)
}

func snapshotHasLabel(snap *snapshot, label string) bool {
	label = strings.ToLower(strings.TrimSpace(label))
	if snap == nil || label == "" {
		return false
	}
	for _, w := range snap.Widgets {
		if widgetHasLabel(w, label) {
			return true
		}
	}
	return false
}

func widgetHasLabel(w widgetInfo, label string) bool {
	if containsLabel(w.Label, label) || containsLabel(w.Description, label) || containsLabel(w.Value, label) {
		return true
	}
	for _, child := range w.Children {
		if widgetHasLabel(child, label) {
			return true
		}
	}
	return false
}

func containsLabel(value, label string) bool {
	if strings.TrimSpace(value) == "" || label == "" {
		return false
	}
	return strings.Contains(strings.ToLower(value), label)
}

type agentClient struct {
	conn net.Conn
	enc  *json.Encoder
	rd   *bufio.Reader
}

func newClient(conn net.Conn) *agentClient {
	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)
	return &agentClient{
		conn: conn,
		enc:  enc,
		rd:   bufio.NewReader(conn),
	}
}

func (c *agentClient) send(req request) (*response, error) {
	if *verbose {
		fmt.Fprintf(os.Stderr, "-> %s\n", mustJSON(req))
	}
	if err := c.enc.Encode(req); err != nil {
		return nil, err
	}
	line, err := c.rd.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "<- %s", string(line))
	}
	var resp response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func dialAgent(addr string, timeout time.Duration) (net.Conn, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		conn, err := dialOnce(addr)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func dialOnce(addr string) (net.Conn, error) {
	switch {
	case strings.HasPrefix(addr, "unix:"):
		path := strings.TrimPrefix(addr, "unix:")
		if strings.TrimSpace(path) == "" {
			return nil, errors.New("unix socket path is required")
		}
		return net.Dial("unix", path)
	case strings.HasPrefix(addr, "tcp:"):
		host := strings.TrimPrefix(addr, "tcp:")
		if strings.TrimSpace(host) == "" {
			return nil, errors.New("tcp address is required")
		}
		return net.Dial("tcp", host)
	default:
		return nil, fmt.Errorf("unsupported address %q (use unix: or tcp:)", addr)
	}
}

func waitWithTimeout(cmd *exec.Cmd, timeout time.Duration) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
}

func intValue(req request, key string, fallback int) int {
	raw, ok := req[key]
	if !ok {
		return fallback
	}
	switch v := raw.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func sleepWithContext(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func mustJSON(req request) string {
	data, err := json.Marshal(req)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
