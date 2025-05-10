package ratelimit

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity   int
	tokens     int
	refillRate int // токенов в секунду
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimiter struct {
	buckets     map[string]*Bucket
	mu          sync.RWMutex
	defaultCap  int
	defaultRate int
}

func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	rl := &RateLimiter{
		buckets:     make(map[string]*Bucket),
		defaultCap:  capacity,
		defaultRate: refillRate,
	}
	go rl.refillAll()
	return rl
}

func (rl *RateLimiter) refillAll() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		rl.mu.RLock()
		for _, b := range rl.buckets {
			b.mu.Lock()
			now := time.Now()
			elapsed := now.Sub(b.lastRefill).Seconds()
			newTokens := int(elapsed * float64(b.refillRate))
			if newTokens > 0 {
				b.tokens = min(b.capacity, b.tokens+newTokens)
				b.lastRefill = now
			}
			b.mu.Unlock()
		}
		rl.mu.RUnlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &Bucket{
			capacity:   rl.defaultCap,
			tokens:     rl.defaultCap,
			refillRate: rl.defaultRate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
