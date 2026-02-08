package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/terminal"
)

func TestPriorityQueue_HighBeforeNormalAndLow(t *testing.T) {
	pq := NewPriorityQueue(16)

	// Enqueue low, normal, then high.
	pq.Send(TickMsg{Time: time.Now()})
	pq.Send(MouseMsg{X: 1, Y: 2})
	pq.Send(KeyMsg{Key: terminal.KeyEnter})

	ctx := context.Background()

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message, got none")
	}
	if _, isKey := msg.(KeyMsg); !isKey {
		t.Fatalf("expected KeyMsg first (high priority), got %T", msg)
	}

	msg, ok = pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message, got none")
	}
	if _, isMouse := msg.(MouseMsg); !isMouse {
		t.Fatalf("expected MouseMsg second (normal priority), got %T", msg)
	}

	msg, ok = pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message, got none")
	}
	if _, isTick := msg.(TickMsg); !isTick {
		t.Fatalf("expected TickMsg third (low priority), got %T", msg)
	}
}

func TestPriorityQueue_NormalBeforeLow(t *testing.T) {
	pq := NewPriorityQueue(16)

	pq.Send(InvalidateMsg{})
	pq.Send(PasteMsg{Text: "hello"})

	ctx := context.Background()

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message, got none")
	}
	if _, isPaste := msg.(PasteMsg); !isPaste {
		t.Fatalf("expected PasteMsg first (normal priority), got %T", msg)
	}

	msg, ok = pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message, got none")
	}
	if _, isInvalidate := msg.(InvalidateMsg); !isInvalidate {
		t.Fatalf("expected InvalidateMsg second (low priority), got %T", msg)
	}
}

func TestPriorityQueue_ContextCancellation(t *testing.T) {
	pq := NewPriorityQueue(16)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	msg, ok := pq.Recv(ctx)
	if ok {
		t.Fatalf("expected no message on cancelled context, got %T", msg)
	}
	if msg != nil {
		t.Fatalf("expected nil message, got %T", msg)
	}
}

func TestPriorityQueue_RoundTrip_High(t *testing.T) {
	pq := NewPriorityQueue(4)
	ctx := context.Background()
	sent := KeyMsg{Key: terminal.KeyRune, Rune: 'x'}
	pq.Send(sent)

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	got, isKey := msg.(KeyMsg)
	if !isKey {
		t.Fatalf("expected KeyMsg, got %T", msg)
	}
	if got.Rune != 'x' {
		t.Fatalf("expected Rune 'x', got %c", got.Rune)
	}
}

func TestPriorityQueue_RoundTrip_Normal(t *testing.T) {
	pq := NewPriorityQueue(4)
	ctx := context.Background()
	sent := MouseMsg{X: 42, Y: 7}
	pq.Send(sent)

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	got, isMouse := msg.(MouseMsg)
	if !isMouse {
		t.Fatalf("expected MouseMsg, got %T", msg)
	}
	if got.X != 42 || got.Y != 7 {
		t.Fatalf("expected (42,7), got (%d,%d)", got.X, got.Y)
	}
}

func TestPriorityQueue_RoundTrip_Low(t *testing.T) {
	pq := NewPriorityQueue(4)
	ctx := context.Background()
	now := time.Now()
	pq.Send(TickMsg{Time: now})

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	got, isTick := msg.(TickMsg)
	if !isTick {
		t.Fatalf("expected TickMsg, got %T", msg)
	}
	if got.Time != now {
		t.Fatalf("expected time %v, got %v", now, got.Time)
	}
}

func TestPriorityQueue_ConcurrentSends(t *testing.T) {
	const goroutines = 10
	const messagesPerGoroutine = 50
	total := goroutines * messagesPerGoroutine * 3

	// Use a buffer large enough to hold all messages so none are dropped.
	pq := NewPriorityQueue(total)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // 3 priority levels

	// Spawn goroutines sending high-priority messages.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				pq.Send(KeyMsg{Rune: 'a'})
			}
		}()
	}
	// Spawn goroutines sending normal-priority messages.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				pq.Send(MouseMsg{X: j})
			}
		}()
	}
	// Spawn goroutines sending low-priority messages.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				pq.Send(TickMsg{})
			}
		}()
	}

	wg.Wait()

	// Drain all messages and count them.
	received := 0
	for received < total {
		_, ok := pq.Recv(ctx)
		if !ok {
			t.Fatalf("context cancelled unexpectedly after %d messages", received)
		}
		received++
	}
	if received != total {
		t.Fatalf("expected %d messages, got %d", total, received)
	}
}

func TestPriorityQueue_TrySend_Full(t *testing.T) {
	pq := NewPriorityQueue(1) // size 1 per channel

	// First send should succeed.
	if !pq.TrySend(KeyMsg{Rune: 'a'}) {
		t.Fatal("expected first TrySend to succeed")
	}
	// Second send to same (high) channel should fail.
	if pq.TrySend(KeyMsg{Rune: 'b'}) {
		t.Fatal("expected second TrySend to fail (channel full)")
	}
}

func TestPriorityQueue_FocusChangedMsg_IsHigh(t *testing.T) {
	pq := NewPriorityQueue(4)
	ctx := context.Background()

	pq.Send(TickMsg{})
	pq.Send(FocusChangedMsg{})

	msg, ok := pq.Recv(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	if _, isFocus := msg.(FocusChangedMsg); !isFocus {
		t.Fatalf("expected FocusChangedMsg first (high priority), got %T", msg)
	}
}

func TestPriorityQueue_CustomMsg_IsNormal(t *testing.T) {
	pri := classifyMessage(CustomMsg{Value: "test"})
	if pri != PriorityNormal {
		t.Fatalf("expected CustomMsg to be PriorityNormal, got %d", pri)
	}
}

func TestPriorityQueue_QueueFlushMsg_IsLow(t *testing.T) {
	pri := classifyMessage(QueueFlushMsg{})
	if pri != PriorityLow {
		t.Fatalf("expected QueueFlushMsg to be PriorityLow, got %d", pri)
	}
}

func TestPriorityQueue_RecvBlocksUntilMessage(t *testing.T) {
	pq := NewPriorityQueue(4)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	var got Message
	var ok bool
	go func() {
		got, ok = pq.Recv(ctx)
		close(done)
	}()

	// Small delay to ensure Recv is blocking.
	time.Sleep(20 * time.Millisecond)
	pq.Send(KeyMsg{Rune: 'z'})

	<-done
	if !ok {
		t.Fatal("expected message")
	}
	if km, isKey := got.(KeyMsg); !isKey || km.Rune != 'z' {
		t.Fatalf("expected KeyMsg{Rune:'z'}, got %T", got)
	}
}

func TestClassifyMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want MessagePriority
	}{
		{"KeyMsg", KeyMsg{}, PriorityHigh},
		{"FocusChangedMsg", FocusChangedMsg{}, PriorityHigh},
		{"MouseMsg", MouseMsg{}, PriorityNormal},
		{"PasteMsg", PasteMsg{}, PriorityNormal},
		{"ResizeMsg", ResizeMsg{}, PriorityNormal},
		{"CustomMsg", CustomMsg{}, PriorityNormal},
		{"TickMsg", TickMsg{}, PriorityLow},
		{"InvalidateMsg", InvalidateMsg{}, PriorityLow},
		{"QueueFlushMsg", QueueFlushMsg{}, PriorityLow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyMessage(tt.msg)
			if got != tt.want {
				t.Errorf("classifyMessage(%T) = %d, want %d", tt.msg, got, tt.want)
			}
		})
	}
}

func TestNewPriorityQueue_DefaultSize(t *testing.T) {
	pq := NewPriorityQueue(0)
	if cap(pq.high) != 64 {
		t.Fatalf("expected default cap 64 for high channel, got %d", cap(pq.high))
	}
	if cap(pq.normal) != 64 {
		t.Fatalf("expected default cap 64 for normal channel, got %d", cap(pq.normal))
	}
	if cap(pq.low) != 64 {
		t.Fatalf("expected default cap 64 for low channel, got %d", cap(pq.low))
	}
}

func TestNewPriorityQueue_NegativeSize(t *testing.T) {
	pq := NewPriorityQueue(-5)
	if cap(pq.high) != 64 {
		t.Fatalf("expected default cap 64 for negative size, got %d", cap(pq.high))
	}
}
