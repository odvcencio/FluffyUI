package state

import (
	"sync"
	"sync/atomic"
)

var (
	batchDepth int32
	batchMu    sync.Mutex
	batchQueue []subscriber
)

// Batch defers subscriber notifications until fn completes.
func Batch(fn func()) {
	if fn == nil {
		return
	}
	atomic.AddInt32(&batchDepth, 1)
	defer func() {
		// Use mutex to coordinate with enqueueBatch - decrement and drain
		// must be atomic to prevent race where enqueue sees depth=0 but
		// queue hasn't been drained yet.
		batchMu.Lock()
		newDepth := atomic.AddInt32(&batchDepth, -1)
		var flush []subscriber
		if newDepth == 0 {
			flush = batchQueue
			batchQueue = nil
		}
		batchMu.Unlock()
		runSubscribers(flush)
	}()
	fn()
}

func enqueueBatch(subs []subscriber) bool {
	if len(subs) == 0 {
		return true
	}
	batchMu.Lock()
	if atomic.LoadInt32(&batchDepth) == 0 {
		batchMu.Unlock()
		return false
	}
	batchQueue = append(batchQueue, subs...)
	batchMu.Unlock()
	return true
}

func runSubscribers(subs []subscriber) {
	if len(subs) == 0 {
		return
	}
	for _, sub := range subs {
		if sub.fn == nil {
			continue
		}
		if sub.scheduler == nil {
			sub.fn()
			continue
		}
		sub.scheduler.Schedule(sub.fn)
	}
}
