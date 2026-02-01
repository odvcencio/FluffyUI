package audio

import "testing"

func TestNormalizeCueAndClamp(t *testing.T) {
	cue := normalizeCue(Cue{ID: "x", Volume: 0, Cooldown: -1})
	if cue.Volume != DefaultVolume {
		t.Fatalf("volume = %d, want %d", cue.Volume, DefaultVolume)
	}
	if cue.Cooldown != 0 {
		t.Fatalf("cooldown = %v, want 0", cue.Cooldown)
	}
	if clampPercent(-5) != 0 || clampPercent(200) != 100 {
		t.Fatalf("clampPercent bounds failed")
	}
	if applyPercent(10, 0) != 0 || applyPercent(0, 50) != 0 {
		t.Fatalf("applyPercent should clamp to 0")
	}
	if applyPercent(50, 50) != 25 {
		t.Fatalf("applyPercent = %d, want 25", applyPercent(50, 50))
	}
}

func TestManagerMutedStopsMusic(t *testing.T) {
	driver := &testDriver{}
	manager := NewManager(driver, Cue{ID: "track", Kind: KindMusic})
	if !manager.PlayMusic("track") {
		t.Fatalf("expected music to play")
	}
	manager.SetMuted(true)
	if !manager.Muted() {
		t.Fatalf("expected manager muted")
	}
	if len(driver.stops) != 1 || driver.stops[0] != KindMusic {
		t.Fatalf("expected music stop on mute, got %#v", driver.stops)
	}
	if manager.PlayMusic("track") {
		t.Fatalf("expected muted manager to block playback")
	}
}

func TestManagerStopMusic(t *testing.T) {
	driver := &testDriver{}
	manager := NewManager(driver)
	if manager.StopMusic() {
		t.Fatalf("expected StopMusic false without current track")
	}
	manager.Register(Cue{ID: "song", Kind: KindMusic})
	manager.PlayMusic("song")
	if !manager.StopMusic() {
		t.Fatalf("expected StopMusic true when track active")
	}
}
