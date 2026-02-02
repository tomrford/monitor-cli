package relay

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu       sync.Mutex
	rate     float64
	capacity float64
	tokens   float64
	last     time.Time
}

func NewTokenBucket(rps, burst float64) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		rate:     rps,
		capacity: burst,
		tokens:   burst,
		last:     now,
	}
}

func (b *TokenBucket) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.last = now
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}
