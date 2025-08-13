
# Lightweight DNS Proxy with dnsmasq and DNSCrypt-Proxy

A secure, minimal DNS resolver container with DNS-over-HTTPS (DoH), caching, and split DNS — ideal for cloud and local infrastructure.

## 🐧 About

`edelux/dns-proxy` is a lightweight container built from [Debian Stable](https://www.debian.org/releases/stable/) and designed for high-security, low-footprint DNS resolution.

This image combines [`dnscrypt-proxy`](https://github.com/DNSCrypt/dnscrypt-proxy) and [`dnsmasq`](https://thekelleys.org.uk/dnsmasq/doc.html) to provide:

- **DNS-over-HTTPS (DoH)** and **DNSCrypt v2** secure upstreams
- **Anonymized DNSCrypt** and **Oblivious DoH (ODoH)** support
- **Caching resolver** using `dnsmasq`
- **Split DNS**: forward selected domains to traditional DNS servers
- **Non-root execution** using `nobody` user
- Built image **from scratch** for minimal footprint
- Based on Debian packages unpacked with `dpkg -x` during build (no `apt install` needed at runtime)

---

## 🚀 Quick Start

Run the container with default ports and custom parameters:
```zsh
docker run --rm -p 53:53/udp -d edelux/dns-proxy
```

```zsh
docker run --rm -p 53:53/udp -d edelux/dns-proxy \
  --anonymized \
  --server=/ec2.internal/10.18.0.2 \
  --server=/amazonaws.com/10.18.0.2
```

This configuration enables:

- DNS queries to DoH providers via anonymizing relays
- Split DNS resolution for AWS internal domains via plain DNS (recommended for cloud infra)

---
## ⚙️ Configuration via Parameters
Configuration is handled at runtime using command-line flags:

| Flag | Description |
| :--- | --- |
| --server=     | Specifies a plain DNS server (e.g. --server=/amazonaws.com/10.18.0.2).  Recommended <br>for internal or cloud-specific domains. |
| --doh-server= | Defines the secure DoH or DNSCrypt v2 server. Supports DNSCrypt, DoH, Anonymized <br>DNSCrypt, and ODoH. |
| --doh-route=  | Specifies which anonymized resolver to use when querying DoH providers. |
| --nocache     | Disables all DNS caching. Useful for debugging or environments where caching is <br>not desirable. |
| --cachesize=  | Sets the maximum number of DNS entries to cache. Set to 0 to disable caching <br>entirely. |
| --anonymized  | Enables anonymized routing of DoH queries using relay resolvers. |

All parameters are optional and can be combined freely.

## ♻️ Default Settings

**dmsmasq**
```conf
no-poll
no-hosts
no-resolv
bogus-priv
user=nobody
cache-size=128
keep-in-foreground
server=127.0.0.1#5300
```

**dnscrypt-proxy**
```toml
listen_addresses = ['127.0.0.1:5300']
user_name = 'nobody'
keepalive = 30

server_names = ['cloudflare', 'odoh-cloudflare', 'wikimedia', 'nextdns', 'libredns', 'fdn', 'comss.one', 'bortzmeyer', 'scaleway-fr', 'anon-cs-berlin', 'anon-cs-ch', 'anon-cs-dc', 'anon-cs-fl']
lb_strategy = 'ph'
lb_estimator = true

log_level = 0
require_nolog = true
require_nofilter = true
ignore_system_dns = true

require_dnssec = true
dnscrypt_servers = true
odoh_servers = true
doh_servers = true

[sources]
  [sources.'public-resolvers']
    cache_file = '/var/cache/dnscrypt-proxy/public-resolvers.md'
    minisign_key = 'RWQf6LRCGA9i53mlYecO4IzT51TGPpvWucNSCh1CBM0QTaLn73Y7GFO3'
    refresh_delay = 72
    urls = ['https://raw.githubusercontent.com/DNSCrypt/dnscrypt-resolvers/master/v3/public-resolvers.md',
      'https://download.dnscrypt.info/resolvers-list/v3/public-resolvers.md']
  [sources.relays]
    cache_file = '/var/cache/dnscrypt-proxy/relays.md'
    minisign_key = 'RWQf6LRCGA9i53mlYecO4IzT51TGPpvWucNSCh1CBM0QTaLn73Y7GFO3'
    refresh_delay = 73
    urls = ['https://raw.githubusercontent.com/DNSCrypt/dnscrypt-resolvers/master/v3/relays.md',
      'https://raw.githubusercontent.com/DNSCrypt/dnscrypt-resolvers/master/v3/relays.md',
      'https://download.dnscrypt.info/resolvers-list/v3/relays.md']
```

---
### 📦 Architecture Support
- arm64
- riscv64
- ppc64le
- s390x
- amd64

### ✅ Use Cases
- Lightweight encrypted DNS proxy for secure-by-default setups
- Internal DNS resolver for cloud environments (e.g. AWS, GCP)
- Drop-in replacement for public resolvers in private infrastructure
- Self-hosted DNS gateway for IoT, edge, or containerized environments

### 🔐 Security & Footprint
- Runs as unprivileged user (nobody)
- Uses only statically unpacked system files
- No package manager, cron, or extra services
- No unnecessary binaries or language runtimes

### 🛠 Build Philosophy
- Based on Debian packages
- Runtime built from scratch
- Uses dpkg -x to extract only required files
- No runtime apt install or package manager
- Focused on minimalism, clarity, and reproducibility

### 📎 Links
[edelux/dns-proxy](https://hub.docker.com/repository/docker/edelux/dns-proxy)

### ✨ License
This project is released under the [`MIT`](https://github.com/edelux/dns-proxy#MIT-1-ov-file)

---
#### 🔁 Repository Renaming Notice
This project was formerly published as:
- [`Docker Hub:`edelux/dnsmasq](https://hub.docker.com/repository/docker/edelux/dnsmasq)
- [`GitHub:` edelux/dnsmasq](https://github.com/edelux/dnsmasq)


Support for new images will continue under the new name edelux/dns-proxy. The previous image and repository will remain available but will only mirror updates made to this project until December 31, 2025.
