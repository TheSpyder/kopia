//go:build !darwin

package powerassert

import "context"

func hold(ctx context.Context, reason string) Release {
	log(ctx).Debugf("power assertions are not implemented on this platform, not holding one for %q", reason)

	return func() {}
}
