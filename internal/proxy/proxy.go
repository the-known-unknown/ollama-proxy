package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New(target *url.URL) *httputil.ReverseProxy {
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.FlushInterval = -1

	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.Header.Del("Authorization")
		req.Header.Del("X-API-Key")
	}
	return rp
}
