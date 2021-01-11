package main

import (
	"fmt"

	"github.com/rnikhori/go-rate-limiter/ratelimit"
)

func main() {
	limiter := ratelimit.NewLimiter(5, 10) // 5 tokens/sec, burst of 10

	for i := 1; i <= 20; i++ {
		if limiter.Allow() {
			fmt.Printf("request %2d: allowed\n", i)
		} else {
			fmt.Printf("request %2d: denied\n", i)
		}
	}
}
