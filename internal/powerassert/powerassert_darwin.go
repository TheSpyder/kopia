//go:build darwin

package powerassert

import (
	"context"
	"sync"

	"github.com/ebitengine/purego"
	"github.com/pkg/errors"
)

// Assertion types held while a snapshot is running. Two are needed because they cover
// different situations and neither one is sufficient on its own:
//
//   - BackgroundTask tells powerd that we are performing work on behalf of the user
//     during a maintenance wake. It is what keeps a lid-closed dark wake from ending as
//     soon as macOS has finished its own housekeeping, and unlike PreventSystemSleep it
//     is honored on battery.
//   - PreventUserIdleSystemSleep stops the idle timer from putting a fully awake machine
//     to sleep during a long upload the user is not interacting with.
//
// Neither can defeat an explicit sleep (closing the lid on an awake machine, or the Sleep
// menu item) - nothing in user space can.
const (
	assertionTypeBackgroundTask             = "BackgroundTask"
	assertionTypePreventUserIdleSystemSleep = "PreventUserIdleSystemSleep"
)

const (
	// kIOPMAssertionLevelOn.
	assertionLevelOn = 255

	// kIOReturnSuccess.
	ioReturnSuccess = 0

	// kCFStringEncodingUTF8.
	cfStringEncodingUTF8 = 0x0800_0100

	coreFoundationPath = "/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation"
	ioKitPath          = "/System/Library/Frameworks/IOKit.framework/IOKit"
)

//nolint:gochecknoglobals
var (
	initOnce sync.Once
	errInit  error

	// CFStringRef CFStringCreateWithCString(CFAllocatorRef, const char *, CFStringEncoding).
	cfStringCreateWithCString func(alloc uintptr, cstr string, encoding uint32) uintptr

	// void CFRelease(CFTypeRef).
	cfRelease func(ref uintptr)

	// IOReturn IOPMAssertionCreateWithName(CFStringRef, IOPMAssertionLevel, CFStringRef, IOPMAssertionID *).
	iopmAssertionCreateWithName func(assertionType uintptr, level uint32, name uintptr, id *uint32) int32

	// IOReturn IOPMAssertionRelease(IOPMAssertionID).
	iopmAssertionRelease func(id uint32) int32
)

// initFrameworks resolves the CoreFoundation and IOKit entry points we need. purego is used
// rather than cgo because kopia release binaries are built with CGO_ENABLED=0.
func initFrameworks() (err error) {
	// purego.RegisterLibFunc panics when a symbol cannot be resolved.
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("unable to resolve power management symbols: %v", r)
		}
	}()

	cf, err := purego.Dlopen(coreFoundationPath, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return errors.Wrap(err, "unable to load CoreFoundation")
	}

	iokit, err := purego.Dlopen(ioKitPath, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		return errors.Wrap(err, "unable to load IOKit")
	}

	purego.RegisterLibFunc(&cfStringCreateWithCString, cf, "CFStringCreateWithCString")
	purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
	purego.RegisterLibFunc(&iopmAssertionCreateWithName, iokit, "IOPMAssertionCreateWithName")
	purego.RegisterLibFunc(&iopmAssertionRelease, iokit, "IOPMAssertionRelease")

	return nil
}

func hold(ctx context.Context, reason string) Release {
	initOnce.Do(func() {
		errInit = initFrameworks()
	})

	if errInit != nil {
		log(ctx).Debugf("not holding a power assertion for %q: %v", reason, errInit)

		return func() {}
	}

	// the assertion name is retained by IOPMAssertionCreateWithName, so it is released here.
	name := cfStringCreateWithCString(0, reason, cfStringEncodingUTF8)
	if name == 0 {
		log(ctx).Debugf("not holding a power assertion for %q: unable to create assertion name", reason)

		return func() {}
	}

	defer cfRelease(name)

	var ids []uint32

	for _, assertionType := range []string{assertionTypeBackgroundTask, assertionTypePreventUserIdleSystemSleep} {
		t := cfStringCreateWithCString(0, assertionType, cfStringEncodingUTF8)
		if t == 0 {
			continue
		}

		var id uint32

		rc := iopmAssertionCreateWithName(t, assertionLevelOn, name, &id)

		cfRelease(t)

		if rc != ioReturnSuccess {
			log(ctx).Debugf("unable to hold %v power assertion: IOReturn 0x%08x", assertionType, uint32(rc)) //nolint:gosec // IOReturn is a bit pattern, not a magnitude

			continue
		}

		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return func() {}
	}

	log(ctx).Debugf("holding %v power assertion(s) for %q", len(ids), reason)

	var once sync.Once

	return func() {
		once.Do(func() {
			for _, id := range ids {
				if rc := iopmAssertionRelease(id); rc != ioReturnSuccess {
					log(ctx).Debugf("unable to release power assertion %v: IOReturn 0x%08x", id, uint32(rc)) //nolint:gosec // IOReturn is a bit pattern, not a magnitude
				}
			}

			log(ctx).Debugf("released %v power assertion(s) for %q", len(ids), reason)
		})
	}
}
