# go-rate-limiter

A minimal, thread-safe token-bucket rate limiter for Go using only the standard library.

## Usage

```go
import "github.com/rnikhori/go-rate-limiter/ratelimit"

limiter := ratelimit.NewLimiter(5, 10) // 5 tokens/sec, burst of 10

if limiter.Allow() {
    // proceed
} else {
    // rejected, over the rate limit
}

// Or block until a token is available (or the context is cancelled):
if err := limiter.Wait(ctx); err != nil {
    // ctx was cancelled/timed out
}
```

### Metrics

Every limiter tracks how many requests it has allowed and denied:

```go
stats := limiter.Stats()
fmt.Println(stats.Allowed, stats.Denied, stats.Total())
```

### Runtime configuration

Rate and burst can be changed after the limiter is created, and a limiter
can be restored to a fresh state:

```go
limiter.SetRate(10)   // tokens/sec from now on
limiter.SetBurst(20)  // new maximum burst capacity
limiter.Reset()       // full bucket, zeroed stats
```

`Rate()` and `Burst()` report the current configuration.

### Per-key limiting

`KeyedLimiter` keeps one bucket per key (an IP, a user ID, an API token),
creating buckets lazily the first time a key is seen:

```go
keyed := ratelimit.NewKeyedLimiter(5, 10) // same rate/burst for every key

if keyed.Allow(clientIP) {
    // proceed
}

if err := keyed.Wait(ctx, clientIP); err != nil {
    // ctx was cancelled/timed out
}
```

Keys that go quiet can be cleaned up so memory doesn't grow unbounded:

```go
keyed.EvictIdle(10 * time.Minute) // drop keys unused for 10+ minutes
keyed.Remove(clientIP)            // or drop one key explicitly
keyed.Len()                       // keys currently tracked
keyed.Keys()                      // snapshot of tracked keys
```

Per-key and aggregated metrics are available via `Stats(key)` and
`TotalStats()`.

### Sliding window limiting

`SlidingWindowLimiter` is an alternative algorithm that caps requests to
a fixed count per rolling time window, instead of refilling tokens
continuously. It implements the same `RateLimiter` interface
(`Allow`, `Wait`, `SetRate`, `Reset`, `Stats`) as `Limiter`, so the two are
interchangeable wherever `ratelimit.RateLimiter` is accepted:

```go
window := ratelimit.NewSlidingWindowLimiter(100, time.Minute) // 100 req/min

if window.Allow() {
    // proceed
}
```

Run the demo:

```sh
go run .
```

## Tests

```sh
go test ./...
```
