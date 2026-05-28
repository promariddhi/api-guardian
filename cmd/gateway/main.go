package main

import (
	"log"
	"net/http"
	"net/url"

	"api_guardian/internal/config"
	"api_guardian/internal/gateway"
	"api_guardian/internal/middleware"
	"api_guardian/internal/proxy"
)

func main() {
	cfg := config.Config{
		Routes: map[string]string{
			"/auth":     "http://localhost:8082",
			"/payments": "http://localhost:8081",
		},
		TrimPrefix: true,
	}

	g := gateway.Gateway{Proxies: make(map[string]http.Handler)}

	for path, backend := range cfg.Routes {
		targetUrl, err := url.Parse(backend)
		if err != nil {
			log.Fatal()
		}

		g.Proxies[path] = proxy.NewReverseProxy(path, targetUrl, cfg.TrimPrefix)
	}

	loggedGateway := middleware.Logging(&g)

	log.Println("Gateway started...")

	if err := http.ListenAndServe(":8090", loggedGateway); err != nil {
		log.Fatal("Server Failure")
	}
}
