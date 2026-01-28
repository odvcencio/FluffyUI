package theme

import "github.com/odvcencio/fluffyui/style"

// DefaultStylesheet returns the default stylesheet using the theme palette.
func DefaultStylesheet() *style.Stylesheet {
	t := DefaultTheme()
	return style.NewStylesheet().
		Add(style.Select("*"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Background.BG,
		}).
		Add(style.Select("Label"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Text"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Button"), style.Style{
			Foreground: t.Accent.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Button").Class("primary"), style.Style{
			Foreground: t.TextInverse.FG,
			Background: t.Accent.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Button").Class("danger"), style.Style{
			Foreground: t.Coral.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Button").Pseudo(style.PseudoFocus), style.Style{
			Reverse: style.Bool(true),
		}).
		Add(style.Select("Input"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Input").Pseudo(style.PseudoFocus), style.Style{
			Bold: style.Bool(true),
		}).
		Add(style.Select("MultilineInput"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("MultilineInput").Pseudo(style.PseudoFocus), style.Style{
			Bold: style.Bool(true),
		}).
		Add(style.Select("TextArea"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("TextArea").Pseudo(style.PseudoFocus), style.Style{
			Bold: style.Bool(true),
		}).
		Add(style.Select("Select"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Select").Pseudo(style.PseudoFocus), style.Style{
			Reverse: style.Bool(true),
		}).
		Add(style.Select("Checkbox"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Checkbox").Pseudo(style.PseudoFocus), style.Style{
			Reverse: style.Bool(true),
		}).
		Add(style.Select("Radio"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Radio").Pseudo(style.PseudoFocus), style.Style{
			Reverse: style.Bool(true),
		}).
		Add(style.Select("Tabs"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Menu"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Table"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Tree"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("List"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("ScrollView"), style.Style{
			Background: t.Surface.BG,
		}).
		Add(style.Select("Palette"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.SurfaceRaised.BG,
		}).
		Add(style.Select("Search"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("ToastStack"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Dialog"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.SurfaceRaised.BG,
		}).
		Add(style.Select("Progress"), style.Style{
			Foreground: t.Accent.FG,
		}).
		Add(style.Select("Sparkline"), style.Style{
			Foreground: t.Accent.FG,
		}).
		Add(style.Select("BarChart"), style.Style{
			Foreground: t.Accent.FG,
		}).
		Add(style.Select("Section"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Stepper"), style.Style{
			Foreground: t.TextPrimary.FG,
		}).
		Add(style.Select("Spinner"), style.Style{
			Foreground: t.Accent.FG,
		}).
		Add(style.Select("Alert"), style.Style{
			Foreground: t.TextPrimary.FG,
			Background: t.Surface.BG,
		}).
		Add(style.Select("Alert").Class("info"), style.Style{
			Foreground: t.Info.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Alert").Class("success"), style.Style{
			Foreground: t.Success.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Alert").Class("warning"), style.Style{
			Foreground: t.Warning.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Alert").Class("error"), style.Style{
			Foreground: t.Error.FG,
			Bold:       style.Bool(true),
		}).
		Add(style.Select("Panel"), style.Style{
			Background: t.Surface.BG,
			Border: &style.Border{
				Style: style.BorderRounded,
				Color: t.Border.FG,
			},
		}).
		Add(style.Select("Box"), style.Style{
			Background: t.Surface.BG,
		})
}
