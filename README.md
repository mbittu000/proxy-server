# Go Proxy Server with Ad Blocking

A simple custom HTTP(S) proxy server built in Go that forwards requests and blocks popular ad networks to reduce ads and improve privacy.

## Features

- Supports both HTTP and HTTPS proxy tunneling
- Blocks the top 15 popular ad network domains
- Efficient bidirectional TCP forwarding using `io.Copy`
- Lightweight and easy to run

## Getting Started

### Prerequisites

- Go 1.20+ installed

### Running the Server

1. Build the proxy server:
```
go build -o proxyserver main.go
```
2. Run the server (defaults to listening on port 8080):
```
./proxyserver
```

### Configure Your Device

- Connect the device to the same network as the proxy server
- Set the proxy IP to the server IP (e.g., `192.168.0.101`) and port `8080`
- Configure manual proxy settings on your device's WiFi

## Ad Blocking

This proxy blocks the following ad domains:
```
doubleclick.net
googlesyndication.com
googleadservices.com
googletagmanager.com
facebook.com
connect.facebook.net
amazon-adsystem.com
propellerads.com
adsterra.com
outbrain.com
taboola.com
criteo.com
media.net
infolinks.com
revenuehits.com
```

You can customize the blocked domains in the code's `blockedAds` map.

## How It Works

- HTTP traffic is proxied transparently
- HTTPS traffic uses a TCP tunnel via the CONNECT method
- The proxy establishes TCP connections to remote servers using Go's `net.Dial`
- Proxies bidirectional encrypted data streams without decrypting

## Limitations

- No HTTPS traffic decryption (no man-in-the-middle)
- Simple domain matching (no wildcards or regex support)
- No authentication or encryption between client and proxy

