package ratelimit

import "context"

// RateLimiter is the common behavior shared by every limiter implementation
// in this package (token bucket, sliding window, ...). Code that only needs
// to check or wait for permission should depend on this interface rather
// than a concrete type, so the underlying algorithm can be swapped freely.
type RateLimiter interface {
	// Allow reports whether a request may proceed right now and, if so,
	// records it against the limit.
	Allow() bool

	// Wait blocks until a request may proceed or ctx is done, in which
	// case it returns ctx.Err().
	Wait(ctx context.Context) error
}

var _ RateLimiter = (*Limiter)(nil)
