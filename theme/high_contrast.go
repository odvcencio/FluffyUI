package theme

import (
	"m31labs.dev/fluffyui/compositor"
)

// HighContrastTheme returns a theme using only the 16 ANSI colors for maximum
// terminal compatibility. All color combinations target WCAG AAA (7:1) contrast
// ratios: bright white text on a black background, with bright yellow for focus
// indicators, bright cyan for links/info, and bright red for errors.
func HighContrastTheme() *Theme {
	return &Theme{
		// Core palette - pure black backgrounds
		Background:    compositor.DefaultStyle().WithBG(compositor.ColorBlack),
		Surface:       compositor.DefaultStyle().WithBG(compositor.ColorBlack),
		SurfaceRaised: compositor.DefaultStyle().WithBG(compositor.ColorBlack),
		SurfaceDim:    compositor.DefaultStyle().WithBG(compositor.ColorBlack),

		// Text hierarchy - bright white for maximum contrast
		TextPrimary:   compositor.DefaultStyle().WithFG(compositor.ColorBrightWhite),
		TextSecondary: compositor.DefaultStyle().WithFG(compositor.ColorWhite),
		TextMuted:     compositor.DefaultStyle().WithFG(compositor.ColorBrightBlack),
		TextInverse:   compositor.DefaultStyle().WithFG(compositor.ColorBlack),

		// Accent colors - bright yellow is highly visible
		Accent:       compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow),
		AccentDim:    compositor.DefaultStyle().WithFG(compositor.ColorYellow),
		AccentGlow:   compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow).WithBold(true),
		ElectricBlue: compositor.DefaultStyle().WithFG(compositor.ColorBrightCyan),
		Coral:        compositor.DefaultStyle().WithFG(compositor.ColorBrightRed),
		Teal:         compositor.DefaultStyle().WithFG(compositor.ColorBrightCyan),

		// Glow variants - use bold instead of dim for visibility
		BlueGlow:   compositor.DefaultStyle().WithFG(compositor.ColorBrightCyan).WithBold(true),
		PurpleGlow: compositor.DefaultStyle().WithFG(compositor.ColorBrightMagenta).WithBold(true),
		CoralGlow:  compositor.DefaultStyle().WithFG(compositor.ColorBrightRed).WithBold(true),

		// Semantic colors - maximally distinct
		Success: compositor.DefaultStyle().WithFG(compositor.ColorBrightGreen),
		Warning: compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow),
		Error:   compositor.DefaultStyle().WithFG(compositor.ColorBrightRed),
		Info:    compositor.DefaultStyle().WithFG(compositor.ColorBrightCyan),

		// Message sources - distinct per source
		User:      compositor.DefaultStyle().WithFG(compositor.ColorBrightGreen),
		Assistant: compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow),
		System:    compositor.DefaultStyle().WithFG(compositor.ColorWhite).WithBold(true),
		Tool:      compositor.DefaultStyle().WithFG(compositor.ColorBrightMagenta),
		Thinking:  compositor.DefaultStyle().WithFG(compositor.ColorBrightBlack).WithBold(true),

		// UI elements - high visibility borders and selections
		Border:      compositor.DefaultStyle().WithFG(compositor.ColorWhite),
		BorderFocus: compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow).WithBold(true),
		Selection:   compositor.DefaultStyle().WithBG(compositor.ColorWhite).WithFG(compositor.ColorBlack),
		SearchMatch: compositor.DefaultStyle().WithBG(compositor.ColorBrightYellow).WithFG(compositor.ColorBlack),
		Scrollbar:   compositor.DefaultStyle().WithFG(compositor.ColorBrightBlack),
		ScrollThumb: compositor.DefaultStyle().WithFG(compositor.ColorBrightWhite),

		// Mode indicators - bold for emphasis
		ModeNormal: compositor.DefaultStyle().WithFG(compositor.ColorBrightWhite),
		ModeShell:  compositor.DefaultStyle().WithFG(compositor.ColorBrightGreen).WithBold(true),
		ModeEnv:    compositor.DefaultStyle().WithFG(compositor.ColorBrightCyan).WithBold(true),
		ModeSearch: compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow).WithBold(true),

		// Special
		Logo:    compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow).WithBold(true),
		Spinner: compositor.DefaultStyle().WithFG(compositor.ColorBrightYellow),
	}
}
