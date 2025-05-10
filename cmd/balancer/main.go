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
)

func main() {
	cfgPath := flag.String("config", "config/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	pool := &balancer.Pool{}
	for _, addr := range cfg.Backends {
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

	lb := &proxy.LoadBalancer{Pool: pool}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: lb,
	}

	log.Printf("Load balancer running on port %d", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
