package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"proxy/domains"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	server := &http.Server{
		Addr: "0.0.0.0:" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				httpsHandeler(w, r)
			} else {
				httpHandeler(w, r)
			}
		}),
	}
	server.ListenAndServe()
}

func httpHandeler(w http.ResponseWriter, r *http.Request) {
	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return
	}
	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func httpsHandeler(w http.ResponseWriter, r *http.Request) {
	if adBloker(r.Host) {
		return
	}

	dest, err := net.Dial("tcp", r.Host)
	if err != nil {
		return
	}
	defer dest.Close()
	w.WriteHeader(http.StatusOK)

	hijack, ok := w.(http.Hijacker)
	if !ok {
		return
	}
	client, _, err := hijack.Hijack()
	if err != nil {
		return
	}
	defer client.Close()

	go func() {
		io.Copy(dest, client)
	}()
	io.Copy(client, dest)
}

func adBloker(host string) bool {
	sp := strings.Split(host, ":")
	if domains.BlockedAds[sp[0]] {
		return true
	}
	return false
}
