package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
			// Log every request so you can debug in Render Dashboard
			log.Printf("Request: %s %s", r.Method, r.Host)
			
			if r.Method == http.MethodConnect {
				httpsHandler(w, r)
			} else {
				httpHandler(w, r)
			}
		}),
	}

	log.Printf("Proxy starting on port %s...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""
	if r.URL.Scheme == "" { r.URL.Scheme = "http" }
	if r.URL.Host == "" { r.URL.Host = r.Host }

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func httpsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Try to connect to the destination FIRST
	// Standard CONNECT requests normally include the port (e.g. google.com:443)
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("Failed to dial %s: %v", r.Host, err)
		http.Error(w, "Failed to connect: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	
	// 2. NOW Hijack the connection.
	// We do this BEFORE writing 200 OK to ensure we control the raw stream.
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 3. MANUALLY write the 200 OK to the raw client connection
	// This is the magic fix for "ERR_TUNNEL_CONNECTION_FAILED"
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		destConn.Close()
		clientConn.Close()
		return
	}

	// 4. Start the tunnels
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
