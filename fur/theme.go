package fur

import "github.com/odvcencio/fluffy-ui/theme"

// Theme is an alias to the FluffyUI theme.
type Theme = theme.Theme

// DefaultTheme returns the default theme.
func DefaultTheme() *Theme {
	return theme.DefaultTheme()
}
