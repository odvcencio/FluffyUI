// Package video provides basic video decoding helpers backed by ffmpeg.
package video

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// VideoInfo describes the core characteristics of a video stream.
type VideoInfo struct {
	Width     int
	Height    int
	Duration  time.Duration
	FrameRate float64
	Codec     string
}

// Decoder extracts frames from a video file using ffmpeg.
type Decoder struct {
	path      string
	info      VideoInfo
	frameSize int
}

// NewDecoder creates a Decoder for the given video file.
func NewDecoder(path string) (*Decoder, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("video path is required")
	}
	info, err := probe(path)
	if err != nil {
		return nil, err
	}
	if info.Width <= 0 || info.Height <= 0 {
		return nil, fmt.Errorf("invalid video dimensions %dx%d", info.Width, info.Height)
	}
	frameSize := info.Width * info.Height * 3
	if frameSize <= 0 {
		return nil, errors.New("invalid frame size")
	}
	return &Decoder{
		path:      path,
		info:      info,
		frameSize: frameSize,
	}, nil
}

// Info returns the decoded video metadata.
func (d *Decoder) Info() VideoInfo {
	if d == nil {
		return VideoInfo{}
	}
	return d.info
}

// Path returns the video file path.
func (d *Decoder) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// ExtractFrame decodes a single frame at the specified timestamp.
func (d *Decoder) ExtractFrame(at time.Duration) (image.Image, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if d.path == "" {
		return nil, errors.New("video path is required")
	}
	var buf bytes.Buffer
	err := ffmpeg.Input(d.path, ffmpeg.KwArgs{"ss": fmt.Sprintf("%.3f", at.Seconds())}).
		Output("pipe:", ffmpeg.KwArgs{
			"vframes": 1,
			"f":       "rawvideo",
			"pix_fmt": "rgb24",
		}).
		WithOutput(&buf, nil).
		Run()
	if err != nil {
		return nil, err
	}
	data := buf.Bytes()
	if len(data) < d.frameSize {
		return nil, fmt.Errorf("short frame: got %d bytes, want %d", len(data), d.frameSize)
	}
	return d.decodeRGB24(data[:d.frameSize])
}

// ExtractFrames streams frames at the requested fps. Frames may be dropped if the receiver is slow.
func (d *Decoder) ExtractFrames(fps float64) (<-chan image.Image, error) {
	return d.ExtractFramesContext(context.Background(), fps)
}

// ExtractFramesContext streams frames with cancellation support.
func (d *Decoder) ExtractFramesContext(ctx context.Context, fps float64) (<-chan image.Image, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if d.path == "" {
		return nil, errors.New("video path is required")
	}
	if fps <= 0 {
		fps = d.info.FrameRate
	}
	if fps <= 0 {
		fps = 30
	}
	if d.frameSize <= 0 {
		return nil, errors.New("invalid frame size")
	}

	pipeReader, pipeWriter := io.Pipe()
	frames := make(chan image.Image, 12)

	go func() {
		defer pipeReader.Close()
		defer close(frames)

		buf := make([]byte, d.frameSize)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, err := io.ReadFull(pipeReader, buf)
			if err != nil {
				return
			}
			frame, err := d.decodeRGB24(buf)
			if err != nil {
				return
			}
			select {
			case frames <- frame:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		var stderr bytes.Buffer
		err := ffmpeg.Input(d.path).
			Output("pipe:", ffmpeg.KwArgs{
				"f":       "rawvideo",
				"pix_fmt": "rgb24",
				"r":       fmt.Sprintf("%.3f", fps),
			}).
			WithOutput(pipeWriter, &stderr).
			Run()
		_ = pipeWriter.CloseWithError(err)
	}()

	return frames, nil
}

func (d *Decoder) decodeRGB24(data []byte) (image.Image, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if d.info.Width <= 0 || d.info.Height <= 0 {
		return nil, errors.New("invalid video dimensions")
	}
	expected := d.info.Width * d.info.Height * 3
	if len(data) < expected {
		return nil, fmt.Errorf("short frame data: got %d bytes, want %d", len(data), expected)
	}
	img := image.NewRGBA(image.Rect(0, 0, d.info.Width, d.info.Height))
	pixels := d.info.Width * d.info.Height
	for i := 0; i < pixels; i++ {
		src := i * 3
		dst := i * 4
		img.Pix[dst] = data[src]
		img.Pix[dst+1] = data[src+1]
		img.Pix[dst+2] = data[src+2]
		img.Pix[dst+3] = 0xff
	}
	return img, nil
}

type probeOutput struct {
	Streams []probeStream `json:"streams"`
	Format  probeFormat   `json:"format"`
}

type probeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
	RFrameRate   string `json:"r_frame_rate"`
	Duration     string `json:"duration"`
}

type probeFormat struct {
	Duration string `json:"duration"`
}

func probe(path string) (VideoInfo, error) {
	raw, err := ffmpeg.Probe(path)
	if err != nil {
		return VideoInfo{}, err
	}
	return parseProbeInfo(raw)
}

func parseProbeInfo(raw string) (VideoInfo, error) {
	var output probeOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return VideoInfo{}, fmt.Errorf("parse ffprobe output: %w", err)
	}
	for _, stream := range output.Streams {
		if stream.CodecType != "video" {
			continue
		}
		info := VideoInfo{
			Width:  stream.Width,
			Height: stream.Height,
			Codec:  stream.CodecName,
		}
		info.FrameRate = parseFrameRate(stream.AvgFrameRate)
		if info.FrameRate == 0 {
			info.FrameRate = parseFrameRate(stream.RFrameRate)
		}
		info.Duration = parseDuration(stream.Duration)
		if info.Duration == 0 {
			info.Duration = parseDuration(output.Format.Duration)
		}
		return info, nil
	}
	return VideoInfo{}, errors.New("no video stream found")
}

func parseFrameRate(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "0/0" {
		return 0
	}
	if strings.Contains(value, "/") {
		parts := strings.SplitN(value, "/", 2)
		num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0
		}
		den, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil || den == 0 {
			return 0
		}
		return num / den
	}
	rate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return rate
}

func parseDuration(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
