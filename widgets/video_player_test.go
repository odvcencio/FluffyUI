package widgets

import (
	"image"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestVideoPlayerSeek(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 500 * time.Millisecond,
		frames:        make([]image.Image, 4),
	}
	player.Seek(1200 * time.Millisecond)
	if player.currentFrame != 2 {
		t.Fatalf("currentFrame = %d, want 2", player.currentFrame)
	}
	player.Seek(-time.Second)
	if player.currentFrame != 0 {
		t.Fatalf("currentFrame = %d, want 0", player.currentFrame)
	}
}

func TestVideoPlayerTickAdvancesFrames(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 500 * time.Millisecond,
		frames:        make([]image.Image, 5),
		playing:       true,
	}
	start := time.Now()
	player.HandleMessage(runtime.TickMsg{Time: start})
	player.HandleMessage(runtime.TickMsg{Time: start.Add(1200 * time.Millisecond)})
	if player.currentFrame != 2 {
		t.Fatalf("currentFrame = %d, want 2", player.currentFrame)
	}
}

func TestVideoPlayerOnEnd(t *testing.T) {
	called := false
	player := &VideoPlayer{
		frameDuration: time.Second,
		frames:        make([]image.Image, 2),
		framesDone:    true,
		playing:       true,
		onEnd: func() {
			called = true
		},
	}
	start := time.Now()
	player.HandleMessage(runtime.TickMsg{Time: start})
	player.HandleMessage(runtime.TickMsg{Time: start.Add(3 * time.Second)})
	if player.playing {
		t.Fatalf("playing = true, want false")
	}
	if !called {
		t.Fatalf("onEnd callback was not called")
	}
	if player.currentFrame != 1 {
		t.Fatalf("currentFrame = %d, want 1", player.currentFrame)
	}
}

func TestVideoPlayerTogglePlay(t *testing.T) {
	player := &VideoPlayer{}
	player.HandleMessage(runtime.KeyMsg{Rune: ' '})
	if !player.playing {
		t.Fatalf("playing = false, want true")
	}
	player.HandleMessage(runtime.KeyMsg{Rune: ' '})
	if player.playing {
		t.Fatalf("playing = true, want false")
	}
}
