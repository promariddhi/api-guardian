package proxy

import (
	"api_guardian/internal/config"
	"api_guardian/internal/middleware"
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReverseProxy struct {
	Path         string
	Handler      http.Handler
	AllowedRoles []string
	BackendPool  *BackendPool
}

func NewReverseProxy(path string, route config.Route, rdb *redis.Client, ctx context.Context) *ReverseProxy {
	pool := NewBackendPool(route.Backends)
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			target := pool.Next()
			log.Println("Forwarding to:", target.String())

			pr.SetURL(target)
			if route.TrimPrefix {
				pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, path)
			}
			pr.SetXForwarded()
		},
	}

	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 3 * time.Second,
	}

	handler := http.Handler(proxy)
	if route.Protected {
		handler = middleware.Auth(middleware.UserRateLimiter(rdb, ctx, handler), route.AllowedRoles)
	}

	return &ReverseProxy{
		Path:         path,
		Handler:      handler,
		AllowedRoles: route.AllowedRoles,
		BackendPool:  pool,
	}
}
