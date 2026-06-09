package main

import (
	"context"
	"log"
	"net/http"

	"api_guardian/internal/config"
	"api_guardian/internal/gateway"
	"api_guardian/internal/middleware"
	"api_guardian/internal/proxy"

	"github.com/redis/go-redis/v9"
)

func main() {
	auth_route := config.Route{
		Backends:     []string{"http://localhost:8081", "http://localhost:8082"},
		TrimPrefix:   true,
		Protected:    false,
		AllowedRoles: nil,
	}
	payments_route := config.Route{
		Backends:     []string{"http://localhost:8083", "http://localhost:8084"},
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

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	g := gateway.Gateway{Proxies: make(map[string]*proxy.ReverseProxy)}

	for path, route := range cfg.Routes {

		g.Proxies[path] = proxy.NewReverseProxy(path, route, client, ctx)
	}

	log.Println("Gateway started...")

	if err := http.ListenAndServe(":8090", middleware.Tracer(middleware.Logging(middleware.IPRateLimiter(client, ctx, &g)))); err != nil {
		log.Fatal("Server Failure")
	}
}
