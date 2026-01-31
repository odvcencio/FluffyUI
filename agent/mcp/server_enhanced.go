package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/runtime"
)

// EnhancedServer extends the MCP server with advanced session management,
// request queuing, and background task support.
type EnhancedServer struct {
	*Server // Embed original server for compatibility

	// Enhanced components
	sessionPool *agent.SessionPool
	queue       *agent.RequestQueue
	taskManager *agent.BackgroundTaskManager

	// Configuration
	enhancedOpts EnhancedOptions

	// State
	ctx       context.Context
	cancel    context.CancelFunc
	running   atomic.Bool
	closeOnce sync.Once
	wg        sync.WaitGroup

	// Health monitoring
	healthMu     sync.RWMutex
	healthStatus agent.HealthStatus
	lastHealth   time.Time

	// Metrics
	totalRequests  atomic.Int64
	failedRequests atomic.Int64
	avgLatency     atomic.Int64 // microseconds
}

// EnhancedOptions configures the enhanced MCP server
type EnhancedOptions struct {
	// Session pool
	PoolLimits agent.PoolLimits

	// Request queue
	QueueOptions agent.QueueOptions

	// Background tasks
	MaxBackgroundTasks int
	MaxTasksPerSession int

	// Health
	EnableHealthCheck bool
	HealthInterval    time.Duration

	// Request handling
	RequestTimeout   time.Duration
	EnableAsyncTools bool // Enable tools that support async execution
}

// DefaultEnhancedOptions returns reasonable defaults
func DefaultEnhancedOptions() EnhancedOptions {
	return EnhancedOptions{
		PoolLimits:         agent.DefaultPoolLimits(),
		QueueOptions:       agent.DefaultQueueOptions(),
		MaxBackgroundTasks: 50,
		MaxTasksPerSession: 5,
		EnableHealthCheck:  true,
		HealthInterval:     30 * time.Second,
		RequestTimeout:     30 * time.Second,
		EnableAsyncTools:   true,
	}
}

// NewEnhancedServer creates a new enhanced MCP server
func NewEnhancedServer(app *runtime.App, mcpOpts runtime.MCPOptions, enhancedOpts EnhancedOptions) (*EnhancedServer, error) {
	if app == nil {
		return nil, errors.New("mcp server requires app")
	}

	// Create base server
	baseServer, err := NewServer(app, mcpOpts)
	if err != nil {
		return nil, err
	}

	// Normalize options
	if enhancedOpts.HealthInterval <= 0 {
		enhancedOpts.HealthInterval = 30 * time.Second
	}
	if enhancedOpts.RequestTimeout <= 0 {
		enhancedOpts.RequestTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	es := &EnhancedServer{
		Server:       baseServer,
		enhancedOpts: enhancedOpts,
		sessionPool:  agent.NewSessionPool(enhancedOpts.PoolLimits),
		queue:        agent.NewRequestQueue(enhancedOpts.QueueOptions),
		taskManager:  agent.NewBackgroundTaskManager(enhancedOpts.MaxBackgroundTasks, enhancedOpts.MaxTasksPerSession),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Set up callbacks
	es.queue.SetQueueFullCallback(es.onQueueFull)
	es.queue.SetRequestStartCallback(es.onRequestStart)
	es.queue.SetRequestDoneCallback(es.onRequestDone)

	return es, nil
}

// Start begins the enhanced MCP server
func (es *EnhancedServer) Start() error {
	if es == nil {
		return errors.New("server is nil")
	}

	if !es.running.CompareAndSwap(false, true) {
		return errors.New("server already running")
	}

	// Start base server
	if err := es.Server.Start(); err != nil {
		es.running.Store(false)
		return err
	}

	// Start session pool
	es.sessionPool.Start()

	// Start health checks
	if es.enhancedOpts.EnableHealthCheck {
		es.wg.Add(1)
		go es.healthCheckLoop()
	}

	// Start metrics collection
	es.wg.Add(1)
	go es.metricsLoop()

	return nil
}

// Close gracefully shuts down the enhanced server
func (es *EnhancedServer) Close() error {
	if es == nil {
		return nil
	}

	es.closeOnce.Do(func() {
		es.running.Store(false)
		es.cancel()

		// Stop queue
		es.queue.Stop()

		// Stop session pool
		es.sessionPool.Stop()

		// Close base server
		_ = es.Server.Close()

		// Wait for all goroutines
		es.wg.Wait()
	})

	return nil
}

// Health returns current health status
func (es *EnhancedServer) Health() agent.HealthStatus {
	if es == nil {
		return agent.HealthStatus{Healthy: false, Message: "server is nil"}
	}
	es.healthMu.RLock()
	defer es.healthMu.RUnlock()
	return es.healthStatus
}

// Stats returns comprehensive server statistics
func (es *EnhancedServer) Stats() EnhancedStats {
	if es == nil {
		return EnhancedStats{}
	}

	poolStats := es.sessionPool.Stats()
	queueStats := es.queue.Stats()

	return EnhancedStats{
		Running:        es.running.Load(),
		SessionStats:   poolStats,
		QueueStats:     queueStats,
		ActiveTasks:    es.taskManager.Count(),
		TotalRequests:  es.totalRequests.Load(),
		FailedRequests: es.failedRequests.Load(),
		Health:         es.Health(),
	}
}

// EnhancedStats contains comprehensive server statistics
type EnhancedStats struct {
	Running        bool               `json:"running"`
	SessionStats   agent.PoolStats    `json:"sessions"`
	QueueStats     agent.QueueStats   `json:"queue"`
	ActiveTasks    int                `json:"active_tasks"`
	TotalRequests  int64              `json:"total_requests"`
	FailedRequests int64              `json:"failed_requests"`
	Health         agent.HealthStatus `json:"health"`
}

// SubmitBackgroundTask submits a background task via the MCP server
func (es *EnhancedServer) SubmitBackgroundTask(name, description, sessionID string, fn agent.BackgroundTaskFunc) (*agent.BackgroundTask, error) {
	if es == nil {
		return nil, errors.New("server is nil")
	}
	id := generateTaskID()
	return es.taskManager.Submit(id, name, description, sessionID, fn)
}

// AsyncExecute submits a function for async execution
func (es *EnhancedServer) AsyncExecute(fn func(ctx context.Context) (any, error), priority agent.RequestPriority) *agent.AsyncResult {
	if es == nil {
		return nil
	}
	return es.queue.AsyncExecute(fn, priority)
}

// GetSession retrieves a session from the pool
func (es *EnhancedServer) GetSession(id string) *agent.Session {
	if es == nil {
		return nil
	}
	return es.sessionPool.GetSession(id)
}

// CreateSession creates a new session in the pool
func (es *EnhancedServer) CreateSession(id string, mode agent.SessionMode, limits agent.SessionLimits) (*agent.Session, error) {
	if es == nil {
		return nil, errors.New("server is nil")
	}
	return es.sessionPool.CreateSession(id, mode, limits)
}

// RemoveSession removes a session from the pool
func (es *EnhancedServer) RemoveSession(id string) {
	if es == nil {
		return
	}
	es.sessionPool.RemoveSession(id)
}

// CancelSessionTasks cancels all background tasks for a session
func (es *EnhancedServer) CancelSessionTasks(sessionID string) int {
	if es == nil {
		return 0
	}
	return es.taskManager.CancelSession(sessionID)
}

// processRequest processes a request through the queue with proper session management
func (es *EnhancedServer) processRequest(sessionID string, fn func(ctx context.Context) error) error {
	if es == nil {
		return errors.New("server is nil")
	}

	session := es.sessionPool.GetSession(sessionID)
	if session == nil {
		return agent.ErrSessionNotFound
	}

	if err := session.StartRequest(); err != nil {
		return err
	}
	defer session.EndRequest(true)

	// Check global rate limit
	if err := es.sessionPool.CheckGlobalRate(); err != nil {
		return err
	}

	// Create request
	req := &agent.Request{
		SessionID: sessionID,
		Priority:  agent.RequestPriorityNormal,
		Execute:   fn,
	}

	// Determine priority based on session mode
	switch session.Mode() {
	case agent.ModeBackground:
		req.Priority = agent.RequestPriorityLow
	case agent.ModeInteractive:
		req.Priority = agent.RequestPriorityHigh
	}

	// Enqueue and wait
	if err := es.queue.Enqueue(req); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(es.ctx, es.enhancedOpts.RequestTimeout)
	defer cancel()

	_, err := req.Wait(ctx)
	return err
}

// healthCheckLoop performs periodic health checks
func (es *EnhancedServer) healthCheckLoop() {
	defer es.wg.Done()

	ticker := time.NewTicker(es.enhancedOpts.HealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-es.ctx.Done():
			return
		case <-ticker.C:
			es.updateHealth()
		}
	}
}

// updateHealth updates the health status
func (es *EnhancedServer) updateHealth() {
	stats := es.Stats()

	healthy := true
	message := "healthy"

	// Check thresholds
	if stats.QueueStats.CriticalSize > 100 {
		healthy = false
		message = "high critical queue backlog"
	} else if stats.SessionStats.TotalPendingRequests > 200 {
		healthy = false
		message = "high pending request count"
	}

	es.healthMu.Lock()
	es.healthStatus = agent.HealthStatus{
		Healthy:        healthy,
		Message:        message,
		ActiveSessions: stats.SessionStats.TotalSessions,
		QueueSize:      stats.QueueStats.CriticalSize + stats.QueueStats.HighSize + stats.QueueStats.NormalSize + stats.QueueStats.LowSize + stats.QueueStats.BackgroundSize,
		ActiveTasks:    stats.ActiveTasks,
		Timestamp:      time.Now(),
	}
	es.lastHealth = time.Now()
	es.healthMu.Unlock()
}

// metricsLoop periodically updates metrics
func (es *EnhancedServer) metricsLoop() {
	defer es.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-es.ctx.Done():
			return
		case <-ticker.C:
			es.updateMetrics()
		}
	}
}

// updateMetrics updates server metrics
func (es *EnhancedServer) updateMetrics() {
	// Could integrate with monitoring system here
}

// Callbacks
func (es *EnhancedServer) onQueueFull(req *agent.Request) {
	// Log queue full condition
}

func (es *EnhancedServer) onRequestStart(req *agent.Request) {
	es.totalRequests.Add(1)
}

func (es *EnhancedServer) onRequestDone(req *agent.Request, duration time.Duration, err error) {
	if err != nil {
		es.failedRequests.Add(1)
	}

	// Update average latency (simple exponential moving average)
	latencyMicros := duration.Microseconds()
	oldAvg := es.avgLatency.Load()
	if oldAvg == 0 {
		es.avgLatency.Store(latencyMicros)
	} else {
		// EMA with alpha = 0.1
		newAvg := (oldAvg*9 + latencyMicros) / 10
		es.avgLatency.Store(newAvg)
	}
}

// AverageLatency returns the average request latency in microseconds
func (es *EnhancedServer) AverageLatency() int64 {
	if es == nil {
		return 0
	}
	return es.avgLatency.Load()
}

// MCPRequestHandler wraps a handler function with session and queue management
func (es *EnhancedServer) MCPRequestHandler(sessionID string, handler func(ctx context.Context) (*mcp.CallToolResult, error)) (*mcp.CallToolResult, error) {
	if es == nil {
		return nil, errors.New("server is nil")
	}

	var result *mcp.CallToolResult
	var handlerErr error

	execFn := func(ctx context.Context) error {
		result, handlerErr = handler(ctx)
		return handlerErr
	}

	if err := es.processRequest(sessionID, execFn); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, newMCPError(-32005, "request timeout", map[string]any{
				"timeout_ms": es.enhancedOpts.RequestTimeout.Milliseconds(),
			})
		}
		var rateErr *agent.RateLimitError
		if errors.As(err, &rateErr) {
			return nil, newMCPError(rateErrorCode, "rate limit exceeded", map[string]any{
				"retry_after_ms": rateErr.RetryAfter.Milliseconds(),
			})
		}
		return nil, newMCPError(-32006, err.Error(), nil)
	}

	return result, handlerErr
}

// RegisterEnhancedTools registers tools that use the enhanced server's capabilities
func (es *EnhancedServer) RegisterEnhancedTools() {
	if es == nil || es.mcpServer == nil {
		return
	}

	// Register health check tool
	es.mcpServer.AddTool(
		mcp.NewTool("server_health",
			mcp.WithDescription("Get server health and statistics"),
		),
		es.handleHealth,
	)

	// Register background task tool
	if es.enhancedOpts.EnableAsyncTools {
		es.mcpServer.AddTool(
			mcp.NewTool("submit_background_task",
				mcp.WithDescription("Submit a background task"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Task name")),
				mcp.WithString("description", mcp.Description("Task description")),
			),
			es.handleSubmitBackgroundTask,
		)

		es.mcpServer.AddTool(
			mcp.NewTool("get_task_status",
				mcp.WithDescription("Get background task status"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
			),
			es.handleGetTaskStatus,
		)

		es.mcpServer.AddTool(
			mcp.NewTool("cancel_task",
				mcp.WithDescription("Cancel a background task"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
			),
			es.handleCancelTask,
		)
	}
}

func (es *EnhancedServer) handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	health := es.Health()
	stats := es.Stats()

	data := map[string]any{
		"health": health,
		"stats": map[string]any{
			"active_sessions": stats.SessionStats.TotalSessions,
			"queue_size":      stats.QueueStats.CriticalSize + stats.QueueStats.HighSize + stats.QueueStats.NormalSize + stats.QueueStats.LowSize,
			"active_tasks":    stats.ActiveTasks,
			"total_requests":  stats.TotalRequests,
			"failed_requests": stats.FailedRequests,
			"avg_latency_us":  es.AverageLatency(),
		},
	}

	return mcp.NewToolResultJSON(data)
}

func (es *EnhancedServer) handleSubmitBackgroundTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !es.enhancedOpts.EnableAsyncTools {
		return nil, newMCPError(-32007, "async tools disabled", nil)
	}

	args, _ := req.Params.Arguments.(map[string]any)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)

	if name == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "name is required", nil)
	}

	taskID := generateTaskID()

	// For demo purposes, create a simple task that waits
	taskFn := func(ctx context.Context, task *agent.BackgroundTask) error {
		// Simulate work
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				task.SetProgress(i + 1)
			}
		}
		return nil
	}

	task, err := es.taskManager.Submit(taskID, name, description, "", taskFn)
	if err != nil {
		return nil, newMCPError(-32008, err.Error(), nil)
	}

	return mcp.NewToolResultJSON(map[string]any{
		"task_id": task.ID,
		"status":  task.Status().String(),
	})
}

func (es *EnhancedServer) handleGetTaskStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "task_id is required", nil)
	}

	task := es.taskManager.Get(taskID)
	if task == nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, "task not found", nil)
	}

	stats := task.Stats()
	return mcp.NewToolResultJSON(stats)
}

func (es *EnhancedServer) handleCancelTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "task_id is required", nil)
	}

	if !es.taskManager.Cancel(taskID) {
		return nil, newMCPError(mcp.INVALID_PARAMS, "task not found", nil)
	}

	return mcp.NewToolResultJSON(map[string]any{
		"cancelled": true,
	})
}

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// ensure EnhancedServer implements the closer interface
var _ interface{ Close() error } = (*EnhancedServer)(nil)
