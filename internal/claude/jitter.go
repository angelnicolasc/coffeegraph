package claude

import (
	rand2 "math/rand/v2"
	"time"
)

// jitteredBackoff adds ±25% random jitter to a base duration to prevent
// thundering herd when multiple clients hit rate limits simultaneously.
func jitteredBackoff(base time.Duration) time.Duration {
	// jitter range: 0.75 to 1.25 of base
	factor := 0.75 + rand2.Float64()*0.5 // [0.75, 1.25)
	return time.Duration(float64(base) * factor)
}
