package widgets

import (
	"hash/fnv"
	"strings"
	"unicode"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

// AvatarSize controls the avatar display dimensions.
type AvatarSize int

const (
	// AvatarSmall renders as 3 wide x 1 tall.
	AvatarSmall AvatarSize = iota
	// AvatarMedium renders as 5 wide x 3 tall (with border).
	AvatarMedium
	// AvatarLarge renders as 7 wide x 3 tall (with border).
	AvatarLarge
)

// AvatarOption configures an Avatar widget.
type AvatarOption = Option[Avatar]

// Avatar displays user initials or a placeholder character in a styled block.
// The background color is deterministically derived from the name.
type Avatar struct {
	Base

	name     string
	initials string
	size     AvatarSize
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewAvatar creates an avatar widget for the given name.
func NewAvatar(name string, opts ...AvatarOption) *Avatar {
	a := &Avatar{
		name:  name,
		size:  AvatarSmall,
		style: backend.DefaultStyle(),
	}
	a.initials = computeInitials(name)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(a)
	}
	a.syncA11y()
	return a
}

// WithAvatarSize sets the display size.
func WithAvatarSize(size AvatarSize) AvatarOption {
	return func(a *Avatar) {
		if a == nil {
			return
		}
		a.size = size
	}
}

// WithAvatarInitials overrides the computed initials.
func WithAvatarInitials(initials string) AvatarOption {
	return func(a *Avatar) {
		if a == nil {
			return
		}
		a.initials = initials
	}
}

// Name returns the avatar name.
func (a *Avatar) Name() string {
	if a == nil {
		return ""
	}
	return a.name
}

// Initials returns the displayed initials.
func (a *Avatar) Initials() string {
	if a == nil {
		return ""
	}
	return a.initials
}

// Size returns the avatar size.
func (a *Avatar) Size() AvatarSize {
	if a == nil {
		return AvatarSmall
	}
	return a.size
}

// SetStyle sets the widget style.
func (a *Avatar) SetStyle(style backend.Style) {
	if a == nil {
		return
	}
	a.style = style
	a.styleSet = true
}

// StyleType returns the selector type name.
func (a *Avatar) StyleType() string {
	return "Avatar"
}

// Measure returns the desired size.
func (a *Avatar) Measure(constraints runtime.Constraints) runtime.Size {
	return a.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		w, h := avatarDimensions(a.size)
		return contentConstraints.Constrain(runtime.Size{Width: w, Height: h})
	})
}

// Layout stores the assigned bounds.
func (a *Avatar) Layout(bounds runtime.Rect) {
	if a == nil {
		return
	}
	a.Base.Layout(bounds)
}

// Render draws the avatar.
func (a *Avatar) Render(ctx runtime.RenderContext) {
	if a == nil {
		return
	}
	a.syncA11y()
	bounds := a.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, a, a.style, a.styleSet)
	bgColor := avatarColor(a.name)
	style = style.Background(bgColor).Foreground(backend.ColorWhite)

	switch a.size {
	case AvatarSmall:
		text := "[" + a.initials + "]"
		writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
	case AvatarMedium:
		w := 5
		if bounds.Width < w {
			w = bounds.Width
		}
		if bounds.Height >= 3 {
			writePadded(ctx.Buffer, bounds.X, bounds.Y, w, truncateString("\u250c\u2500\u2500\u2500\u2510", w), style)
			inner := "\u2502 " + padRight(a.initials, 2) + "\u2502"
			writePadded(ctx.Buffer, bounds.X, bounds.Y+1, w, truncateString(inner, w), style)
			writePadded(ctx.Buffer, bounds.X, bounds.Y+2, w, truncateString("\u2514\u2500\u2500\u2500\u2518", w), style)
		} else {
			text := "[" + a.initials + "]"
			writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
		}
	case AvatarLarge:
		w := 7
		if bounds.Width < w {
			w = bounds.Width
		}
		if bounds.Height >= 3 {
			writePadded(ctx.Buffer, bounds.X, bounds.Y, w, truncateString("\u250c\u2500\u2500\u2500\u2500\u2500\u2510", w), style)
			inner := "\u2502  " + padRight(a.initials, 2) + " \u2502"
			writePadded(ctx.Buffer, bounds.X, bounds.Y+1, w, truncateString(inner, w), style)
			writePadded(ctx.Buffer, bounds.X, bounds.Y+2, w, truncateString("\u2514\u2500\u2500\u2500\u2500\u2500\u2518", w), style)
		} else {
			text := "[ " + a.initials + " ]"
			writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
		}
	}
}

// HandleMessage returns unhandled (non-interactive widget).
func (a *Avatar) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// Bind attaches app services.
func (a *Avatar) Bind(services runtime.Services) {
	if a == nil {
		return
	}
	a.services = services
}

// Unbind releases app services.
func (a *Avatar) Unbind() {
	if a == nil {
		return
	}
	a.services = runtime.Services{}
}

func (a *Avatar) syncA11y() {
	if a == nil {
		return
	}
	if a.Base.Role == "" {
		a.Base.Role = accessibility.RoleImage
	}
	a.Base.Label = a.name
	a.Base.Description = "Avatar for " + a.name
}

func computeInitials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	words := strings.Fields(name)
	var sb strings.Builder
	count := 0
	for _, w := range words {
		if count >= 2 {
			break
		}
		for _, r := range w {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sb.WriteRune(unicode.ToUpper(r))
				count++
				break
			}
		}
	}
	if sb.Len() == 0 {
		return "?"
	}
	return sb.String()
}

func avatarColor(name string) backend.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	colors := []backend.Color{
		backend.ColorRed,
		backend.ColorGreen,
		backend.ColorBlue,
		backend.ColorMagenta,
		backend.ColorCyan,
		backend.ColorYellow,
		backend.ColorBrightRed,
		backend.ColorBrightBlue,
	}
	idx := int(h.Sum32()) % len(colors)
	if idx < 0 {
		idx = -idx
	}
	return colors[idx]
}

func avatarDimensions(size AvatarSize) (int, int) {
	switch size {
	case AvatarSmall:
		return 4, 1
	case AvatarMedium:
		return 5, 3
	case AvatarLarge:
		return 7, 3
	default:
		return 4, 1
	}
}

var _ runtime.Widget = (*Avatar)(nil)
var _ runtime.Bindable = (*Avatar)(nil)
var _ runtime.Unbindable = (*Avatar)(nil)
