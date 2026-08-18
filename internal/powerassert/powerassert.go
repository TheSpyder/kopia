// Package powerassert holds operating system power assertions.
//
// On macOS laptops a scheduled snapshot can start during one of the short maintenance
// ("dark") wakes. This package holds activity power assertions allowing
// snapshots to complete without losing their network connection.
package powerassert

import (
	"context"
	"os"
	"strconv"

	"github.com/kopia/kopia/repo/logging"
)

var log = logging.Module("powerassert")

// Allow bypass for users who prefer to manage system sleep themselves.
const disableEnvVar = "KOPIA_DISABLE_POWER_ASSERTIONS"

// Release previous power assertions.
type Release func()

// Hold asks macOS to keep the machine awake until the returned Release is called. The reason
// is visible with OS tooling (`pmset -g assertions` on macOS). This function never fails; when
// assertions are unsupported, disabled or rejected by the OS it returns a no-op Release.
func Hold(ctx context.Context, reason string) Release {
	if v, err := strconv.ParseBool(os.Getenv(disableEnvVar)); err == nil && v {
		log(ctx).Debugf("not holding a power assertion for %q because %v is set", reason, disableEnvVar)

		return func() {}
	}

	return hold(ctx, reason)
}
