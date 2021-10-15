package httplimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnikhori/go-rate-limiter/ratelimit"
)

func TestMiddlewareReturns429WhenDenied(t *testing.T) {
	limiter := ratelimit.NewKeyedLimiter(1, 1)
	handler := Middleware(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.1:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got status %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 once the limit is exhausted, got status %d", rec.Code)
	}
}
