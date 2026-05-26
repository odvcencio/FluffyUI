package html

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/compositor"
	"m31labs.dev/fluffyui/style"
)

// StylesheetToCSS converts an FSS stylesheet to CSS.
func StylesheetToCSS(sheet *style.Stylesheet) string {
	if sheet == nil {
		return ""
	}
	var b strings.Builder
	for _, rule := range sheet.Rules() {
		sel := selectorToCSS(rule.Selector)
		props := styleToCSS(rule.Style)
		if sel != "" && props != "" {
			fmt.Fprintf(&b, "%s { %s }\n", sel, props)
		}
	}
	return b.String()
}

func selectorToCSS(sel style.Selector) string {
	var chain []string
	current := &sel
	for current != nil {
		part := selectorPartToCSS(*current)
		if part != "" {
			chain = append([]string{part}, chain...)
		}
		current = current.Parent
	}
	return strings.Join(chain, " ")
}

func selectorPartToCSS(sel style.Selector) string {
	if sel.Type == "*" && sel.ID == "" && len(sel.Classes) == 0 {
		var b strings.Builder
		b.WriteString("*")
		for _, p := range sel.Pseudo {
			b.WriteString(":" + string(p))
		}
		return b.String()
	}
	var b strings.Builder
	if sel.Type != "" && sel.Type != "*" {
		b.WriteString(".fluffy-" + sel.Type)
	}
	if sel.ID != "" {
		b.WriteString("#" + sel.ID)
	}
	for _, cls := range sel.Classes {
		b.WriteString("." + cls)
	}
	for _, p := range sel.Pseudo {
		b.WriteString(":" + string(p))
	}
	return b.String()
}

func colorToCSS(c compositor.Color) string {
	if c.Mode != compositor.ColorModeRGB {
		return ""
	}
	r := uint8((c.Value >> 16) & 0xff)
	g := uint8((c.Value >> 8) & 0xff)
	bl := uint8(c.Value & 0xff)
	return fmt.Sprintf("#%02x%02x%02x", r, g, bl)
}

func spacingToCSS(s *style.Spacing, prop string) string {
	if s == nil {
		return ""
	}
	if s.Top == s.Bottom && s.Left == s.Right {
		if s.Top == s.Left {
			return fmt.Sprintf("%s: %dch", prop, s.Top)
		}
		return fmt.Sprintf("%s: %dch %dch", prop, s.Top, s.Right)
	}
	return fmt.Sprintf("%s: %dch %dch %dch %dch", prop, s.Top, s.Right, s.Bottom, s.Left)
}

func styleToCSS(s style.Style) string {
	var props []string

	if fg := colorToCSS(s.Foreground); fg != "" {
		props = append(props, "color: "+fg)
	}
	if bg := colorToCSS(s.Background); bg != "" {
		props = append(props, "background-color: "+bg)
	}
	if s.Bold != nil && *s.Bold {
		props = append(props, "font-weight: bold")
	}
	if s.Italic != nil && *s.Italic {
		props = append(props, "font-style: italic")
	}
	if s.Underline != nil && *s.Underline {
		props = append(props, "text-decoration: underline")
	}
	if s.Dim != nil && *s.Dim {
		props = append(props, "opacity: 0.5")
	}
	if p := spacingToCSS(s.Padding, "padding"); p != "" {
		props = append(props, p)
	}
	if m := spacingToCSS(s.Margin, "margin"); m != "" {
		props = append(props, m)
	}
	if s.Border != nil && s.Border.ColorSet {
		if bc := colorToCSS(s.Border.Color); bc != "" {
			props = append(props, "border-color: "+bc)
		}
	}
	switch s.TextAlign {
	case style.TextAlignCenter:
		props = append(props, "text-align: center")
	case style.TextAlignRight:
		props = append(props, "text-align: right")
	case style.TextAlignLeft:
		props = append(props, "text-align: left")
	}
	if s.Display == style.DisplayHidden {
		props = append(props, "display: none")
	}
	switch s.Overflow {
	case style.OverflowHidden:
		props = append(props, "overflow: hidden")
	case style.OverflowScroll:
		props = append(props, "overflow: auto")
	}

	return strings.Join(props, "; ")
}
