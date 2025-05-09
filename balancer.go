package main

import (
	"log"
	"net/http"
	"sync"
)

type Balancer struct {
	backends []*Backend
	mu       sync.Mutex
	current  int
}

func NewBalancer(urls []string) *Balancer {
	backends := make([]*Backend, 0, len(urls))
	for _, url := range urls {
		b := NewBackend(url)
		backends = append(backends, b)
	}
	return &Balancer{backends: backends}
}

func (b *Balancer) getNextBackend() *Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := len(b.backends)
	start := b.current

	for {
		backend := b.backends[b.current]
		b.current = (b.current + 1) % n

		if backend.IsAlive() {
			return backend
		}

		if b.current == start {
			// все бэкенды мертвы
			return nil
		}
	}
}

func (b *Balancer) Serve(w http.ResponseWriter, r *http.Request) {
	backend := b.getNextBackend()
	if backend == nil {
		log.Println("No backend available")
		http.Error(w, "503 Service unavailable: no backend alive", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Forwarding request to: %s", backend.URL)
	backend.ReverseProxy().ServeHTTP(w, r)
}
