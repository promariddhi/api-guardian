package middleware

import (
	"net"
	"net/http"
	"sync"
)

type IpRateLimiter struct {
	mu       sync.Mutex
	requests map[string]uint64
}

var ipRateLimiter = NewIpRateLimiter()

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromUrl, _, _ := net.SplitHostPort(r.RemoteAddr)
		ipRateLimiter.Increment(fromUrl)
		next.ServeHTTP(w, r)
	})
}

func NewIpRateLimiter() *IpRateLimiter {
	return &IpRateLimiter{
		requests: make(map[string]uint64),
	}
}

func (r *IpRateLimiter) Increment(fromUrl string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[fromUrl]++
}
