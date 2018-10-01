package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
)

var (
	skipHeaders = map[string]interface{}{"Origin": nil} //, "Host": nil}
)

type service struct {
	m *http.ServeMux
	b string
	c *http.Client
	s *http.Server
}

func newService(listenAddress, b string) *service {
	s := &service{m: http.NewServeMux(), b: b, c: &http.Client{
		// http client that does not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}
	s.routes(b)
	s.s = &http.Server{
		Addr: listenAddress, Handler: s.Mux(),
		TLSNextProto: make(map[string]func(s *http.Server, c *tls.Conn, h http.Handler)), // disable HTTP/2
	}

	return s
}

func (s *service) Server() *http.Server {
	return s.s
}

func (s *service) Mux() *http.ServeMux {
	return s.m
}

func (s *service) BackendURL(r *http.Request) string {
	return s.b + r.RequestURI
}

func (s *service) NormalizeRedirects(loc []string) (*[]string, error) {
	// remove scheme://host:port from backend redirects
	location := make([]string, len(loc))
	for i, l := range loc {
		u, err := url.Parse(l)
		if err != nil {
			return nil, err
		}
		if strings.Contains(s.b, "://") && strings.Contains(s.b, u.Host) {
			location[i] = u.RequestURI()
		} else {
			location[i] = l
		}
	}
	return &location, nil
}

func (s *service) ProxyRequest(r *http.Request) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, s.BackendURL(r), r.Body)
	if err != nil {
		return nil, err
	}
	req.TransferEncoding = []string{"identity"}
	req.Close = true
	req.Header = make(http.Header)

	for h, val := range r.Header {
		if _, ok := skipHeaders[h]; ok {
			continue
		}
		req.Header[h] = val
	}

	return req, nil
}

func (s *service) Client() *http.Client {
	return s.c
}
