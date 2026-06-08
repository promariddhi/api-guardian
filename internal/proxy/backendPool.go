package proxy

import (
	"log"
	"net/url"
	"sync"
)

type BackendPool struct {
	backends []*url.URL
	current  int
	mu       sync.Mutex
}

func NewBackendPool(urls []string) *BackendPool {
	backends := make([]*url.URL, len(urls))
	for i, urlString := range urls {
		if url, err := url.Parse(urlString); err == nil {
			backends[i] = url
		} else {
			log.Fatal("error when making backend pool")
		}
	}
	return &BackendPool{
		backends: backends,
		current:  0,
	}
}

func (p *BackendPool) Next() *url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = (p.current + 1) % len(p.backends)
	return p.backends[p.current]
}
