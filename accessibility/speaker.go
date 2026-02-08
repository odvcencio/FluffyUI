package accessibility

import "context"

// Speaker provides text-to-speech output.
// Implementations wrap platform-specific TTS engines.
type Speaker interface {
	// Speak sends text to the TTS engine. It blocks until speech completes
	// or ctx is cancelled.
	Speak(ctx context.Context, text string) error

	// Stop cancels any in-progress speech.
	Stop() error

	// Close releases TTS resources.
	Close() error
}
