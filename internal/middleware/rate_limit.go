package middleware

import (
	"api_guardian/internal/ratelimiter"
	"context"
	"net"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func IPRateLimiter(rdb *redis.Client, ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromUrl, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ok := ratelimiter.RedisAllow(fromUrl, ctx, rdb); !ok {
			http.Error(w, "ip rate limit reached", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserRateLimiter(userRateLimiter *ratelimiter.RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(claimsKey).(*Claims)
		if ok := userRateLimiter.Allow(ctx.Subject); !ok {
			http.Error(w, " rate limit reached", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
