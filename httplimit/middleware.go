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

// Middleware wraps next so that requests are checked against limiter,
// keyed by the client's remote IP address. Requests the limiter denies
// receive a 429 Too Many Requests response instead of reaching next.
func Middleware(limiter KeyedAllower, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := remoteIP(r)
		if !limiter.Allow(key) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
