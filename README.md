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

Run the demo:

```sh
go run .
```

## Tests

```sh
go test ./...
```
