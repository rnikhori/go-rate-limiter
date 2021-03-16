// Package ratelimit provides a thread-safe token-bucket rate limiter.
package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Limiter is a token-bucket rate limiter. Tokens are added at rate per
// second, up to a maximum of burst, and each call to Allow or Wait
// consumes one token.
type Limiter struct {
	mu         sync.Mutex
	rate       float64
	burst      float64
	tokens     float64
	lastRefill time.Time

	allowed atomic.Uint64
	denied  atomic.Uint64
}

// NewLimiter creates a Limiter that refills at rate tokens per second with
// a maximum burst capacity of burst tokens. The bucket starts full.
func NewLimiter(rate float64, burst int) *Limiter {
	return &Limiter{
		rate:       rate,
		burst:      float64(burst),
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

// Allow reports whether a token is available and, if so, consumes it.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refillLocked(time.Now())
	if l.tokens < 1 {
		l.denied.Add(1)
		return false
	}
	l.tokens--
	l.allowed.Add(1)
	return true
}

// Wait blocks until a token becomes available or ctx is cancelled, in
// which case it returns ctx.Err().
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		l.refillLocked(now)

		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			l.allowed.Add(1)
			return nil
		}

		missing := 1 - l.tokens
		wait := time.Duration(missing / l.rate * float64(time.Second))
		l.mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// Reset restores the limiter to a full bucket and zeroes its Stats
// counters, as if it had just been created with NewLimiter.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tokens = l.burst
	l.lastRefill = time.Now()
	l.allowed.Store(0)
	l.denied.Store(0)
}

// SetRate changes the limiter's refill rate at runtime. Pending token
// accumulation under the previous rate is settled first, so the new rate
// only affects refills from this point forward.
func (l *Limiter) SetRate(rate float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refillLocked(time.Now())
	l.rate = rate
}

// SetBurst changes the limiter's maximum burst capacity at runtime. If the
// new burst is smaller than the current token count, the available tokens
// are capped down to the new burst immediately.
func (l *Limiter) SetBurst(burst int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refillLocked(time.Now())
	l.burst = float64(burst)
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
}

// Rate returns the limiter's current refill rate in tokens per second.
func (l *Limiter) Rate() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rate
}

// Burst returns the limiter's current maximum burst capacity.
func (l *Limiter) Burst() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return int(l.burst)
}

// Stats returns a snapshot of the limiter's allowed/denied request counts.
// It is safe to call concurrently with Allow and Wait.
func (l *Limiter) Stats() Stats {
	return Stats{
		Allowed: l.allowed.Load(),
		Denied:  l.denied.Load(),
	}
}

// refillLocked adds tokens accumulated since lastRefill, capped at burst.
// Callers must hold l.mu.
func (l *Limiter) refillLocked(now time.Time) {
	elapsed := now.Sub(l.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	l.tokens += elapsed * l.rate
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.lastRefill = now
}
