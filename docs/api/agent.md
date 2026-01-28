# Agent API

The agent package provides semantic access to a FluffyUI app for automation,
AI workflows, and test harnesses.

## Agent Basics

```go
agt := agent.New(agent.Config{App: app})
if err := agt.WaitForText("Ready", 2*time.Second); err != nil {
    // handle timeout
}

snap := agt.Snapshot()
```

Common actions include focus, typing, clicks, and snapshot capture.

## JSONL Agent Server

Expose the agent over a local socket:

```go
srv, _ := agent.NewServer(agent.ServerOptions{
    Addr:      "tcp:127.0.0.1:7777",
    App:       app,
    AllowText: true,
})

go srv.Serve(context.Background())
```

Supported request types include `hello`, `ping`, `snapshot`, `text`, `key`,
`mouse`, `paste`, and `resize`.

## WebSocket Server

The WebSocket server mirrors the JSONL API over HTTP:

```go
wsServer, _ := agent.NewWebSocketServer(agent.ServerOptions{App: app})
http.ListenAndServe(":7778", wsServer)
```

## MCP Integration

The `agent/mcp` package exposes a Model Context Protocol server and Go client
for tool-driven observation and control. See `agent/mcp` for details.
