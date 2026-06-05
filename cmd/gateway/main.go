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
	auth_route := config.Route{
		Path:       "/auth",
		Url:        "http://localhost:8082",
		TrimPrefix: true,
		Protected:  false,
	}
	payments_route := config.Route{
		Path:       "/payments",
		Url:        "http://localhost:8081",
		TrimPrefix: true,
		Protected:  true,
	}
	cfg := config.Config{
		Routes: map[string]config.Route{
			"/auth":     auth_route,
			"/payments": payments_route,
		},
	}

	g := gateway.Gateway{Proxies: make(map[string]*proxy.ReverseProxy)}

	for path, backendRoute := range cfg.Routes {
		targetUrl, err := url.Parse(backendRoute.Url)
		if err != nil {
			log.Fatal()
		}

		g.Proxies[path] = proxy.NewReverseProxy(path, cfg.Routes[path].Protected, targetUrl, cfg.Routes[path].TrimPrefix)
	}

	log.Println("Gateway started...")

	if err := http.ListenAndServe(":8090", middleware.Logging(&g)); err != nil {
		log.Fatal("Server Failure")
	}
}
