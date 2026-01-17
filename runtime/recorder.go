package runtime

import "time"

// Recorder captures rendered frames for playback/export.
type Recorder interface {
	Start(width, height int, now time.Time) error
	Resize(width, height int) error
	Frame(buffer *Buffer, now time.Time) error
	Close() error
}
