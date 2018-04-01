package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type W http.HandlerFunc

func (f W) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

type WebView struct {
	server *http.Server

	// TODO: not replaced safely
	revision int
	source   string
}

func NewWebView() *WebView {
	wv := &WebView{}
	wv.server = &http.Server{
		Addr: ":8421",
		Handler: W(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/shader/fragment" {
				if r.URL.RawQuery != "" {
					rev, _ := strconv.Atoi(r.URL.RawQuery)
					for i := 0; i < 300 && rev == wv.revision; i++ {
						time.Sleep(100 * time.Millisecond)
					}
				}
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("X-Revision", fmt.Sprintf("%d", wv.revision))
				w.WriteHeader(200)
				w.Write([]byte(wv.source))
				return
			} else if page, ok := webSource[r.URL.Path]; ok {
				// TODO: store content-type
				if r.URL.Path == "" {
					w.Header().Set("Content-Type", "text/html")
				} else if strings.HasSuffix(r.URL.Path, ".js") {
					w.Header().Set("Content-Type", "text/javascript")
				}
				w.WriteHeader(200)
				w.Write([]byte(page))
			} else {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(404)
				w.Write([]byte("<!DOCTYPE html><p>Not found"))
			}
		}),
	}
	log.Println("Starting webserver on port 8421")
	go wv.server.ListenAndServe()
	return wv
}

func (w *WebView) Update(source string) {
	w.source = source
	w.revision += 1
	if w.revision < 0 {
		w.revision = 1
	}
}
