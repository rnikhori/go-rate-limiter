package ratelimit

// Stats is a snapshot of how many requests a limiter has allowed and
// denied since it was created or last Reset.
type Stats struct {
	Allowed uint64
	Denied  uint64
}

// Total returns the total number of requests observed (Allowed + Denied).
func (s Stats) Total() uint64 {
	return s.Allowed + s.Denied
}
