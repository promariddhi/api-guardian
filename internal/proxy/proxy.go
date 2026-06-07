package proxy

import (
	"api_guardian/internal/middleware"
	"api_guardian/internal/ratelimiter"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ReverseProxy struct {
	Path         string
	Handler      http.Handler
	AllowedRoles []string
}

func NewReverseProxy(path string, protected bool, targetUrl *url.URL, trim bool, allowedRoles []string, userRateLimiter *ratelimiter.RateLimiter) *ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(targetUrl)
			if trim {
				pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, path)
			}
			pr.SetXForwarded()
		},
	}

	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 3 * time.Second,
	}

	handler := http.Handler(proxy)
	if protected {
		handler = middleware.Auth(middleware.UserRateLimiter(userRateLimiter, handler), allowedRoles)
	}

	return &ReverseProxy{
		Path:         path,
		Handler:      handler,
		AllowedRoles: allowedRoles,
	}
}
