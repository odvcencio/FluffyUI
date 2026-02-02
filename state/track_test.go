package state

import (
	"testing"
)

// TestTrackDependencies_Nested tests that nested trackDependencies calls
// properly isolate their dependency tracking (fix for 1.3).
func TestTrackDependencies_Nested(t *testing.T) {
	s1 := NewSignal(1)
	s2 := NewSignal(2)
	s3 := NewSignal(3)

	// Outer tracking reads s1 and calls inner tracking which reads s2
	outerResult, outerDeps := trackDependencies(func() int {
		_ = s1.Get() // Should be tracked by outer

		innerResult, innerDeps := trackDependencies(func() int {
			_ = s2.Get() // Should be tracked by inner only
			return 100
		})

		if innerResult != 100 {
			t.Fatalf("expected inner result 100, got %d", innerResult)
		}
		if len(innerDeps) != 1 {
			t.Fatalf("expected 1 inner dep, got %d", len(innerDeps))
		}

		_ = s3.Get() // Should be tracked by outer (after inner returns)
		return 42
	})

	if outerResult != 42 {
		t.Fatalf("expected outer result 42, got %d", outerResult)
	}
	if len(outerDeps) != 2 {
		t.Fatalf("expected 2 outer deps (s1 and s3), got %d", len(outerDeps))
	}

	// Verify s2 is not in outer deps
	foundS1, foundS3 := false, false
	for _, dep := range outerDeps {
		if dep == s1 {
			foundS1 = true
		}
		if dep == s3 {
			foundS3 = true
		}
		if dep == s2 {
			t.Fatal("s2 should not be in outer deps")
		}
	}
	if !foundS1 || !foundS3 {
		t.Fatal("expected both s1 and s3 in outer deps")
	}
}

// TestTrackDependencies_DeeplyNested tests deeply nested tracking calls.
func TestTrackDependencies_DeeplyNested(t *testing.T) {
	s1 := NewSignal(1)
	s2 := NewSignal(2)
	s3 := NewSignal(3)

	_, deps1 := trackDependencies(func() int {
		_ = s1.Get()

		_, deps2 := trackDependencies(func() int {
			_ = s2.Get()

			_, deps3 := trackDependencies(func() int {
				_ = s3.Get()
				return 3
			})
			if len(deps3) != 1 {
				t.Fatalf("expected 1 dep in level 3, got %d", len(deps3))
			}

			return 2
		})
		if len(deps2) != 1 {
			t.Fatalf("expected 1 dep in level 2, got %d", len(deps2))
		}

		return 1
	})
	if len(deps1) != 1 {
		t.Fatalf("expected 1 dep in level 1, got %d", len(deps1))
	}
}

// TestRecordDependency_NoTracker tests that recordDependency is safe
// when no tracker is active.
func TestRecordDependency_NoTracker(t *testing.T) {
	s := NewSignal(1)
	// Should not panic
	recordDependency(s)
}
