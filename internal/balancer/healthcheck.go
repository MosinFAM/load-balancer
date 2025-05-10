package balancer

import (
	"log"
	"net"
	"net/url"
	"time"
)

func isAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func StartHealthCheck(pool *Pool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			log.Println("Running health check...")
			for _, b := range pool.Backends() {
				alive := isAlive(b.URL)
				b.SetAlive(alive)
			}
		}
	}()
}
