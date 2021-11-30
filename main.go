package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rnikhori/go-rate-limiter/httplimit"
	"github.com/rnikhori/go-rate-limiter/ratelimit"
)

func main() {
	demoSingleLimiter()
	demoKeyedLimiter()
	demoMiddleware()
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

func demoMiddleware() {
	fmt.Println("-- httplimit.Middleware --")
	keyed := ratelimit.NewKeyedLimiter(1, 2) // 1 token/sec, burst of 2, per client IP

	handler := httplimit.Middleware(keyed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}), httplimit.WithRetryAfter(1*time.Second))

	for i := 1; i <= 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.42:5555"

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		fmt.Printf("request %d: status=%d\n", i, rec.Code)
	}
}
