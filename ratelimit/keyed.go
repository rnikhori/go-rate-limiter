package ratelimit

import (
	"context"
	"sync"
	"time"
)

// KeyedLimiter maintains one token-bucket Limiter per key (for example, a
// client IP or user ID), creating buckets lazily on first use. It is safe
// for concurrent use by multiple goroutines.
type KeyedLimiter struct {
	rate  float64
	burst int

	mu      sync.Mutex
	buckets map[string]*keyedEntry
}

type keyedEntry struct {
	limiter  *Limiter
	lastSeen time.Time
}

// NewKeyedLimiter creates a KeyedLimiter whose per-key limiters all share
// the given rate and burst.
func NewKeyedLimiter(rate float64, burst int) *KeyedLimiter {
	return &KeyedLimiter{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*keyedEntry),
	}
}

// Allow reports whether a request for key may proceed right now, creating
// a new limiter for key on first use.
func (k *KeyedLimiter) Allow(key string) bool {
	return k.limiterFor(key).Allow()
}

// Wait blocks until a request for key may proceed or ctx is done, in
// which case it returns ctx.Err(). A new limiter for key is created on
// first use if necessary.
func (k *KeyedLimiter) Wait(ctx context.Context, key string) error {
	return k.limiterFor(key).Wait(ctx)
}

func (k *KeyedLimiter) limiterFor(key string) *Limiter {
	now := time.Now()

	k.mu.Lock()
	defer k.mu.Unlock()

	entry, ok := k.buckets[key]
	if !ok {
		entry = &keyedEntry{limiter: NewLimiter(k.rate, k.burst)}
		k.buckets[key] = entry
	}
	entry.lastSeen = now
	return entry.limiter
}
