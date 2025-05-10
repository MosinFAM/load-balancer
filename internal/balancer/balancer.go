package balancer

import (
	"log"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.Alive = alive
	b.mu.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

type Pool struct {
	backends []*Backend
	current  uint64
}

func (p *Pool) AddBackend(backend *Backend) {
	p.backends = append(p.backends, backend)
}

func (p *Pool) NextBackend() *Backend {
	n := len(p.backends)
	for i := 0; i < n; i++ {
		idx := int(atomic.AddUint64(&p.current, 1)) % n
		b := p.backends[idx]
		if b.IsAlive() {
			return b
		}
	}
	return nil
}

func (p *Pool) MarkBackendStatus(u *url.URL, alive bool) {
	for _, b := range p.backends {
		if b.URL.String() == u.String() {
			b.SetAlive(alive)
			status := "up"
			if !alive {
				status = "down"
			}
			log.Printf("Backend %s marked as %s\n", u, status)
			break
		}
	}
}

func (p *Pool) Backends() []*Backend {
	return p.backends
}
