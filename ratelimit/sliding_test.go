package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindowLimiterAllowsUpToLimit(t *testing.T) {
	s := NewSlidingWindowLimiter(3, 50*time.Millisecond)

	for i := 0; i < 3; i++ {
		if !s.Allow() {
			t.Fatalf("request %d: expected allowed within limit", i)
		}
	}
	if s.Allow() {
		t.Fatal("expected request to be denied once limit is reached")
	}
}

func TestSlidingWindowLimiterWeightsPreviousWindow(t *testing.T) {
	window := 40 * time.Millisecond
	s := NewSlidingWindowLimiter(2, window)

	if !s.Allow() || !s.Allow() {
		t.Fatal("expected both requests in the first window to be allowed")
	}
	if s.Allow() {
		t.Fatal("expected third request in the first window to be denied")
	}

	time.Sleep(window + window/2) // move well into the next window

	if !s.Allow() {
		t.Fatal("expected a request early in the new window to be allowed")
	}
}

func TestSlidingWindowLimiterSetRateAndReset(t *testing.T) {
	s := NewSlidingWindowLimiter(1, 50*time.Millisecond)

	if !s.Allow() {
		t.Fatal("expected first request to be allowed")
	}
	if s.Allow() {
		t.Fatal("expected second request to be denied at limit 1")
	}

	s.SetRate(5)
	if !s.Allow() {
		t.Fatal("expected request to be allowed after raising the limit")
	}

	stats := s.Stats()
	if stats.Allowed != 2 || stats.Denied != 1 {
		t.Fatalf("expected stats {2,1}, got %+v", stats)
	}

	s.Reset()
	stats = s.Stats()
	if stats.Allowed != 0 || stats.Denied != 0 {
		t.Fatalf("expected zeroed stats after Reset, got %+v", stats)
	}
}

func TestSlidingWindowLimiterWaitUnblocksAfterWindow(t *testing.T) {
	window := 30 * time.Millisecond
	s := NewSlidingWindowLimiter(1, window)

	if !s.Allow() {
		t.Fatal("expected first request to be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := s.Wait(ctx); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
}
