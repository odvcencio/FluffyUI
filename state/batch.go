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
		if atomic.AddInt32(&batchDepth, -1) != 0 {
			return
		}
		flush := drainBatch()
		runSubscribers(flush)
	}()
	fn()
}

func batching() bool {
	return atomic.LoadInt32(&batchDepth) > 0
}

func enqueueBatch(subs []subscriber) bool {
	if len(subs) == 0 {
		return true
	}
	if !batching() {
		return false
	}
	batchMu.Lock()
	if batchDepth == 0 {
		batchMu.Unlock()
		return false
	}
	batchQueue = append(batchQueue, subs...)
	batchMu.Unlock()
	return true
}

func drainBatch() []subscriber {
	batchMu.Lock()
	flush := batchQueue
	batchQueue = nil
	batchMu.Unlock()
	return flush
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
