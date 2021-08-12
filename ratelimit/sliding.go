package ratelimit

import (
	"sync"
	"time"
)

// SlidingWindowLimiter limits requests to at most limit occurrences per
// window, using a weighted count of the current and previous fixed
// windows to approximate a true sliding window without storing individual
// request timestamps. It implements the RateLimiter interface.
type SlidingWindowLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration

	windowStart time.Time
	currentCt   int
	previousCt  int

	allowed uint64
	denied  uint64
}

// NewSlidingWindowLimiter creates a SlidingWindowLimiter that permits at
// most limit requests per window duration.
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:       limit,
		window:      window,
		windowStart: time.Now(),
	}
}

// Allow reports whether a request may proceed right now under the
// weighted sliding window count, and if so, records it.
func (s *SlidingWindowLimiter) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.advanceLocked(time.Now())

	elapsedInWindow := time.Since(s.windowStart)
	weight := float64(s.window-elapsedInWindow) / float64(s.window)
	if weight < 0 {
		weight = 0
	}

	estimate := float64(s.previousCt)*weight + float64(s.currentCt)
	if estimate >= float64(s.limit) {
		s.denied++
		return false
	}

	s.currentCt++
	s.allowed++
	return true
}

// advanceLocked rolls the window forward if enough time has elapsed.
// Callers must hold s.mu.
func (s *SlidingWindowLimiter) advanceLocked(now time.Time) {
	elapsed := now.Sub(s.windowStart)
	if elapsed < s.window {
		return
	}

	windows := int(elapsed / s.window)
	if windows == 1 {
		s.previousCt = s.currentCt
	} else {
		s.previousCt = 0
	}
	s.currentCt = 0
	s.windowStart = s.windowStart.Add(time.Duration(windows) * s.window)
}
