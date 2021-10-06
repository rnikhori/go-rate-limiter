// Package httplimit provides an HTTP middleware that enforces a
// per-client rate limit using a keyed limiter from the ratelimit package.
package httplimit

import (
	"net"
	"net/http"
	"strings"
)

// remoteIP returns the client's IP address, preferring the first address
// in an X-Forwarded-For header (set by a trusted reverse proxy) and
// falling back to the request's RemoteAddr.
func remoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, found := strings.Cut(xff, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// KeyedAllower is the subset of ratelimit.KeyedLimiter's behavior the
// middleware depends on. ratelimit.KeyedLimiter satisfies this interface,
// and so does any other type with a matching Allow method.
type KeyedAllower interface {
	Allow(key string) bool
}

// KeyFunc derives the rate-limit key (for example, a client IP or an API
// token) from an incoming request.
type KeyFunc func(*http.Request) string

// Option configures the middleware returned by Middleware.
type Option func(*config)

type config struct {
	keyFunc KeyFunc
}

// WithKeyFunc overrides how the rate-limit key is derived from a request.
// By default, the client's IP address is used (see remoteIP).
func WithKeyFunc(fn KeyFunc) Option {
	return func(c *config) {
		c.keyFunc = fn
	}
}

// Middleware wraps next so that requests are checked against limiter,
// keyed by remoteIP (or a custom KeyFunc from WithKeyFunc). Requests the
// limiter denies receive a 429 Too Many Requests response instead of
// reaching next.
func Middleware(limiter KeyedAllower, next http.Handler, opts ...Option) http.Handler {
	cfg := config{keyFunc: remoteIP}
	for _, opt := range opts {
		opt(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := cfg.keyFunc(r)
		if !limiter.Allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
