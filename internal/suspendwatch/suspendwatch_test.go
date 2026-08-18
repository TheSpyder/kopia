package suspendwatch_test

import (
	"testing"
	"time"

	"github.com/kopia/kopia/internal/suspendwatch"
)

// tolerance is the largest gap between the two clocks this test treats as normal.
// The wall and monotonic readings are sampled separately, so they diverge by a few
// nanoseconds even when nothing sleeps, and NTP slewing CLOCK_REALTIME against
// CLOCK_MONOTONIC widens that further under a VM. Anything in that range is noise;
// the value that matters is whether the gap approaches MinDetectable.
const tolerance = time.Second

func TestNoSuspendWhenRunningNormally(t *testing.T) {
	t.Parallel()

	w := suspendwatch.Start()

	time.Sleep(10 * time.Millisecond)

	if got := w.Suspended(); got > tolerance {
		t.Errorf("Suspended() = %v, want at most %v when the machine did not sleep", got, tolerance)
	}

	if w.DidSuspend() {
		t.Error("DidSuspend() = true when the machine did not sleep")
	}
}

func TestMinDetectableIsSane(t *testing.T) {
	t.Parallel()

	// guards against someone lowering the threshold into the range of ordinary
	// scheduling jitter, which would make every slow snapshot look like a suspend.
	if suspendwatch.MinDetectable < time.Second {
		t.Errorf("MinDetectable = %v, too small to distinguish a suspend from jitter", suspendwatch.MinDetectable)
	}

	// the tolerance above only means something if it stays well under the threshold.
	if tolerance >= suspendwatch.MinDetectable {
		t.Errorf("tolerance = %v, must stay below MinDetectable = %v", tolerance, suspendwatch.MinDetectable)
	}
}
