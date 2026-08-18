// Package powerassert holds operating system power assertions for the duration of
// long-running background operations, such as scheduled snapshot uploads.
//
// On laptops a scheduled snapshot frequently becomes due while the machine is asleep
// and starts during one of the short maintenance ("dark") wakes that the OS performs
// on its own. Nothing is holding the machine up for kopia's benefit, so the OS is
// free to go back to sleep in the middle of the upload. The kopia process is frozen
// mid-request and its TCP connections are silently torn down; when the machine wakes
// again the upload resumes on dead sockets and fails with an EOF-style error, long
// after the snapshot would normally have finished.
//
// Holding a power assertion for as long as a snapshot is running tells the OS not to
// do that. The implementation is OS-specific and is a no-op on platforms where kopia
// does not know how to talk to the power manager.
package powerassert

import (
	"context"
	"os"
	"strconv"

	"github.com/kopia/kopia/repo/logging"
)

var log = logging.Module("powerassert")

// disableEnvVar can be set to a truthy value to make Hold() do nothing, for users who
// prefer to manage system wakefulness themselves.
const disableEnvVar = "KOPIA_DISABLE_POWER_ASSERTIONS"

// Release releases power assertions previously acquired by Hold. It is always safe to
// call, including more than once.
type Release func()

// Hold asks the operating system to keep the machine awake until the returned Release
// is called. The reason is a human-readable description of the work being performed and
// is surfaced by OS tooling (`pmset -g assertions` on macOS).
//
// Hold never fails the caller: when assertions are unsupported, disabled or rejected by
// the OS it logs the reason and returns a Release that does nothing.
func Hold(ctx context.Context, reason string) Release {
	if v, err := strconv.ParseBool(os.Getenv(disableEnvVar)); err == nil && v {
		log(ctx).Debugf("not holding a power assertion for %q because %v is set", reason, disableEnvVar)

		return func() {}
	}

	return hold(ctx, reason)
}
