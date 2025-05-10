package ratelimit

import (
	"sync"
	"time"

	"github.com/MosinFAM/load-balancer/internal/model"
)

type Bucket struct {
	capacity   int
	tokens     int
	refillRate int // токенов в секунду
	lastRefill time.Time
	mu         sync.Mutex
}

type ClientLimit struct {
	Capacity   int
	RefillRate int
}

type Storage interface {
	GetClientLimit(key string) (*model.ClientLimit, error)
}

type RateLimiter struct {
	buckets     map[string]*Bucket
	mu          sync.RWMutex
	defaultCap  int
	defaultRate int
	Store       Storage
}

func NewRateLimiter(capacity, refillRate int, store Storage) *RateLimiter {
	rl := &RateLimiter{
		buckets:     make(map[string]*Bucket),
		defaultCap:  capacity,
		defaultRate: refillRate,
		Store:       store,
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

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	b, ok := rl.buckets[key]
	if !ok {
		capacity := rl.defaultCap
		refillRate := rl.defaultRate
		if rl.Store != nil {
			if cl, err := rl.Store.GetClientLimit(key); err == nil && cl != nil {
				capacity = cl.Capacity
				refillRate = cl.RefillRate
			}
		}
		b = &Bucket{
			capacity:   capacity,
			tokens:     capacity,
			refillRate: refillRate,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
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
