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
	"os/signal"
	"sort"
	"strings"
	"syscall"
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
	Timestamp time.Time    `json:"timestamp,omitempty"`
	Width     int          `json:"width,omitempty"`
	Height    int          `json:"height,omitempty"`
	Text      string       `json:"text,omitempty"`
	Widgets   []widgetInfo `json:"widgets,omitempty"`
	FocusedID string       `json:"focused_id,omitempty"`
	Focused   *widgetInfo  `json:"focused,omitempty"`
}

type widgetInfo struct {
	ID          string       `json:"id,omitempty"`
	Role        string       `json:"type,omitempty"`
	Label       string       `json:"label,omitempty"`
	Description string       `json:"description,omitempty"`
	Value       string       `json:"value,omitempty"`
	Children    []widgetInfo `json:"children,omitempty"`
}

type Action struct {
	Type        string
	Key         string
	Text        string
	X           int
	Y           int
	Button      string
	MouseAction string
	Width       int
	Height      int
	Delay       time.Duration
}

type Decision struct {
	Reason  string
	Actions []Action
}

type Policy interface {
	Name() string
	Reset()
	Decide(context.Context, *snapshot) (Decision, error)
}

type policyFactory func() Policy

type runner struct {
	client      *agentClient
	policy      Policy
	interval    time.Duration
	maxSteps    int
	maxRuntime  time.Duration
	includeText bool
	logger      *logWriter
	verbose     bool
}

type logWriter struct {
	file *os.File
	enc  *json.Encoder
}

type logEvent struct {
	Time     time.Time   `json:"time"`
	Type     string      `json:"type"`
	Step     int         `json:"step,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Snapshot *snapshot   `json:"snapshot,omitempty"`
	Action   *actionLog  `json:"action,omitempty"`
	Actions  []actionLog `json:"actions,omitempty"`
	Response *response   `json:"response,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type actionLog struct {
	Type        string `json:"type"`
	Key         string `json:"key,omitempty"`
	Text        string `json:"text,omitempty"`
	X           int    `json:"x,omitempty"`
	Y           int    `json:"y,omitempty"`
	Button      string `json:"button,omitempty"`
	MouseAction string `json:"action,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	DelayMs     int    `json:"delay_ms,omitempty"`
}

var (
	addr         = flag.String("addr", "unix:/tmp/fluffyui.sock", "agent server address (unix:/path or tcp:host:port)")
	token        = flag.String("token", "", "agent server token (if required)")
	policyName   = flag.String("policy", "noop", "policy to drive actions")
	listPolicies = flag.Bool("list-policies", false, "list available policies and exit")
	interval     = flag.Duration("interval", 250*time.Millisecond, "sleep between decision loops")
	maxSteps     = flag.Int("max-steps", 0, "max decision steps (0 = unlimited)")
	maxRuntime   = flag.Duration("max-runtime", 0, "max runtime (0 = unlimited)")
	includeText  = flag.Bool("include-text", false, "request raw screen text in snapshots (requires AllowText)")
	logPath      = flag.String("log", "", "write JSONL decision/action log to path")
	dialTimeout  = flag.Duration("dial-timeout", 5*time.Second, "agent server dial timeout")
	startupDelay = flag.Duration("startup-delay", 0, "extra delay before dialing the agent")
	recordPath   = flag.String("record", "", "record output path (sets FLUFFYUI_RECORD for child cmd)")
	exportPath   = flag.String("export", "", "record export output path (sets FLUFFYUI_RECORD_EXPORT for child cmd)")
	recordTitle  = flag.String("record-title", "FluffyUI Demo", "recording title (sets FLUFFYUI_RECORD_TITLE)")
	backend      = flag.String("backend", "", "backend override for child cmd (e.g., sim)")
	width        = flag.Int("width", 0, "backend width for child cmd (sim only)")
	height       = flag.Int("height", 0, "backend height for child cmd (sim only)")
	verbose      = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()

	if *listPolicies {
		printPolicies()
		return
	}

	factory, ok := policyRegistry()[strings.ToLower(strings.TrimSpace(*policyName))]
	if !ok {
		fail("unknown policy %q", *policyName)
	}
	policy := factory()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
		sleepWithContext(ctx, *startupDelay)
	}

	conn, err := dialAgent(*addr, *dialTimeout)
	if err != nil {
		fail("dial agent: %v", err)
	}
	defer conn.Close()

	client := newClient(conn, *verbose)
	if err := client.hello(*token); err != nil {
		fail("hello failed: %v", err)
	}

	logger, err := newLogWriter(*logPath)
	if err != nil {
		fail("log init: %v", err)
	}
	if logger != nil {
		defer logger.Close()
	}

	if *includeText {
		fmt.Fprintln(os.Stderr, "include-text enabled: requires AllowText on the server")
	}

	policy.Reset()
	r := &runner{
		client:      client,
		policy:      policy,
		interval:    *interval,
		maxSteps:    *maxSteps,
		maxRuntime:  *maxRuntime,
		includeText: *includeText,
		logger:      logger,
		verbose:     *verbose,
	}

	if err := r.run(ctx); err != nil {
		fail("run failed: %v", err)
	}

	if cmd != nil {
		waitWithTimeout(cmd, 2*time.Second)
	}
}

func (r *runner) run(ctx context.Context) error {
	start := time.Now()
	steps := 0
	for {
		if r.maxRuntime > 0 && time.Since(start) >= r.maxRuntime {
			return nil
		}
		if r.maxSteps > 0 && steps >= r.maxSteps {
			return nil
		}

		snap, err := r.client.snapshot(r.includeText)
		if err != nil {
			return err
		}
		if r.logger != nil {
			if err := r.logger.write(logEvent{
				Time:     time.Now(),
				Type:     "snapshot",
				Step:     steps,
				Snapshot: snap,
			}); err != nil {
				return err
			}
		}

		decision, err := r.policy.Decide(ctx, snap)
		if err != nil {
			return err
		}
		if r.logger != nil {
			if err := r.logger.write(logEvent{
				Time:    time.Now(),
				Type:    "decision",
				Step:    steps,
				Reason:  decision.Reason,
				Actions: toActionLogs(decision.Actions),
			}); err != nil {
				return err
			}
		}

		if len(decision.Actions) == 0 {
			sleepWithContext(ctx, r.interval)
			continue
		}

		for _, action := range decision.Actions {
			if strings.EqualFold(action.Type, "sleep") {
				sleepWithContext(ctx, action.Delay)
				continue
			}
			resp, err := r.client.send(actionToRequest(action))
			if err != nil {
				return err
			}
			if resp.Error != "" || !resp.OK {
				return fmt.Errorf("action %s failed: %s %s", action.Type, resp.Error, resp.Message)
			}
			if r.logger != nil {
				if err := r.logger.write(logEvent{
					Time:     time.Now(),
					Type:     "action",
					Step:     steps,
					Action:   toActionLog(action),
					Response: resp,
				}); err != nil {
					return err
				}
			}
			if action.Delay > 0 {
				sleepWithContext(ctx, action.Delay)
			}
		}

		steps++
		sleepWithContext(ctx, r.interval)
	}
}

func toActionLogs(actions []Action) []actionLog {
	if len(actions) == 0 {
		return nil
	}
	out := make([]actionLog, 0, len(actions))
	for _, action := range actions {
		out = append(out, *toActionLog(action))
	}
	return out
}

func toActionLog(action Action) *actionLog {
	return &actionLog{
		Type:        action.Type,
		Key:         action.Key,
		Text:        action.Text,
		X:           action.X,
		Y:           action.Y,
		Button:      action.Button,
		MouseAction: action.MouseAction,
		Width:       action.Width,
		Height:      action.Height,
		DelayMs:     int(action.Delay / time.Millisecond),
	}
}

func newLogWriter(path string) (*logWriter, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(file)
	enc.SetEscapeHTML(false)
	return &logWriter{file: file, enc: enc}, nil
}

func (l *logWriter) write(event logEvent) error {
	if l == nil || l.enc == nil {
		return nil
	}
	return l.enc.Encode(event)
}

func (l *logWriter) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func actionToRequest(action Action) request {
	req := request{
		"type": strings.ToLower(strings.TrimSpace(action.Type)),
	}
	switch req["type"] {
	case "key":
		req["key"] = action.Key
	case "text":
		req["text"] = action.Text
	case "mouse":
		req["x"] = action.X
		req["y"] = action.Y
		req["button"] = action.Button
		req["action"] = action.MouseAction
	case "paste":
		req["text"] = action.Text
	case "resize":
		req["width"] = action.Width
		req["height"] = action.Height
	}
	return req
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
	if *includeText {
		env = append(env, "FLUFFYUI_AGENT_ALLOW_TEXT=1")
	}
	return env
}

func printPolicies() {
	keys := make([]string, 0, len(policyRegistry()))
	for key := range policyRegistry() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Println(key)
	}
}

func policyRegistry() map[string]policyFactory {
	return map[string]policyFactory{
		"noop":            func() Policy { return &noopPolicy{} },
		"candy-wars-demo": func() Policy { return newCandyWarsPolicy() },
	}
}

type noopPolicy struct{}

func (n *noopPolicy) Name() string { return "noop" }

func (n *noopPolicy) Reset() {}

func (n *noopPolicy) Decide(ctx context.Context, snap *snapshot) (Decision, error) {
	return Decision{Reason: "noop"}, nil
}

type candyWarsPolicy struct {
	step int
	plan []actionPlan
}

type actionPlan struct {
	Name    string
	Actions []Action
}

func newCandyWarsPolicy() Policy {
	return &candyWarsPolicy{
		plan: []actionPlan{
			{
				Name: "buy",
				Actions: []Action{
					keyAction("b", 150*time.Millisecond),
					textAction("3", 100*time.Millisecond),
					keyAction("enter", 250*time.Millisecond),
				},
			},
			{
				Name: "travel",
				Actions: []Action{
					keyAction("2", 200*time.Millisecond),
				},
			},
			{
				Name: "sell",
				Actions: []Action{
					keyAction("s", 150*time.Millisecond),
					textAction("2", 100*time.Millisecond),
					keyAction("enter", 250*time.Millisecond),
				},
			},
			{
				Name: "end_day",
				Actions: []Action{
					keyAction("e", 250*time.Millisecond),
				},
			},
			{
				Name: "tabs",
				Actions: []Action{
					keyAction("right", 150*time.Millisecond),
					keyAction("right", 150*time.Millisecond),
					keyAction("left", 150*time.Millisecond),
				},
			},
		},
	}
}

func (c *candyWarsPolicy) Name() string { return "candy-wars-demo" }

func (c *candyWarsPolicy) Reset() { c.step = 0 }

func (c *candyWarsPolicy) Decide(ctx context.Context, snap *snapshot) (Decision, error) {
	if snapHasAnyLabel(snap, "normal -", "nightmare -", "hell -") {
		return Decision{
			Reason: "start game",
			Actions: []Action{
				keyAction("enter", 300*time.Millisecond),
			},
		}, nil
	}

	plan := c.plan[c.step%len(c.plan)]
	c.step++
	return Decision{Reason: plan.Name, Actions: plan.Actions}, nil
}

func keyAction(key string, delay time.Duration) Action {
	return Action{Type: "key", Key: key, Delay: delay}
}

func textAction(text string, delay time.Duration) Action {
	return Action{Type: "text", Text: text, Delay: delay}
}

func snapHasLabelContains(snap *snapshot, label string) bool {
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

func snapHasAnyLabel(snap *snapshot, labels ...string) bool {
	for _, label := range labels {
		if snapHasLabelContains(snap, label) {
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
	conn    net.Conn
	enc     *json.Encoder
	reader  *bufio.Reader
	verbose bool
}

func newClient(conn net.Conn, verbose bool) *agentClient {
	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)
	return &agentClient{
		conn:    conn,
		enc:     enc,
		reader:  bufio.NewReader(conn),
		verbose: verbose,
	}
}

func (c *agentClient) hello(token string) error {
	resp, err := c.send(request{
		"type":  "hello",
		"token": token,
	})
	if err != nil {
		return err
	}
	if resp.Error != "" || !resp.OK {
		return fmt.Errorf("%s %s", resp.Error, resp.Message)
	}
	return nil
}

func (c *agentClient) snapshot(includeText bool) (*snapshot, error) {
	resp, err := c.send(request{
		"type":         "snapshot",
		"include_text": includeText,
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" || !resp.OK {
		return nil, fmt.Errorf("snapshot failed: %s %s", resp.Error, resp.Message)
	}
	return resp.Snapshot, nil
}

func (c *agentClient) send(req request) (*response, error) {
	if c.verbose {
		fmt.Fprintf(os.Stderr, "-> %s\n", mustJSON(req))
	}
	if err := c.enc.Encode(req); err != nil {
		return nil, err
	}
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if c.verbose {
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
