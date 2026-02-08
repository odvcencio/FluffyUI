package runtime

import "context"

// MessagePriority indicates how urgently a message should be processed.
type MessagePriority int

const (
	// PriorityHigh is for keyboard, focus, and accessibility messages
	// that must be processed with minimal latency.
	PriorityHigh MessagePriority = iota

	// PriorityNormal is for mouse, paste, resize, and custom messages.
	PriorityNormal

	// PriorityLow is for tick, invalidate, and queue flush messages
	// that can tolerate some delay.
	PriorityLow
)

// PriorityQueue routes messages into priority-stratified channels so that
// high-priority events (keyboard input) are never blocked behind bulk
// low-priority traffic (timer ticks, invalidations).
type PriorityQueue struct {
	high   chan Message
	normal chan Message
	low    chan Message
}

// NewPriorityQueue creates a PriorityQueue with buffered channels of the
// given size per priority level.
func NewPriorityQueue(size int) *PriorityQueue {
	if size <= 0 {
		size = 64
	}
	return &PriorityQueue{
		high:   make(chan Message, size),
		normal: make(chan Message, size),
		low:    make(chan Message, size),
	}
}

// Send classifies msg and sends it to the appropriate priority channel.
// If the target channel is full, Send drops the message (non-blocking).
func (q *PriorityQueue) Send(msg Message) {
	ch := q.channelFor(msg)
	select {
	case ch <- msg:
	default:
		// Drop rather than block, matching tryPost semantics.
	}
}

// TrySend classifies msg and attempts a non-blocking send.
// It returns true if the message was enqueued, false if the channel was full.
func (q *PriorityQueue) TrySend(msg Message) bool {
	ch := q.channelFor(msg)
	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}

// Recv returns the highest-priority pending message, blocking only when all
// channels are empty. It returns (nil, false) when ctx is cancelled.
//
// Drain order: high before normal, normal before low.
func (q *PriorityQueue) Recv(ctx context.Context) (Message, bool) {
	// First, try high priority non-blocking.
	select {
	case msg := <-q.high:
		return msg, true
	default:
	}
	// Then try high or normal non-blocking.
	select {
	case msg := <-q.high:
		return msg, true
	case msg := <-q.normal:
		return msg, true
	default:
	}
	// Finally block on all three + context.
	select {
	case <-ctx.Done():
		return nil, false
	case msg := <-q.high:
		return msg, true
	case msg := <-q.normal:
		return msg, true
	case msg := <-q.low:
		return msg, true
	}
}

// classifyMessage determines the priority of a message based on its type.
func classifyMessage(msg Message) MessagePriority {
	switch msg.(type) {
	// High priority: keyboard input and focus changes need immediate processing.
	case KeyMsg, FocusChangedMsg:
		return PriorityHigh

	// Low priority: periodic and housekeeping messages that can wait.
	case TickMsg, InvalidateMsg, QueueFlushMsg:
		return PriorityLow

	// Normal priority: everything else (mouse, paste, resize, custom, call).
	default:
		return PriorityNormal
	}
}

// channelFor returns the channel corresponding to the message's priority.
func (q *PriorityQueue) channelFor(msg Message) chan Message {
	switch classifyMessage(msg) {
	case PriorityHigh:
		return q.high
	case PriorityLow:
		return q.low
	default:
		return q.normal
	}
}
