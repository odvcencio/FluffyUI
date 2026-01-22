package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func (s *Server) startTransport() error {
	switch s.opts.Transport {
	case "stdio":
		return s.startStdio()
	case "sse":
		return s.startSSE()
	case "unix":
		return s.startUnix()
	default:
		return errors.New("unsupported MCP transport")
	}
}

func (s *Server) startStdio() error {
	s.stdioServer = mcpserver.NewStdioServer(s.mcpServer)
	if s.opts.MaxPendingEvents > 0 {
		mcpserver.WithStdioNotificationQueueSize(s.opts.MaxPendingEvents)(s.stdioServer)
	}
	go func() {
		if err := s.stdioServer.Listen(s.ctx, os.Stdin, os.Stdout); err != nil && s.ctx.Err() == nil {
			log.Printf("mcp stdio server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) startSSE() error {
	s.sseServer = mcpserver.NewSSEServer(
		s.mcpServer,
		mcpserver.WithSSEContextFunc(s.sseContext),
		mcpserver.WithNotificationQueueSize(s.opts.MaxPendingEvents),
	)
	addr := s.opts.Addr
	go func() {
		if err := s.sseServer.Start(addr); err != nil && s.ctx.Err() == nil {
			log.Printf("mcp sse server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) startUnix() error {
	path := s.opts.Addr
	if strings.TrimSpace(path) == "" {
		return errors.New("unix transport requires socket path")
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	s.unixListener = ln
	s.unixPath = path
	go s.acceptUnix(ln)
	return nil
}

func (s *Server) acceptUnix(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.ctx.Err() != nil {
				return
			}
			continue
		}
		go s.serveUnixConn(s.ctx, conn)
	}
}

func (s *Server) serveUnixConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	session := newSocketSession(s.opts.MaxPendingEvents)
	session.setCloseFn(func() {
		_ = conn.Close()
	})
	if err := s.mcpServer.RegisterSession(ctx, session); err != nil {
		return
	}
	defer s.mcpServer.UnregisterSession(ctx, session.SessionID())
	sessionCtx := s.mcpServer.WithContext(ctx, session)

	writer := bufio.NewWriter(conn)
	var writeMu sync.Mutex
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case notification := <-session.notifications:
				_ = writeJSON(writer, &writeMu, notification)
			case <-done:
				return
			case <-sessionCtx.Done():
				return
			}
		}
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var raw json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			_ = writeJSON(writer, &writeMu, mcp.NewJSONRPCError(mcp.NewRequestId(nil), mcp.PARSE_ERROR, "Parse error", nil))
			continue
		}
		response := s.mcpServer.HandleMessage(sessionCtx, raw)
		if response != nil {
			_ = writeJSON(writer, &writeMu, response)
		}
	}
}

func writeJSON(writer *bufio.Writer, mu *sync.Mutex, payload any) error {
	if writer == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	mu.Lock()
	defer mu.Unlock()
	if _, err := writer.Write(data); err != nil {
		return err
	}
	return writer.Flush()
}

type socketSession struct {
	sessionID          string
	notifications      chan mcp.JSONRPCNotification
	initialized        atomic.Bool
	loggingLevel       atomic.Value
	tools              sync.Map
	resources          sync.Map
	resourceTemplates  sync.Map
	clientInfo         atomic.Value
	clientCapabilities atomic.Value
	closeOnce          sync.Once
	closeFn            func()
}

func newSocketSession(queueSize int) *socketSession {
	if queueSize <= 0 {
		queueSize = 100
	}
	return &socketSession{
		sessionID:     mcpserver.GenerateInProcessSessionID(),
		notifications: make(chan mcp.JSONRPCNotification, queueSize),
	}
}

func (s *socketSession) SessionID() string {
	return s.sessionID
}

func (s *socketSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notifications
}

func (s *socketSession) NotificationQueue() chan mcp.JSONRPCNotification {
	return s.notifications
}

func (s *socketSession) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		if s.closeFn != nil {
			s.closeFn()
		}
	})
	return nil
}

func (s *socketSession) setCloseFn(fn func()) {
	s.closeFn = fn
}

func (s *socketSession) Initialize() {
	s.loggingLevel.Store(mcp.LoggingLevelError)
	s.initialized.Store(true)
}

func (s *socketSession) Initialized() bool {
	return s.initialized.Load()
}

func (s *socketSession) SetLogLevel(level mcp.LoggingLevel) {
	s.loggingLevel.Store(level)
}

func (s *socketSession) GetLogLevel() mcp.LoggingLevel {
	level := s.loggingLevel.Load()
	if level == nil {
		return mcp.LoggingLevelError
	}
	return level.(mcp.LoggingLevel)
}

func (s *socketSession) GetSessionTools() map[string]mcpserver.ServerTool {
	out := make(map[string]mcpserver.ServerTool)
	s.tools.Range(func(key, value any) bool {
		if tool, ok := value.(mcpserver.ServerTool); ok {
			out[key.(string)] = tool
		}
		return true
	})
	return out
}

func (s *socketSession) SetSessionTools(tools map[string]mcpserver.ServerTool) {
	s.tools.Clear()
	for name, tool := range tools {
		s.tools.Store(name, tool)
	}
}

func (s *socketSession) GetSessionResources() map[string]mcpserver.ServerResource {
	out := make(map[string]mcpserver.ServerResource)
	s.resources.Range(func(key, value any) bool {
		if resource, ok := value.(mcpserver.ServerResource); ok {
			out[key.(string)] = resource
		}
		return true
	})
	return out
}

func (s *socketSession) SetSessionResources(resources map[string]mcpserver.ServerResource) {
	s.resources.Clear()
	for name, resource := range resources {
		s.resources.Store(name, resource)
	}
}

func (s *socketSession) GetSessionResourceTemplates() map[string]mcpserver.ServerResourceTemplate {
	out := make(map[string]mcpserver.ServerResourceTemplate)
	s.resourceTemplates.Range(func(key, value any) bool {
		if template, ok := value.(mcpserver.ServerResourceTemplate); ok {
			out[key.(string)] = template
		}
		return true
	})
	return out
}

func (s *socketSession) SetSessionResourceTemplates(templates map[string]mcpserver.ServerResourceTemplate) {
	s.resourceTemplates.Clear()
	for name, template := range templates {
		s.resourceTemplates.Store(name, template)
	}
}

func (s *socketSession) GetClientInfo() mcp.Implementation {
	if value := s.clientInfo.Load(); value != nil {
		if clientInfo, ok := value.(mcp.Implementation); ok {
			return clientInfo
		}
	}
	return mcp.Implementation{}
}

func (s *socketSession) SetClientInfo(clientInfo mcp.Implementation) {
	s.clientInfo.Store(clientInfo)
}

func (s *socketSession) GetClientCapabilities() mcp.ClientCapabilities {
	if value := s.clientCapabilities.Load(); value != nil {
		if caps, ok := value.(mcp.ClientCapabilities); ok {
			return caps
		}
	}
	return mcp.ClientCapabilities{}
}

func (s *socketSession) SetClientCapabilities(clientCapabilities mcp.ClientCapabilities) {
	s.clientCapabilities.Store(clientCapabilities)
}

var (
	_ mcpserver.ClientSession                = (*socketSession)(nil)
	_ mcpserver.SessionWithTools             = (*socketSession)(nil)
	_ mcpserver.SessionWithResources         = (*socketSession)(nil)
	_ mcpserver.SessionWithResourceTemplates = (*socketSession)(nil)
	_ mcpserver.SessionWithLogging           = (*socketSession)(nil)
	_ mcpserver.SessionWithClientInfo        = (*socketSession)(nil)
)
