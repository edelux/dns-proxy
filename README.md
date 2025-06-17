
# Lightweight DNS Proxy with dnsmasq and DNSCrypt-Proxy

This container provides a minimal Alpine-based DNS proxy using `dnsmasq` and `dnscrypt-proxy`, combining traditional DNS caching and resolution with secure DNS-over-HTTPS (DoH) upstream queries.

## 🛠 Features

- Alpine Linux base image (tiny footprint)
- `dnscrypt-proxy` configured to use Cloudflare DoH resolvers
- `dnsmasq` running in foreground with DNS caching enabled
- DNS requests forwarded to 127.0.0.1:5300 (dnscrypt-proxy)
- Custom `--server=` parameters accepted at runtime
- Designed for multi-architecture builds

## 🚀 Quick Start

```zsh
docker run -d \
  --name secure-dns \
  -p 53:53/udp \
  -p 53:53/tcp \
  edelux/dnsmasq:0.0
```
