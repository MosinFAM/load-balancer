package ratelimit

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/MosinFAM/load-balancer/internal/model"
)

// мок-хранилище лимитов
type mockStorage struct {
	data map[string]*model.ClientLimit
}

func (m *mockStorage) GetClientLimit(key string) (*model.ClientLimit, error) {
	limit, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return limit, nil
}

func TestAllow_DefaultLimit(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(2, 1, nil)
	key := "user1"

	if !rl.Allow(key) {
		t.Errorf("first request should be allowed")
	}
	if !rl.Allow(key) {
		t.Errorf("second request should be allowed")
	}
	if rl.Allow(key) {
		t.Errorf("third request should be blocked due to token depletion")
	}
}

func TestAllow_WithCustomLimitFromStorage(t *testing.T) {
	t.Parallel()

	mock := &mockStorage{
		data: map[string]*model.ClientLimit{
			"user42": {
				Capacity:   1,
				RefillRate: 1,
			},
		},
	}
	rl := NewRateLimiter(5, 5, mock)

	if !rl.Allow("user42") {
		t.Errorf("request should be allowed from custom limit")
	}
	if rl.Allow("user42") {
		t.Errorf("second request should be blocked due to capacity=1")
	}
}

func TestRefillTokens(t *testing.T) {
	t.Parallel()

	mock := &mockStorage{
		data: map[string]*model.ClientLimit{
			"userX": {
				Capacity:   1,
				RefillRate: 1,
			},
		},
	}
	rl := NewRateLimiter(10, 10, mock) // дефолтные значения не важны, используется кастомный лимит из mock

	key := "userX"
	if !rl.Allow(key) {
		t.Errorf("should allow first request")
	}
	if rl.Allow(key) {
		t.Errorf("should block second request immediately")
	}

	time.Sleep(1100 * time.Millisecond)

	if !rl.Allow(key) {
		t.Errorf("should allow after refill")
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(10, 5, nil)
	key := "concurrent_user"
	var wg sync.WaitGroup
	allowedCount := 0
	mu := sync.Mutex{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow(key) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != 10 {
		t.Errorf("expected 10 allowed requests, got %d", allowedCount)
	}
}
