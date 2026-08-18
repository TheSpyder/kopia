// Package powerassert holds operating system power assertions for the duration of
// long-running background operations, such as scheduled snapshot uploads.
//
// On laptops a scheduled snapshot often starts during one of the short maintenance
// ("dark") wakes the OS performs on its own. Unless something holds the machine up,
// it goes back to sleep mid-upload and the snapshot fails with an EOF-style error
// once the machine wakes again. The implementation is OS-specific and is a no-op on
// platforms where kopia does not know how to talk to the power manager.
package powerassert

import (
	"context"
	"os"
	"strconv"

	"github.com/kopia/kopia/repo/logging"
)

var log = logging.Module("powerassert")

// disableEnvVar can be set to a truthy value to make Hold a no-op, for users who
// prefer to manage system sleep themselves.
const disableEnvVar = "KOPIA_DISABLE_POWER_ASSERTIONS"

// Release releases power assertions previously acquired by Hold. It is always safe to
// call, including more than once.
type Release func()

// Hold asks the operating system to keep the machine awake until the returned Release
// is called. The reason is a human-readable description of the work being performed,
// surfaced by OS tooling (`pmset -g assertions` on macOS). Hold never fails: when
// assertions are unsupported, disabled or rejected by the OS it returns a Release
// that does nothing.
func Hold(ctx context.Context, reason string) Release {
	if v, err := strconv.ParseBool(os.Getenv(disableEnvVar)); err == nil && v {
		log(ctx).Debugf("not holding a power assertion for %q because %v is set", reason, disableEnvVar)

		return func() {}
	}

	return hold(ctx, reason)
}
