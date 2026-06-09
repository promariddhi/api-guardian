package proxy

import (
	"api_guardian/internal/metrics"
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type backendPool struct {
	backends []backend
	current  int
	mu       sync.Mutex
}

type backend struct {
	url          *url.URL
	alive        bool
	failureCount int
}

func newBackendPool(urls []string) *backendPool {
	backends := make([]backend, len(urls))
	for i, urlString := range urls {
		if url, err := url.Parse(urlString); err == nil {
			backends[i] = backend{url: url, alive: true}
		} else {
			log.Fatal("error when making backend pool")
		}
	}
	return &backendPool{
		backends: backends,
		current:  0,
	}
}

func (p *backendPool) next() *backend {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 1; i <= len(p.backends); i++ {
		current := (p.current + i) % len(p.backends)
		if p.backends[current].alive {
			p.current = current
			return &p.backends[p.current]
		}
	}
	return nil
}

func (b *backend) markFailure() {
	b.failureCount++
	metrics.BackendFailures.WithLabelValues(b.url.String()).Inc()
	if b.failureCount >= 3 {
		b.alive = false
		metrics.BackendUp.WithLabelValues(b.url.String()).Set(0)
		metrics.CircuitBreakerTrips.WithLabelValues(b.url.String()).Inc()
	}
}

func (p *backendPool) startHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Health Checker Stopped")
				return
			case <-ticker.C:
				p.checkDeadBackends()

			}
		}
	}()
}

func (p *backendPool) checkDeadBackends() {
	p.mu.Lock()
	var dead []*backend
	for i := range p.backends {
		if !p.backends[i].alive {
			dead = append(dead, &p.backends[i])

		}
	}
	p.mu.Unlock()
	for _, backend := range dead {
		backend.probeBackend()
	}
}

func (b *backend) probeBackend() {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(b.url.String())
	if err != nil {
		return
	}
	resp.Body.Close()
	b.alive = true
	b.failureCount = 0
	log.Printf("backend recovered: %s", b.url.String())
	metrics.BackendUp.WithLabelValues(b.url.String()).Set(1)
	metrics.BackendsRecovered.WithLabelValues(b.url.String()).Inc()
}
