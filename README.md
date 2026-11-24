
# dns-proxy

A secure, minimal DNS resolver container with DNS-over-HTTPS (DoH), caching, and split DNS — ideal for cloud and local infrastructure.

## 🐧 About

`edelux/dns-proxy` is a lightweight container built from [Debian Stable](https://www.debian.org/releases/stable/) and designed for high-security, low-footprint DNS resolution.

This project began as a practical solution for securing DNS resolution on my own Linux laptop.  I wanted to reduce how much of my internet activity could be inferred from DNS queries by my Internet Service Provider (ISP) and the DNS resolver operator — often the same organization. Although most of my connections already use encrypted protocols such as TLS (historically known as SSL), SSH, and other secure transports, traditional DNS queries are typically transmitted without encryption. That makes domain lookups observable and potentially loggable, even when the application traffic itself is encrypted.

Encrypting DNS protects DNS queries in transit to the selected resolver, but **does not make internet browsing anonymous**. ISPs may still observe destination IP addresses, traffic patterns, and other metadata. The DNS resolver itself may also see queries, depending on the protocol and anonymization configuration. Anonymized DNSCrypt can separate the client IP address from the DNS query as seen by the upstream resolver when correctly configured with compatible relays; it is not the same thing as ordinary DoH.

Over time, this personal project evolved into a reusable DNS service that I also incorporate into infrastructure environments for organizations I advise as a Platform Engineer and Information Security specialist.  I hope it proves useful in your own work, too! ;-)

This image combines [`dnscrypt-proxy`](https://github.com/DNSCrypt/dnscrypt-proxy) and [`dnsmasq`](https://thekelleys.org.uk/dnsmasq/doc.html) to provide:

- **DNS-over-HTTPS (DoH)** and **DNSCrypt v2** secure upstreams
- **Anonymized DNSCrypt** and **Oblivious DoH (ODoH)** support
- **Caching resolver** using `dnsmasq`
- **Split DNS**: forward selected domains to traditional DNS servers
- **Non-root execution** using `nobody` user
- Runtime image built **from scratch** for a minimal footprint
- **Based on Debian** packages unpacked with `dpkg -x` during build (no `apt install` needed at runtime)
- Includes /var/lib/dpkg/list.txt with an inventory of extracted Debian packages and their versions, and /var/lib/dpkg/status with package metadata for auditing and vulnerability scanning.

---
## 🚀 Quick Start

For a local workstation, bind the published DNS port to the host loopback interface only:
```zsh
docker run --rm -p 127.0.0.1:53:53/udp -d edelux/dns-proxy
```

For infrastructure deployments, the original example remains available with the intended network exposure and split-DNS options:
```zsh
docker run --rm -p 53:53/udp -d edelux/dns-proxy \
  --anonymized \
  --server=/ec2.internal/10.18.0.2 \
  --server=/amazonaws.com/10.18.0.2
```

This configuration enables:

- Anonymized upstream resolution **where supported by the configured protocol and relay routes**
- Split DNS resolution for AWS, for example, internal domains via traditional DNS, as required by the cloud network

**Exposure warning:** `-p 53:53/udp` publishes the DNS port on host interfaces by default. Use this only when network access is deliberately controlled (for example, by host firewalls and cloud security groups). The loopback-bound command is preferred for a single laptop. Publishing on the host's `127.0.0.1` is distinct from the container's own loopback interface.

---
## 💻 Configure Your Operating System to Use the Local DNS Proxy
### Why configure the system DNS resolver?

Running the container does not automatically cause the operating system or applications to use it. The host's DNS configuration must point to the locally published resolver, so that or dinary system DNS queries are sent to `127.0.0.1:53` and then forwarded by `dnsmasq` to `dnscrypt-proxy` for encrypted upstream resolution (except domains explicitly configured for split DNS).

The loopback address is local to **each machine**: configuring `127.0.0.1` on another device will not make that device use the resolver running on your laptop. Start the container and verify its published port before switching system DNS, or name resolution may fail.

### Linux: `/etc/resolv.conf`

On systems where `/etc/resolv.conf` is managed directly, configure:

```conf
nameserver 127.0.0.1
```

For example, edit with `sudoedit /etc/resolv.conf`. **First check whether it is generated or symlinked** (`ls -l /etc/resolv.conf`); on many distributions, NetworkManager, `systemd-resolved`, or `resolvconf` owns this file and manual changes will be overwritten. Use the managing service's configuration instead. Avoid adding an external fallback resolver if your goal is to ensure that all ordinary system DNS queries use the local proxy.

### Debian and derivatives using `resolvconf`
If your system **actually uses the `resolvconf` package** to generate `/etc/resolv.conf`, edit:

```text
/etc/resolvconf/resolv.conf.d/base
```

Add:
```conf
nameserver 127.0.0.1
```

Then regenerate the file:
```zsh
sudo resolvconf -u
cat /etc/resolv.conf
```

Check the generated result: other DNS entries may be injected by DHCP, VPNs, or other `resolvconf` fragments. This method does not apply to every Debian/Ubuntu installation; many use `systemd-resolved` or NetworkManager instead.

### Linux with NetworkManager (including many RHEL-family systems)
Use the graphical **Network Settings → active connection → IPv4/IPv6 → DNS** controls to set the DNS server to `127.0.0.1` and disable automatic DNS for that connection, where appropriate. Reconnect and verify the active resolver. For a command-line setup, inspect your active profile with `nmcli connection show --active` and configure its `ipv4.dns`, `ipv4.ignore-auto-dns`, and corresponding IPv6 settings as required.

On RHEL, Fedora, Rocky Linux, AlmaLinux, and other NetworkManager-managed systems, prefer changing the active connection profile rather than editing generated `/etc/resolv.conf` directly.

### Linux with `systemd-resolved`
Configure DNS through `systemd-resolved` or its network manager rather than overwriting a managed symlink. For example, a system-wide drop-in can be created at `/etc/systemd/resolved.conf.d/dns-proxy.conf`:

```ini
[Resolve]
DNS=127.0.0.1
FallbackDNS=
```

Restart the service (`sudo systemctl restart systemd-resolved`) and inspect `resolvectl status`. Be aware that per-link DNS from DHCP/VPN and routing domains can override or supplement global settings; review these before assuming all DNS uses the proxy. A system stub address such as `127.0.0.53` in `/etc/resolv.conf` is normal when `systemd-resolved` is active and does not necessarily mean the proxy is unused.

### macOS
On macOS, open **System Settings → Network → your active connection → Details → DNS**, and add `127.0.0.1` as the DNS server. Remove or adjust competing servers only if appropriate for your network. The proxy must be reachable on the macOS host (for example, via Docker Desktop with the port published on loopback). VPNs, managed profiles, and scoped DNS resolvers may take precedence.

### Windows 10/11
pen **Settings → Network & Internet → your active Wi-Fi/Ethernet connection → DNS server assignment → Edit** and configure manual DNS to `127.0.0.1` where supported. Ensure Docker Desktop publishes UDP port 53 to the Windows host loopback and that no other service occupies the port. Corporate policies, VPN clients, browser-specific DNS settings, and IPv6 may affect which resolver is actually used.

### Verify and troubleshoot
```zsh
# Confirm Docker publishes UDP/53 on the host loopback
docker ps

# Query the local DNS proxy directly (requires dig)
dig @127.0.0.1 x.org

# Inspect resolver ownership/configuration on Linux
ls -l /etc/resolv.conf
cat /etc/resolv.conf
```

A successful direct `dig` test confirms the proxy responds, **not** that every application uses it. Test normal OS resolution separately (`getent hosts example.org` on Linux), and inspect `resolvectl status` where applicable. Some browsers enable their own secure DNS; VPNs may inject dedicated resolvers; and IPv6 DNS configuration may bypass the intended path. Check both IPv4 and IPv6 behavior. DNS encryption also does not prevent destination-IP or traffic-metadata observation.

**Availability:** This local setup creates a DNS dependency on the container and Docker daemon. If they stop, ordinary DNS resolution can fail. For production, consider health checks, restart policies, monitoring, controlled network access, and redundant resolvers. The examples use UDP/53 as supported by the documented container invocation; verify TCP/53 support before relying on TCP fallback or large DNS responses.

---
## 🐳 Docker Compose
For a reproducible deployment with additional container security controls, use the provided docker/compose.yaml configuration:

```zsh
docker compose -f docker/compose.yaml up -d
```

This configuration enables a read-only filesystem, drops unnecessary Linux capabilities, prevents privilege escalation, and binds DNS exclusively to the host loopback interface (127.0.0.1).

---
## ⚙️ Configuration via Parameters
Configuration is handled at runtime using command-line flags:

| Flag | Description |
| :--- | --- |
| `--server=`     | Specifies a plain DNS server (e.g. --server=/amazonaws.com/10.18.0.2).  Recommended <br>for internal or cloud-specific domains. |
| `--doh-server=` | Defines the secure DoH or DNSCrypt v2 server. Supports DNSCrypt, DoH, Anonymized <br>DNSCrypt, and ODoH. |
| `--doh-route=`  | Specifies which anonymized resolver to use when querying DoH providers. |
| `--nocache`     | Disables all DNS caching. Useful for debugging or environments where caching is <br>not desirable. |
| `--cachesize=`  | Sets the maximum number of DNS entries to cache. Set to 0 to disable caching <br>entirely. |
| `--anonymized`  | Enables anonymized routing of DoH queries using relay resolvers. |

All parameters are optional and can be combined as supported by the image.
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
### 📦 Package Inventory
The image includes Debian package metadata to support software inventory, security auditing, and vulnerability assessment.

Inspect the extracted package inventory:
```zsh
docker run --rm --entrypoint /usr/bin/cat edelux/dns-proxy /var/lib/dpkg/list.txt
```

Inspect the Debian package status metadata:
```zsh
docker run --rm --entrypoint /usr/bin/cat edelux/dns-proxy /var/lib/dpkg/status
```

These files provide package information without requiring a package manager in the runtime image.

---
### 📦 Architecture Support
- amd64
- arm64
- riscv64 
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
- Based on Debian packages
- Runtime built from scratch
- Uses dpkg -x to extract only required files
- No runtime apt install or package manager
- Focused on minimalism, clarity, and reproducibility

### 📎 Links
[edelux/dns-proxy](https://hub.docker.com/repository/docker/edelux/dns-proxy)

### ✨ License
This project is released under the [`MIT`](https://github.com/edelux/dns-proxy#MIT-1-ov-file)
