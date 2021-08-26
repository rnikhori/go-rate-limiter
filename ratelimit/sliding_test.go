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
