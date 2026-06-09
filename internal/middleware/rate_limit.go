package middleware

import (
	"api_guardian/internal/metrics"
	"api_guardian/internal/ratelimiter"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func IPRateLimiter(rdb *redis.Client, ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromUrl, _, _ := net.SplitHostPort(r.RemoteAddr)
		key := fmt.Sprintf("ip:%s", fromUrl)
		if ok := ratelimiter.RedisAllow(key, ctx, rdb); !ok {
			http.Error(w, "ip rate limit reached", http.StatusTooManyRequests)
			metrics.RateLimitedRequests.WithLabelValues(r.URL.String()).Inc()
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserRateLimiter(rdb *redis.Client, ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claimsCtx, ok := r.Context().Value(claimsKey).(*Claims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		key := fmt.Sprintf("userId:%s", claimsCtx.Subject)
		if ok := ratelimiter.RedisAllow(key, ctx, rdb); !ok {
			http.Error(w, "user rate limit reached", http.StatusTooManyRequests)
			metrics.RateLimitedRequests.WithLabelValues(r.URL.String()).Inc()
			return
		}
		next.ServeHTTP(w, r)
	})
}
