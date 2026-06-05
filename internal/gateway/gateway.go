package gateway

import (
	"api_guardian/internal/proxy"
	"net/http"
	"strings"
)

type Gateway struct {
	Proxies map[string]*proxy.ReverseProxy
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, proxy := range g.Proxies {
		if strings.HasPrefix(r.URL.Path, path) {
			proxy.Handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
