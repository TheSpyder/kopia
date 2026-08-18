package suspendwatch

import (
	"strings"
	"testing"
	"time"
)

// TestStartRetainsMonotonicReading guards the single property this package rests on:
// the timestamp Start captures must carry a monotonic clock reading.
//
// Suspended subtracts monotonic elapsed time from wall elapsed time. Replacing
// time.Now with clock.Now — which the forbidigo linter asks for, and which is
// suppressed in suspendwatch.go on purpose — routes through discardMonotonicTime and
// strips that reading. Both subtractions then read the same wall clock, the difference
// is permanently zero, and DidSuspend can never fire again.
//
// This has to live in-package: the field it inspects is unexported, and no external
// behavior distinguishes the two cases on a machine that did not actually sleep.
func TestStartRetainsMonotonicReading(t *testing.T) {
	t.Parallel()

	// time.Time.String documents that it appends " m=±<value>" when, and only when,
	// the value carries a monotonic reading.
	if s := Start().start.String(); !strings.Contains(s, " m=") {
		t.Fatalf("Start().start = %s, which carries no monotonic reading; "+
			"suspend detection is a silent no-op without it", s)
	}
}

// TestStrippedMonotonicReadingIsInert pins down why the guard above matters. It does
// not assert desirable behavior: it demonstrates the regression, so that the cost of
// stripping the monotonic reading is visible in the suite rather than implied.
func TestStrippedMonotonicReadingIsInert(t *testing.T) {
	t.Parallel()

	// Round(0) is what discardMonotonicTime does to the timestamp.
	w := Watch{start: time.Now().Round(0)} //nolint:forbidigo

	time.Sleep(10 * time.Millisecond)

	// With no monotonic reading on either side, both subtractions in Suspended read the
	// wall clock, so they cancel exactly, however long the machine was actually asleep.
	if got := w.Suspended(); got != 0 {
		t.Errorf("Suspended() = %v, want exactly 0: a stripped start cannot observe a gap", got)
	}

	if w.DidSuspend() {
		t.Error("DidSuspend() = true, but a stripped start cannot detect a suspend at all")
	}
}
