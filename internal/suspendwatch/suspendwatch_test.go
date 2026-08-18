package suspendwatch_test

import (
	"testing"
	"time"

	"github.com/kopia/kopia/internal/suspendwatch"
)

func TestNoSuspendWhenRunningNormally(t *testing.T) {
	t.Parallel()

	w := suspendwatch.Start()

	time.Sleep(10 * time.Millisecond)

	if got := w.Suspended(); got != 0 {
		t.Errorf("Suspended() = %v, want 0 when the machine did not sleep", got)
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
}
