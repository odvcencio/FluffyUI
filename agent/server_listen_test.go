package agent

import (
	"crypto/tls"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
)

func TestListenAgentAddrErrors(t *testing.T) {
	if _, _, err := listenAgentAddr("unix:"); err == nil {
		t.Fatalf("expected unix path error")
	}
	if _, _, err := listenAgentAddr("tcp:"); err == nil {
		t.Fatalf("expected tcp address error")
	}
	if _, _, err := listenAgentAddr("bad"); err == nil {
		t.Fatalf("expected unsupported address")
	}
}

func TestListenAgentAddrUnixAndTCP(t *testing.T) {
	path := t.TempDir() + "/agent.sock"
	ln, socketPath, err := listenAgentAddr("unix:" + path)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	if socketPath != path {
		t.Fatalf("unexpected unix socket path")
	}
	_ = ln.Close()

	ln, socketPath, err = listenAgentAddr("tcp:127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	if socketPath != "" {
		t.Fatalf("expected empty tcp socket path")
	}
	_ = ln.Close()
}

func TestParseKeyPress(t *testing.T) {
	if _, err := parseKeyPress("enter"); err != nil {
		t.Fatalf("expected key press, got %v", err)
	}
	if _, err := parseKeyPress("g g"); err == nil {
		t.Fatalf("expected error for sequence")
	}
	if _, err := parseKeyPress(""); err == nil {
		t.Fatalf("expected error for empty")
	}
}

func TestParseMouseButtonAndAction(t *testing.T) {
	btn, err := parseMouseButton("left")
	if err != nil || btn != runtime.MouseLeft {
		t.Fatalf("unexpected mouse button")
	}
	btn, err = parseMouseButton("wheel_up")
	if err != nil || btn != runtime.MouseWheelUp {
		t.Fatalf("unexpected wheel up")
	}
	if _, err := parseMouseButton("unknown"); err == nil {
		t.Fatalf("expected unknown button error")
	}

	act, err := parseMouseAction("release")
	if err != nil || act != runtime.MouseRelease {
		t.Fatalf("unexpected action")
	}
	if _, err := parseMouseAction("bad"); err == nil {
		t.Fatalf("expected action error")
	}
}

func TestServerHelperUtilities(t *testing.T) {
	encoded := encodeJSON(map[string]string{"hello": "world"})
	if !strings.Contains(encoded, "\"hello\"") {
		t.Fatalf("unexpected json: %s", encoded)
	}

	id1 := generateSessionID()
	id2 := generateSessionID()
	if id1 == "" || id2 == "" || id1 == id2 {
		t.Fatalf("unexpected ids: %q %q", id1, id2)
	}
	if !strings.Contains(id1, "-") {
		t.Fatalf("expected dash in session id: %q", id1)
	}
}

func TestResolveTLSConfig(t *testing.T) {
	if _, err := resolveTLSConfig(nil, "cert.pem", ""); err == nil {
		t.Fatalf("expected TLS config error for missing key")
	}
	cfg, err := resolveTLSConfig(&tls.Config{MinVersion: tls.VersionTLS13}, "", "")
	if err != nil {
		t.Fatalf("unexpected tls config error: %v", err)
	}
	if cfg == nil || cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected TLS config to preserve min version")
	}
}

func TestListenAgentAddrWithTLS(t *testing.T) {
	ln, socketPath, err := listenAgentAddrWithTLS("tcp:127.0.0.1:0", &tls.Config{MinVersion: tls.VersionTLS12})
	if err != nil {
		t.Fatalf("listen with tls: %v", err)
	}
	if socketPath != "" {
		t.Fatalf("expected empty socket path for tcp listener")
	}
	_ = ln.Close()
}
