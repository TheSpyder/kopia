// Package suspendwatch detects whether the machine was suspended (put to sleep) while
// an operation was in progress.
//
// clock.Now specifically discards the monotonic clock because it stops ticking while
// asleep. This does the reverse, leveraging that fact to detect the process was suspended.
package suspendwatch

import "time"

// MinDetectable is the smallest gap between the two clocks that is reported as a suspend.
const MinDetectable = 5 * time.Second

// Watch observes an interval of time for evidence that the machine was suspended.
type Watch struct {
	start time.Time
}

// Start begins watching.
func Start() Watch {
	// Use time.Now directly as clock.Now discards the monotonic reading.
	return Watch{start: time.Now()} //nolint:forbidigo
}

// Suspended returns how long the machine spent suspended, rounded down to zero.
func (w Watch) Suspended() time.Duration {
	// Use time.Now directly as clock.Now discards the monotonic reading.
	now := time.Now() //nolint:forbidigo

	// Strip the monotonic reading so the first subtraction uses the wall
	// clock and the second the monotonic clock.
	wall := now.Round(0).Sub(w.start.Round(0))
	monotonic := now.Sub(w.start)

	if d := wall - monotonic; d > 0 {
		return d
	}

	return 0
}

// DidSuspend reports whether the machine was suspended for a detectable amount of time.
func (w Watch) DidSuspend() bool {
	return w.Suspended() >= MinDetectable
}
