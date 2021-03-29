package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestAllowBurst(t *testing.T) {
	l := NewLimiter(1, 3)

	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("request %d: expected allowed within burst", i)
		}
	}
	if l.Allow() {
		t.Fatal("expected request to be denied once burst is exhausted")
	}
}

func TestAllowRefillsOverTime(t *testing.T) {
	l := NewLimiter(100, 1) // 100 tokens/sec, burst 1 -> refill every 10ms

	if !l.Allow() {
		t.Fatal("expected first request to be allowed")
	}
	if l.Allow() {
		t.Fatal("expected immediate second request to be denied")
	}

	time.Sleep(20 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("expected request to be allowed after refill")
	}
}

func TestWaitUnblocksAfterRefill(t *testing.T) {
	l := NewLimiter(200, 1) // refill every 5ms

	if !l.Allow() {
		t.Fatal("expected first request to be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("Wait took too long: %v", elapsed)
	}
}

func TestWaitReturnsOnContextCancel(t *testing.T) {
	l := NewLimiter(0.001, 1) // effectively no refill within the test window

	if !l.Allow() {
		t.Fatal("expected first request to be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := l.Wait(ctx); err != context.DeadlineExceeded {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestResetRestoresFullBucketAndStats(t *testing.T) {
	l := NewLimiter(1, 3)

	for i := 0; i < 4; i++ {
		l.Allow()
	}

	stats := l.Stats()
	if stats.Total() == 0 {
		t.Fatal("expected some requests to be recorded before Reset")
	}

	l.Reset()

	stats = l.Stats()
	if stats.Allowed != 0 || stats.Denied != 0 {
		t.Fatalf("expected zeroed stats after Reset, got %+v", stats)
	}

	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("request %d: expected allowed with full burst after Reset", i)
		}
	}
	if l.Allow() {
		t.Fatal("expected burst to be exhausted again after Reset replenished it")
	}
}

func TestSetRateTakesEffectImmediately(t *testing.T) {
	l := NewLimiter(1, 1) // refill every 1s

	if !l.Allow() {
		t.Fatal("expected first request to be allowed")
	}
	if l.Allow() {
		t.Fatal("expected second request to be denied before refill")
	}

	l.SetRate(1000) // refill every 1ms from now on

	time.Sleep(10 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("expected request to be allowed after rate increase")
	}
}

func TestSetBurstCapsTokensWhenShrunk(t *testing.T) {
	l := NewLimiter(1, 5)

	l.SetBurst(1)
	if got := l.Burst(); got != 1 {
		t.Fatalf("expected burst 1, got %d", got)
	}

	if !l.Allow() {
		t.Fatal("expected first request to be allowed")
	}
	if l.Allow() {
		t.Fatal("expected second request to be denied, tokens should be capped at new burst")
	}
}

func TestRateAndBurstAccessors(t *testing.T) {
	l := NewLimiter(5, 10)

	if got := l.Rate(); got != 5 {
		t.Fatalf("expected rate 5, got %v", got)
	}
	if got := l.Burst(); got != 10 {
		t.Fatalf("expected burst 10, got %v", got)
	}
}

func TestStatsTracksAllowedAndDenied(t *testing.T) {
	l := NewLimiter(1, 2)

	l.Allow()
	l.Allow()
	l.Allow()
	l.Allow()

	stats := l.Stats()
	if stats.Allowed != 2 {
		t.Fatalf("expected 2 allowed, got %d", stats.Allowed)
	}
	if stats.Denied != 2 {
		t.Fatalf("expected 2 denied, got %d", stats.Denied)
	}
	if total := stats.Total(); total != 4 {
		t.Fatalf("expected total of 4, got %d", total)
	}
}
