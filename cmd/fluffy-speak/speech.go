package main

import (
	"context"
	"sync"
	"time"
)

// Utterance is a speech request with priority.
type Utterance struct {
	Text     string
	Priority string // "assertive" or "polite"
}

// SpeechQueue manages prioritized speech output.
// Assertive utterances preempt current speech.
// Polite utterances are debounced and queued.
type SpeechQueue struct {
	speaker    Speaker
	assertive  chan Utterance
	polite     chan Utterance
	debounce   time.Duration
	mu         sync.Mutex
	cancelFunc context.CancelFunc
}

// NewSpeechQueue creates a speech queue with the given speaker.
func NewSpeechQueue(speaker Speaker, debounce time.Duration) *SpeechQueue {
	if debounce <= 0 {
		debounce = 50 * time.Millisecond
	}
	return &SpeechQueue{
		speaker:   speaker,
		assertive: make(chan Utterance, 32),
		polite:    make(chan Utterance, 64),
		debounce:  debounce,
	}
}

// Enqueue adds an utterance to the queue.
func (q *SpeechQueue) Enqueue(u Utterance) {
	if u.Priority == "assertive" {
		select {
		case q.assertive <- u:
		default:
		}
	} else {
		select {
		case q.polite <- u:
		default:
		}
	}
}

// Run processes the speech queue until ctx is cancelled.
func (q *SpeechQueue) Run(ctx context.Context) {
	var debounceTimer *time.Timer
	var pending *Utterance

	for {
		select {
		case <-ctx.Done():
			_ = q.speaker.Stop()
			return

		case u := <-q.assertive:
			// Assertive: cancel current speech and speak immediately
			q.mu.Lock()
			if q.cancelFunc != nil {
				q.cancelFunc()
			}
			q.mu.Unlock()
			_ = q.speaker.Stop()
			q.speak(ctx, u)

		case u := <-q.polite:
			// Polite: debounce - reset timer on each new polite utterance
			pending = &u
			if debounceTimer == nil {
				debounceTimer = time.NewTimer(q.debounce)
			} else {
				if !debounceTimer.Stop() {
					select {
					case <-debounceTimer.C:
					default:
					}
				}
				debounceTimer.Reset(q.debounce)
			}

		default:
			// Check debounce timer
			if debounceTimer != nil && pending != nil {
				select {
				case <-debounceTimer.C:
					u := *pending
					pending = nil
					q.speak(ctx, u)
				case u := <-q.assertive:
					pending = nil
					q.mu.Lock()
					if q.cancelFunc != nil {
						q.cancelFunc()
					}
					q.mu.Unlock()
					_ = q.speaker.Stop()
					q.speak(ctx, u)
				case u := <-q.polite:
					pending = &u
					if !debounceTimer.Stop() {
						select {
						case <-debounceTimer.C:
						default:
						}
					}
					debounceTimer.Reset(q.debounce)
				case <-ctx.Done():
					_ = q.speaker.Stop()
					return
				}
			} else {
				// Nothing pending, block until next utterance
				select {
				case u := <-q.assertive:
					q.mu.Lock()
					if q.cancelFunc != nil {
						q.cancelFunc()
					}
					q.mu.Unlock()
					_ = q.speaker.Stop()
					q.speak(ctx, u)
				case u := <-q.polite:
					pending = &u
					if debounceTimer == nil {
						debounceTimer = time.NewTimer(q.debounce)
					} else {
						debounceTimer.Reset(q.debounce)
					}
				case <-ctx.Done():
					_ = q.speaker.Stop()
					return
				}
			}
		}
	}
}

func (q *SpeechQueue) speak(ctx context.Context, u Utterance) {
	speakCtx, cancel := context.WithCancel(ctx)
	q.mu.Lock()
	q.cancelFunc = cancel
	q.mu.Unlock()
	_ = q.speaker.Speak(speakCtx, u.Text)
	cancel()
}
