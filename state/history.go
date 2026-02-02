package state

import (
	"sync"
	"time"
	"unsafe"
)

// HistoryOption configures a History instance.
type HistoryOption func(*historyConfig)

type historyConfig struct {
	maxDepth    int
	maxBytes    int64
	groupWindow time.Duration
}

// WithMaxDepth sets the maximum number of undo steps (0 = unlimited).
func WithMaxDepth(n int) HistoryOption {
	return func(c *historyConfig) {
		c.maxDepth = n
	}
}

// WithMaxBytes sets the memory limit for history (0 = unlimited).
func WithMaxBytes(n int64) HistoryOption {
	return func(c *historyConfig) {
		c.maxBytes = n
	}
}

// WithGroupWindow sets the time window for grouping operations.
// Operations within this window will be merged into a single undo step.
func WithGroupWindow(d time.Duration) HistoryOption {
	return func(c *historyConfig) {
		c.groupWindow = d
	}
}

// historyNode represents a state in the history tree.
type historyNode[T any] struct {
	state     T
	parent    *historyNode[T]
	children  []*historyNode[T]
	timestamp time.Time
}

// History provides generic undo/redo with branching history support.
type History[T any] struct {
	mu          sync.RWMutex
	root        *historyNode[T]
	current     *historyNode[T]
	redoStack   []*historyNode[T] // Stack of nodes we can redo to
	config      historyConfig
	subscribers map[int]func()
	nextSubID   int
	sizeFunc    func(T) int64 // Optional function to calculate state size
	lastPush    time.Time     // Time of last push for grouping
}

// NewHistory creates a new history with an initial state.
func NewHistory[T any](initial T, opts ...HistoryOption) *History[T] {
	config := historyConfig{
		maxDepth:    0, // unlimited
		maxBytes:    0, // unlimited
		groupWindow: 0, // no grouping
	}
	for _, opt := range opts {
		opt(&config)
	}

	root := &historyNode[T]{
		state:     initial,
		timestamp: time.Now(),
	}

	return &History[T]{
		root:        root,
		current:     root,
		config:      config,
		subscribers: make(map[int]func()),
	}
}

// SetSizeFunc sets a custom function to calculate the size of a state.
// This is used for memory limit enforcement.
func (h *History[T]) SetSizeFunc(fn func(T) int64) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.sizeFunc = fn
	h.mu.Unlock()
}

// Current returns the current state.
func (h *History[T]) Current() T {
	if h == nil {
		var zero T
		return zero
	}
	h.mu.RLock()
	state := h.current.state
	h.mu.RUnlock()
	return state
}

// Push records a new state, clearing the redo stack.
func (h *History[T]) Push(state T) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.pushLocked(state, false)
	h.mu.Unlock()
}

// PushGrouped groups with recent pushes if within the group window.
func (h *History[T]) PushGrouped(state T) {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.pushLocked(state, true)
	h.mu.Unlock()
}

func (h *History[T]) pushLocked(state T, grouped bool) {
	now := time.Now()

	// Check if we should group with the previous state
	if grouped && h.config.groupWindow > 0 && !h.lastPush.IsZero() {
		if now.Sub(h.lastPush) <= h.config.groupWindow {
			// Update the current node's state instead of creating a new one
			h.current.state = state
			h.current.timestamp = now
			h.lastPush = now
			h.notifyLocked()
			return
		}
	}

	// Create a new node
	node := &historyNode[T]{
		state:     state,
		parent:    h.current,
		timestamp: now,
	}

	// Add as child of current node
	h.current.children = append(h.current.children, node)

	// Move to new node
	h.current = node

	// Clear redo stack (branching behavior)
	h.redoStack = nil

	h.lastPush = now

	// Enforce limits
	h.enforceLimitsLocked()

	h.notifyLocked()
}

// Undo moves back one state and returns the new current state.
// Returns false if already at root.
func (h *History[T]) Undo() (T, bool) {
	if h == nil {
		var zero T
		return zero, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current.parent == nil {
		return h.current.state, false
	}

	// Push current to redo stack before moving back
	h.redoStack = append(h.redoStack, h.current)
	h.current = h.current.parent

	// Reset grouping on undo
	h.lastPush = time.Time{}

	h.notifyLocked()
	return h.current.state, true
}

// Redo moves forward one state and returns the new current state.
// Returns false if no redo is available.
func (h *History[T]) Redo() (T, bool) {
	if h == nil {
		var zero T
		return zero, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.redoStack) == 0 {
		return h.current.state, false
	}

	// Pop from redo stack
	n := len(h.redoStack) - 1
	node := h.redoStack[n]
	h.redoStack = h.redoStack[:n]

	h.current = node

	// Reset grouping on redo
	h.lastPush = time.Time{}

	h.notifyLocked()
	return h.current.state, true
}

// CanUndo returns true if undo is possible.
func (h *History[T]) CanUndo() bool {
	if h == nil {
		return false
	}
	h.mu.RLock()
	canUndo := h.current.parent != nil
	h.mu.RUnlock()
	return canUndo
}

// CanRedo returns true if redo is possible.
func (h *History[T]) CanRedo() bool {
	if h == nil {
		return false
	}
	h.mu.RLock()
	canRedo := len(h.redoStack) > 0
	h.mu.RUnlock()
	return canRedo
}

// Clear resets history to the initial state only.
func (h *History[T]) Clear() {
	if h == nil {
		return
	}
	h.mu.Lock()
	// Keep only the root
	h.root.children = nil
	h.current = h.root
	h.redoStack = nil
	h.lastPush = time.Time{}
	h.mu.Unlock()
	h.notify()
}

// UndoDepth returns the number of undo steps available.
func (h *History[T]) UndoDepth() int {
	if h == nil {
		return 0
	}
	h.mu.RLock()
	depth := 0
	node := h.current
	for node.parent != nil {
		depth++
		node = node.parent
	}
	h.mu.RUnlock()
	return depth
}

// RedoDepth returns the number of redo steps available.
func (h *History[T]) RedoDepth() int {
	if h == nil {
		return 0
	}
	h.mu.RLock()
	depth := len(h.redoStack)
	h.mu.RUnlock()
	return depth
}

// Subscribe registers a listener for history changes.
// Returns an unsubscribe function.
func (h *History[T]) Subscribe(fn func()) func() {
	if h == nil || fn == nil {
		return func() {}
	}
	h.mu.Lock()
	id := h.nextSubID
	h.nextSubID++
	h.subscribers[id] = fn
	h.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subscribers, id)
			h.mu.Unlock()
		})
	}
}

func (h *History[T]) notify() {
	h.mu.RLock()
	subs := make([]func(), 0, len(h.subscribers))
	for _, fn := range h.subscribers {
		subs = append(subs, fn)
	}
	h.mu.RUnlock()

	for _, fn := range subs {
		fn()
	}
}

func (h *History[T]) notifyLocked() {
	subs := make([]func(), 0, len(h.subscribers))
	for _, fn := range h.subscribers {
		subs = append(subs, fn)
	}

	// Notify outside the lock to prevent deadlocks
	go func() {
		for _, fn := range subs {
			fn()
		}
	}()
}

func (h *History[T]) enforceLimitsLocked() {
	// Enforce max depth
	if h.config.maxDepth > 0 {
		h.enforceDepthLocked()
	}

	// Enforce max bytes
	if h.config.maxBytes > 0 {
		h.enforceBytesLocked()
	}
}

func (h *History[T]) enforceDepthLocked() {
	depth := 0
	node := h.current
	for node.parent != nil {
		depth++
		node = node.parent
	}

	// If over limit, we need to prune old states
	for depth > h.config.maxDepth {
		// Remove the oldest child of root (that leads to current)
		// Find the path from root to current
		path := h.pathFromRoot()
		if len(path) < 2 {
			break
		}

		// The second node in path is the child of root we should keep
		// Remove all other children from root and make that child the new root
		keepChild := path[1]

		// Create new root from the keep child
		keepChild.parent = nil
		h.root = keepChild
		depth--
	}
}

func (h *History[T]) enforceBytesLocked() {
	totalSize := h.calculateTotalSizeLocked()

	for totalSize > h.config.maxBytes && h.root != h.current {
		// Similar pruning as depth
		path := h.pathFromRoot()
		if len(path) < 2 {
			break
		}

		keepChild := path[1]
		keepChild.parent = nil
		h.root = keepChild

		totalSize = h.calculateTotalSizeLocked()
	}
}

func (h *History[T]) pathFromRoot() []*historyNode[T] {
	// Build path from current to root, then reverse
	var path []*historyNode[T]
	node := h.current
	for node != nil {
		path = append(path, node)
		node = node.parent
	}

	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func (h *History[T]) calculateTotalSizeLocked() int64 {
	return h.calculateNodeSize(h.root)
}

func (h *History[T]) calculateNodeSize(node *historyNode[T]) int64 {
	if node == nil {
		return 0
	}

	var size int64
	if h.sizeFunc != nil {
		size = h.sizeFunc(node.state)
	} else {
		// Estimate size using unsafe.Sizeof (rough approximation)
		size = int64(unsafe.Sizeof(node.state))
	}

	for _, child := range node.children {
		size += h.calculateNodeSize(child)
	}

	return size
}

// Snapshot represents a serializable state of the history.
type Snapshot[T any] struct {
	States  []T
	Current int
}

// TakeSnapshot returns a linear snapshot of states from root to current.
// Useful for serialization.
func (h *History[T]) TakeSnapshot() Snapshot[T] {
	if h == nil {
		return Snapshot[T]{}
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	path := h.pathFromRoot()
	states := make([]T, len(path))
	for i, node := range path {
		states[i] = node.state
	}

	return Snapshot[T]{
		States:  states,
		Current: len(states) - 1,
	}
}

// RestoreSnapshot restores history from a snapshot.
func (h *History[T]) RestoreSnapshot(snap Snapshot[T]) {
	if h == nil || len(snap.States) == 0 {
		return
	}
	h.mu.Lock()

	// Rebuild linear history from snapshot
	h.root = &historyNode[T]{
		state:     snap.States[0],
		timestamp: time.Now(),
	}
	h.current = h.root
	h.redoStack = nil

	for i := 1; i < len(snap.States); i++ {
		node := &historyNode[T]{
			state:     snap.States[i],
			parent:    h.current,
			timestamp: time.Now(),
		}
		h.current.children = append(h.current.children, node)
		h.current = node
	}

	// Move to correct position
	if snap.Current >= 0 && snap.Current < len(snap.States) {
		// Walk back from current to the target position
		targetDepth := snap.Current
		currentDepth := len(snap.States) - 1

		for currentDepth > targetDepth {
			h.redoStack = append(h.redoStack, h.current)
			h.current = h.current.parent
			currentDepth--
		}
	}

	h.lastPush = time.Time{}
	h.mu.Unlock()
	h.notify()
}
