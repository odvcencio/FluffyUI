package widgets

import (
	"image"
	"math"
	"sync"
	"time"

	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/video"
)

// VideoPlayer renders video frames onto a canvas.
type VideoPlayer struct {
	Component

	decoder       *video.Decoder
	blitter       graphics.Blitter
	canvas        *graphics.Canvas
	cellWidth     int
	cellHeight    int
	frameRate     float64
	frameDuration time.Duration

	frames        []image.Image
	framesDone    bool
	framesDropped int64
	framesMu      sync.RWMutex

	playing      bool
	playhead     time.Duration
	lastTick     time.Time
	currentFrame int
	onEnd        func()
}

// NewVideoPlayer creates a player and starts decoding frames.
func NewVideoPlayer(path string) (*VideoPlayer, error) {
	decoder, err := video.NewDecoder(path)
	if err != nil {
		return nil, err
	}
	player := &VideoPlayer{
		decoder: decoder,
		blitter: graphics.BestBlitter(nil),
	}
	player.initTiming(decoder.Info())
	if err := player.startFrameLoader(); err != nil {
		return nil, err
	}
	return player, nil
}

// WithBlitter configures the blitter used for rendering frames.
func (v *VideoPlayer) WithBlitter(blitter graphics.Blitter) *VideoPlayer {
	if v == nil || blitter == nil {
		return v
	}
	v.blitter = blitter
	v.canvas = nil
	return v
}

// SetOnEnd registers a callback for when playback completes.
func (v *VideoPlayer) SetOnEnd(fn func()) {
	if v == nil {
		return
	}
	v.onEnd = fn
}

// Play starts playback.
func (v *VideoPlayer) Play() {
	if v == nil {
		return
	}
	v.playing = true
	v.lastTick = time.Time{}
}

// Pause stops playback.
func (v *VideoPlayer) Pause() {
	if v == nil {
		return
	}
	v.playing = false
}

// Seek moves the playhead to the given position.
func (v *VideoPlayer) Seek(pos time.Duration) {
	if v == nil {
		return
	}
	if pos < 0 {
		pos = 0
	}
	v.playhead = pos
	v.lastTick = time.Time{}
	v.currentFrame = v.frameIndexFor(pos)
}

// IsPlaying reports whether the player is currently playing.
func (v *VideoPlayer) IsPlaying() bool {
	if v == nil {
		return false
	}
	return v.playing
}

// DroppedFrames returns the count of frames dropped during loading.
func (v *VideoPlayer) DroppedFrames() int64 {
	if v == nil {
		return 0
	}
	v.framesMu.RLock()
	defer v.framesMu.RUnlock()
	return v.framesDropped
}

// StyleType returns the selector type name.
func (v *VideoPlayer) StyleType() string {
	return "VideoPlayer"
}

// Measure returns the desired size for the video player.
func (v *VideoPlayer) Measure(constraints runtime.Constraints) runtime.Size {
	return v.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := contentConstraints.MaxWidth
		height := contentConstraints.MaxHeight
		if width == 0 {
			width = contentConstraints.MinWidth
		}
		if height == 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout updates layout bounds and canvas size.
func (v *VideoPlayer) Layout(bounds runtime.Rect) {
	v.Component.Layout(bounds)
	content := v.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		v.canvas = nil
		v.cellWidth = 0
		v.cellHeight = 0
		return
	}
	if v.canvas == nil || content.Width != v.cellWidth || content.Height != v.cellHeight {
		v.canvas = graphics.NewCanvasWithBlitter(content.Width, content.Height, v.blitter)
		v.cellWidth = content.Width
		v.cellHeight = content.Height
	}
}

// Render draws the current video frame.
func (v *VideoPlayer) Render(ctx runtime.RenderContext) {
	if v == nil || v.canvas == nil {
		return
	}
	content := v.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	frame := v.currentFrameImage()
	if frame == nil {
		return
	}
	v.canvas.Clear()
	v.drawFrame(frame)
	v.canvas.Render(ctx.Buffer, content.X, content.Y)
}

// HandleMessage advances playback on ticks and toggles play on spacebar.
func (v *VideoPlayer) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if v == nil {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.TickMsg:
		if !v.playing || v.frameDuration <= 0 {
			return runtime.Unhandled()
		}
		if v.lastTick.IsZero() {
			v.lastTick = m.Time
			return runtime.Handled()
		}
		if m.Time.After(v.lastTick) {
			v.playhead += m.Time.Sub(v.lastTick)
		}
		v.lastTick = m.Time
		v.advanceFrame()
		return runtime.Handled()
	case runtime.KeyMsg:
		if m.Rune == ' ' {
			if v.playing {
				v.Pause()
			} else {
				v.Play()
			}
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (v *VideoPlayer) initTiming(info video.VideoInfo) {
	v.frameRate = info.FrameRate
	if v.frameRate <= 0 {
		v.frameRate = 30
	}
	v.frameDuration = time.Duration(float64(time.Second) / v.frameRate)
}

func (v *VideoPlayer) startFrameLoader() error {
	if v.decoder == nil {
		return nil
	}
	frames, err := v.decoder.ExtractFrames(v.frameRate)
	if err != nil {
		return err
	}
	go func() {
		for frame := range frames {
			v.framesMu.Lock()
			wasEmpty := len(v.frames) == 0
			v.frames = append(v.frames, frame)
			v.framesMu.Unlock()
			if wasEmpty {
				v.Invalidate()
			}
		}
		v.framesMu.Lock()
		v.framesDone = true
		v.framesMu.Unlock()
		v.Invalidate()
	}()
	return nil
}

func (v *VideoPlayer) advanceFrame() {
	if v == nil || v.frameDuration <= 0 {
		return
	}
	target := v.frameIndexFor(v.playhead)
	count, done := v.frameSnapshot()
	if count == 0 {
		return
	}
	if target >= count {
		v.currentFrame = count - 1
		if done && v.playing {
			v.playing = false
			if v.onEnd != nil {
				v.onEnd()
			}
		}
		return
	}
	v.currentFrame = target
}

func (v *VideoPlayer) frameSnapshot() (int, bool) {
	v.framesMu.RLock()
	defer v.framesMu.RUnlock()
	return len(v.frames), v.framesDone
}

func (v *VideoPlayer) frameIndexFor(pos time.Duration) int {
	if v.frameDuration <= 0 {
		return 0
	}
	index := int(pos / v.frameDuration)
	if index < 0 {
		return 0
	}
	return index
}

func (v *VideoPlayer) currentFrameImage() image.Image {
	v.framesMu.RLock()
	defer v.framesMu.RUnlock()
	if len(v.frames) == 0 {
		return nil
	}
	if v.currentFrame < 0 {
		return v.frames[0]
	}
	if v.currentFrame >= len(v.frames) {
		return v.frames[len(v.frames)-1]
	}
	return v.frames[v.currentFrame]
}

func (v *VideoPlayer) drawFrame(frame image.Image) {
	if v == nil || v.canvas == nil || frame == nil {
		return
	}
	canvasW, canvasH := v.canvas.Size()
	if canvasW <= 0 || canvasH <= 0 {
		return
	}
	bounds := frame.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}
	scale := math.Min(float64(canvasW)/float64(srcW), float64(canvasH)/float64(srcH))
	if scale <= 0 {
		return
	}
	targetW := int(math.Round(float64(srcW) * scale))
	targetH := int(math.Round(float64(srcH) * scale))
	if targetW <= 0 || targetH <= 0 {
		return
	}
	offsetX := (canvasW - targetW) / 2
	offsetY := (canvasH - targetH) / 2
	v.canvas.DrawImageScaled(offsetX, offsetY, targetW, targetH, frame)
}
