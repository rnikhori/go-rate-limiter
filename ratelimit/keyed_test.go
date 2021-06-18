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

func TestKeyedLimiterRemove(t *testing.T) {
	k := NewKeyedLimiter(1, 1)

	k.Allow("alice")
	if k.Len() != 1 {
		t.Fatalf("expected 1 key, got %d", k.Len())
	}

	k.Remove("alice")
	if k.Len() != 0 {
		t.Fatalf("expected key to be removed, Len is %d", k.Len())
	}

	if !k.Allow("alice") {
		t.Fatal("expected a fresh limiter with full burst after Remove")
	}
}

func TestKeyedLimiterKeysAndLen(t *testing.T) {
	k := NewKeyedLimiter(1, 1)

	if k.Len() != 0 {
		t.Fatalf("expected empty KeyedLimiter to have Len 0, got %d", k.Len())
	}

	k.Allow("alice")
	k.Allow("bob")

	if got := k.Len(); got != 2 {
		t.Fatalf("expected Len 2, got %d", got)
	}

	keys := k.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %v", keys)
	}

	seen := map[string]bool{}
	for _, key := range keys {
		seen[key] = true
	}
	if !seen["alice"] || !seen["bob"] {
		t.Fatalf("expected keys to contain alice and bob, got %v", keys)
	}
}

func TestKeyedLimiterEvictIdle(t *testing.T) {
	k := NewKeyedLimiter(1, 1)

	k.Allow("stale")
	time.Sleep(20 * time.Millisecond)
	k.Allow("fresh")

	evicted := k.EvictIdle(10 * time.Millisecond)
	if evicted != 1 {
		t.Fatalf("expected 1 key evicted, got %d", evicted)
	}

	if _, ok := k.buckets["stale"]; ok {
		t.Fatal("expected stale key to be removed")
	}
	if _, ok := k.buckets["fresh"]; !ok {
		t.Fatal("expected fresh key to remain")
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
