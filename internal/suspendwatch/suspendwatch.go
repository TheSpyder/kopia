// Package suspendwatch detects whether the machine was suspended (put to sleep) while
// an operation was in progress.
//
// It compares the two clocks that time.Time carries: the wall clock keeps running
// while the machine is asleep, the monotonic clock (on macOS and Linux) does not, so
// time that elapsed on one but not the other was spent suspended. No OS-specific API
// is needed; on platforms whose monotonic clock does run across suspend, Suspended()
// simply reports zero.
package suspendwatch

import "time"

// MinDetectable is the smallest gap between the two clocks that is reported as a
// suspend: well above the clock skew of normal operation, well below the shortest
// realistic sleep.
const MinDetectable = 5 * time.Second

// Watch observes an interval of time for evidence that the machine was suspended.
// The zero value is not usable, call Start.
type Watch struct {
	start time.Time
}

// Start begins watching. The returned Watch may be copied and read any number of times.
func Start() Watch {
	// time.Now rather than clock.Now: clock.Now discards the monotonic reading,
	// which is the very thing this package compares against the wall clock.
	return Watch{start: time.Now()} //nolint:forbidigo
}

// Suspended returns how long the machine spent suspended since Start was called,
// rounded down to zero when the two clocks agree.
func (w Watch) Suspended() time.Duration {
	// see the note in Start about time.Now versus clock.Now.
	now := time.Now() //nolint:forbidigo

	// Round(0) strips the monotonic reading, so the first subtraction uses the wall
	// clock and the second the monotonic clock.
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
