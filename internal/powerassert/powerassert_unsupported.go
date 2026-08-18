//go:build !darwin

package powerassert

import "context"

func hold(_ context.Context, _ string) Release {
	// do nothing on platforms other than macOS
	return func() {}
}
