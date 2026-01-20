package audio

import (
	"testing"
	"time"
)

type testDriver struct {
	plays []Cue
	stops []Kind
}

func (d *testDriver) Play(cue Cue) error {
	d.plays = append(d.plays, cue)
	return nil
}

func (d *testDriver) Stop(kind Kind) error {
	d.stops = append(d.stops, kind)
	return nil
}

func TestManagerPlayAppliesVolumeAndCooldown(t *testing.T) {
	driver := &testDriver{}
	manager := NewManager(driver, Cue{
		ID:       "click",
		Kind:     KindSFX,
		Volume:   100,
		Cooldown: time.Second,
	})
	manager.SetMasterVolume(80)
	manager.SetSFXVolume(50)
	manager.clockNow = func() time.Time { return time.Unix(0, 0) }

	if !manager.PlaySFX("click") {
		t.Fatal("expected first play to succeed")
	}
	if len(driver.plays) != 1 {
		t.Fatalf("expected 1 play, got %d", len(driver.plays))
	}
	if driver.plays[0].Volume != 40 {
		t.Fatalf("expected volume 40, got %d", driver.plays[0].Volume)
	}
	if manager.PlaySFX("click") {
		t.Fatal("expected cooldown to block replay")
	}

	manager.clockNow = func() time.Time { return time.Unix(2, 0) }
	if !manager.PlaySFX("click") {
		t.Fatal("expected play after cooldown")
	}
	if len(driver.plays) != 2 {
		t.Fatalf("expected 2 plays, got %d", len(driver.plays))
	}
}

func TestManagerPlayMusicStopsPreviousTrack(t *testing.T) {
	driver := &testDriver{}
	manager := NewManager(driver,
		Cue{ID: "track-a", Kind: KindMusic},
		Cue{ID: "track-b", Kind: KindMusic},
	)

	if !manager.PlayMusic("track-a") {
		t.Fatal("expected first track to play")
	}
	if len(driver.stops) != 0 {
		t.Fatalf("expected 0 stops on first play, got %d", len(driver.stops))
	}
	if !manager.PlayMusic("track-b") {
		t.Fatal("expected second track to play")
	}
	if len(driver.stops) != 1 || driver.stops[0] != KindMusic {
		t.Fatalf("expected 1 music stop, got %#v", driver.stops)
	}
	if manager.PlayMusic("track-b") {
		t.Fatal("expected same track to be ignored")
	}
	if len(driver.plays) != 2 {
		t.Fatalf("expected 2 plays, got %d", len(driver.plays))
	}
}
