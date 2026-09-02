package spectrocloud

import "time"

// waitDelayOverride, when non-nil, replaces the Delay field of every
// retry.StateChangeConf constructed by the provider (13 call sites at
// last count). Tests set this to 0 via TestMain so wait loops fire on
// the first refresh instead of blocking for 30 seconds each; production
// leaves it nil, preserving the original 30-second default.
//
// The variable is set once at package init (or once by TestMain) before
// any goroutines read it, so a plain pointer read is safe without a
// mutex. The waiter helpers call resolveWaitDelay in their Delay field;
// production code paths therefore observe no behavior change unless the
// override is explicitly assigned.
var waitDelayOverride *time.Duration

// resolveWaitDelay returns the test-only override when set, otherwise
// the caller's production default. Every retry.StateChangeConf built by
// the provider funnels its Delay field through this helper — search
// the codebase for `resolveWaitDelay(` to enumerate the call sites.
func resolveWaitDelay(fallback time.Duration) time.Duration {
	if waitDelayOverride != nil {
		return *waitDelayOverride
	}
	return fallback
}
