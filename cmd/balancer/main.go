package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/MosinFAM/load-balancer/internal/balancer"
	"github.com/MosinFAM/load-balancer/internal/config"
	"github.com/MosinFAM/load-balancer/internal/proxy"
	"github.com/MosinFAM/load-balancer/internal/ratelimit"
	"github.com/MosinFAM/load-balancer/internal/storage"
)

func main() {
	cfgPath := flag.String("config", "config/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	store, err := storage.NewStore("postgres://user:password@db:5432/rate_limiter?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	pool := &balancer.Pool{}
	log.Println("Loaded backends from config:")
	for _, addr := range cfg.Backends {
		log.Printf(" - %s", addr)
		u, err := url.Parse(addr)
		if err != nil {
			log.Fatalf("Invalid backend address: %s", addr)
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		pool.AddBackend(&balancer.Backend{
			URL:          u,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Added backend: %s\n", u)
	}

	balancer.StartHealthCheck(pool, 30*time.Second)

	rl := ratelimit.NewRateLimiter(5, 1, store) // default: 5 tokens, 1 per second
	lb := &proxy.LoadBalancer{
		Pool:    pool,
		Limiter: rl,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: lb,
	}

	log.Printf("Load balancer running on port %d", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
