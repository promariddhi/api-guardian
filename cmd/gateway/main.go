package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api_guardian/internal/config"
	"api_guardian/internal/gateway"
	"api_guardian/internal/middleware"
	"api_guardian/internal/proxy"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := gateway.Gateway{Proxies: make(map[string]*proxy.ReverseProxy)}

	for path, route := range cfg.Routes {

		g.Proxies[path] = proxy.NewReverseProxy(path, route, client, ctx, appCtx)
	}

	log.Println("Gateway started...")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", middleware.Tracer(middleware.Logging(middleware.IPRateLimiter(client, ctx, &g))))

	server := http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server Failure")
		}
	}()

	handleShutdown(&server, client)
}

func handleShutdown(s *http.Server, client *redis.Client) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, os.Interrupt)

	<-sig
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	if err := client.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	log.Println("Shutdown Complete")
}
