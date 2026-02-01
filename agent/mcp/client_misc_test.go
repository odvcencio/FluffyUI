package mcp

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClientConnectionHelpers(t *testing.T) {
	if _, err := Connect(); !errors.Is(err, ErrInvalidTransport) {
		t.Fatalf("expected invalid transport error")
	}
	if _, err := Connect("stdio"); err == nil {
		t.Fatalf("expected stdio command error")
	}
	if _, err := ConnectWithOptions(ClientOptions{}); !errors.Is(err, ErrInvalidTransport) {
		t.Fatalf("expected invalid transport for empty options")
	}
	if _, err := connectStdio(ClientOptions{}); err == nil {
		t.Fatalf("expected stdio error")
	}
	if _, err := connectUnix(ClientOptions{}); err == nil {
		t.Fatalf("expected unix error")
	}
	if _, err := connectSSE(ClientOptions{}); err == nil {
		t.Fatalf("expected sse error")
	}
}

func TestClientUtilityHelpers(t *testing.T) {
	transport := defaultClientTransport()
	if transport == "" {
		t.Fatalf("expected default transport")
	}

	url, err := normalizeSSEURL("sse://localhost:8080")
	if err != nil || !strings.HasSuffix(url, "/sse") {
		t.Fatalf("unexpected sse url: %v", err)
	}
	if _, err := normalizeSSEURL(":bad"); err == nil {
		t.Fatalf("expected parse error")
	}

	host, err := sseHostPort("http://localhost:9090/sse")
	if err != nil || host != "localhost:9090" {
		t.Fatalf("unexpected host: %v", err)
	}
	if _, err := sseHostPort("http://"); err == nil {
		t.Fatalf("expected invalid host error")
	}

	path, err := tempSocketPath()
	if err != nil {
		t.Fatalf("temp socket path error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected temp socket path removed")
	}

	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen unix error: %v", err)
	}
	defer ln.Close()

	conn, err := waitForUnix(path, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("wait for unix error: %v", err)
	}
	_ = conn.Close()

	if got := effectiveTimeout(0); got != defaultConnectTimeout {
		t.Fatalf("expected default timeout")
	}
	if got := effectiveTimeout(10 * time.Millisecond); got != 10*time.Millisecond {
		t.Fatalf("unexpected timeout")
	}

	env := buildEnv(map[string]string{"A": "1", "B": "2"}, map[string]string{"B": "3", "C": "4"})
	if env["A"] != "1" || env["B"] != "3" || env["C"] != "4" {
		t.Fatalf("unexpected env merge")
	}
	slice := envSlice(env)
	if len(slice) != 3 || slice[0] != "A=1" || slice[1] != "B=3" || slice[2] != "C=4" {
		t.Fatalf("unexpected env slice ordering: %#v", slice)
	}
}

func TestClientContextsAndPayloads(t *testing.T) {
	client := &Client{timeout: 10 * time.Millisecond}
	ctx, cancel := client.callContext()
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatalf("expected deadline")
	}

	ctx, cancel = (&Client{}).callContext()
	cancel()
	if ctx == nil {
		t.Fatalf("expected background context")
	}

	if timeoutMs(0) != 0 {
		t.Fatalf("expected zero timeout ms")
	}
	if timeoutMs(150*time.Millisecond) != 150 {
		t.Fatalf("unexpected timeout ms")
	}

	labelArgs := waitLabelPayload("label", 0)
	if labelArgs["label"] != "label" {
		t.Fatalf("unexpected label payload")
	}
	if _, ok := labelArgs["timeout_ms"]; ok {
		t.Fatalf("unexpected timeout ms entry")
	}

	textArgs := waitTextPayload("text", time.Millisecond)
	if textArgs["text"] != "text" || textArgs["timeout_ms"].(int) == 0 {
		t.Fatalf("unexpected text payload")
	}
}

func TestClientOptionsAndSubscriptions(t *testing.T) {
	opts := ClientOptions{}
	WithTextAccess()(&opts)
	WithClipboardAccess()(&opts)
	WithStrictLabels()(&opts)
	if opts.Env["FLUFFY_MCP_ALLOW_TEXT"] != "1" || opts.Env["FLUFFY_MCP_ALLOW_CLIPBOARD"] != "1" || opts.Env["FLUFFY_MCP_STRICT_LABELS"] != "1" {
		t.Fatalf("expected env options to be set")
	}

	client, _ := newMockClient()
	if err := client.Subscribe("fluffy://widgets", func(ResourceEvent) {}); err != nil {
		t.Fatalf("subscribe error: %v", err)
	}
	if err := client.Resubscribe("fluffy://widgets", "fluffy://widgets/new"); err != nil {
		t.Fatalf("resubscribe error: %v", err)
	}
	if err := client.Unsubscribe("fluffy://widgets/new"); err != nil {
		t.Fatalf("unsubscribe error: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
}
