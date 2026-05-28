package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func NewReverseProxy(path string, targetUrl *url.URL, trim bool) *httputil.ReverseProxy {
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

	return proxy
}
