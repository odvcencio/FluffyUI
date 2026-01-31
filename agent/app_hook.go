//go:build !js

package agent

import (
	"sync"
	"sync/atomic"

	"github.com/odvcencio/fluffyui/runtime"
)

// AppHook integrates the agent server with the app's message/render loop
type AppHook struct {
	app            *runtime.App
	server         *RealTimeServer
	renderObserver runtime.RenderObserver

	// State tracking
	lastRenderFrame atomic.Int64
	mu              sync.Mutex
	onRender        []func()
	onMessage       []func(runtime.Message)
}

// NewAppHook creates a new app hook
func NewAppHook(app *runtime.App, server *RealTimeServer) *AppHook {
	return &AppHook{
		app:    app,
		server: server,
	}
}

// Install installs the hook into the app
func (h *AppHook) Install() {
	if h == nil || h.app == nil {
		return
	}

	// Create render observer to detect render cycles
	h.renderObserver = runtime.RenderObserverFunc(func(stats runtime.RenderStats) {
		h.handleRender(stats)
	})

	// Note: The app doesn't have a SetRenderObserver method, so we'll
	// use a different approach - polling the app's state in the notifier
}

// Uninstall removes the hook
func (h *AppHook) Uninstall() {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.onRender = nil
	h.onMessage = nil
	h.mu.Unlock()
}

// OnRender registers a callback for render events
func (h *AppHook) OnRender(fn func()) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onRender = append(h.onRender, fn)
}

// OnMessage registers a callback for message events
func (h *AppHook) OnMessage(fn func(runtime.Message)) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onMessage = append(h.onMessage, fn)
}

func (h *AppHook) handleRender(stats runtime.RenderStats) {
	if h == nil {
		return
	}

	// Track render frame
	h.lastRenderFrame.Store(stats.Frame)

	// Notify callbacks
	h.mu.Lock()
	callbacks := make([]func(), len(h.onRender))
	copy(callbacks, h.onRender)
	h.mu.Unlock()

	for _, fn := range callbacks {
		fn()
	}

	// Trigger server notification
	if h.server != nil && h.server.notifier != nil {
		h.server.notifier.checkForChanges()
	}
}

// AppIntegration provides high-level integration between App and Agent
type AppIntegration struct {
	App    *runtime.App
	Server *RealTimeServer
	Hook   *AppHook
}

// NewAppIntegration creates and initializes app-agent integration
func NewAppIntegration(app *runtime.App, opts EnhancedServerOptions) (*AppIntegration, error) {
	server, err := NewRealTimeServer(opts)
	if err != nil {
		return nil, err
	}

	hook := NewAppHook(app, server)
	hook.Install()

	return &AppIntegration{
		App:    app,
		Server: server,
		Hook:   hook,
	}, nil
}

// Start starts the integration
func (i *AppIntegration) Start() error {
	if i == nil || i.Server == nil {
		return nil
	}
	return i.Server.Start()
}

// Stop stops the integration
func (i *AppIntegration) Stop() error {
	if i == nil {
		return nil
	}
	if i.Hook != nil {
		i.Hook.Uninstall()
	}
	if i.Server != nil {
		return i.Server.Stop()
	}
	return nil
}
