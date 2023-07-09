package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"simple-load-balancer/backend"
	"simple-load-balancer/server"
	"time"
)

const (
	Attempts int = iota // implement constants incremetally
	Retry
)

func main() {

	backendURLs := []*url.URL{
		parseURL("http://localhost:8080"),
		parseURL("http://localhost:8081"),
		parseURL("http://localhost:8082"),
	}

	backends := make([]*backend.Backend, len(backendURLs))
	for i, backendURL := range backendURLs {
		backends[i] = &backend.Backend{
			URL:          backendURL,
			Alive:        true,
			ReverseProxy: httputil.NewSingleHostReverseProxy(backendURL),
		}
	}

	var server server.ServerPool = server.ServerPool{
		Backends: backends,
		Current:  0,
	}

	go healthCheck(&server)
}

func healthCheck(s *server.ServerPool) {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			s.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

func parseURL(urlstring string) *url.URL {
	parsedURL, _ := url.Parse(urlstring)
	return parsedURL
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func GetAttemptFromContext(r *http.Request) int {
	if attempt, ok := r.Context().Value(Attempts).(int); ok {
		return attempt
	}
	return 0
}
