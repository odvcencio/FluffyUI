package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	mcp "m31labs.dev/fluffyui/agent/mcp"
)

type speakRunner struct {
	client  *mcp.Client
	queue   *SpeechQueue
	logger  *logWriter
	verbose bool
}

type logWriter struct {
	file *os.File
	enc  *json.Encoder
}

type logEvent struct {
	Time      time.Time      `json:"time"`
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	Priority  string         `json:"priority,omitempty"`
	FocusedID string         `json:"focused_id,omitempty"`
	Added     int            `json:"added,omitempty"`
	Removed   int            `json:"removed,omitempty"`
	Modified  int            `json:"modified,omitempty"`
	Snapshot  *mcp.Snapshot  `json:"snapshot,omitempty"`
	Error     string         `json:"error,omitempty"`
}

func newLogWriter(path string) (*logWriter, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return &logWriter{file: f, enc: enc}, nil
}

func (l *logWriter) write(event logEvent) {
	if l == nil || l.enc == nil {
		return
	}
	_ = l.enc.Encode(event)
}

func (l *logWriter) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (r *speakRunner) run(ctx context.Context) error {
	// Get initial snapshot
	snap, err := r.client.Snapshot()
	if err != nil {
		return fmt.Errorf("initial snapshot: %w", err)
	}
	r.log(logEvent{Time: time.Now(), Type: "connected"})

	// Announce the initially focused widget
	if snap.FocusedID != "" {
		for _, w := range snap.Widgets {
			if w.ID == snap.FocusedID {
				text := formatWidget(w)
				if text != "" {
					r.queue.Enqueue(Utterance{Text: text, Priority: "assertive"})
					r.log(logEvent{
						Time:     time.Now(),
						Type:     "focus",
						Text:     text,
						Priority: "assertive",
					})
				}
				break
			}
		}
	}

	// Subscribe to resource changes
	changed := make(chan struct{}, 1)
	notify := func(_ mcp.ResourceEvent) {
		select {
		case changed <- struct{}{}:
		default:
		}
	}
	_ = r.client.Subscribe("fluffy://focused", notify)
	_ = r.client.Subscribe("fluffy://widgets", notify)

	// Poll/event loop
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	prev := snap
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-changed:
		case <-ticker.C:
		}

		current, err := r.client.Snapshot()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if r.verbose {
				fmt.Fprintf(os.Stderr, "snapshot error: %v\n", err)
			}
			continue
		}

		diff := diffSnapshots(prev, current)
		if !hasMeaningfulChanges(diff) {
			prev = current
			continue
		}

		utterances := Announce(prev, current, diff)
		for _, u := range utterances {
			r.queue.Enqueue(u)
			r.log(logEvent{
				Time:      time.Now(),
				Type:      "announce",
				Text:      u.Text,
				Priority:  u.Priority,
				FocusedID: current.FocusedID,
				Added:     len(diff.WidgetsAdded),
				Removed:   len(diff.WidgetsRemoved),
				Modified:  len(diff.WidgetsModified),
			})
		}

		prev = current
	}
}

func (r *speakRunner) log(event logEvent) {
	if r.logger != nil {
		r.logger.write(event)
	}
	if r.verbose && event.Text != "" {
		fmt.Fprintf(os.Stderr, "[%s] %s (%s)\n", event.Type, event.Text, event.Priority)
	}
}

func hasMeaningfulChanges(diff mcp.Diff) bool {
	return diff.FocusChanged ||
		len(diff.WidgetsAdded) > 0 ||
		len(diff.WidgetsRemoved) > 0 ||
		len(diff.WidgetsModified) > 0
}
