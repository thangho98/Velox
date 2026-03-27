package subprovider

import (
	"sync"
	"time"
)

// Throttle enforces a minimum interval between API calls for a provider.
type Throttle struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

// NewThrottle creates a rate limiter with the given minimum interval.
func NewThrottle(interval time.Duration) *Throttle {
	return &Throttle{interval: interval}
}

// Wait blocks until enough time has passed since the last call.
func (t *Throttle) Wait() {
	t.mu.Lock()
	defer t.mu.Unlock()
	since := time.Since(t.lastCall)
	if since < t.interval {
		time.Sleep(t.interval - since)
	}
	t.lastCall = time.Now()
}
