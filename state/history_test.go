package state

import (
	"sync"
	"testing"
	"time"
)

func TestHistory_NewHistory(t *testing.T) {
	h := NewHistory("initial")
	if h == nil {
		t.Fatal("expected non-nil history")
	}
	if got := h.Current(); got != "initial" {
		t.Errorf("expected initial state 'initial', got %q", got)
	}
}

func TestHistory_Push(t *testing.T) {
	h := NewHistory("state1")
	h.Push("state2")

	if got := h.Current(); got != "state2" {
		t.Errorf("expected current state 'state2', got %q", got)
	}
	if !h.CanUndo() {
		t.Error("expected CanUndo to be true after push")
	}
}

func TestHistory_UndoRedo(t *testing.T) {
	h := NewHistory("state1")
	h.Push("state2")
	h.Push("state3")

	// Test undo
	state, ok := h.Undo()
	if !ok {
		t.Fatal("expected undo to succeed")
	}
	if state != "state2" {
		t.Errorf("expected state2 after undo, got %q", state)
	}

	// Test redo
	state, ok = h.Redo()
	if !ok {
		t.Fatal("expected redo to succeed")
	}
	if state != "state3" {
		t.Errorf("expected state3 after redo, got %q", state)
	}

	// Redo when at end
	_, ok = h.Redo()
	if ok {
		t.Error("expected redo to fail at end of history")
	}
}

func TestHistory_UndoAtRoot(t *testing.T) {
	h := NewHistory("initial")

	state, ok := h.Undo()
	if ok {
		t.Error("expected undo to fail at root")
	}
	if state != "initial" {
		t.Errorf("expected initial state after failed undo, got %q", state)
	}
}

func TestHistory_BranchingBehavior(t *testing.T) {
	// State1 -> State2 -> State3
	//                |
	//                +--> State4 (after undo and new push)
	h := NewHistory("state1")
	h.Push("state2")
	h.Push("state3")

	// Undo back to state2
	h.Undo()

	// Push new state - should clear redo stack
	h.Push("state4")

	if got := h.Current(); got != "state4" {
		t.Errorf("expected state4, got %q", got)
	}

	// Redo should not be possible (redo stack cleared)
	if h.CanRedo() {
		t.Error("expected CanRedo to be false after new push")
	}

	// Can still undo
	state, ok := h.Undo()
	if !ok || state != "state2" {
		t.Errorf("expected state2 after undo, got %q, ok=%v", state, ok)
	}
}

func TestHistory_Grouping(t *testing.T) {
	h := NewHistory("initial", WithGroupWindow(100*time.Millisecond))

	// Rapid pushes should be grouped
	h.PushGrouped("state1") // t=0ms
	h.PushGrouped("state2") // Should replace state1
	h.PushGrouped("state3") // Should replace state2

	// Current should be state3
	if got := h.Current(); got != "state3" {
		t.Errorf("expected state3, got %q", got)
	}

	// Should only need one undo to get back to initial
	state, ok := h.Undo()
	if !ok || state != "initial" {
		t.Errorf("expected initial after single undo (grouped), got %q, ok=%v", state, ok)
	}
}

func TestHistory_GroupingWithDelay(t *testing.T) {
	h := NewHistory("initial", WithGroupWindow(50*time.Millisecond))

	h.PushGrouped("state1")
	h.PushGrouped("state2") // Grouped with state1

	// Wait for group window to expire
	time.Sleep(60 * time.Millisecond)

	h.PushGrouped("state3") // New group

	// Should need two undos
	if depth := h.UndoDepth(); depth != 2 {
		t.Errorf("expected undo depth 2, got %d", depth)
	}
}

func TestHistory_MaxDepth(t *testing.T) {
	h := NewHistory("state0", WithMaxDepth(3))

	h.Push("state1")
	h.Push("state2")
	h.Push("state3")
	h.Push("state4") // Should prune oldest

	// Should only be able to undo 3 times
	if depth := h.UndoDepth(); depth > 3 {
		t.Errorf("expected undo depth <= 3, got %d", depth)
	}
}

func TestHistory_Clear(t *testing.T) {
	h := NewHistory("initial")
	h.Push("state1")
	h.Push("state2")
	h.Undo()

	h.Clear()

	if got := h.Current(); got != "initial" {
		t.Errorf("expected initial after clear, got %q", got)
	}
	if h.CanUndo() {
		t.Error("expected CanUndo to be false after clear")
	}
	if h.CanRedo() {
		t.Error("expected CanRedo to be false after clear")
	}
}

func TestHistory_UndoDepth(t *testing.T) {
	h := NewHistory("state0")
	if depth := h.UndoDepth(); depth != 0 {
		t.Errorf("expected initial undo depth 0, got %d", depth)
	}

	h.Push("state1")
	if depth := h.UndoDepth(); depth != 1 {
		t.Errorf("expected undo depth 1, got %d", depth)
	}

	h.Push("state2")
	h.Push("state3")
	if depth := h.UndoDepth(); depth != 3 {
		t.Errorf("expected undo depth 3, got %d", depth)
	}
}

func TestHistory_RedoDepth(t *testing.T) {
	h := NewHistory("state0")
	h.Push("state1")
	h.Push("state2")

	if depth := h.RedoDepth(); depth != 0 {
		t.Errorf("expected redo depth 0, got %d", depth)
	}

	h.Undo()
	if depth := h.RedoDepth(); depth != 1 {
		t.Errorf("expected redo depth 1 after one undo, got %d", depth)
	}

	h.Undo()
	if depth := h.RedoDepth(); depth != 2 {
		t.Errorf("expected redo depth 2 after two undos, got %d", depth)
	}
}

func TestHistory_Subscribe(t *testing.T) {
	h := NewHistory("initial")
	calls := 0
	var mu sync.Mutex

	unsub := h.Subscribe(func() {
		mu.Lock()
		calls++
		mu.Unlock()
	})

	h.Push("state1")

	// Give time for async notification
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if calls != 1 {
		t.Errorf("expected 1 notification, got %d", calls)
	}
	mu.Unlock()

	unsub()
	h.Push("state2")

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if calls != 1 {
		t.Errorf("expected still 1 notification after unsub, got %d", calls)
	}
	mu.Unlock()
}

func TestHistory_SubscribeMultiple(t *testing.T) {
	h := NewHistory("initial")
	calls1 := 0
	calls2 := 0
	var mu sync.Mutex

	h.Subscribe(func() {
		mu.Lock()
		calls1++
		mu.Unlock()
	})
	h.Subscribe(func() {
		mu.Lock()
		calls2++
		mu.Unlock()
	})

	h.Push("state1")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if calls1 != 1 || calls2 != 1 {
		t.Errorf("expected both subscribers notified, got calls1=%d, calls2=%d", calls1, calls2)
	}
	mu.Unlock()
}

func TestHistory_NilSafety(t *testing.T) {
	var h *History[string]

	// All methods should be safe on nil
	if got := h.Current(); got != "" {
		t.Errorf("expected empty string from nil history, got %q", got)
	}
	h.Push("test")
	h.PushGrouped("test")
	if _, ok := h.Undo(); ok {
		t.Error("expected undo to fail on nil history")
	}
	if _, ok := h.Redo(); ok {
		t.Error("expected redo to fail on nil history")
	}
	if h.CanUndo() {
		t.Error("expected CanUndo false on nil history")
	}
	if h.CanRedo() {
		t.Error("expected CanRedo false on nil history")
	}
	h.Clear()
	if depth := h.UndoDepth(); depth != 0 {
		t.Errorf("expected 0 undo depth on nil history, got %d", depth)
	}
	if depth := h.RedoDepth(); depth != 0 {
		t.Errorf("expected 0 redo depth on nil history, got %d", depth)
	}
	unsub := h.Subscribe(func() {})
	unsub() // Should not panic
}

func TestHistory_Snapshot(t *testing.T) {
	h := NewHistory("state0")
	h.Push("state1")
	h.Push("state2")

	snap := h.TakeSnapshot()

	if len(snap.States) != 3 {
		t.Errorf("expected 3 states in snapshot, got %d", len(snap.States))
	}
	if snap.Current != 2 {
		t.Errorf("expected current index 2, got %d", snap.Current)
	}
}

func TestHistory_RestoreSnapshot(t *testing.T) {
	h := NewHistory("temp")

	snap := Snapshot[string]{
		States:  []string{"state0", "state1", "state2"},
		Current: 1,
	}

	h.RestoreSnapshot(snap)

	if got := h.Current(); got != "state1" {
		t.Errorf("expected state1 after restore, got %q", got)
	}
	if !h.CanUndo() {
		t.Error("expected CanUndo true after restore")
	}
	if !h.CanRedo() {
		t.Error("expected CanRedo true after restore")
	}
}

func TestHistory_ConcurrentAccess(t *testing.T) {
	h := NewHistory(0)
	var wg sync.WaitGroup

	// Concurrent pushes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			h.Push(val)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.Current()
			_ = h.CanUndo()
			_ = h.CanRedo()
		}()
	}

	wg.Wait()
}

func TestHistory_UndoRedoNotifications(t *testing.T) {
	h := NewHistory("state0")
	h.Push("state1")
	h.Push("state2")

	notifications := 0
	var mu sync.Mutex

	h.Subscribe(func() {
		mu.Lock()
		notifications++
		mu.Unlock()
	})

	h.Undo()
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if notifications != 1 {
		t.Errorf("expected 1 notification after undo, got %d", notifications)
	}
	mu.Unlock()

	h.Redo()
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if notifications != 2 {
		t.Errorf("expected 2 notifications after redo, got %d", notifications)
	}
	mu.Unlock()
}

func TestHistory_ComplexState(t *testing.T) {
	type State struct {
		Text      string
		CursorPos int
		Selection struct {
			Start int
			End   int
		}
	}

	h := NewHistory(State{Text: "", CursorPos: 0})
	h.Push(State{Text: "h", CursorPos: 1})
	h.Push(State{Text: "he", CursorPos: 2})
	h.Push(State{Text: "hel", CursorPos: 3})

	state, ok := h.Undo()
	if !ok {
		t.Fatal("expected undo to succeed")
	}
	if state.Text != "he" || state.CursorPos != 2 {
		t.Errorf("unexpected state after undo: %+v", state)
	}
}

func TestHistory_WithMaxBytes(t *testing.T) {
	h := NewHistory("a", WithMaxBytes(100))
	h.SetSizeFunc(func(s string) int64 {
		return int64(len(s))
	})

	// Push many states
	for i := 0; i < 200; i++ {
		h.Push(string(make([]byte, 10)))
	}

	// History should be pruned but still functional
	if !h.CanUndo() {
		t.Error("expected some history to remain")
	}
}

func TestHistory_GroupingResetOnUndo(t *testing.T) {
	h := NewHistory("initial", WithGroupWindow(100*time.Millisecond))

	h.PushGrouped("state1")
	h.Undo()

	// After undo, next push should not be grouped
	h.PushGrouped("state2")

	// Should be at state2
	if got := h.Current(); got != "state2" {
		t.Errorf("expected state2, got %q", got)
	}

	// Undo should go back to initial, not state1
	state, ok := h.Undo()
	if !ok || state != "initial" {
		t.Errorf("expected initial after undo, got %q", state)
	}
}

func TestHistory_EmptySnapshot(t *testing.T) {
	var h *History[string]
	snap := h.TakeSnapshot()

	if len(snap.States) != 0 {
		t.Errorf("expected empty snapshot from nil history, got %d states", len(snap.States))
	}
}

func TestHistory_RestoreEmptySnapshot(t *testing.T) {
	h := NewHistory("initial")
	h.Push("state1")

	// Restore empty snapshot should not change anything
	h.RestoreSnapshot(Snapshot[string]{})

	if got := h.Current(); got != "state1" {
		t.Errorf("expected state1 unchanged after empty restore, got %q", got)
	}
}

func TestHistory_NonGroupedPush(t *testing.T) {
	h := NewHistory("initial", WithGroupWindow(100*time.Millisecond))

	// Non-grouped push should always create new entry
	h.Push("state1")
	h.Push("state2")

	if depth := h.UndoDepth(); depth != 2 {
		t.Errorf("expected undo depth 2 for non-grouped pushes, got %d", depth)
	}
}

func TestHistory_SubscribeNilFunction(t *testing.T) {
	h := NewHistory("initial")
	unsub := h.Subscribe(nil)
	unsub() // Should not panic
}

func BenchmarkHistory_Push(b *testing.B) {
	h := NewHistory(0)
	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
}

func BenchmarkHistory_PushGrouped(b *testing.B) {
	h := NewHistory(0, WithGroupWindow(time.Second))
	for i := 0; i < b.N; i++ {
		h.PushGrouped(i)
	}
}

func BenchmarkHistory_UndoRedo(b *testing.B) {
	h := NewHistory(0)
	for i := 0; i < 100; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if h.CanUndo() {
			h.Undo()
		}
		if h.CanRedo() {
			h.Redo()
		}
	}
}
