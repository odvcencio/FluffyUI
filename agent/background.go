//go:build !js

package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// BackgroundTask represents a long-running background task
type BackgroundTask struct {
	ID          string
	Name        string
	Description string
	SessionID   string

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	// State
	startedAt   time.Time
	completedAt atomic.Value // time.Time
	status      atomic.Value // TaskStatus
	err         atomic.Value // error
	progress    atomic.Int64 // 0-100

	// Execution
	fn BackgroundTaskFunc
}

// TaskStatus represents the status of a background task
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskPaused
	TaskCompleted
	TaskFailed
	TaskCancelled
)

func (s TaskStatus) String() string {
	switch s {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskPaused:
		return "paused"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// BackgroundTaskFunc is the function signature for background tasks
type BackgroundTaskFunc func(ctx context.Context, task *BackgroundTask) error

// NewBackgroundTask creates a new background task
func NewBackgroundTask(id, name, description string, sessionID string, fn BackgroundTaskFunc) *BackgroundTask {
	ctx, cancel := context.WithCancel(context.Background())

	t := &BackgroundTask{
		ID:          id,
		Name:        name,
		Description: description,
		SessionID:   sessionID,
		ctx:         ctx,
		cancel:      cancel,
		done:        make(chan struct{}),
		startedAt:   time.Now(),
		fn:          fn,
	}

	t.status.Store(TaskPending)
	return t
}

// Start begins task execution
func (t *BackgroundTask) Start() error {
	if t == nil {
		return errors.New("task is nil")
	}

	if !t.setStatus(TaskPending, TaskRunning) {
		return errors.New("task already started")
	}

	go t.run()
	return nil
}

// Cancel stops the task
func (t *BackgroundTask) Cancel() {
	if t == nil {
		return
	}
	t.cancel()
}

// Wait blocks until the task completes
func (t *BackgroundTask) Wait(ctx context.Context) error {
	if t == nil {
		return errors.New("task is nil")
	}
	select {
	case <-t.done:
		if err := t.Error(); err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WaitTimeout waits for the task with a timeout
func (t *BackgroundTask) WaitTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return t.Wait(ctx)
}

// Status returns the current task status
func (t *BackgroundTask) Status() TaskStatus {
	if t == nil {
		return TaskFailed
	}
	if s, ok := t.status.Load().(TaskStatus); ok {
		return s
	}
	return TaskFailed
}

// IsDone returns true if the task has completed, failed, or been cancelled
func (t *BackgroundTask) IsDone() bool {
	switch t.Status() {
	case TaskCompleted, TaskFailed, TaskCancelled:
		return true
	default:
		return false
	}
}

// Progress returns the current progress (0-100)
func (t *BackgroundTask) Progress() int {
	if t == nil {
		return 0
	}
	return int(t.progress.Load())
}

// SetProgress updates the progress (0-100)
func (t *BackgroundTask) SetProgress(p int) {
	if t == nil {
		return
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	t.progress.Store(int64(p))
}

// Error returns the error if the task failed
func (t *BackgroundTask) Error() error {
	if t == nil {
		return nil
	}
	if err, ok := t.err.Load().(error); ok {
		return err
	}
	return nil
}

// StartedAt returns when the task started
func (t *BackgroundTask) StartedAt() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.startedAt
}

// CompletedAt returns when the task completed (zero if not complete)
func (t *BackgroundTask) CompletedAt() time.Time {
	if t == nil {
		return time.Time{}
	}
	if ts, ok := t.completedAt.Load().(time.Time); ok {
		return ts
	}
	return time.Time{}
}

// Duration returns how long the task has been running
func (t *BackgroundTask) Duration() time.Duration {
	if t == nil {
		return 0
	}
	if t.IsDone() {
		return t.CompletedAt().Sub(t.startedAt)
	}
	return time.Since(t.startedAt)
}

// run executes the task function
func (t *BackgroundTask) run() {
	defer close(t.done)

	err := t.fn(t.ctx, t)

	if err != nil {
		t.err.Store(err)
		if errors.Is(err, context.Canceled) {
			t.status.Store(TaskCancelled)
		} else {
			t.status.Store(TaskFailed)
		}
	} else {
		t.status.Store(TaskCompleted)
	}

	t.completedAt.Store(time.Now())
	t.progress.Store(100)
}

// setStatus atomically updates the status
func (t *BackgroundTask) setStatus(from, to TaskStatus) bool {
	for {
		current := t.status.Load().(TaskStatus)
		if current != from {
			return false
		}
		if t.status.CompareAndSwap(current, to) {
			return true
		}
	}
}

// TaskStats returns task statistics
type TaskStats struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	SessionID   string        `json:"session_id,omitempty"`
	Status      string        `json:"status"`
	Progress    int           `json:"progress"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
}

// Stats returns task statistics
func (t *BackgroundTask) Stats() TaskStats {
	if t == nil {
		return TaskStats{}
	}

	stats := TaskStats{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		SessionID:   t.SessionID,
		Status:      t.Status().String(),
		Progress:    t.Progress(),
		StartedAt:   t.startedAt,
		Duration:    t.Duration(),
	}

	if t.IsDone() {
		completedAt := t.CompletedAt()
		stats.CompletedAt = &completedAt
	}

	if err := t.Error(); err != nil {
		stats.Error = err.Error()
	}

	return stats
}

// BackgroundTaskManager manages background tasks
type BackgroundTaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*BackgroundTask

	// Limits
	maxTasks        int
	maxTasksPerSession int

	// Callbacks
	onTaskStart func(t *BackgroundTask)
	onTaskDone  func(t *BackgroundTask)
}

// NewBackgroundTaskManager creates a new task manager
func NewBackgroundTaskManager(maxTasks, maxPerSession int) *BackgroundTaskManager {
	if maxTasks <= 0 {
		maxTasks = 50
	}
	if maxPerSession <= 0 {
		maxPerSession = 5
	}

	return &BackgroundTaskManager{
		tasks:              make(map[string]*BackgroundTask),
		maxTasks:           maxTasks,
		maxTasksPerSession: maxPerSession,
	}
}

// Submit creates and starts a new background task
func (m *BackgroundTaskManager) Submit(id, name, description, sessionID string, fn BackgroundTaskFunc) (*BackgroundTask, error) {
	if m == nil {
		return nil, errors.New("task manager is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check global limits
	if len(m.tasks) >= m.maxTasks {
		return nil, errors.New("max background tasks reached")
	}

	// Check per-session limits
	if sessionID != "" && m.maxTasksPerSession > 0 {
		sessionCount := 0
		for _, t := range m.tasks {
			if t.SessionID == sessionID {
				sessionCount++
			}
		}
		if sessionCount >= m.maxTasksPerSession {
			return nil, errors.New("max background tasks for session reached")
		}
	}

	// Check for duplicate ID
	if _, exists := m.tasks[id]; exists {
		return nil, errors.New("task with this ID already exists")
	}

	task := NewBackgroundTask(id, name, description, sessionID, fn)

	// Start the task
	if err := task.Start(); err != nil {
		return nil, err
	}

	m.tasks[id] = task

	// Clean up when done
	go func() {
		<-task.done
		m.mu.Lock()
		delete(m.tasks, id)
		m.mu.Unlock()

		if m.onTaskDone != nil {
			m.onTaskDone(task)
		}
	}()

	if m.onTaskStart != nil {
		m.onTaskStart(task)
	}

	return task, nil
}

// Get retrieves a task by ID
func (m *BackgroundTaskManager) Get(id string) *BackgroundTask {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[id]
}

// Cancel cancels a task by ID
func (m *BackgroundTaskManager) Cancel(id string) bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if task, ok := m.tasks[id]; ok {
		task.Cancel()
		return true
	}
	return false
}

// CancelSession cancels all tasks for a session
func (m *BackgroundTaskManager) CancelSession(sessionID string) int {
	if m == nil || sessionID == "" {
		return 0
	}

	m.mu.RLock()
	var toCancel []*BackgroundTask
	for _, t := range m.tasks {
		if t.SessionID == sessionID {
			toCancel = append(toCancel, t)
		}
	}
	m.mu.RUnlock()

	for _, t := range toCancel {
		t.Cancel()
	}

	return len(toCancel)
}

// List returns all active task IDs
func (m *BackgroundTaskManager) List() []string {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.tasks))
	for id := range m.tasks {
		ids = append(ids, id)
	}
	return ids
}

// ListSession returns all task IDs for a session
func (m *BackgroundTaskManager) ListSession(sessionID string) []string {
	if m == nil || sessionID == "" {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ids []string
	for id, t := range m.tasks {
		if t.SessionID == sessionID {
			ids = append(ids, id)
		}
	}
	return ids
}

// Stats returns statistics for all tasks
func (m *BackgroundTaskManager) Stats() []TaskStats {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]TaskStats, 0, len(m.tasks))
	for _, t := range m.tasks {
		stats = append(stats, t.Stats())
	}
	return stats
}

// Count returns the number of active tasks
func (m *BackgroundTaskManager) Count() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks)
}

// SetTaskStartCallback sets a callback for when tasks start
func (m *BackgroundTaskManager) SetTaskStartCallback(fn func(t *BackgroundTask)) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onTaskStart = fn
}

// SetTaskDoneCallback sets a callback for when tasks complete
func (m *BackgroundTaskManager) SetTaskDoneCallback(fn func(t *BackgroundTask)) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onTaskDone = fn
}

// BackgroundJob is a convenience type for simple background jobs
type BackgroundJob struct {
	manager *BackgroundTaskManager
	task    *BackgroundTask
}

// IsRunning returns true if the job is still running
func (j *BackgroundJob) IsRunning() bool {
	if j == nil || j.task == nil {
		return false
	}
	return !j.task.IsDone()
}

// Wait blocks until the job completes
func (j *BackgroundJob) Wait(ctx context.Context) error {
	if j == nil || j.task == nil {
		return errors.New("no job")
	}
	return j.task.Wait(ctx)
}

// Cancel stops the job
func (j *BackgroundJob) Cancel() {
	if j == nil || j.task == nil {
		return
	}
	j.task.Cancel()
}

// Progress returns the job progress
func (j *BackgroundJob) Progress() int {
	if j == nil || j.task == nil {
		return 0
	}
	return j.task.Progress()
}

// SubmitSimple submits a simple background job
func (m *BackgroundTaskManager) SubmitSimple(name string, fn func(ctx context.Context) error) (*BackgroundJob, error) {
	if m == nil {
		return nil, errors.New("task manager is nil")
	}

	id := generateTaskID()
	taskFn := func(ctx context.Context, task *BackgroundTask) error {
		return fn(ctx)
	}

	task, err := m.Submit(id, name, "", "", taskFn)
	if err != nil {
		return nil, err
	}

	return &BackgroundJob{manager: m, task: task}, nil
}

// generateTaskID generates a unique task ID
var taskIDCounter atomic.Int64

func generateTaskID() string {
	return time.Now().Format("20060102-150405") + "-" + string(rune(taskIDCounter.Add(1)))
}
