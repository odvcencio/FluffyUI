package agent

import (
	"errors"
	"testing"
	"time"
)

func TestSessionRequestLifecycleAndLimits(t *testing.T) {
	limits := SessionLimits{MaxPendingRequests: 1}
	s := NewSession("s1", ModeInteractive, limits)
	if s.Priority() != PriorityHigh {
		t.Fatalf("expected interactive session to be high priority")
	}

	if err := s.StartRequest(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if err := s.StartRequest(); !errors.Is(err, ErrTooManyRequests) {
		t.Fatalf("expected too many requests, got %v", err)
	}

	s.EndRequest(true)
	stats := s.Stats()
	if stats.CompletedRequests != 1 || stats.PendingRequests != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}

	s.Reject()
	if !s.IsRejected() {
		t.Fatalf("expected rejected session")
	}
	if err := s.CanAcceptRequest(); !errors.Is(err, ErrSessionRejected) {
		t.Fatalf("expected rejected error, got %v", err)
	}

	s.Close()
	if !s.IsClosed() {
		t.Fatalf("expected closed session")
	}
}

func TestSessionID(t *testing.T) {
	if (*Session)(nil).ID() != "" {
		t.Fatal("expected empty id for nil session")
	}
	s := NewSession("session-id", ModeNormal, DefaultSessionLimits())
	if s.ID() != "session-id" {
		t.Fatalf("id = %q", s.ID())
	}
}

func TestSessionMetadataAndAuth(t *testing.T) {
	var nilSession *Session
	if nilSession.Mode() != ModeNormal {
		t.Fatalf("expected ModeNormal for nil session")
	}
	if nilSession.Context() == nil {
		t.Fatal("expected non-nil context for nil session")
	}
	if nilSession.IsAuthed() {
		t.Fatal("expected nil session to be unauthenticated")
	}

	s := NewSession("meta", ModeBackground, DefaultSessionLimits())
	if s.Mode() != ModeBackground {
		t.Fatalf("mode = %v", s.Mode())
	}
	s.SetPriority(PriorityHigh)
	if s.Priority() != PriorityHigh {
		t.Fatalf("priority = %v", s.Priority())
	}
	if s.IsAuthed() {
		t.Fatal("expected authed false before Auth")
	}
	s.Auth()
	if !s.IsAuthed() {
		t.Fatal("expected authed true")
	}
	if s.Context() == nil {
		t.Fatal("expected session context")
	}

	err := (&RateLimitError{}).Error()
	if err == "" {
		t.Fatal("expected rate limit error message")
	}
}

func TestSessionRateLimit(t *testing.T) {
	limits := SessionLimits{MaxRequestsPerSec: 1, BurstLimit: 1}
	s := NewSession("s2", ModeNormal, limits)
	if err := s.StartRequest(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if err := s.StartRequest(); err == nil {
		t.Fatalf("expected rate limit error")
	} else {
		if _, ok := err.(*RateLimitError); !ok {
			t.Fatalf("expected RateLimitError, got %T", err)
		}
	}
	if s.Stats().RateLimited != 1 {
		t.Fatalf("expected rate limited count to be 1")
	}
	s.EndRequest(true)
}

func TestSessionExpiration(t *testing.T) {
	limits := SessionLimits{IdleTimeout: 500 * time.Millisecond}
	s := NewSession("s3", ModeNormal, limits)
	past := time.Now().Add(-time.Second)
	s.lastSeen.Store(past)
	if !s.IsExpired(time.Now()) {
		t.Fatalf("expected session to be expired")
	}
}

func TestSessionPoolLimitsAndStats(t *testing.T) {
	pool := NewSessionPool(PoolLimits{MaxSessions: 1, MaxBackgroundTasks: 1})
	s1, err := pool.CreateSession("one", ModeNormal, DefaultSessionLimits())
	if err != nil || s1 == nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if _, err := pool.CreateSession("two", ModeNormal, DefaultSessionLimits()); !errors.Is(err, ErrSessionRejected) {
		t.Fatalf("expected session rejected, got %v", err)
	}
	stats := pool.Stats()
	if stats.TotalSessions != 1 || stats.NormalSessions != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(pool.ListSessions()) != 1 {
		t.Fatalf("expected one session in list")
	}
	pool.RemoveSession("one")
	if pool.GetSession("one") != nil {
		t.Fatalf("expected session removed")
	}
}

func TestSessionPoolBackgroundLimitAndCleanup(t *testing.T) {
	pool := NewSessionPool(PoolLimits{MaxBackgroundTasks: 1})
	if _, err := pool.CreateSession("bg1", ModeBackground, DefaultSessionLimits()); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if _, err := pool.CreateSession("bg2", ModeBackground, DefaultSessionLimits()); !errors.Is(err, ErrSessionRejected) {
		t.Fatalf("expected background session rejected, got %v", err)
	}

	session, _ := pool.CreateSession("normal", ModeNormal, SessionLimits{IdleTimeout: time.Millisecond})
	session.lastSeen.Store(time.Now().Add(-time.Second))
	pool.cleanupExpired()
	if pool.GetSession("normal") != nil {
		t.Fatalf("expected expired session removed")
	}
}

func TestSessionPoolGlobalRate(t *testing.T) {
	pool := NewSessionPool(PoolLimits{GlobalRateLimit: 1, GlobalBurstLimit: 1})
	if err := pool.CheckGlobalRate(); err != nil {
		t.Fatalf("unexpected rate error: %v", err)
	}
	if err := pool.CheckGlobalRate(); err == nil {
		t.Fatalf("expected global rate limit error")
	}
	stats := pool.Stats()
	if stats.RateLimitedGlobal != 1 {
		t.Fatalf("expected global rate limited count to be 1")
	}
}
