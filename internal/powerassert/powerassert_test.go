package powerassert_test

import (
	"testing"

	"github.com/kopia/kopia/internal/powerassert"
	"github.com/kopia/kopia/internal/testlogging"
)

// TestHoldAndRelease verifies that Hold works on every platform and that the returned
// Release is idempotent. Whether an assertion is actually created depends on the OS
// and on the sandbox the test runs in, so that is not asserted.
func TestHoldAndRelease(t *testing.T) {
	ctx := testlogging.Context(t)

	release := powerassert.Hold(ctx, "kopia unit test")
	if release == nil {
		t.Fatal("Hold() returned nil Release")
	}

	release()
	release()
}

func TestHoldDisabledByEnvar(t *testing.T) {
	t.Setenv("KOPIA_DISABLE_POWER_ASSERTIONS", "true")

	ctx := testlogging.Context(t)

	powerassert.Hold(ctx, "kopia unit test")()
}
