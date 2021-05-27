package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestKeyedLimiterIsolatesKeys(t *testing.T) {
	k := NewKeyedLimiter(1, 1)

	if !k.Allow("alice") {
		t.Fatal("expected alice's first request to be allowed")
	}
	if k.Allow("alice") {
		t.Fatal("expected alice's second request to be denied, burst exhausted")
	}

	if !k.Allow("bob") {
		t.Fatal("expected bob's first request to be allowed, independent of alice")
	}
}

func TestKeyedLimiterWaitPerKey(t *testing.T) {
	k := NewKeyedLimiter(200, 1) // refill every 5ms

	if !k.Allow("alice") {
		t.Fatal("expected first request to be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := k.Wait(ctx, "alice"); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
}
