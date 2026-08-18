package suspendwatch

import (
	"strings"
	"testing"
	"time"
)

// TestStartRetainsMonotonicReading ensures the timestamp captured by Start carries a
// monotonic clock reading. Replacing time.Now with clock.Now would strip it (via
// discardMonotonicTime), making both subtractions in Suspended read the wall clock
// and silently disabling suspend detection. This must be an in-package test: the
// field is unexported and an unsuspended machine behaves the same either way.
func TestStartRetainsMonotonicReading(t *testing.T) {
	t.Parallel()

	// time.Time.String appends " m=..." only when the value carries a monotonic reading.
	if s := Start().start.String(); !strings.Contains(s, " m=") {
		t.Fatalf("Start().start = %s, want a monotonic reading", s)
	}
}

// TestStrippedMonotonicReadingIsInert demonstrates the failure mode guarded against
// above: with the monotonic reading stripped, Suspended is permanently zero.
func TestStrippedMonotonicReadingIsInert(t *testing.T) {
	t.Parallel()

	// Round(0) is what discardMonotonicTime does to the timestamp.
	w := Watch{start: time.Now().Round(0)} //nolint:forbidigo

	time.Sleep(10 * time.Millisecond)

	if got := w.Suspended(); got != 0 {
		t.Errorf("Suspended() = %v, want 0", got)
	}

	if w.DidSuspend() {
		t.Error("DidSuspend() = true for a start time without a monotonic reading")
	}
}
