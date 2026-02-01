package video

import (
	"image"
	"image/color"
	"testing"
	"time"
)

func TestParseProbeInfo(t *testing.T) {
	raw := `{
  "streams": [
    {
      "codec_type": "video",
      "codec_name": "h264",
      "width": 1920,
      "height": 1080,
      "avg_frame_rate": "30000/1001",
      "duration": "12.5"
    }
  ],
  "format": {
    "duration": "12.5"
  }
}`
	info, err := parseProbeInfo(raw)
	if err != nil {
		t.Fatalf("parseProbeInfo error: %v", err)
	}
	if info.Width != 1920 || info.Height != 1080 {
		t.Fatalf("dimensions = %dx%d, want 1920x1080", info.Width, info.Height)
	}
	if info.Codec != "h264" {
		t.Fatalf("codec = %q, want h264", info.Codec)
	}
	if info.FrameRate < 29.9 || info.FrameRate > 30.1 {
		t.Fatalf("frame rate = %f, want ~29.97", info.FrameRate)
	}
	if info.Duration != 12*time.Second+500*time.Millisecond {
		t.Fatalf("duration = %s, want 12.5s", info.Duration)
	}
}

func TestDecodeRGB24(t *testing.T) {
	decoder := &Decoder{
		info: VideoInfo{Width: 2, Height: 1},
	}
	data := []byte{
		255, 0, 0,
		0, 255, 0,
	}
	img, err := decoder.decodeRGB24(data)
	if err != nil {
		t.Fatalf("decodeRGB24 error: %v", err)
	}
	rgba, ok := img.(*image.RGBA)
	if !ok {
		t.Fatalf("expected *image.RGBA, got %T", img)
	}
	if got := rgba.RGBAAt(0, 0); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("pixel(0,0) = %#v, want red", got)
	}
	if got := rgba.RGBAAt(1, 0); got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("pixel(1,0) = %#v, want green", got)
	}
}

func TestParseFrameRate(t *testing.T) {
	if rate := parseFrameRate("24000/1001"); rate < 23.9 || rate > 24.1 {
		t.Fatalf("rate = %f, want ~23.98", rate)
	}
	if rate := parseFrameRate("30"); rate != 30 {
		t.Fatalf("rate = %f, want 30", rate)
	}
	if rate := parseFrameRate(""); rate != 0 {
		t.Fatalf("rate = %f, want 0", rate)
	}
}

func TestNewDecoderEmptyPath(t *testing.T) {
	if _, err := NewDecoder(" "); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

func TestDecoderInfoPathNil(t *testing.T) {
	var d *Decoder
	if d.Info() != (VideoInfo{}) {
		t.Fatalf("expected zero info for nil decoder")
	}
	if d.Path() != "" {
		t.Fatalf("expected empty path for nil decoder")
	}
}

func TestDecodeRGB24Errors(t *testing.T) {
	var d *Decoder
	if _, err := d.decodeRGB24([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for nil decoder")
	}

	d = &Decoder{info: VideoInfo{Width: 0, Height: 1}}
	if _, err := d.decodeRGB24([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for invalid dimensions")
	}

	d = &Decoder{info: VideoInfo{Width: 2, Height: 2}}
	if _, err := d.decodeRGB24([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for short data")
	}
}

func TestParseProbeInfoNoVideoStream(t *testing.T) {
	raw := `{"streams":[{"codec_type":"audio","duration":"1.0"}],"format":{"duration":"1.0"}}`
	if _, err := parseProbeInfo(raw); err == nil {
		t.Fatalf("expected error for missing video stream")
	}
}

func TestParseFrameRateInvalid(t *testing.T) {
	if rate := parseFrameRate("0/0"); rate != 0 {
		t.Fatalf("rate = %f, want 0", rate)
	}
	if rate := parseFrameRate("30/0"); rate != 0 {
		t.Fatalf("rate = %f, want 0", rate)
	}
	if rate := parseFrameRate("abc"); rate != 0 {
		t.Fatalf("rate = %f, want 0", rate)
	}
}

func TestParseDurationInvalid(t *testing.T) {
	if parseDuration("") != 0 {
		t.Fatalf("expected zero duration for empty string")
	}
	if parseDuration("oops") != 0 {
		t.Fatalf("expected zero duration for invalid value")
	}
}

func TestExtractFramesErrors(t *testing.T) {
	var d *Decoder
	if _, err := d.ExtractFramesContext(nil, 24); err == nil {
		t.Fatalf("expected error for nil decoder")
	}

	d = &Decoder{path: "", info: VideoInfo{FrameRate: 24}, frameSize: 12}
	if _, err := d.ExtractFramesContext(nil, 24); err == nil {
		t.Fatalf("expected error for empty path")
	}

	d = &Decoder{path: "video.mp4", info: VideoInfo{FrameRate: 24}, frameSize: 0}
	if _, err := d.ExtractFramesContext(nil, 24); err == nil {
		t.Fatalf("expected error for invalid frame size")
	}
}

func TestExtractFrameErrors(t *testing.T) {
	var d *Decoder
	if _, err := d.ExtractFrame(0); err == nil {
		t.Fatalf("expected error for nil decoder")
	}

	d = &Decoder{path: ""}
	if _, err := d.ExtractFrame(0); err == nil {
		t.Fatalf("expected error for empty path")
	}
}
