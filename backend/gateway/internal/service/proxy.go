package service

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ReverseProxy struct {
	proxy  *httputil.ReverseProxy
	target *url.URL
	client *http.Client
}

func NewReverseProxy(coreServiceURL string) (*ReverseProxy, error) {
	target, err := url.Parse(coreServiceURL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = client.Transport

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	return &ReverseProxy{
		proxy:  proxy,
		target: target,
		client: client,
	}, nil
}

func (r *ReverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.proxy.ServeHTTP(w, req)
}
