package proxy

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MosinFAM/load-balancer/internal/balancer"
)

const (
	MaxAttempts = 3
	MaxRetries  = 3
)

type contextKey string

var (
	AttemptsKey = contextKey("attempts")
	RetryKey    = contextKey("retry")
)

type LoadBalancer struct {
	Pool *balancer.Pool
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	attempts := getCtxValue(r, AttemptsKey)
	if attempts >= MaxAttempts {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	backend := lb.Pool.NextBackend()
	if backend == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	backend.ReverseProxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Error from %s: %v", backend.URL.Host, err)
		retries := getCtxValue(req, RetryKey)
		if retries < MaxRetries {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(req.Context(), RetryKey, retries+1)
			backend.ReverseProxy.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		lb.Pool.MarkBackendStatus(backend.URL, false)
		ctx := context.WithValue(r.Context(), AttemptsKey, attempts+1)
		lb.ServeHTTP(w, r.WithContext(ctx))
	}

	backend.ReverseProxy.ServeHTTP(w, r)
}

func getCtxValue(r *http.Request, key contextKey) int {
	val, _ := r.Context().Value(key).(int)
	return val
}
