package server

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"simple-load-balancer/backend"
	"sync/atomic"
	"time"
)

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.Current, uint64(1)) % uint64(len(s.Backends)))
}

func (s *ServerPool) GetNextPeer() *backend.Backend {
	next := s.NextIndex()
	length := len(s.Backends) + next

	for i := next; i < length; i++ {
		idx := i % len(s.Backends)
		if s.Backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.Current, uint64(idx))
			}
			return s.Backends[idx]
		}
	}

	return nil
}

func (s *ServerPool) LoadBalancer(w http.ResponseWriter, r *http.Request) {
	peer := s.GetNextPeer()

	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.Backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}

		log.Printf("%s [%s]\n", b.URL, status)
	}
}

// Check whether a backand still alive or not by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}

	_ = conn.Close()

	return true
}
