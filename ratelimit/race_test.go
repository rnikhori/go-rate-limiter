package ratelimit

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentAllowAcrossLimiterTypes exercises Limiter, KeyedLimiter,
// and SlidingWindowLimiter under concurrent load from many goroutines.
// Run with -race to catch any data races; functionally it just checks
// that no call panics and that counts stay internally consistent.
func TestConcurrentAllowAcrossLimiterTypes(t *testing.T) {
	const goroutines = 50
	const perGoroutine = 200

	limiter := NewLimiter(1000, 100)
	keyed := NewKeyedLimiter(1000, 100)
	window := NewSlidingWindowLimiter(10000, time.Second)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			key := "worker"
			for i := 0; i < perGoroutine; i++ {
				limiter.Allow()
				keyed.Allow(key)
				window.Allow()
			}
		}(g)
	}
	wg.Wait()

	total := goroutines * perGoroutine

	if stats := limiter.Stats(); int(stats.Total()) != total {
		t.Fatalf("Limiter: expected %d total requests recorded, got %d", total, stats.Total())
	}

	if stats, ok := keyed.Stats("worker"); !ok || int(stats.Total()) != total {
		t.Fatalf("KeyedLimiter: expected %d total requests recorded, got %+v (ok=%v)", total, stats, ok)
	}

	if stats := window.Stats(); int(stats.Total()) != total {
		t.Fatalf("SlidingWindowLimiter: expected %d total requests recorded, got %d", total, stats.Total())
	}
}
