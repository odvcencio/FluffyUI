// Package gosxscene displays typed GoSX Scene3D frames in a FluffyUI terminal UI.
package gosxscene

import (
	"fmt"

	"m31labs.dev/fluffyui/widgets"
	"m31labs.dev/gosx/render/bundle"
	"m31labs.dev/gosx/scene"
	"m31labs.dev/gosx/scene/preview"
)

const (
	defaultWidth  = 640
	defaultHeight = 360
)

// Options controls the initial rasterized Scene3D frame and terminal image.
// Preview options retain their normal GoSX meaning; Width and Height default
// from Props before falling back to a compact terminal-friendly frame.
type Options struct {
	Preview  preview.Options
	Fit      widgets.ImageFit
	Protocol widgets.ImageProtocol
	Label    string
}

// Scene is a FluffyUI widget backed by a GoSX Scene3D render. Its embedded
// Image satisfies runtime.Widget, so it can be used directly in a widget tree.
type Scene struct {
	*widgets.Image

	props   scene.Props
	options Options
	result  *preview.Result
}

// New lowers props through GoSX's native RenderBundle and exposes the resulting
// frame as a terminal image widget.
func New(props scene.Props, options Options) (*Scene, error) {
	view := &Scene{props: props, options: options}
	if err := view.RenderFrame(options.Preview.Time); err != nil {
		return nil, err
	}
	return view, nil
}

// RenderFrame updates the terminal image with a frame rendered at time.
// Call it from the application loop when a scene animation advances.
func (s *Scene) RenderFrame(time float64) error {
	if s == nil {
		return fmt.Errorf("gosxscene: nil scene")
	}

	opts := s.previewOptions()
	opts.Time = time
	result, err := preview.Render(s.props, opts)
	if err != nil {
		return fmt.Errorf("gosxscene: render scene: %w", err)
	}
	if s.Image == nil {
		s.Image = widgets.NewImage(result.Image)
		s.Image.SetFit(s.options.Fit)
		s.Image.SetProtocol(s.options.Protocol)
		label := s.options.Label
		if label == "" {
			label = s.props.AriaLabel
		}
		if label == "" {
			label = s.props.Label
		}
		if label != "" {
			s.Image.SetLabel(label)
		}
	} else {
		s.Image.SetImage(result.Image)
	}
	s.result = result
	s.Image.Invalidate()
	return nil
}

// Result returns the native GoSX renderer result for the latest frame.
func (s *Scene) Result() *preview.Result {
	if s == nil {
		return nil
	}
	return s.result
}

// Stats returns the renderer measurements for the latest Scene3D frame.
func (s *Scene) Stats() bundle.FrameStats {
	if s == nil || s.result == nil {
		return bundle.FrameStats{}
	}
	return s.result.Stats
}

func (s *Scene) previewOptions() preview.Options {
	opts := s.options.Preview
	if opts.Width <= 0 {
		opts.Width = s.props.Width
	}
	if opts.Height <= 0 {
		opts.Height = s.props.Height
	}
	if opts.Width <= 0 {
		opts.Width = defaultWidth
	}
	if opts.Height <= 0 {
		opts.Height = defaultHeight
	}
	return opts
}
