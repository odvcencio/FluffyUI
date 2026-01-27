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
