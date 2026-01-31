# FluffyUI Agent Server

The FluffyUI agent server provides AI-friendly, real-time interaction with terminal applications. It operates in **real-time mode by default**, streaming UI events to connected clients as they happen.

## Overview

The agent server enables:
- **Real-time UI streaming** - Live notifications of UI changes
- **Bidirectional communication** - WebSocket support for interactive sessions
- **Async operations** - Wait for UI conditions without polling
- **Background tasks** - Long-running operations with progress tracking
- **Session management** - Rate limiting and resource controls

## Quick Start

### Environment Variables

```bash
# Enable the agent server (required)
export FLUFFYUI_AGENT=unix:/tmp/fluffy-agent.sock

# Optional: Enable WebSocket endpoint
export FLUFFYUI_AGENT_WS=:8765

# Optional: Authentication
export FLUFFYUI_AGENT_TOKEN=my-secret-token

# Optional: Allow text capture in snapshots
export FLUFFYUI_AGENT_ALLOW_TEXT=1
```

### Basic Usage

```go
package main

import (
    "github.com/odvcencio/fluffyui/agent"
    "github.com/odvcencio/fluffyui/runtime"
)

func main() {
    app := runtime.NewApp(runtime.AppConfig{
        Backend: myBackend,
        Root:    myRootWidget,
    })

    // Enable agent server from environment (real-time mode by default)
    server, err := agent.EnableFromEnv(app)
    if err != nil {
        log.Fatal(err)
    }
    if server != nil {
        defer server.Stop()
    }

    app.Run(context.Background())
}
```

### Fluent Configuration

```go
server, err := agent.NewConfig().
    WithAddress("unix:/tmp/myapp.sock").
    WithWebSocketAddress(":8765").
    WithToken("secret").
    WithTextAccess().
    WithMaxSessions(50).
    Build(app)

if err != nil {
    log.Fatal(err)
}

server.Start()
defer server.Stop()
```

## Real-Time Features

### Event Streaming

The server automatically detects and streams UI changes:

```go
// Subscribe to real-time events
sub := server.Subscribe("session-id", agent.DefaultEventFilters())

for event := range sub.Events {
    switch event.Type {
    case agent.EventWidgetChanged:
        // Widget tree changed
    case agent.EventFocusChanged:
        // Focus moved
    case agent.EventValueChanged:
        // Widget value changed
    case agent.EventTextChanged:
        // Screen text changed
    }
}
```

### Async Wait Operations

Wait for UI conditions without polling:

```go
// Wait for a widget to appear
widget, err := server.WaitForWidget(ctx, "Submit Button", 5*time.Second)

// Wait for text to appear
err := server.WaitForText(ctx, "Loading complete", 5*time.Second)

// Wait for focus change
err := server.WaitForFocus(ctx, widgetID, 5*time.Second)

// Wait for specific value
err := server.WaitForValue(ctx, widgetID, "completed", 5*time.Second)

// Custom condition
snapshot, err := server.WaitForCondition(ctx, func(s agent.Snapshot) bool {
    return s.FocusedID == "submit-button"
}, timeout)
```

### Background Tasks

Submit long-running tasks:

```go
task, err := server.SubmitBackgroundTask(
    "data-processing",
    "Process large dataset",
    sessionID,
    func(ctx context.Context, task *agent.BackgroundTask) error {
        for i := 0; i < 100; i++ {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                // Do work
                task.SetProgress(i + 1)
            }
        }
        return nil
    },
)

// Monitor progress
progress := task.Progress()  // 0-100
status := task.Status()      // pending, running, completed, failed, cancelled
```

## Protocols

### JSONL Protocol (Default)

Connect via Unix socket or TCP:

```bash
# Unix socket
nc -U /tmp/fluffy-agent.sock

# TCP
telnet localhost 8716
```

**Request:**
```json
{"id": 1, "type": "hello", "token": "my-secret"}
```

**Response:**
```json
{"id": 1, "ok": true, "capabilities": {"allow_text": true}}
```

### WebSocket Protocol

Connect via WebSocket when `FLUFFYUI_AGENT_WS` is set:

```javascript
const ws = new WebSocket('ws://localhost:8765/agent');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'event') {
        // Handle real-time UI event
        console.log('UI changed:', message.data);
    }
};

// Send commands
ws.send(JSON.stringify({
    type: 'action',
    id: '123',
    payload: {action: 'focus', label: 'Submit'}
}));
```

## Event Types

| Event | Description |
|-------|-------------|
| `widget_changed` | Widget tree structure changed |
| `focus_changed` | Focus moved to different widget |
| `text_changed` | Screen text content changed |
| `value_changed` | Widget value changed |
| `state_changed` | Widget state changed (enabled/checked/etc) |
| `layout_changed` | Widget bounds/layout changed |
| `snapshot` | Full UI snapshot |
| `heartbeat` | Connection health check |

## Request Types

### Core Operations

| Type | Description |
|------|-------------|
| `hello` | Authenticate and get capabilities |
| `ping` | Health check |
| `close` | Close the connection |

### UI Interaction

| Type | Description |
|------|-------------|
| `snapshot` | Capture UI state |
| `key` | Send key press |
| `text` | Send text input |
| `mouse` | Send mouse event |
| `paste` | Paste text |
| `resize` | Resize terminal |

### Server Management

| Type | Description |
|------|-------------|
| `health` | Get health status |
| `stats` | Get server statistics |

### Background Tasks

| Type | Description |
|------|-------------|
| `task_status` | Get task status |
| `task_cancel` | Cancel a task |

## Event Filters

Control which events subscribers receive:

```go
// Default: widget, focus, value, state changes
filters := agent.DefaultEventFilters()

// All events
filters := agent.AllEventsFilter()

// Custom
filters := agent.EventFilters{
    WidgetChanges: true,
    FocusChanges:  true,
    TextChanges:   true,
    ValueChanges:  true,
    StateChanges:  true,
    LayoutChanges: false,
}
```

## Session Management

Sessions have configurable limits:

```go
limits := agent.SessionLimits{
    MaxPendingRequests: 50,
    MaxRequestsPerSec:  100,
    BurstLimit:         200,
    IdleTimeout:        30 * time.Minute,
    MaxRequestDuration: 5 * time.Minute,
}
```

### Session Modes

- **Normal** (`ModeNormal`): Standard session with default limits
- **Interactive** (`ModeInteractive`): Higher priority for user-facing sessions
- **Background** (`ModeBackground`): Lower priority, more restrictive limits

## Configuration Options

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `FLUFFYUI_AGENT` | Server address (unix:/path or tcp::port) | - |
| `FLUFFYUI_AGENT_WS` | WebSocket server address | - |
| `FLUFFYUI_AGENT_TOKEN` | Authentication token | - |
| `FLUFFYUI_AGENT_ALLOW_TEXT` | Allow text capture | false |
| `FLUFFYUI_AGENT_MAX_SESSIONS` | Max concurrent sessions | 100 |
| `FLUFFYUI_AGENT_RATE_LIMIT` | Global rate limit (req/sec) | 1000 |
| `FLUFFYUI_AGENT_DISABLE_HEALTH` | Disable health checks | false |

## Health Monitoring

```go
health := server.Health()
fmt.Printf("Healthy: %v\n", health.Healthy)
fmt.Printf("Active Sessions: %d\n", health.ActiveSessions)
fmt.Printf("Queue Size: %d\n", health.QueueSize)

stats := server.Stats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
fmt.Printf("Active Tasks: %d\n", stats.ActiveTasks)
```

## Migration from Legacy Mode

The agent server now operates in real-time mode **by default**. The legacy non-real-time mode has been removed.

If you were using the agent before:

1. Update imports - no changes needed
2. Update environment variables - `FLUFFYUI_AGENT` works the same
3. Your code will automatically get real-time capabilities

```go
// Before (still works, now with real-time)
server, err := agent.EnableEnhancedServerFromEnv(app)

// After (recommended)
server, err := agent.EnableFromEnv(app)
```

## Example: Full Integration

```go
package main

import (
    "context"
    "log"
    
    "github.com/odvcencio/fluffyui/agent"
    "github.com/odvcencio/fluffyui/backend/tcell"
    "github.com/odvcencio/fluffyui/runtime"
    "github.com/odvcencio/fluffyui/widgets"
)

func main() {
    // Create UI
    label := widgets.NewLabel("Hello, Agent!")
    button := widgets.NewButton("Click Me", widgets.WithOnClick(func() {
        label.SetText("Button clicked!")
    }))
    
    root := widgets.VBox(
        widgets.FlexFixed(label),
        widgets.FlexFixed(button),
    )
    
    // Create app
    be, _ := tcell.New()
    app := runtime.NewApp(runtime.AppConfig{
        Backend: be,
        Root:    root,
    })
    
    // Enable real-time agent server
    server, err := agent.EnableFromEnv(app)
    if err != nil {
        log.Fatal(err)
    }
    if server != nil {
        defer server.Stop()
        
        // Subscribe to events
        sub := server.Subscribe("demo", agent.DefaultEventFilters())
        go func() {
            for event := range sub.Events {
                log.Printf("UI Event: %s", event.Type)
            }
        }()
    }
    
    app.Run(context.Background())
}
```

## See Also

- [Agent Enhanced Documentation](agent-enhanced.md) - Detailed real-time features
- [Examples](../examples/agent-enhanced/) - Working example code
