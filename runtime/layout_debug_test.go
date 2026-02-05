package runtime

import (
	"bytes"
	"strings"
	"testing"
)

func TestWarnInvalidConstraints_Deduplicates(t *testing.T) {
	var buf bytes.Buffer
	restore := setupLayoutDebugTest(t, &buf, true)
	defer restore()

	c := Constraints{
		MinWidth:  10,
		MaxWidth:  2,
		MinHeight: 1,
		MaxHeight: 1,
	}
	WarnInvalidConstraints("test.scope", c)
	WarnInvalidConstraints("test.scope", c)

	output := buf.String()
	if count := strings.Count(output, "unsatisfiable constraints"); count != 1 {
		t.Fatalf("warning count = %d, want 1\noutput:\n%s", count, output)
	}
}

func TestWarnZeroMeasure_Enabled(t *testing.T) {
	var buf bytes.Buffer
	restore := setupLayoutDebugTest(t, &buf, true)
	defer restore()

	c := Constraints{MinWidth: 0, MaxWidth: 20, MinHeight: 0, MaxHeight: 5}
	WarnZeroMeasure("test.zero", c, Size{})

	output := buf.String()
	if !strings.Contains(output, "zero-size measurement") {
		t.Fatalf("missing zero-size warning in output: %q", output)
	}
}

func TestWarnZeroMeasure_Disabled(t *testing.T) {
	var buf bytes.Buffer
	restore := setupLayoutDebugTest(t, &buf, false)
	defer restore()

	c := Constraints{MinWidth: 0, MaxWidth: 20, MinHeight: 0, MaxHeight: 5}
	WarnZeroMeasure("test.zero.disabled", c, Size{})

	if buf.Len() != 0 {
		t.Fatalf("expected no output when layout debug disabled, got: %q", buf.String())
	}
}

func setupLayoutDebugTest(t *testing.T, writer *bytes.Buffer, enabled bool) func() {
	t.Helper()

	layoutDebugOnce.Do(initLayoutDebug)
	resetLayoutWarningsForTest()

	layoutDebugMu.Lock()
	prevEnabled := layoutDebugEnabled
	prevWriter := layoutDebugWriter
	layoutDebugEnabled = enabled
	layoutDebugWriter = writer
	layoutDebugMu.Unlock()

	return func() {
		layoutDebugMu.Lock()
		layoutDebugEnabled = prevEnabled
		layoutDebugWriter = prevWriter
		layoutDebugMu.Unlock()
		resetLayoutWarningsForTest()
	}
}
