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

Run the demo:

```sh
go run .
```

## Tests

```sh
go test ./...
```
