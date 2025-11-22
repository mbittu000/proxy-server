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
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// LOGGING: This is critical to know if your phone is even connecting!
			log.Printf("👉 INCOMING: %s %s", r.Method, r.Host)

			if r.Method == http.MethodConnect {
				handleTunnel(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}

	log.Printf("✅ Proxy LIVE on port %s. Waiting for connections...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Standard HTTP handling
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

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	// 1. Set a timeout for the connection (prevents hanging)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("❌ FAIL: Could not dial destination %s: %v", r.Host, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 2. Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		log.Printf("❌ FAIL: Webserver does not support Hijacking")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, clientBuffer, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		log.Printf("❌ FAIL: Could not hijack connection: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 3. Send 200 OK
	// We simply write to the raw connection.
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		destConn.Close()
		clientConn.Close()
		return
	}

	// 4. Start Tunnels
	// IMPORTANT: We use 'clientBuffer' for reading from client, in case data is stuck there.
	go transfer(destConn, clientBuffer) // Read from buffer+conn, Write to Dest
	go transfer(clientConn, destConn)   // Read from Dest, Write to Client
	
	log.Printf("🚀 SUCCESS: Tunnel established to %s", r.Host)
}

// Helper to copy data between connections
func transfer(destination io.WriteCloser, source io.Reader) {
	defer destination.Close()
	io.Copy(destination, source)
}
