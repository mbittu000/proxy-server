# Go Proxy Server with Ad Blocking

This project is a **custom HTTP(S) proxy server built in Go**. It forwards requests from clients (such as your phone) to the internet while optionally **blocking popular online ad networks** to reduce ads and improve privacy.

## Features

- Supports both HTTP and HTTPS proxy tunneling
- Blocks top 15 popular ad networks by domain
- Efficient bidirectional TCP data forwarding using `io.Copy`
- Lightweight and portable

## Getting Started

### Prerequisites

- Go 1.20+ installed
- Basic knowledge of Go and networking

### Running the Server

1. Clone this repository
2. Build the proxy server:
```
go build -o proxyserver main.go
```
3. Run the server (listens on port 8080 by default):
```
./proxyserver
```

### Configuring Your Phone or Device

- Connect your phone to the same WiFi network as the computer running the proxy
- Set the proxy server IP to your computer's local IP (e.g., `192.168.0.101`) and port `8080`
- On Android or iOS, configure the WiFi proxy settings manually

## Ad Blocking

The proxy blocks the following popular ad network domains by default:

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

You can add or remove domains easily by editing the `blockedAds` map in the code.

## How It Works

- For HTTP traffic, the proxy transparently forwards requests and responses
- For HTTPS traffic, the proxy establishes a TCP tunnel using the CONNECT method
- It uses Go `net.Dial` to connect to destination servers
- Hijacks the HTTP connection to access raw TCP streams
- Forwards encrypted data bidirectionally between client and destination

## Limitations

- Does not decrypt HTTPS traffic (no man-in-the-middle)
- Simple domain matching for ad blocking (can be improved with wildcards/regex)
- No authentication or encryption between client and proxy

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT License

---

This project combines networking fundamentals with ad-blocking features, making it ideal for educational and practical solopreneur use cases.