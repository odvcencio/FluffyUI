//go:build !js

package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Session errors
var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRejected    = errors.New("session rejected: server at capacity")
	ErrTooManyRequests    = errors.New("too many pending requests")
	ErrServerShuttingDown = errors.New("server is shutting down")
)

// SessionPriority defines the priority level for a session
type SessionPriority int

const (
	PriorityLow SessionPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityBackground
)

// SessionMode defines how the session operates
type SessionMode int

const (
	ModeNormal      SessionMode = iota
	ModeBackground              // Lower priority, can be throttled more aggressively
	ModeInteractive             // Higher priority for user-interactive sessions
)

// SessionLimits defines resource limits for a session
type SessionLimits struct {
	MaxPendingRequests int           // Max requests in queue (0 = unlimited)
	MaxRequestsPerSec  int           // Rate limit (0 = unlimited)
	BurstLimit         int           // Burst allowance (0 = MaxRequestsPerSec * 2)
	IdleTimeout        time.Duration // Session idle timeout
	MaxRequestDuration time.Duration // Max duration for a single request
}

// DefaultSessionLimits returns reasonable default limits
func DefaultSessionLimits() SessionLimits {
	return SessionLimits{
		MaxPendingRequests: 50,
		MaxRequestsPerSec:  100,
		BurstLimit:         200,
		IdleTimeout:        30 * time.Minute,
		MaxRequestDuration: 5 * time.Minute,
	}
}

// BackgroundSessionLimits returns limits optimized for background sessions
func BackgroundSessionLimits() SessionLimits {
	return SessionLimits{
		MaxPendingRequests: 20,
		MaxRequestsPerSec:  10,
		BurstLimit:         20,
		IdleTimeout:        10 * time.Minute,
		MaxRequestDuration: 10 * time.Minute,
	}
}

// Session represents an active agent session
type Session struct {
	id        string
	priority  SessionPriority
	mode      SessionMode
	limits    SessionLimits
	createdAt time.Time
	lastSeen  atomic.Value // time.Time

	// Request tracking
	pendingRequests   atomic.Int64
	completedRequests atomic.Int64
	failedRequests    atomic.Int64

	// Rate limiting
	limiter *rate.Limiter

	// State
	authed   atomic.Bool
	rejected atomic.Bool
	closed   atomic.Bool

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Metadata
}

// NewSession creates a new session with the given parameters
func NewSession(id string, mode SessionMode, limits SessionLimits) *Session {
	if limits.MaxRequestDuration <= 0 {
		limits.MaxRequestDuration = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Session{
		id:        id,
		mode:      mode,
		limits:    limits,
		createdAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	s.lastSeen.Store(time.Now())

	// Set priority based on mode
	switch mode {
	case ModeBackground:
		s.priority = PriorityBackground
	case ModeInteractive:
		s.priority = PriorityHigh
	default:
		s.priority = PriorityNormal
	}

	// Initialize rate limiter
	if limits.MaxRequestsPerSec > 0 {
		burst := limits.BurstLimit
		if burst <= 0 {
			burst = limits.MaxRequestsPerSec * 2
		}
		s.limiter = rate.NewLimiter(rate.Limit(limits.MaxRequestsPerSec), burst)
	}

	return s
}

// ID returns the session ID
func (s *Session) ID() string {
	if s == nil {
		return ""
	}
	return s.id
}

// Priority returns the session priority
func (s *Session) Priority() SessionPriority {
	if s == nil {
		return PriorityNormal
	}
	return s.priority
}

// Mode returns the session mode
func (s *Session) Mode() SessionMode {
	if s == nil {
		return ModeNormal
	}
	return s.mode
}

// SetPriority updates the session priority
func (s *Session) SetPriority(p SessionPriority) {
	if s == nil {
		return
	}
	s.priority = p
}

// Auth marks the session as authenticated
func (s *Session) Auth() {
	if s == nil {
		return
	}
	s.authed.Store(true)
	s.Touch()
}

// IsAuthed returns true if the session is authenticated
func (s *Session) IsAuthed() bool {
	if s == nil {
		return false
	}
	return s.authed.Load()
}

// Reject marks the session as rejected (server at capacity)
func (s *Session) Reject() {
	if s == nil {
		return
	}
	s.rejected.Store(true)
}

// IsRejected returns true if the session was rejected
func (s *Session) IsRejected() bool {
	if s == nil {
		return false
	}
	return s.rejected.Load()
}

// Touch updates the last seen timestamp
func (s *Session) Touch() {
	if s == nil {
		return
	}
	s.lastSeen.Store(time.Now())
}

// LastSeen returns the last seen timestamp
func (s *Session) LastSeen() time.Time {
	if s == nil {
		return time.Time{}
	}
	if ts, ok := s.lastSeen.Load().(time.Time); ok {
		return ts
	}
	return time.Time{}
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired(now time.Time) bool {
	if s == nil {
		return true
	}
	if s.closed.Load() {
		return true
	}
	if s.limits.IdleTimeout <= 0 {
		return false
	}
	return now.Sub(s.LastSeen()) > s.limits.IdleTimeout
}

// CanAcceptRequest checks if the session can accept a new request
func (s *Session) CanAcceptRequest() error {
	if s == nil {
		return ErrSessionNotFound
	}
	if s.closed.Load() {
		return ErrSessionNotFound
	}
	if s.rejected.Load() {
		return ErrSessionRejected
	}
	if s.limits.MaxPendingRequests > 0 {
		if s.pendingRequests.Load() >= int64(s.limits.MaxPendingRequests) {
			return ErrTooManyRequests
		}
	}
	return nil
}

// StartRequest marks the beginning of a request
func (s *Session) StartRequest() error {
	if err := s.CanAcceptRequest(); err != nil {
		return err
	}

	// Check rate limit
	if s.limiter != nil && !s.limiter.Allow() {
		return &RateLimitError{
			RetryAfter: rateRetryDelay(s.limiter),
		}
	}

	s.pendingRequests.Add(1)
	s.Touch()
	return nil
}

// EndRequest marks the end of a request
func (s *Session) EndRequest(success bool) {
	if s == nil {
		return
	}
	s.pendingRequests.Add(-1)
	if success {
		s.completedRequests.Add(1)
	} else {
		s.failedRequests.Add(1)
	}
	s.Touch()
}

// Context returns the session context
func (s *Session) Context() context.Context {
	if s == nil {
		return context.Background()
	}
	return s.ctx
}

// Close closes the session
func (s *Session) Close() {
	if s == nil {
		return
	}
	if s.closed.CompareAndSwap(false, true) {
		s.cancel()
	}
}

// IsClosed returns true if the session is closed
func (s *Session) IsClosed() bool {
	if s == nil {
		return true
	}
	return s.closed.Load()
}

// Stats returns session statistics
func (s *Session) Stats() SessionStats {
	if s == nil {
		return SessionStats{}
	}
	return SessionStats{
		ID:                s.id,
		CreatedAt:         s.createdAt,
		LastSeen:          s.LastSeen(),
		PendingRequests:   int(s.pendingRequests.Load()),
		CompletedRequests: int(s.completedRequests.Load()),
		FailedRequests:    int(s.failedRequests.Load()),
		Authed:            s.authed.Load(),
		Mode:              s.mode,
		Priority:          s.priority,
	}
}

// SessionStats contains session statistics
type SessionStats struct {
	ID                string          `json:"id"`
	CreatedAt         time.Time       `json:"created_at"`
	LastSeen          time.Time       `json:"last_seen"`
	PendingRequests   int             `json:"pending_requests"`
	CompletedRequests int             `json:"completed_requests"`
	FailedRequests    int             `json:"failed_requests"`
	Authed            bool            `json:"authed"`
	Mode              SessionMode     `json:"mode"`
	Priority          SessionPriority `json:"priority"`
}

// RateLimitError is returned when rate limit is exceeded
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}

// SessionPool manages a pool of sessions with resource limits
type SessionPool struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	limits   PoolLimits

	// Global rate limiting
	globalLimiter *rate.Limiter

	// Housekeeping
	ticker *time.Ticker
	stop   chan struct{}
	wg     sync.WaitGroup
}

// PoolLimits defines global pool limits
type PoolLimits struct {
	MaxSessions        int           // Max total sessions
	MaxBackgroundTasks int           // Max background sessions
	GlobalRateLimit    int           // Global requests per second (0 = unlimited)
	GlobalBurstLimit   int           // Global burst allowance
	CleanupInterval    time.Duration // How often to check for expired sessions
}

// DefaultPoolLimits returns reasonable default pool limits
func DefaultPoolLimits() PoolLimits {
	return PoolLimits{
		MaxSessions:        100,
		MaxBackgroundTasks: 20,
		GlobalRateLimit:    1000,
		GlobalBurstLimit:   2000,
		CleanupInterval:    30 * time.Second,
	}
}

// NewSessionPool creates a new session pool
func NewSessionPool(limits PoolLimits) *SessionPool {
	if limits.CleanupInterval <= 0 {
		limits.CleanupInterval = 30 * time.Second
	}

	pool := &SessionPool{
		sessions: make(map[string]*Session),
		limits:   limits,
		stop:     make(chan struct{}),
	}

	if limits.GlobalRateLimit > 0 {
		burst := limits.GlobalBurstLimit
		if burst <= 0 {
			burst = limits.GlobalRateLimit * 2
		}
		pool.globalLimiter = rate.NewLimiter(rate.Limit(limits.GlobalRateLimit), burst)
	}

	return pool
}

// Start begins background housekeeping
func (p *SessionPool) Start() {
	if p == nil {
		return
	}
	p.ticker = time.NewTicker(p.limits.CleanupInterval)
	p.wg.Add(1)
	go p.housekeeping()
}

// Stop stops background housekeeping
func (p *SessionPool) Stop() {
	if p == nil {
		return
	}
	close(p.stop)
	if p.ticker != nil {
		p.ticker.Stop()
	}
	p.wg.Wait()

	// Close all sessions
	p.mu.Lock()
	for _, s := range p.sessions {
		s.Close()
	}
	clear(p.sessions)
	p.mu.Unlock()
}

// CreateSession creates a new session in the pool
func (p *SessionPool) CreateSession(id string, mode SessionMode, limits SessionLimits) (*Session, error) {
	if p == nil {
		return nil, errors.New("session pool is nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check global limits
	if p.limits.MaxSessions > 0 && len(p.sessions) >= p.limits.MaxSessions {
		return nil, ErrSessionRejected
	}

	// Check background task limits
	if mode == ModeBackground && p.limits.MaxBackgroundTasks > 0 {
		backgroundCount := 0
		for _, s := range p.sessions {
			if s.mode == ModeBackground {
				backgroundCount++
			}
		}
		if backgroundCount >= p.limits.MaxBackgroundTasks {
			return nil, ErrSessionRejected
		}
	}

	session := NewSession(id, mode, limits)
	p.sessions[id] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (p *SessionPool) GetSession(id string) *Session {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sessions[id]
}

// RemoveSession removes a session from the pool
func (p *SessionPool) RemoveSession(id string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if s, ok := p.sessions[id]; ok {
		s.Close()
		delete(p.sessions, id)
	}
}

// CheckGlobalRate returns nil if the global rate limit allows the request
func (p *SessionPool) CheckGlobalRate() error {
	if p == nil || p.globalLimiter == nil {
		return nil
	}
	if !p.globalLimiter.Allow() {
		return &RateLimitError{
			RetryAfter: rateRetryDelay(p.globalLimiter),
		}
	}
	return nil
}

// Stats returns pool statistics
func (p *SessionPool) Stats() PoolStats {
	if p == nil {
		return PoolStats{}
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		TotalSessions: len(p.sessions),
		MaxSessions:   p.limits.MaxSessions,
	}

	for _, s := range p.sessions {
		switch s.mode {
		case ModeBackground:
			stats.BackgroundSessions++
		case ModeInteractive:
			stats.InteractiveSessions++
		default:
			stats.NormalSessions++
		}
		stats.TotalPendingRequests += int(s.pendingRequests.Load())
	}

	return stats
}

// ListSessions returns all active session IDs
func (p *SessionPool) ListSessions() []string {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]string, 0, len(p.sessions))
	for id := range p.sessions {
		ids = append(ids, id)
	}
	return ids
}

// housekeeping performs periodic cleanup of expired sessions
func (p *SessionPool) housekeeping() {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		case <-p.ticker.C:
			p.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired sessions
func (p *SessionPool) cleanupExpired() {
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, s := range p.sessions {
		if s.IsExpired(now) {
			s.Close()
			delete(p.sessions, id)
		}
	}
}

// PoolStats contains pool statistics
type PoolStats struct {
	TotalSessions        int `json:"total_sessions"`
	MaxSessions          int `json:"max_sessions"`
	NormalSessions       int `json:"normal_sessions"`
	BackgroundSessions   int `json:"background_sessions"`
	InteractiveSessions  int `json:"interactive_sessions"`
	TotalPendingRequests int `json:"total_pending_requests"`
}

// Helper function to calculate retry delay
func rateRetryDelay(limiter *rate.Limiter) time.Duration {
	if limiter == nil {
		return time.Second
	}
	res := limiter.Reserve()
	if !res.OK() {
		return time.Second
	}
	delay := res.Delay()
	res.CancelAt(time.Now())
	if delay < 0 {
		return 0
	}
	return delay
}
