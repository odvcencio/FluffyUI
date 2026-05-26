package fluffy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/backend/web"
)

func TestBuildBackendWebAddrFile(t *testing.T) {
	tmp := t.TempDir()
	addrFile := filepath.Join(tmp, "addr")
	t.Setenv("FLUFFYUI_BACKEND", "web")
	t.Setenv("FLUFFYUI_WEB_ADDR", ":0")
	t.Setenv("FLUFFYUI_WEB_ADDR_FILE", addrFile)

	b, err := buildBackendFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer b.Fini()

	if err := b.Init(); err != nil {
		t.Fatal(err)
	}

	// The addr file is written by the OnReady callback, which requires
	// running the full app. Test that buildBackendFromEnv returns a web
	// backend, and that we can manually write the addr file.
	wb, ok := b.(*web.Backend)
	if !ok {
		t.Fatal("expected *web.Backend")
	}
	addr := wb.Addr()
	if addr == nil {
		t.Fatal("expected non-nil addr")
	}
	if err := os.WriteFile(addrFile, []byte(addr.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(addrFile)
	if err != nil {
		t.Fatal("addr file not written:", err)
	}
	if !strings.Contains(string(data), ":") {
		t.Fatalf("expected host:port, got %q", data)
	}
}

func TestNewBuilderWebAddrFileOnReady(t *testing.T) {
	tmp := t.TempDir()
	addrFile := filepath.Join(tmp, "addr")
	t.Setenv("FLUFFYUI_BACKEND", "web")
	t.Setenv("FLUFFYUI_WEB_ADDR", ":0")
	t.Setenv("FLUFFYUI_WEB_ADDR_FILE", addrFile)

	builder, err := newBuilder()
	if err != nil {
		t.Fatal(err)
	}
	defer builder.cfg.Backend.Fini()

	if builder.cfg.OnReady == nil {
		t.Fatal("expected OnReady to be set when FLUFFYUI_WEB_ADDR_FILE is set")
	}

	// Init the backend so Addr() returns a real address
	if err := builder.cfg.Backend.Init(); err != nil {
		t.Fatal(err)
	}

	// Simulate what runtime.App.Run does: call OnReady
	builder.cfg.OnReady(nil)

	data, err := os.ReadFile(addrFile)
	if err != nil {
		t.Fatal("addr file not written:", err)
	}
	if !strings.Contains(string(data), ":") {
		t.Fatalf("expected host:port, got %q", data)
	}
}
