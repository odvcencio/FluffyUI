package terminal

import "testing"

func FuzzParseKittyKey(f *testing.F) {
	f.Add([]byte("\x1b[65u"))         // Plain 'A'
	f.Add([]byte("\x1b[65;2u"))       // 'A' with shift
	f.Add([]byte("\x1b[65:97;1u"))    // Shifted key
	f.Add([]byte("\x1b[57358u"))      // F1
	f.Add([]byte("\x1b[13u"))         // Enter
	f.Add([]byte(""))                 // Empty
	f.Add([]byte("\x1b"))             // Just ESC
	f.Add([]byte("\x1b["))            // Incomplete CSI
	f.Add([]byte("hello"))            // Not a CSI sequence

	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = ParseKittyKey(input) // Should never panic
	})
}
