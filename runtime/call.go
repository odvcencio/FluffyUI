package runtime

import (
	"context"
	"errors"
)

type callMsg struct {
	fn   func(*App) error
	done chan error
}

func (callMsg) isMessage() {}

// Call runs fn on the app's event loop and waits for it to finish.
// Call blocks until the function completes or the context is done.
// Call must not be invoked from the app's update/render goroutine.
func (a *App) Call(ctx context.Context, fn func(*App) error) error {
	if a == nil {
		return errors.New("app is nil")
	}
	if fn == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if a.messages == nil {
		return errors.New("app not initialized")
	}

	done := make(chan error, 1)
	msg := callMsg{
		fn:   fn,
		done: done,
	}

	select {
	case a.messages <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
