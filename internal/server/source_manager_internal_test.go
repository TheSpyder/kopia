package server

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/suspendwatch"
)

// TestShouldBackoffAfter guards the distinction the sleep handling rests on: a snapshot
// that failed because the machine slept must not be put on the failure backoff, while a
// genuine failure must be.
//
// Getting this wrong is quiet in both directions. Backing off an interrupted snapshot
// pushes its retry past the next short wake, so a sleeping laptop never finishes one;
// failing to back off a genuine failure turns a broken source into a hot loop.
func TestShouldBackoffAfter(t *testing.T) {
	t.Parallel()

	genuine := errors.New("error putting manifest: EOF")

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"success", nil, false},
		{"genuine failure", genuine, true},
		{"wrapped genuine failure", errors.Wrap(genuine, "snapshot task"), true},
		{"bare interruption", errInterruptedBySuspend, false},
		{
			// the shape wrapIfInterruptedBySuspend actually produces.
			name: "interruption wrapping a genuine failure",
			err:  errors.Wrapf(errInterruptedBySuspend, "%v (machine was asleep for %v)", genuine, 17*time.Minute),
			want: false,
		},
		{
			// and once more through the layer runSnapshotTask adds.
			name: "interruption wrapped again by the task runner",
			err: errors.Wrap(
				errors.Wrapf(errInterruptedBySuspend, "%v (machine was asleep for %v)", genuine, time.Minute),
				"snapshot task"),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := shouldBackoffAfter(tc.err); got != tc.want {
				t.Errorf("shouldBackoffAfter(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestInterruptedBySuspendErrorMessage pins the wording the user actually reads against
// the task in the UI. The interruption must lead; when it trailed, the message opened
// with a raw EOF and only explained itself at the end, which reads as an unexplained
// failure - and is the first thing truncated in a narrow column.
func TestInterruptedBySuspendErrorMessage(t *testing.T) {
	t.Parallel()

	cause := errors.New("unable to save snapshot: error putting manifest: EOF")
	err := &interruptedBySuspendError{cause: cause, slept: 17*time.Minute + 16*time.Second}

	const want = "snapshot interrupted by system sleep after 17m16s: unable to save snapshot: error putting manifest: EOF"

	if got := err.Error(); got != want {
		t.Errorf("Error() =\n  %q\nwant\n  %q", got, want)
	}

	if !errors.Is(err, errInterruptedBySuspend) {
		t.Error("does not match the errInterruptedBySuspend sentinel, so callers cannot detect it")
	}

	if !errors.Is(err, cause) {
		t.Error("the original failure was dropped from the chain")
	}
}

// TestWrapIfInterruptedBySuspendLeavesGenuineFailuresAlone covers the paths reachable
// without a real suspend: a machine that did not sleep must not have its errors
// reclassified, or genuine failures would be silently hidden from the user.
func TestWrapIfInterruptedBySuspendLeavesGenuineFailuresAlone(t *testing.T) {
	t.Parallel()

	w := suspendwatch.Start() // this machine is awake, so DidSuspend() is false

	if got := wrapIfInterruptedBySuspend(nil, w); got != nil {
		t.Errorf("wrapIfInterruptedBySuspend(nil) = %v, want nil", got)
	}

	genuine := errors.New("boom")

	got := wrapIfInterruptedBySuspend(genuine, w)
	if !errors.Is(got, genuine) {
		t.Errorf("error was replaced rather than passed through: %v", got)
	}

	if errors.Is(got, errInterruptedBySuspend) {
		t.Errorf("a failure with no suspend was misreported as an interruption: %v", got)
	}
}
