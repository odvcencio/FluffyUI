package style

// DarkTheme returns a stylesheet with dark terminal colors.
// It provides a cohesive dark palette suitable for most terminals,
// using the standard 16 terminal colors for maximum compatibility.
func DarkTheme() *Stylesheet {
	s := NewStylesheet()

	// -- Label / Text --
	s.Add(Select("Label"), Style{
		Foreground: ColorWhite,
	})
	s.Add(Select("Text"), Style{
		Foreground: ColorWhite,
	})

	// -- Button --
	s.Add(Select("Button"), Style{
		Foreground: ColorBlack,
		Background: ColorWhite,
		Bold:       Bool(true),
		Padding:    PadXY(2, 0),
	})
	s.Add(Select("Button").Pseudo(PseudoFocus), Style{
		Foreground: ColorBlack,
		Background: ColorBrightCyan,
		Bold:       Bool(true),
	})
	s.Add(Select("Button").Class("primary"), Style{
		Foreground: ColorBlack,
		Background: ColorBlue,
	})
	s.Add(Select("Button").Class("danger"), Style{
		Foreground: ColorWhite,
		Background: ColorRed,
	})
	s.Add(Select("Button").Pseudo(PseudoDisabled), Style{
		Foreground: ColorBrightBlack,
		Background: ColorBrightBlack,
		Dim:        Bool(true),
	})

	// -- Input --
	s.Add(Select("Input"), Style{
		Foreground: ColorWhite,
		Background: ColorBrightBlack,
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Input").Pseudo(PseudoFocus), Style{
		Border: &Border{Style: BorderSingle, Color: ColorCyan, StyleSet: true, ColorSet: true},
	})

	// -- Panel --
	s.Add(Select("Panel"), Style{
		Foreground: ColorWhite,
		Background: ColorBlack,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Dialog --
	s.Add(Select("Dialog"), Style{
		Foreground: ColorWhite,
		Background: ColorBrightBlack,
		Border:     &Border{Style: BorderRounded, Color: ColorCyan, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Tabs --
	s.Add(Select("Tabs"), Style{
		Foreground: ColorWhite,
		Background: ColorBlack,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})

	// -- Alert --
	s.Add(Select("Alert"), Style{
		Foreground: ColorWhite,
		Bold:       Bool(true),
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: ColorWhite, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("info"), Style{
		Foreground: ColorCyan,
		Border:     &Border{Style: BorderSingle, Color: ColorCyan, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("warning"), Style{
		Foreground: ColorYellow,
		Border:     &Border{Style: BorderSingle, Color: ColorYellow, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("error"), Style{
		Foreground: ColorRed,
		Border:     &Border{Style: BorderSingle, Color: ColorRed, StyleSet: true, ColorSet: true},
	})

	// -- Table --
	s.Add(Select("Table"), Style{
		Foreground: ColorWhite,
		Background: ColorBlack,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})

	// -- Progress --
	s.Add(Select("Progress"), Style{
		Foreground: ColorGreen,
		Background: ColorBrightBlack,
	})

	s.SetVariable("theme-name", "dark")
	s.SetVariable("theme-variant", "dark")

	return s
}

// LightTheme returns a stylesheet with light terminal colors.
// It provides a clean, high-contrast palette for light terminal backgrounds.
func LightTheme() *Stylesheet {
	s := NewStylesheet()

	// -- Label / Text --
	s.Add(Select("Label"), Style{
		Foreground: ColorBlack,
	})
	s.Add(Select("Text"), Style{
		Foreground: ColorBlack,
	})

	// -- Button --
	s.Add(Select("Button"), Style{
		Foreground: ColorWhite,
		Background: ColorBlue,
		Bold:       Bool(true),
		Padding:    PadXY(2, 0),
	})
	s.Add(Select("Button").Pseudo(PseudoFocus), Style{
		Foreground: ColorWhite,
		Background: ColorBrightBlue,
		Bold:       Bool(true),
	})
	s.Add(Select("Button").Class("primary"), Style{
		Foreground: ColorWhite,
		Background: ColorBlue,
	})
	s.Add(Select("Button").Class("danger"), Style{
		Foreground: ColorWhite,
		Background: ColorRed,
	})
	s.Add(Select("Button").Pseudo(PseudoDisabled), Style{
		Foreground: ColorBrightWhite,
		Background: ColorBrightWhite,
		Dim:        Bool(true),
	})

	// -- Input --
	s.Add(Select("Input"), Style{
		Foreground: ColorBlack,
		Background: ColorWhite,
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Input").Pseudo(PseudoFocus), Style{
		Border: &Border{Style: BorderSingle, Color: ColorBlue, StyleSet: true, ColorSet: true},
	})

	// -- Panel --
	s.Add(Select("Panel"), Style{
		Foreground: ColorBlack,
		Background: ColorWhite,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Dialog --
	s.Add(Select("Dialog"), Style{
		Foreground: ColorBlack,
		Background: ColorBrightWhite,
		Border:     &Border{Style: BorderRounded, Color: ColorBlue, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Tabs --
	s.Add(Select("Tabs"), Style{
		Foreground: ColorBlack,
		Background: ColorWhite,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})

	// -- Alert --
	s.Add(Select("Alert"), Style{
		Foreground: ColorBlack,
		Bold:       Bool(true),
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: ColorBlack, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("info"), Style{
		Foreground: ColorBlue,
		Border:     &Border{Style: BorderSingle, Color: ColorBlue, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("warning"), Style{
		Foreground: ColorYellow,
		Border:     &Border{Style: BorderSingle, Color: ColorYellow, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("error"), Style{
		Foreground: ColorRed,
		Border:     &Border{Style: BorderSingle, Color: ColorRed, StyleSet: true, ColorSet: true},
	})

	// -- Table --
	s.Add(Select("Table"), Style{
		Foreground: ColorBlack,
		Background: ColorWhite,
		Border:     &Border{Style: BorderSingle, Color: ColorBrightBlack, StyleSet: true, ColorSet: true},
	})

	// -- Progress --
	s.Add(Select("Progress"), Style{
		Foreground: ColorBlue,
		Background: ColorBrightWhite,
	})

	s.SetVariable("theme-name", "light")
	s.SetVariable("theme-variant", "light")

	return s
}

// MonokaiTheme returns a Monokai-inspired color scheme.
// Monokai is known for its warm tones and vibrant syntax-highlighting palette.
// This theme uses 24-bit RGB colors for accurate Monokai reproduction.
func MonokaiTheme() *Stylesheet {
	s := NewStylesheet()

	// Monokai palette
	monoBG := RGB(39, 40, 34)       // dark olive background
	monoFG := RGB(248, 248, 242)    // warm white foreground
	monoComment := RGB(117, 113, 94) // muted olive
	monoPink := RGB(249, 38, 114)    // vivid pink
	monoGreen := RGB(166, 226, 46)   // bright green
	monoYellow := RGB(230, 219, 116) // warm yellow
	monoOrange := RGB(253, 151, 31)  // bright orange
	monoPurple := RGB(174, 129, 255) // soft purple
	monoCyan := RGB(102, 217, 239)   // sky cyan
	monoGutter := RGB(60, 61, 55)    // gutter/border gray

	// -- Label / Text --
	s.Add(Select("Label"), Style{
		Foreground: monoFG,
	})
	s.Add(Select("Text"), Style{
		Foreground: monoFG,
	})

	// -- Button --
	s.Add(Select("Button"), Style{
		Foreground: monoBG,
		Background: monoFG,
		Bold:       Bool(true),
		Padding:    PadXY(2, 0),
	})
	s.Add(Select("Button").Pseudo(PseudoFocus), Style{
		Foreground: monoBG,
		Background: monoYellow,
		Bold:       Bool(true),
	})
	s.Add(Select("Button").Class("primary"), Style{
		Foreground: monoBG,
		Background: monoGreen,
	})
	s.Add(Select("Button").Class("danger"), Style{
		Foreground: monoFG,
		Background: monoPink,
	})
	s.Add(Select("Button").Pseudo(PseudoDisabled), Style{
		Foreground: monoComment,
		Background: monoGutter,
		Dim:        Bool(true),
	})

	// -- Input --
	s.Add(Select("Input"), Style{
		Foreground: monoFG,
		Background: monoGutter,
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: monoComment, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Input").Pseudo(PseudoFocus), Style{
		Border: &Border{Style: BorderSingle, Color: monoOrange, StyleSet: true, ColorSet: true},
	})

	// -- Panel --
	s.Add(Select("Panel"), Style{
		Foreground: monoFG,
		Background: monoBG,
		Border:     &Border{Style: BorderSingle, Color: monoGutter, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Dialog --
	s.Add(Select("Dialog"), Style{
		Foreground: monoFG,
		Background: monoGutter,
		Border:     &Border{Style: BorderRounded, Color: monoOrange, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Tabs --
	s.Add(Select("Tabs"), Style{
		Foreground: monoFG,
		Background: monoBG,
		Border:     &Border{Style: BorderSingle, Color: monoGutter, StyleSet: true, ColorSet: true},
	})

	// -- Alert --
	s.Add(Select("Alert"), Style{
		Foreground: monoFG,
		Bold:       Bool(true),
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: monoFG, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("info"), Style{
		Foreground: monoCyan,
		Border:     &Border{Style: BorderSingle, Color: monoCyan, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("warning"), Style{
		Foreground: monoOrange,
		Border:     &Border{Style: BorderSingle, Color: monoOrange, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("error"), Style{
		Foreground: monoPink,
		Border:     &Border{Style: BorderSingle, Color: monoPink, StyleSet: true, ColorSet: true},
	})

	// -- Table --
	s.Add(Select("Table"), Style{
		Foreground: monoFG,
		Background: monoBG,
		Border:     &Border{Style: BorderSingle, Color: monoGutter, StyleSet: true, ColorSet: true},
	})

	// -- Progress --
	s.Add(Select("Progress"), Style{
		Foreground: monoPurple,
		Background: monoGutter,
	})

	s.SetVariable("theme-name", "monokai")
	s.SetVariable("theme-variant", "dark")

	return s
}

// NordTheme returns a Nord-inspired color scheme.
// Nord is an arctic, cool-toned palette with blue-gray foundations
// and pastel accents. This theme uses 24-bit RGB colors.
func NordTheme() *Stylesheet {
	s := NewStylesheet()

	// Nord palette (Polar Night, Snow Storm, Frost, Aurora)
	nord0 := RGB(46, 52, 64)    // polar night darkest
	nord1 := RGB(59, 66, 82)    // polar night dark
	nord2 := RGB(67, 76, 94)    // polar night mid
	nord3 := RGB(76, 86, 106)   // polar night light
	nord4 := RGB(216, 222, 233) // snow storm dark
	nord7 := RGB(143, 188, 187)  // frost teal
	nord8 := RGB(136, 192, 208)  // frost light blue
	nord11 := RGB(191, 97, 106)  // aurora red
	nord13 := RGB(235, 203, 139) // aurora yellow
	nord14 := RGB(163, 190, 140) // aurora green

	// -- Label / Text --
	s.Add(Select("Label"), Style{
		Foreground: nord4,
	})
	s.Add(Select("Text"), Style{
		Foreground: nord4,
	})

	// -- Button --
	s.Add(Select("Button"), Style{
		Foreground: nord0,
		Background: nord8,
		Bold:       Bool(true),
		Padding:    PadXY(2, 0),
	})
	s.Add(Select("Button").Pseudo(PseudoFocus), Style{
		Foreground: nord0,
		Background: nord7,
		Bold:       Bool(true),
	})
	s.Add(Select("Button").Class("primary"), Style{
		Foreground: nord0,
		Background: nord8,
	})
	s.Add(Select("Button").Class("danger"), Style{
		Foreground: nord4,
		Background: nord11,
	})
	s.Add(Select("Button").Pseudo(PseudoDisabled), Style{
		Foreground: nord3,
		Background: nord2,
		Dim:        Bool(true),
	})

	// -- Input --
	s.Add(Select("Input"), Style{
		Foreground: nord4,
		Background: nord1,
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: nord3, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Input").Pseudo(PseudoFocus), Style{
		Border: &Border{Style: BorderSingle, Color: nord8, StyleSet: true, ColorSet: true},
	})

	// -- Panel --
	s.Add(Select("Panel"), Style{
		Foreground: nord4,
		Background: nord0,
		Border:     &Border{Style: BorderSingle, Color: nord3, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Dialog --
	s.Add(Select("Dialog"), Style{
		Foreground: nord4,
		Background: nord1,
		Border:     &Border{Style: BorderRounded, Color: nord8, StyleSet: true, ColorSet: true},
		Padding:    Pad(1),
	})

	// -- Tabs --
	s.Add(Select("Tabs"), Style{
		Foreground: nord4,
		Background: nord0,
		Border:     &Border{Style: BorderSingle, Color: nord3, StyleSet: true, ColorSet: true},
	})

	// -- Alert --
	s.Add(Select("Alert"), Style{
		Foreground: nord4,
		Bold:       Bool(true),
		Padding:    PadXY(1, 0),
		Border:     &Border{Style: BorderSingle, Color: nord4, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("info"), Style{
		Foreground: nord8,
		Border:     &Border{Style: BorderSingle, Color: nord8, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("warning"), Style{
		Foreground: nord13,
		Border:     &Border{Style: BorderSingle, Color: nord13, StyleSet: true, ColorSet: true},
	})
	s.Add(Select("Alert").Class("error"), Style{
		Foreground: nord11,
		Border:     &Border{Style: BorderSingle, Color: nord11, StyleSet: true, ColorSet: true},
	})

	// -- Table --
	s.Add(Select("Table"), Style{
		Foreground: nord4,
		Background: nord0,
		Border:     &Border{Style: BorderSingle, Color: nord3, StyleSet: true, ColorSet: true},
	})

	// -- Progress --
	s.Add(Select("Progress"), Style{
		Foreground: nord14,
		Background: nord2,
	})

	s.SetVariable("theme-name", "nord")
	s.SetVariable("theme-variant", "dark")

	return s
}
