package httplimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rnikhori/go-rate-limiter/ratelimit"
)

func TestMiddlewareWithCustomKeyFuncIsolatesClients(t *testing.T) {
	limiter := ratelimit.NewKeyedLimiter(1, 1)
	handler := Middleware(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithKeyFunc(func(r *http.Request) string {
		return r.Header.Get("X-API-Key")
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.Header.Set("X-API-Key", "tenant-a")
	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.Header.Set("X-API-Key", "tenant-b")

	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("expected tenant-a's first request to succeed, got %d", recA.Code)
	}

	recA2 := httptest.NewRecorder()
	handler.ServeHTTP(recA2, reqA)
	if recA2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected tenant-a's second request to be denied, got %d", recA2.Code)
	}

	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("expected tenant-b's request to succeed independently, got %d", recB.Code)
	}
}

func TestMiddlewareAllowsRequestsUnderTheLimit(t *testing.T) {
	limiter := ratelimit.NewKeyedLimiter(1, 5)
	called := 0
	handler := Middleware(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.2:12345"

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200 within burst, got %d", i, rec.Code)
		}
	}

	if called != 5 {
		t.Fatalf("expected the wrapped handler to be called 5 times, got %d", called)
	}
}

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
