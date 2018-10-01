package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

func init() {
	// handlers := map[string]http.HandlerFunc{}
	// handlers["/"] =
}

func (s *service) routes(backend string) {
	s.m.HandleFunc("/", s.log(s.handleIndex()))
}

func (s *service) log(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		le := &logEntry{log: logger, start: nowMillisecond(), r: r}
		defer le.Write()
		dump, _ := httputil.DumpRequest(r, true)
		log.Println(string(dump))
		h.ServeHTTP(sw, r)
		le.statusCode = sw.status
		le.responseLength = sw.length
	})
}

func (s *service) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyReq, err := s.ProxyRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := s.Client().Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if loc, ok := resp.Header["Location"]; ok {
			normLoc, err := s.NormalizeRedirects(loc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			resp.Header["Location"] = *normLoc
		}

		s.copyHeaders(resp.Header, w)
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *service) copyHeaders(headers http.Header, w http.ResponseWriter) {
	for h, val := range headers {
		for _, v := range val {
			w.Header().Set(h, v)
		}
	}
}
