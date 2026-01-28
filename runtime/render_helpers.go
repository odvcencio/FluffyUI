package runtime

// RenderChild renders a child widget if it intersects the current context bounds.
// Returns true if the widget was rendered.
func RenderChild(ctx RenderContext, child Widget) bool {
	if child == nil {
		return false
	}
	if boundsProvider, ok := child.(BoundsProvider); ok {
		bounds := boundsProvider.Bounds()
		if bounds.Width <= 0 || bounds.Height <= 0 {
			return false
		}
		if !ctx.Visible(bounds) {
			return false
		}
		child.Render(ctx.Sub(bounds))
		return true
	}
	child.Render(ctx)
	return true
}
