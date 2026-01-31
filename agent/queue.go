//go:build !js

package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Request errors
var (
	ErrQueueFull     = errors.New("request queue is full")
	ErrQueueClosed   = errors.New("request queue is closed")
	ErrRequestTimeout = errors.New("request timeout")
)

// RequestPriority defines the priority of a request
type RequestPriority int

const (
	RequestPriorityLow RequestPriority = iota
	RequestPriorityNormal
	RequestPriorityHigh
	RequestPriorityCritical
)

// Request represents a queued operation
type Request struct {
	ID        string
	SessionID string
	Priority  RequestPriority
	CreatedAt time.Time
	Deadline  time.Time // Optional deadline

	// Execution
	Execute   func(ctx context.Context) error
	OnSuccess func(result any)
	OnError   func(err error)

	// Internal
	enqueuedAt  time.Time
	startedAt   atomic.Value // time.Time
	completedAt atomic.Value // time.Time
	result      any
	err         error
	done        chan struct{}
}

// IsExpired returns true if the request has exceeded its deadline
func (r *Request) IsExpired(now time.Time) bool {
	if r == nil {
		return true
	}
	if !r.Deadline.IsZero() && now.After(r.Deadline) {
		return true
	}
	return false
}

// Wait blocks until the request is completed or the context is cancelled
func (r *Request) Wait(ctx context.Context) (any, error) {
	if r == nil {
		return nil, errors.New("nil request")
	}
	select {
	case <-r.done:
		return r.result, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RequestQueue manages prioritized request queuing
type RequestQueue struct {
	mu sync.RWMutex

	// Four priority queues
	critical []*Request
	high     []*Request
	normal   []*Request
	low      []*Request

	// Background queue (lowest priority, processed when idle)
	background []*Request

	// Limits
	maxSize        int
	maxPerPriority int

	// Processing
	workers     int
	active      atomic.Int64
	totalQueued atomic.Int64
	totalDone   atomic.Int64

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Synchronization
	notEmpty chan struct{}
	closed   atomic.Bool

	// Callbacks
	onQueueFull    func(req *Request)
	onRequestStart func(req *Request)
	onRequestDone  func(req *Request, duration time.Duration, err error)
}

// QueueOptions configures the request queue
type QueueOptions struct {
	MaxSize        int // Total max requests across all priorities
	MaxPerPriority int // Max per individual priority queue
	Workers        int // Number of concurrent workers
}

// DefaultQueueOptions returns reasonable defaults
func DefaultQueueOptions() QueueOptions {
	return QueueOptions{
		MaxSize:        1000,
		MaxPerPriority: 250,
		Workers:        4,
	}
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(opts QueueOptions) *RequestQueue {
	if opts.MaxSize <= 0 {
		opts.MaxSize = 1000
	}
	if opts.MaxPerPriority <= 0 {
		opts.MaxPerPriority = opts.MaxSize / 4
	}
	if opts.Workers <= 0 {
		opts.Workers = 4
	}

	ctx, cancel := context.WithCancel(context.Background())

	q := &RequestQueue{
		maxSize:        opts.MaxSize,
		maxPerPriority: opts.MaxPerPriority,
		workers:        opts.Workers,
		ctx:            ctx,
		cancel:         cancel,
		notEmpty:       make(chan struct{}, 1),
	}

	// Start workers
	for i := 0; i < opts.Workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	return q
}

// Stop shuts down the queue and waits for all workers
func (q *RequestQueue) Stop() {
	if q == nil {
		return
	}
	q.closed.Store(true)
	q.cancel()
	close(q.notEmpty)
	q.wg.Wait()

	// Complete any remaining requests with error
	q.mu.Lock()
	allReqs := q.drainAll()
	q.mu.Unlock()

	for _, req := range allReqs {
		req.err = ErrQueueClosed
		close(req.done)
	}
}

// Enqueue adds a request to the queue
func (q *RequestQueue) Enqueue(req *Request) error {
	if q == nil {
		return errors.New("queue is nil")
	}
	if q.closed.Load() {
		return ErrQueueClosed
	}
	if req == nil {
		return errors.New("request is nil")
	}

	req.enqueuedAt = time.Now()
	req.done = make(chan struct{})

	q.mu.Lock()
	defer q.mu.Unlock()

	// Check total size
	total := len(q.critical) + len(q.high) + len(q.normal) + len(q.low) + len(q.background)
	if total >= q.maxSize {
		if q.onQueueFull != nil {
			q.onQueueFull(req)
		}
		return ErrQueueFull
	}

	// Add to appropriate queue based on priority
	switch req.Priority {
	case RequestPriorityCritical:
		if len(q.critical) >= q.maxPerPriority {
			return ErrQueueFull
		}
		q.critical = append(q.critical, req)
	case RequestPriorityHigh:
		if len(q.high) >= q.maxPerPriority {
			return ErrQueueFull
		}
		q.high = append(q.high, req)
	case RequestPriorityLow:
		if len(q.low) >= q.maxPerPriority {
			return ErrQueueFull
		}
		q.low = append(q.low, req)
	default:
		if len(q.normal) >= q.maxPerPriority {
			return ErrQueueFull
		}
		q.normal = append(q.normal, req)
	}

	q.totalQueued.Add(1)

	// Signal workers
	select {
	case q.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// EnqueueBackground adds a request to the background queue
func (q *RequestQueue) EnqueueBackground(req *Request) error {
	if q == nil {
		return errors.New("queue is nil")
	}
	if q.closed.Load() {
		return ErrQueueClosed
	}
	if req == nil {
		return errors.New("request is nil")
	}

	req.enqueuedAt = time.Now()
	req.done = make(chan struct{})

	q.mu.Lock()
	defer q.mu.Unlock()

	// Check size
	if len(q.background) >= q.maxPerPriority {
		return ErrQueueFull
	}

	q.background = append(q.background, req)
	q.totalQueued.Add(1)

	// Signal workers
	select {
	case q.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// TryEnqueue attempts to enqueue without blocking, returns immediately
func (q *RequestQueue) TryEnqueue(req *Request) bool {
	err := q.Enqueue(req)
	return err == nil
}

// Size returns the current queue size
func (q *RequestQueue) Size() int {
	if q == nil {
		return 0
	}
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.critical) + len(q.high) + len(q.normal) + len(q.low) + len(q.background)
}

// ActiveCount returns the number of active (processing) requests
func (q *RequestQueue) ActiveCount() int {
	if q == nil {
		return 0
	}
	return int(q.active.Load())
}

// Stats returns queue statistics
func (q *RequestQueue) Stats() QueueStats {
	if q == nil {
		return QueueStats{}
	}
	q.mu.RLock()
	defer q.mu.RUnlock()

	return QueueStats{
		CriticalSize:   len(q.critical),
		HighSize:       len(q.high),
		NormalSize:     len(q.normal),
		LowSize:        len(q.low),
		BackgroundSize: len(q.background),
		Active:         int(q.active.Load()),
		TotalQueued:    int(q.totalQueued.Load()),
		TotalDone:      int(q.totalDone.Load()),
	}
}

// SetQueueFullCallback sets a callback for when the queue is full
func (q *RequestQueue) SetQueueFullCallback(fn func(req *Request)) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onQueueFull = fn
}

// SetRequestStartCallback sets a callback for when a request starts
func (q *RequestQueue) SetRequestStartCallback(fn func(req *Request)) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onRequestStart = fn
}

// SetRequestDoneCallback sets a callback for when a request completes
func (q *RequestQueue) SetRequestDoneCallback(fn func(req *Request, duration time.Duration, err error)) {
	if q == nil {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onRequestDone = fn
}

// worker processes requests from the queue
func (q *RequestQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case _, ok := <-q.notEmpty:
			if !ok {
				return
			}
			q.processOne()
		}
	}
}

// processOne processes a single request from the queues
func (q *RequestQueue) processOne() {
	req := q.dequeue()
	if req == nil {
		return
	}

	// Check if expired
	if req.IsExpired(time.Now()) {
		req.err = ErrRequestTimeout
		close(req.done)
		return
	}

	q.active.Add(1)
	req.startedAt.Store(time.Now())

	if q.onRequestStart != nil {
		q.onRequestStart(req)
	}

	// Create context with deadline if specified
	ctx := q.ctx
	if !req.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, req.Deadline)
		defer cancel()
	}

	// Execute
	start := time.Now()
	if req.Execute != nil {
		req.err = req.Execute(ctx)
	}
	duration := time.Since(start)

	// Complete
	req.completedAt.Store(time.Now())
	q.active.Add(-1)
	q.totalDone.Add(1)

	// Callbacks
	if req.err != nil {
		if req.OnError != nil {
			req.OnError(req.err)
		}
	} else {
		if req.OnSuccess != nil {
			req.OnSuccess(req.result)
		}
	}

	if q.onRequestDone != nil {
		q.onRequestDone(req, duration, req.err)
	}

	close(req.done)
}

// dequeue retrieves the highest priority request
func (q *RequestQueue) dequeue() *Request {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Priority order: critical > high > normal > low > background
	if len(q.critical) > 0 {
		req := q.critical[0]
		q.critical = q.critical[1:]
		return req
	}
	if len(q.high) > 0 {
		req := q.high[0]
		q.high = q.high[1:]
		return req
	}
	if len(q.normal) > 0 {
		req := q.normal[0]
		q.normal = q.normal[1:]
		return req
	}
	if len(q.low) > 0 {
		req := q.low[0]
		q.low = q.low[1:]
		return req
	}
	if len(q.background) > 0 {
		req := q.background[0]
		q.background = q.background[1:]
		return req
	}

	return nil
}

// drainAll removes all pending requests (used during shutdown)
func (q *RequestQueue) drainAll() []*Request {
	var all []*Request
	all = append(all, q.critical...)
	all = append(all, q.high...)
	all = append(all, q.normal...)
	all = append(all, q.low...)
	all = append(all, q.background...)

	q.critical = nil
	q.high = nil
	q.normal = nil
	q.low = nil
	q.background = nil

	return all
}

// QueueStats contains queue statistics
type QueueStats struct {
	CriticalSize   int `json:"critical_size"`
	HighSize       int `json:"high_size"`
	NormalSize     int `json:"normal_size"`
	LowSize        int `json:"low_size"`
	BackgroundSize int `json:"background_size"`
	Active         int `json:"active"`
	TotalQueued    int `json:"total_queued"`
	TotalDone      int `json:"total_done"`
}

// AsyncResult represents a future result from an async operation
type AsyncResult struct {
	request *Request
}

// Wait blocks until the async operation completes
func (ar *AsyncResult) Wait(ctx context.Context) (any, error) {
	if ar == nil || ar.request == nil {
		return nil, errors.New("no async operation pending")
	}
	return ar.request.Wait(ctx)
}

// WaitTimeout waits for the async operation with a timeout
func (ar *AsyncResult) WaitTimeout(timeout time.Duration) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ar.Wait(ctx)
}

// IsDone returns true if the operation is complete
func (ar *AsyncResult) IsDone() bool {
	if ar == nil || ar.request == nil {
		return true
	}
	select {
	case <-ar.request.done:
		return true
	default:
		return false
	}
}

// AsyncExecute submits a function for async execution and returns a handle to the result
func (q *RequestQueue) AsyncExecute(fn func(ctx context.Context) (any, error), priority RequestPriority) *AsyncResult {
	if q == nil {
		return &AsyncResult{request: &Request{err: errors.New("queue is nil"), done: make(chan struct{})}}
	}

	req := &Request{
		Priority: priority,
	}
	req.Execute = func(ctx context.Context) error {
		result, err := fn(ctx)
		req.result = result
		return err
	}

	if err := q.Enqueue(req); err != nil {
		// Return immediately with error
		req.err = err
		req.done = make(chan struct{})
		close(req.done)
		return &AsyncResult{request: req}
	}

	return &AsyncResult{request: req}
}

// AsyncExecuteBackground submits a function for background execution
func (q *RequestQueue) AsyncExecuteBackground(fn func(ctx context.Context) (any, error)) *AsyncResult {
	if q == nil {
		return &AsyncResult{request: &Request{err: errors.New("queue is nil"), done: make(chan struct{})}}
	}

	req := &Request{
		Priority: RequestPriorityLow,
	}
	req.Execute = func(ctx context.Context) error {
		result, err := fn(ctx)
		req.result = result
		return err
	}

	if err := q.EnqueueBackground(req); err != nil {
		req.err = err
		req.done = make(chan struct{})
		close(req.done)
		return &AsyncResult{request: req}
	}

	return &AsyncResult{request: req}
}
