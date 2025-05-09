package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Port     string   `json:"port"`
	Backends []string `json:"backends"`
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	// Переопределение из переменных окружения
	if port := os.Getenv("LB_PORT"); port != "" {
		cfg.Port = port
	}
	if be := os.Getenv("LB_BACKENDS"); be != "" {
		cfg.Backends = strings.Split(be, ",")
	}

	return &cfg, nil
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if len(cfg.Backends) == 0 {
		log.Fatal("No backends configured")
	}

	balancer := NewBalancer(cfg.Backends)

	log.Printf("Starting load balancer on port %s\n", cfg.Port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		balancer.Serve(w, r)
	})

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
