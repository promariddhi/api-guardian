package gateway

import (
	"net/http"
	"strings"
)

type Gateway struct {
	Proxies map[string]http.Handler
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, proxy := range g.Proxies {
		if strings.HasPrefix(r.URL.Path, path) {
			proxy.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
