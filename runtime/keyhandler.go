package runtime

// KeyHandler handles key events before widget dispatch.
type KeyHandler interface {
	HandleKey(app *App, msg KeyMsg, focused Widget) bool
}
