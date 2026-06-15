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
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	env := config.LoadEnv()

	client := redis.NewClient(&redis.Options{
		Addr:     env.RedisAddr,
		Password: env.RedisPassword,
		DB:       env.RedisDB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := gateway.Gateway{Proxies: make(map[string]*proxy.ReverseProxy)}

	for _, route := range cfg.Routes {

		g.Proxies[route.Path] = proxy.NewReverseProxy(route.Path, route, client, ctx, appCtx)
	}

	log.Println("Gateway started...")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", middleware.Tracer(middleware.Logging(middleware.IPRateLimiter(client, ctx, &g))))

	server := http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server Failure: ", err)
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
