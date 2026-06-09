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
	var matched *proxy.ReverseProxy
	longest := 0
	for path, proxy := range g.Proxies {
		if strings.HasPrefix(r.URL.Path, path) && len(path) > longest {
			matched = proxy
			longest = len(path)
		}
	}
	if matched != nil {
		matched.Handler.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}
