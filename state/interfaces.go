package state

// Readable exposes read-only access to a reactive value. Both Signal and
// Computed satisfy this interface.
type Readable[T any] interface {
	Get() T
	Subscribe(fn func()) func()
	SubscribeWithScheduler(scheduler Scheduler, fn func()) func()
}

// Writable exposes read/write access to a reactive value. Signal satisfies
// this interface; Computed does not (it is read-only).
type Writable[T any] interface {
	Readable[T]
	Set(value T) bool
	Update(fn func(T) T) bool
}
