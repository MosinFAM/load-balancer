package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	alive        int32
	reverseProxy *httputil.ReverseProxy
}

func NewBackend(rawURL string) *Backend {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Invalid backend URL: %s", rawURL)
	}

	backend := &Backend{
		URL:   parsedURL,
		alive: 1,
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = parsedURL.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Backend %s failed: %v", rawURL, err)
		backend.SetAlive(false)
		http.Error(w, "Backend not available", http.StatusServiceUnavailable)
	}

	backend.reverseProxy = proxy

	go backend.healthCheck()

	return backend
}

func (b *Backend) ReverseProxy() *httputil.ReverseProxy {
	return b.reverseProxy
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (b *Backend) SetAlive(alive bool) {
	var val int32
	if alive {
		val = 1
	}
	if b.IsAlive() != alive {
		log.Printf("Backend %s is now %v", b.URL, alive)
	}
	atomic.StoreInt32(&b.alive, val)
}

func (b *Backend) healthCheck() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := http.Get(b.URL.String())
		if err != nil || resp.StatusCode >= 500 {
			b.SetAlive(false)
			log.Printf("Health check failed for %s", b.URL)
			continue
		}
		b.SetAlive(true)
	}
}
