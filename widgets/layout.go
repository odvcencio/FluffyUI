package widgets

import (
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
)

type layoutMetrics struct {
	margin  style.Spacing
	padding style.Spacing
	border  int
}

func layoutMetricsFromStyle(s style.Style) layoutMetrics {
	return layoutMetrics{
		margin:  clampSpacing(spacingOrZero(s.Margin)),
		padding: clampSpacing(spacingOrZero(s.Padding)),
		border:  borderWidth(s.Border),
	}
}

func spacingOrZero(spacing *style.Spacing) style.Spacing {
	if spacing == nil {
		return style.Spacing{}
	}
	return *spacing
}

func clampSpacing(spacing style.Spacing) style.Spacing {
	return style.Spacing{
		Top:    max(0, spacing.Top),
		Right:  max(0, spacing.Right),
		Bottom: max(0, spacing.Bottom),
		Left:   max(0, spacing.Left),
	}
}

func borderWidth(border *style.Border) int {
	if border == nil || border.Style == style.BorderNone {
		return 0
	}
	return 1
}

func (m layoutMetrics) marginInsets() (top, right, bottom, left int) {
	return m.margin.Top, m.margin.Right, m.margin.Bottom, m.margin.Left
}

func (m layoutMetrics) contentInsets() (top, right, bottom, left int) {
	return m.padding.Top + m.border,
		m.padding.Right + m.border,
		m.padding.Bottom + m.border,
		m.padding.Left + m.border
}

const layoutMaxInt = int(^uint(0) >> 1)

func (b *Base) measureWithStyle(
	constraints runtime.Constraints,
	measureContent func(runtime.Constraints) runtime.Size,
) runtime.Size {
	if b == nil {
		if measureContent == nil {
			return constraints.MinSize()
		}
		return constraints.Constrain(measureContent(constraints))
	}

	metrics := b.layoutMetrics
	marginTop, marginRight, marginBottom, marginLeft := metrics.marginInsets()
	borderConstraints := shrinkConstraints(constraints, marginTop, marginRight, marginBottom, marginLeft)

	contentTop, contentRight, contentBottom, contentLeft := metrics.contentInsets()
	contentConstraints := shrinkConstraints(borderConstraints, contentTop, contentRight, contentBottom, contentLeft)

	contentSize := runtime.Size{}
	if measureContent != nil {
		contentSize = contentConstraints.Constrain(measureContent(contentConstraints))
	}

	borderSize := runtime.Size{
		Width:  contentSize.Width + contentLeft + contentRight,
		Height: contentSize.Height + contentTop + contentBottom,
	}
	borderSize.Width = resolveStyledSize(b.layoutStyle.Width, borderSize.Width, borderConstraints.MinWidth, borderConstraints.MaxWidth)
	borderSize.Height = resolveStyledSize(b.layoutStyle.Height, borderSize.Height, borderConstraints.MinHeight, borderConstraints.MaxHeight)

	outer := runtime.Size{
		Width:  borderSize.Width + marginLeft + marginRight,
		Height: borderSize.Height + marginTop + marginBottom,
	}
	return constraints.Constrain(outer)
}

func shrinkConstraints(c runtime.Constraints, top, right, bottom, left int) runtime.Constraints {
	return runtime.Constraints{
		MinWidth:  max(0, c.MinWidth-left-right),
		MaxWidth:  max(0, c.MaxWidth-left-right),
		MinHeight: max(0, c.MinHeight-top-bottom),
		MaxHeight: max(0, c.MaxHeight-top-bottom),
	}
}

func resolveStyledSize(spec *style.Size, auto int, minValue int, maxValue int) int {
	if spec == nil || spec.Mode == style.SizeAuto {
		return clampInt(auto, minValue, maxValue)
	}

	switch spec.Mode {
	case style.SizeFixed:
		return clampInt(spec.Value, minValue, maxValue)
	case style.SizePercent:
		if maxValue == layoutMaxInt {
			return clampInt(auto, minValue, maxValue)
		}
		return clampInt(maxValue*spec.Value/100, minValue, maxValue)
	case style.SizeFill:
		if maxValue == layoutMaxInt {
			return clampInt(auto, minValue, maxValue)
		}
		return clampInt(maxValue, minValue, maxValue)
	default:
		return clampInt(auto, minValue, maxValue)
	}
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
