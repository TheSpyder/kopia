package suspendwatch_test

import (
	"testing"
	"time"

	"github.com/kopia/kopia/internal/suspendwatch"
)

// tolerance is the largest gap between the two clocks this test treats as normal.
// The wall and monotonic readings are sampled separately and NTP slewing widens the
// gap further under a VM; what matters is that it stays far below MinDetectable.
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

	// a threshold within the range of ordinary scheduling jitter would make every
	// slow snapshot look like a suspend.
	if suspendwatch.MinDetectable < time.Second {
		t.Errorf("MinDetectable = %v, too small to distinguish a suspend from jitter", suspendwatch.MinDetectable)
	}

	if tolerance >= suspendwatch.MinDetectable {
		t.Errorf("tolerance = %v, must stay below MinDetectable = %v", tolerance, suspendwatch.MinDetectable)
	}
}
