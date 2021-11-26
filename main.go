package main

import (
	"fmt"

	"github.com/rnikhori/go-rate-limiter/ratelimit"
)

func main() {
	demoSingleLimiter()
	demoKeyedLimiter()
}

func demoSingleLimiter() {
	fmt.Println("-- single Limiter --")
	limiter := ratelimit.NewLimiter(5, 10) // 5 tokens/sec, burst of 10

	for i := 1; i <= 20; i++ {
		if limiter.Allow() {
			fmt.Printf("request %2d: allowed\n", i)
		} else {
			fmt.Printf("request %2d: denied\n", i)
		}
	}

	stats := limiter.Stats()
	fmt.Printf("stats: allowed=%d denied=%d\n", stats.Allowed, stats.Denied)
}

func demoKeyedLimiter() {
	fmt.Println("-- KeyedLimiter (per client) --")
	keyed := ratelimit.NewKeyedLimiter(2, 3) // 2 tokens/sec, burst of 3, per key

	clients := []string{"alice", "bob", "alice", "alice", "bob"}
	for _, client := range clients {
		if keyed.Allow(client) {
			fmt.Printf("client %-6s: allowed\n", client)
		} else {
			fmt.Printf("client %-6s: denied\n", client)
		}
	}

	fmt.Printf("tracked keys: %d\n", keyed.Len())
}
