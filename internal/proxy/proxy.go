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
	BackendPool  *backendPool
}

func NewReverseProxy(path string, route config.Route, rdb *redis.Client, ctx context.Context) *ReverseProxy {
	pool := newBackendPool(route.Backends)
	pool.startHealthChecks()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend := pool.next()
		if backend == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := forward(w, r, path, route.TrimPrefix, backend); err != nil {
			backend.markFailure()
			if r.Method == "GET" {
				backend = pool.next()
				if backend == nil {
					http.Error(w, "service unavailable", http.StatusServiceUnavailable)
					return
				}
				if err = forward(w, r, path, route.TrimPrefix, backend); err != nil {
					backend.markFailure()
					http.Error(w, "service unavailable", http.StatusServiceUnavailable)
					return
				}
			}
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		}

	})

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

func forward(w http.ResponseWriter, r *http.Request, path string, trim bool, backend *backend) error {
	target := backend.url
	proxy := &httputil.ReverseProxy{}

	proxy.Rewrite = func(pr *httputil.ProxyRequest) {

		log.Printf("Forwarding to: %s", target.String())

		pr.SetURL(target)
		if trim {
			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, path)
		}
		pr.SetXForwarded()
	}
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 3 * time.Second,
	}

	var proxyErr error
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		proxyErr = err
	}

	proxy.ServeHTTP(w, r)
	return proxyErr
}
