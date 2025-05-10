package balancer

import (
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestBackendAliveStatus(t *testing.T) {
	rawURL := "http://localhost:8080"
	u, _ := url.Parse(rawURL)
	b := &Backend{
		URL:          u,
		Alive:        false,
		ReverseProxy: httputil.NewSingleHostReverseProxy(u),
	}

	b.SetAlive(true)
	if !b.IsAlive() {
		t.Errorf("expected backend to be alive")
	}

	b.SetAlive(false)
	if b.IsAlive() {
		t.Errorf("expected backend to be not alive")
	}
}

func TestPool_NextBackend(t *testing.T) {
	p := &Pool{}
	u1, _ := url.Parse("http://backend1:5678")
	u2, _ := url.Parse("http://backend2:5678")

	b1 := &Backend{URL: u1, Alive: true, ReverseProxy: httputil.NewSingleHostReverseProxy(u1)}
	b2 := &Backend{URL: u2, Alive: true, ReverseProxy: httputil.NewSingleHostReverseProxy(u2)}

	p.AddBackend(b1)
	p.AddBackend(b2)

	for i := 0; i < 10; i++ {
		b := p.NextBackend()
		if b == nil {
			t.Errorf("expected non-nil backend")
		}
	}
}

func TestPool_NextBackend_AllDown(t *testing.T) {
	p := &Pool{}
	u1, _ := url.Parse("http://backend1:5678")
	u2, _ := url.Parse("http://backend2:5678")

	b1 := &Backend{URL: u1, Alive: false, ReverseProxy: httputil.NewSingleHostReverseProxy(u1)}
	b2 := &Backend{URL: u2, Alive: false, ReverseProxy: httputil.NewSingleHostReverseProxy(u2)}

	p.AddBackend(b1)
	p.AddBackend(b2)

	if p.NextBackend() != nil {
		t.Errorf("expected nil when all backends are down")
	}
}

func TestPool_MarkBackendStatus(t *testing.T) {
	p := &Pool{}
	u, _ := url.Parse("http://backend1:5678")
	b := &Backend{URL: u, Alive: true, ReverseProxy: httputil.NewSingleHostReverseProxy(u)}
	p.AddBackend(b)

	p.MarkBackendStatus(u, false)
	if b.IsAlive() {
		t.Errorf("expected backend to be marked as down")
	}

	p.MarkBackendStatus(u, true)
	if !b.IsAlive() {
		t.Errorf("expected backend to be marked as up")
	}
}
