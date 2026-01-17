# Recording Output

FluffyUI can record rendered output to efficient, replayable formats.
See `examples/recording` for a headless recording demo.

## Asciicast (Recommended)

Asciicast v2 is compact, streamable, and widely supported by tools like
asciinema and agg.

```go
recorder, err := recording.NewAsciicastRecorder("session.cast", recording.AsciicastOptions{
    Title: "FluffyUI Demo",
})
if err != nil {
    return err
}

app := runtime.NewApp(runtime.AppConfig{
    Backend:  backend,
    Recorder: recorder,
})
```

To reduce storage, use a gzip suffix:

```
session.cast.gz
```

## Export to Video (Optional)

If you have `agg` installed, you can render the cast file to a video format:

```go
err := recording.ExportWithAgg("session.cast", "session.webm", recording.AggOptions{
    Theme:    "monokai",
    FontSize: 16,
    FPS:      30,
})
```

`agg` decides the output format by file extension. Use `.webm` for efficient
video output; `.mp4` may work depending on your agg build.

## Export to MP4 (Agg + FFmpeg)

To guarantee MP4 output, render with `agg` and transcode with `ffmpeg`:

```go
err := recording.ExportVideo("session.cast", "session.mp4", recording.VideoOptions{
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
})
```

`ExportVideo` uses a temporary `.webm` render from `agg` and transcodes to `.mp4`.
Install both `agg` and `ffmpeg` for this workflow.

## VideoRecorder (One-Step Export)

`VideoRecorder` wraps asciicast recording and exports on close:

```go
recorder, err := recording.NewVideoRecorder("session.mp4", recording.VideoRecorderOptions{
    Cast: recording.AsciicastOptions{Title: "FluffyUI Demo"},
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
if err != nil {
    return err
}

app := runtime.NewApp(runtime.AppConfig{
    Backend:  backend,
    Recorder: recorder,
})
```

Set `KeepCast` to retain the intermediate `.cast` file for debugging or reuse.
