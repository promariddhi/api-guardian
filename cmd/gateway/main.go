package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"api_guardian/internal/config"
	"api_guardian/internal/gateway"
	"api_guardian/internal/middleware"
	"api_guardian/internal/proxy"
	"api_guardian/internal/ratelimiter"

	"github.com/redis/go-redis/v9"
)

var userRateLimiter = ratelimiter.NewRateLimiter()

func main() {
	auth_route := config.Route{
		Path:         "/auth",
		Url:          "http://localhost:8082",
		TrimPrefix:   true,
		Protected:    false,
		AllowedRoles: nil,
	}
	payments_route := config.Route{
		Path:         "/payments",
		Url:          "http://localhost:8081",
		TrimPrefix:   true,
		Protected:    true,
		AllowedRoles: []string{"admin"},
	}
	cfg := config.Config{
		Routes: map[string]config.Route{
			"/auth":     auth_route,
			"/payments": payments_route,
		},
	}

	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	g := gateway.Gateway{Proxies: make(map[string]*proxy.ReverseProxy)}

	for path, backendRoute := range cfg.Routes {
		targetUrl, err := url.Parse(backendRoute.Url)
		if err != nil {
			log.Fatal()
		}

		g.Proxies[path] = proxy.NewReverseProxy(path, cfg.Routes[path].Protected, targetUrl, cfg.Routes[path].TrimPrefix, cfg.Routes[path].AllowedRoles, userRateLimiter)
	}

	log.Println("Gateway started...")

	if err := http.ListenAndServe(":8090", middleware.Logging(middleware.IPRateLimiter(client, ctx, &g))); err != nil {
		log.Fatal("Server Failure")
	}
}
