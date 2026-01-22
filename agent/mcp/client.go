package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptr "github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/term"
)

const (
	defaultConnectTimeout = 5 * time.Second
	defaultSchemaVersion  = "v1"
)

// ClientOptions configures the Go MCP client.
type ClientOptions struct {
	Transport     string
	Command       string
	Args          []string
	Addr          string
	Token         string
	Timeout       time.Duration
	SchemaVersion string
	Env           map[string]string
}

// Client wraps the MCP client with fluffy-specific helpers.
type Client struct {
	inner       *mcpclient.Client
	transport   string
	timeout     time.Duration
	cmd         *exec.Cmd
	ownsProcess bool
	subsMu      sync.RWMutex
	subscribers map[string][]subscriptionHandler
}

// ResourceEvent is delivered for resources/updated notifications.
type ResourceEvent struct {
	URI    string
	Reason string
	NewURI string
}

// Connect creates a client using command or explicit transport.
func Connect(args ...string) (*Client, error) {
	if len(args) == 0 {
		return nil, ErrInvalidTransport
	}
	transport := strings.ToLower(strings.TrimSpace(args[0]))
	switch transport {
	case "stdio":
		if len(args) < 2 {
			return nil, errors.New("stdio transport requires command")
		}
		return ConnectWithOptions(ClientOptions{
			Transport: "stdio",
			Command:   args[1],
			Args:      args[2:],
		})
	case "unix":
		if len(args) < 2 {
			return nil, errors.New("unix transport requires socket path")
		}
		return ConnectWithOptions(ClientOptions{
			Transport: "unix",
			Addr:      args[1],
		})
	case "sse":
		if len(args) < 2 {
			return nil, errors.New("sse transport requires URL")
		}
		return ConnectWithOptions(ClientOptions{
			Transport: "sse",
			Addr:      args[1],
		})
	default:
		return ConnectWithOptions(ClientOptions{
			Command: args[0],
			Args:    args[1:],
		})
	}
}

// ConnectWithOptions creates and initializes a client based on options.
func ConnectWithOptions(opts ClientOptions) (*Client, error) {
	transport := strings.ToLower(strings.TrimSpace(opts.Transport))
	if transport == "" {
		if opts.Command != "" && strings.TrimSpace(opts.Addr) == "" {
			transport = defaultClientTransport()
		} else if opts.Addr != "" {
			transport = "sse"
		} else {
			return nil, ErrInvalidTransport
		}
	}

	switch transport {
	case "stdio":
		return connectStdio(opts)
	case "unix":
		return connectUnix(opts)
	case "sse":
		return connectSSE(opts)
	default:
		return nil, ErrInvalidTransport
	}
}

func defaultClientTransport() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "unix"
	}
	return "stdio"
}

// Close shuts down the underlying client and any managed process.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	var err error
	if c.inner != nil {
		err = c.inner.Close()
	}
	if c.ownsProcess && c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_, _ = c.cmd.Process.Wait()
	}
	return err
}

// Subscribe registers for resource updates and invokes handler on changes.
// The handler may accept a ResourceEvent or a decoded resource payload.
func (c *Client) Subscribe(uri string, handler any) error {
	if c == nil || c.inner == nil {
		return ErrNotConnected
	}
	sub, err := parseSubscriptionHandler(handler)
	if err != nil {
		return err
	}
	c.subsMu.Lock()
	if c.subscribers == nil {
		c.subscribers = make(map[string][]subscriptionHandler)
	}
	c.subscribers[uri] = append(c.subscribers[uri], sub)
	c.subsMu.Unlock()

	ctx, cancel := c.callContext()
	defer cancel()
	return c.inner.Subscribe(ctx, mcp.SubscribeRequest{
		Params: mcp.SubscribeParams{URI: uri},
	})
}

// Unsubscribe removes subscriptions for a resource URI.
func (c *Client) Unsubscribe(uri string) error {
	if c == nil || c.inner == nil {
		return ErrNotConnected
	}
	c.subsMu.Lock()
	delete(c.subscribers, uri)
	c.subsMu.Unlock()

	ctx, cancel := c.callContext()
	defer cancel()
	return c.inner.Unsubscribe(ctx, mcp.UnsubscribeRequest{
		Params: mcp.UnsubscribeParams{URI: uri},
	})
}

// Resubscribe moves handlers from old URI to new URI and re-subscribes.
func (c *Client) Resubscribe(oldURI, newURI string) error {
	if c == nil || c.inner == nil {
		return ErrNotConnected
	}
	c.subsMu.Lock()
	handlers := c.subscribers[oldURI]
	delete(c.subscribers, oldURI)
	if len(handlers) > 0 {
		c.subscribers[newURI] = append(c.subscribers[newURI], handlers...)
	}
	c.subsMu.Unlock()

	ctx, cancel := c.callContext()
	defer cancel()
	if err := c.inner.Unsubscribe(ctx, mcp.UnsubscribeRequest{
		Params: mcp.UnsubscribeParams{URI: oldURI},
	}); err != nil {
		return err
	}
	if len(handlers) == 0 {
		return nil
	}
	return c.inner.Subscribe(ctx, mcp.SubscribeRequest{
		Params: mcp.SubscribeParams{URI: newURI},
	})
}

// CallTool sends a raw tools/call request.
func (c *Client) CallTool(ctx context.Context, name string, args any) (*mcp.CallToolResult, error) {
	if c == nil || c.inner == nil {
		return nil, ErrNotConnected
	}
	return c.inner.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
}

func callTool[T any](c *Client, name string, args any) (T, error) {
	ctx, cancel := c.callContext()
	defer cancel()
	return callToolWithContext[T](c, ctx, name, args)
}

func callToolWithContext[T any](c *Client, ctx context.Context, name string, args any) (T, error) {
	var zero T
	result, err := c.CallTool(ctx, name, args)
	if err != nil {
		return zero, err
	}
	envelope, err := parseToolEnvelope(result)
	if err != nil {
		return zero, err
	}
	if result.IsError {
		if envelope.Error == "" {
			return zero, errors.New("tool returned error")
		}
		return zero, errors.New(envelope.Error)
	}
	if envelope.Error != "" {
		return zero, errors.New(envelope.Error)
	}
	if len(envelope.Data) == 0 {
		return zero, nil
	}
	var out T
	if err := json.Unmarshal(envelope.Data, &out); err != nil {
		return zero, err
	}
	return out, nil
}

func (c *Client) callContext() (context.Context, func()) {
	if c == nil || c.timeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), c.timeout)
}

func connectStdio(opts ClientOptions) (*Client, error) {
	if strings.TrimSpace(opts.Command) == "" {
		return nil, errors.New("stdio transport requires command")
	}
	env := buildEnv(opts.Env, map[string]string{
		"FLUFFY_MCP": "stdio",
	})
	if opts.Token != "" {
		env["FLUFFY_MCP_TOKEN"] = opts.Token
	}
	envSlice := envSlice(env)

	transport := mcptr.NewStdioWithOptions(opts.Command, envSlice, opts.Args)
	inner := mcpclient.NewClient(transport)

	if err := startTransport(inner, opts.Timeout); err != nil {
		return nil, err
	}
	client := &Client{
		inner:     inner,
		transport: "stdio",
		timeout:   opts.Timeout,
	}
	if err := client.initialize(opts); err != nil {
		_ = client.Close()
		return nil, err
	}
	client.inner.OnNotification(client.handleNotification)
	return client, nil
}

func connectUnix(opts ClientOptions) (*Client, error) {
	addr := strings.TrimSpace(opts.Addr)
	var cmd *exec.Cmd
	if opts.Command != "" {
		if addr == "" {
			socketPath, err := tempSocketPath()
			if err != nil {
				return nil, err
			}
			addr = socketPath
		}
		env := buildEnv(opts.Env, map[string]string{
			"FLUFFY_MCP": fmt.Sprintf("unix://%s", addr),
		})
		if opts.Token != "" {
			env["FLUFFY_MCP_TOKEN"] = opts.Token
		}
		envSlice := envSlice(env)
		cmd = exec.Command(opts.Command, opts.Args...)
		cmd.Env = append(os.Environ(), envSlice...)
		if err := cmd.Start(); err != nil {
			return nil, err
		}
	}
	if addr == "" {
		return nil, errors.New("unix transport requires socket path")
	}

	conn, err := waitForUnix(addr, effectiveTimeout(opts.Timeout))
	if err != nil {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil, err
	}
	transport := mcptr.NewIO(conn, conn, nil)
	inner := mcpclient.NewClient(transport)
	if err := startTransport(inner, opts.Timeout); err != nil {
		return nil, err
	}
	client := &Client{
		inner:       inner,
		transport:   "unix",
		timeout:     opts.Timeout,
		cmd:         cmd,
		ownsProcess: cmd != nil,
	}
	if err := client.initialize(opts); err != nil {
		_ = client.Close()
		return nil, err
	}
	client.inner.OnNotification(client.handleNotification)
	return client, nil
}

func connectSSE(opts ClientOptions) (*Client, error) {
	if strings.TrimSpace(opts.Addr) == "" {
		return nil, errors.New("sse transport requires URL")
	}
	baseURL, err := normalizeSSEURL(opts.Addr)
	if err != nil {
		return nil, err
	}
	var cmd *exec.Cmd
	if opts.Command != "" {
		hostPort, err := sseHostPort(baseURL)
		if err != nil {
			return nil, err
		}
		env := buildEnv(opts.Env, map[string]string{
			"FLUFFY_MCP": fmt.Sprintf("sse://%s", hostPort),
		})
		if opts.Token != "" {
			env["FLUFFY_MCP_TOKEN"] = opts.Token
		}
		envSlice := envSlice(env)
		cmd = exec.Command(opts.Command, opts.Args...)
		cmd.Env = append(os.Environ(), envSlice...)
		if err := cmd.Start(); err != nil {
			return nil, err
		}
	}

	headers := map[string]string{}
	if opts.Token != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", opts.Token)
	}
	transport, err := mcptr.NewSSE(baseURL, mcptr.WithHeaders(headers))
	if err != nil {
		return nil, err
	}
	inner := mcpclient.NewClient(transport)
	if err := startTransport(inner, opts.Timeout); err != nil {
		return nil, err
	}
	client := &Client{
		inner:       inner,
		transport:   "sse",
		timeout:     opts.Timeout,
		cmd:         cmd,
		ownsProcess: cmd != nil,
	}
	if err := client.initialize(opts); err != nil {
		_ = client.Close()
		return nil, err
	}
	client.inner.OnNotification(client.handleNotification)
	return client, nil
}

func startTransport(inner *mcpclient.Client, timeout time.Duration) error {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	return inner.Start(ctx)
}

func (c *Client) initialize(opts ClientOptions) error {
	version := strings.TrimSpace(opts.SchemaVersion)
	if version == "" {
		version = defaultSchemaVersion
	}
	caps := mcp.ClientCapabilities{
		Experimental: map[string]any{
			"fluffy": map[string]any{
				"schemaVersion": version,
			},
		},
	}
	params := mcp.InitializeParams{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ClientInfo: mcp.Implementation{
			Name:    "fluffy-mcp-client",
			Version: "dev",
		},
		Capabilities: caps,
	}
	if opts.Token != "" {
		params.Auth = &mcp.AuthParams{Token: opts.Token}
	}
	ctx := context.Background()
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	_, err := c.inner.Initialize(ctx, mcp.InitializeRequest{Params: params})
	return err
}

func normalizeSSEURL(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimPrefix(addr, "sse://")
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	parsed, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/sse"
	}
	return parsed.String(), nil
}

func sseHostPort(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid SSE url %q", baseURL)
	}
	return parsed.Host, nil
}

func tempSocketPath() (string, error) {
	dir := os.TempDir()
	if dir == "" {
		return "", errors.New("temp dir not available")
	}
	file, err := os.CreateTemp(dir, "fluffy-mcp-*.sock")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
}

func waitForUnix(path string, timeout time.Duration) (net.Conn, error) {
	if timeout <= 0 {
		timeout = defaultConnectTimeout
	}
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("unix", path, 100*time.Millisecond)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func effectiveTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultConnectTimeout
	}
	return timeout
}

func buildEnv(base map[string]string, overrides map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}

func envSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s=%s", k, env[k]))
	}
	return out
}

type toolEnvelope struct {
	Schema string          `json:"_schema"`
	Tool   string          `json:"_tool"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func parseToolEnvelope(result *mcp.CallToolResult) (toolEnvelope, error) {
	var envelope toolEnvelope
	if result == nil {
		return envelope, errors.New("empty tool result")
	}
	payload := result.StructuredContent
	if payload == nil {
		for _, content := range result.Content {
			switch value := content.(type) {
			case mcp.TextContent:
				payload = json.RawMessage(value.Text)
			case *mcp.TextContent:
				payload = json.RawMessage(value.Text)
			}
			if payload != nil {
				break
			}
		}
	}
	if payload == nil {
		return envelope, errors.New("missing structured content")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return envelope, err
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return envelope, err
	}
	return envelope, nil
}

type subscriptionHandler struct {
	fn           reflect.Value
	argType      reflect.Type
	expectsEvent bool
}

func parseSubscriptionHandler(handler any) (subscriptionHandler, error) {
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return subscriptionHandler{}, errors.New("handler must be a function")
	}
	handlerType := value.Type()
	if handlerType.NumIn() != 1 || handlerType.NumOut() != 0 {
		return subscriptionHandler{}, errors.New("handler must accept one argument and return nothing")
	}
	argType := handlerType.In(0)
	eventType := reflect.TypeOf(ResourceEvent{})
	if argType == eventType || (argType.Kind() == reflect.Pointer && argType.Elem() == eventType) {
		return subscriptionHandler{fn: value, argType: argType, expectsEvent: true}, nil
	}
	return subscriptionHandler{fn: value, argType: argType}, nil
}

func (c *Client) handleNotification(notification mcp.JSONRPCNotification) {
	if notification.Method != mcp.MethodNotificationResourceUpdated {
		return
	}
	fields := notification.Params.AdditionalFields
	if len(fields) == 0 {
		return
	}
	uri, _ := fields["uri"].(string)
	if uri == "" {
		return
	}
	event := ResourceEvent{URI: uri}
	if reason, ok := fields["reason"].(string); ok {
		event.Reason = reason
	}
	if newURI, ok := fields["new_uri"].(string); ok {
		event.NewURI = newURI
	}

	handlers := c.handlersForURI(uri)
	if len(handlers) == 0 {
		return
	}
	for _, handler := range handlers {
		if handler.expectsEvent {
			c.invokeHandler(handler, reflect.ValueOf(event))
			continue
		}
		if event.Reason != "" {
			continue
		}
		value, err := c.readResourceValue(uri, handler.argType)
		if err != nil {
			continue
		}
		c.invokeHandler(handler, value)
	}
}

func (c *Client) handlersForURI(uri string) []subscriptionHandler {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()
	handlers := c.subscribers[uri]
	if len(handlers) == 0 {
		return nil
	}
	out := make([]subscriptionHandler, len(handlers))
	copy(out, handlers)
	return out
}

func (c *Client) invokeHandler(handler subscriptionHandler, value reflect.Value) {
	if !value.IsValid() {
		return
	}
	if value.Type() != handler.argType {
		if handler.argType.Kind() == reflect.Pointer && handler.argType.Elem() == value.Type() {
			ptr := reflect.New(value.Type())
			ptr.Elem().Set(value)
			value = ptr
		} else if value.Kind() == reflect.Pointer && value.Type().Elem() == handler.argType {
			value = value.Elem()
		} else {
			return
		}
	}
	handler.fn.Call([]reflect.Value{value})
}

func (c *Client) readResourceValue(uri string, typ reflect.Type) (reflect.Value, error) {
	ctx, cancel := c.callContext()
	defer cancel()
	result, err := c.inner.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: uri},
	})
	if err != nil {
		return reflect.Value{}, err
	}
	text, mime, err := resourceTextFromResult(result)
	if err != nil {
		return reflect.Value{}, err
	}
	if typ.Kind() == reflect.String && mime != "application/json" {
		return reflect.ValueOf(text), nil
	}
	target := reflect.New(typ)
	if typ.Kind() == reflect.Pointer {
		target = reflect.New(typ)
		if err := json.Unmarshal([]byte(text), target.Interface()); err != nil {
			return reflect.Value{}, err
		}
		return target.Elem(), nil
	}
	if err := json.Unmarshal([]byte(text), target.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return target.Elem(), nil
}

func resourceTextFromResult(result *mcp.ReadResourceResult) (string, string, error) {
	if result == nil {
		return "", "", errors.New("empty resource result")
	}
	for _, content := range result.Contents {
		switch value := content.(type) {
		case mcp.TextResourceContents:
			return value.Text, value.MIMEType, nil
		case *mcp.TextResourceContents:
			return value.Text, value.MIMEType, nil
		}
	}
	return "", "", errors.New("resource content not text")
}

// TestClient starts an app and returns a ready client for tests.
func TestClient(t testing.TB, command string, options ...TestOption) *Client {
	t.Helper()
	opts := ClientOptions{
		Command: command,
	}
	for _, opt := range options {
		opt(&opts)
	}
	client, err := ConnectWithOptions(opts)
	if err != nil {
		t.Fatalf("mcp test client failed: %v", err)
	}
	return client
}

// TestOption configures TestClient.
type TestOption func(*ClientOptions)

// WithTextAccess enables text access for TestClient.
func WithTextAccess() TestOption {
	return func(opts *ClientOptions) {
		if opts.Env == nil {
			opts.Env = make(map[string]string)
		}
		opts.Env["FLUFFY_MCP_ALLOW_TEXT"] = "1"
	}
}

// WithClipboardAccess enables clipboard access for TestClient.
func WithClipboardAccess() TestOption {
	return func(opts *ClientOptions) {
		if opts.Env == nil {
			opts.Env = make(map[string]string)
		}
		opts.Env["FLUFFY_MCP_ALLOW_CLIPBOARD"] = "1"
	}
}

// WithStrictLabels enforces strict label matching for TestClient.
func WithStrictLabels() TestOption {
	return func(opts *ClientOptions) {
		if opts.Env == nil {
			opts.Env = make(map[string]string)
		}
		opts.Env["FLUFFY_MCP_STRICT_LABELS"] = "1"
	}
}
