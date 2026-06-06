package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const BUCKET_CAPACITY = 50
const BUCKET_FILL_RATE = 5

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
}

var ipRateLimiter = NewRateLimiter()
var userRateLimiter = NewRateLimiter()

func IPRateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromUrl, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ok := ipRateLimiter.Allow(fromUrl); !ok {
			http.Error(w, "ip rate limit reached", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserRateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(claimsKey).(*Claims)
		if ok := userRateLimiter.Allow(ctx.Subject); !ok {
			http.Error(w, " rate limit reached", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	allowed := true
	if bucket, ok := r.buckets[key]; ok {
		allowed = bucket.useToken()
	} else {
		r.buckets[key] = NewTokenBucket()
	}
	return allowed
}

type TokenBucket struct {
	availTokens int64
	lastUsed    time.Time
}

func NewTokenBucket() *TokenBucket {
	return &TokenBucket{
		availTokens: BUCKET_CAPACITY - 1,
		lastUsed:    time.Now(),
	}
}

func (b *TokenBucket) useToken() bool {
	elapsed := time.Since(b.lastUsed).Milliseconds()
	b.availTokens = min(BUCKET_CAPACITY, b.availTokens+elapsed/1000*BUCKET_FILL_RATE)
	b.lastUsed = time.Now()
	if b.availTokens < 1 {
		return false
	}
	b.availTokens -= 1
	return true
}
