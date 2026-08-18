// Package suspendwatch detects whether the machine was suspended (put to sleep) while
// an operation was in progress.
//
// It works by comparing the two clocks that time.Time carries: the wall clock, which
// keeps running while the machine is asleep, and the monotonic clock, which on macOS
// (mach_absolute_time) and Linux (CLOCK_MONOTONIC) does not. Time that elapsed on the
// wall clock but not on the monotonic clock is time the machine spent suspended.
//
// This needs no OS-specific API. On any platform whose monotonic clock does keep
// running across suspend, Suspended() simply reports zero and callers degrade to
// treating the operation as an ordinary one.
package suspendwatch

import "time"

// MinDetectable is the smallest gap between the two clocks that is reported as a
// suspend. It is well above the sub-millisecond skew of normal operation and well below
// the shortest realistic sleep.
const MinDetectable = 5 * time.Second

// Watch observes an interval of time for evidence that the machine was suspended.
// The zero value is not usable, call Start.
type Watch struct {
	start time.Time
}

// Start begins watching. The returned Watch may be copied and read any number of times.
func Start() Watch {
	// time.Now rather than clock.Now: clock.Now deliberately discards the monotonic
	// reading, which is the very thing this package compares against the wall clock.
	return Watch{start: time.Now()} //nolint:forbidigo
}

// Suspended returns how long the machine spent suspended since Start was called,
// rounded down to zero when the two clocks agree.
func (w Watch) Suspended() time.Duration {
	// see the note in Start about time.Now versus clock.Now.
	now := time.Now() //nolint:forbidigo

	// t.Round(0) strips the monotonic reading, so this subtraction uses the wall clock,
	// while the unmodified subtraction below uses the monotonic reading.
	wall := now.Round(0).Sub(w.start.Round(0))
	monotonic := now.Sub(w.start)

	if d := wall - monotonic; d > 0 {
		return d
	}

	return 0
}

// DidSuspend reports whether the machine was suspended for a detectable amount of time
// since Start was called.
func (w Watch) DidSuspend() bool {
	return w.Suspended() >= MinDetectable
}
