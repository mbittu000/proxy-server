package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	server := &http.Server{
		Addr: "0.0.0.0:" + port,
		Handler: http.HandlerFunc(handleRequest),
	}

	log.Printf("🚀 Proxy running on port %s", port)
	log.Fatal(server.ListenAndServe())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleTunnel(w, r)
	} else {
		handleHTTP(w, r)
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Standard HTTP Proxy logic
	resp, err := http.Get(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	// 1. Dial the target (Google, etc.)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	
	// 2. THE FIX: Use standard Go WriteHeader + Flush
	// This tells Render "Hey, the connection is good" using standard HTTP
	w.WriteHeader(http.StatusOK)
	
	// Force the header out to Render immediately
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// 3. NOW Hijack the connection for raw data transfer
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		return
	}

	// 4. Start the bidirectional tunnel
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(dest io.WriteCloser, src io.ReadCloser) {
	defer dest.Close()
	defer src.Close()
	io.Copy(dest, src)
}
