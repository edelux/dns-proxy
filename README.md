
# Lightweight DNS Proxy with dnsmasq and DNSCrypt-Proxy

A secure, minimal DNS resolver container with DNS-over-HTTPS (DoH), caching, and split DNS — ideal for cloud and local infrastructure.

## 🐧 About

`edelux/doh-proxy` is a lightweight container built from [Debian Testing](https://www.debian.org/releases/testing/) and designed for high-security, low-footprint DNS resolution.

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
docker run --rm -p 53:53/udp -d edelux/doh-proxy
```

```zsh
docker run --rm -p 53:53/udp -d edelux/doh-proxy \
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
| --server= | Specifies a plain DNS server (e.g. --server=/amazonaws.com/10.18.0.2).  Recommended <br>for internal or cloud-specific domains. |
| --doh-server= | Defines the secure DoH or DNSCrypt v2 server. Supports DNSCrypt, DoH, Anonymized <br>DNSCrypt, and ODoH. |
| --doh-route= | Specifies which anonymized resolver to use when querying DoH providers. |
| --anonymized | Enables anonymized routing of DoH queries using relay resolvers. |

All parameters are optional and can be combined freely.

---
### 📦 Architecture Support
- amd64
- arm64
- ppc64le
- s390x

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
- Based on Debian Testing packages
- Runtime built from scratch
- Uses dpkg -x to extract only required files
- No runtime apt install or package manager
- Focused on minimalism, clarity, and reproducibility

### 📎 Links
[`DockerHub Repository` edelux/doh-proxy](https://hub.docker.com/repository/docker/edelux/doh-proxy)

### ✨ License
This project is released under the [`MIT`](https://github.com/edelux/doh-proxy#MIT-1-ov-file)

---
#### 🔁 Repository Renaming Notice
This project was formerly published as:
- [`Docker Hub:`edelux/dnsmasq](https://hub.docker.com/repository/docker/edelux/dnsmasq)
- [`GitHub:` edelux/dnsmasq](https://github.com/edelux/dnsmasq)


Support for new images will continue under the new name edelux/doh-proxy. The previous image and repository will remain available but will only mirror updates made to this project until December 31, 2025.
