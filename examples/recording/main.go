package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/recording"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	output := flag.String("out", "recordings/demo.cast", "output file (.cast, .cast.gz, .webm, .mp4)")
	width := flag.Int("width", 72, "recording width")
	height := flag.Int("height", 18, "recording height")
	flag.Parse()

	if err := ensureDir(*output); err != nil {
		fmt.Fprintln(os.Stderr, "create output directory:", err)
		os.Exit(1)
	}

	recorder, err := buildRecorder(*output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recorder error:", err)
		os.Exit(1)
	}

	text := widgets.NewText("")
	panel := widgets.NewPanel(text).WithBorder(backend.DefaultStyle())
	panel.SetTitle("Recording Demo")

	totalFrames := 90
	frame := 0
	spinnerFrames := []string{"-", "\\", "|", "/"}

	update := func(app *runtime.App, msg runtime.Message) bool {
		switch msg.(type) {
		case runtime.TickMsg:
			frame++
			ratio := float64(frame) / float64(totalFrames)
			if ratio > 1 {
				ratio = 1
			}
			gauge := widgets.DrawGaugeString(30, ratio, widgets.GaugeStyle{})
			text.SetText(fmt.Sprintf(
				"FluffyUI recording\n%s %3.0f%% %s\nOutput: %s",
				gauge,
				ratio*100,
				spinnerFrames[frame%len(spinnerFrames)],
				filepath.Base(*output),
			))
			if frame >= totalFrames {
				app.ExecuteCommand(runtime.Quit{})
				return false
			}
			return true
		default:
			return runtime.DefaultUpdate(app, msg)
		}
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  sim.New(*width, *height),
		Root:     panel,
		Update:   update,
		TickRate: time.Second / 15,
		Recorder: recorder,
	})

	if err := app.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "run error:", err)
		os.Exit(1)
	}
}

func buildRecorder(path string) (runtime.Recorder, error) {
	title := recording.AsciicastOptions{Title: "FluffyUI Recording"}
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".cast") || strings.HasSuffix(lower, ".cast.gz") {
		return recording.NewAsciicastRecorder(path, title)
	}
	return recording.NewVideoRecorder(path, recording.VideoRecorderOptions{
		Cast: title,
		Video: recording.VideoOptions{
			Agg: recording.AggOptions{
				Theme:    "monokai",
				FontSize: 16,
				FPS:      30,
			},
			FFmpeg: recording.FFmpegOptions{
				VideoCodec: "libx264",
				Preset:     "medium",
				CRF:        22,
			},
		},
	})
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
