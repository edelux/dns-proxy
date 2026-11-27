
# dns-proxy
A secure, minimal DNS resolver container with DNS-over-HTTPS (DoH), caching, and split DNS — ideal for cloud and local infrastructure.

## 🐧 About
`edelux/dns-proxy` is a lightweight container built from [Debian Stable](https://www.debian.org/releases/stable/) and designed for secure, low-footprint DNS resolution.

This project began as a practical solution for securing DNS resolution on my own Linux laptop.  I wanted to reduce how much of my internet activity could be inferred from DNS queries by my Internet Service Provider (ISP) and the DNS resolver operator — often the same organization. Although most of my connections already use encrypted protocols such as TLS (the successor to SSL), SSH, and other secure transports, traditional DNS queries are typically transmitted without encryption. That makes domain lookups observable and potentially loggable, even when the application traffic itself is encrypted.

Encrypting DNS protects DNS queries in transit to the selected resolver, but **does not make internet browsing anonymous**. ISPs may still observe destination IP addresses, traffic patterns, and other metadata. The DNS resolver itself may also see queries, depending on the protocol and anonymization configuration. Anonymized DNSCrypt can separate the client IP address from the DNS query as seen by the upstream resolver when correctly configured with compatible relays; it is not the same thing as ordinary DoH.

Over time, this personal project evolved into a reusable DNS service that I also incorporate into the IT infrastructure of organizations I advise as a Platform Engineer and Information Security Specialist. I hope you find it useful, whether for personal use or as part of your IT infrastructure!  ;-)

This image combines [`dnscrypt-proxy`](https://github.com/DNSCrypt/dnscrypt-proxy) and [`dnsmasq`](https://thekelleys.org.uk/dnsmasq/doc.html) to provide:

* **DNS-over-HTTPS (DoH)** and **DNSCrypt v2** secure upstreams
* **Anonymized DNSCrypt** and **Oblivious DoH (ODoH)** support
* **Caching resolver** using `dnsmasq`
* **Split DNS**: forward selected domains to traditional DNS servers
* **Unprivileged** DNS services running as the nobody user
* Runtime image built using Docker's **scratch** base image
* **Based on Debian** packages unpacked with `dpkg -x` during build (no `apt install` needed at runtime)
* **Package inventory**: includes `/var/lib/dpkg/list.txt` with an inventory of extracted Debian packages and their versions, and `/var/lib/dpkg/status` with package metadata to support security auditing and vulnerability scanning.

## 🚀 Quick Start
For a local workstation, bind the published DNS port to the host loopback interface only:

```zsh
docker run --restart always --name dns-proxy \
 -p 127.0.0.1:53:53/udp \
 -p 127.0.0.1:53:53/tcp \
 -d edelux/dns-proxy
```

For infrastructure deployments, the DNS proxy can be exposed to other systems and configured with split DNS and anonymized upstream resolution:

```zsh
docker run --restart always --name dns-proxy \
 -p 53:53/udp -p 53:53/tcp \
 -d edelux/dns-prox \
 --server=/amazonaws.com/10.18.0.2 \
 --server=/ec2.internal/10.18.0.2 \
 --anonymized
```

This configuration enables:
* Anonymized upstream resolution **where supported by the configured protocol and relay routes**
* Split DNS resolution for AWS, for example, internal domains via traditional DNS, as required by the cloud network

**Exposure warning:**  Publishing port 53 over TCP and UDP without specifying a host IP address exposes the DNS service on all host interfaces by default. Use this configuration only when network access is deliberately controlled through firewalls, cloud security groups, or equivalent network security controls. Avoid operating an unrestricted public DNS resolver.

## 🐳 Docker Compose
For a reproducible deployment with additional security controls and optimized resource usage, use the provided docker/compose.yaml configuration:

```zsh
docker compose -f docker/compose.yaml up -d
```

The Docker Compose configuration provides:
* **Read-only root filesystem** to minimize runtime modifications.
* **Linux capability restrictions** following the principle of least privilege.
* **Privilege escalation prevention** through no-new-privileges.
* **Resource limits** for CPU, memory, swap, and process/thread usage.
* **Temporary filesystems** with explicit size limits.
* **Log rotation and compression** to control disk usage.
* **Dedicated Docker bridge network** named dns-proxy.
* **Local-only DNS exposure** over TCP and UDP on port 53.
* **Automatic container restart** to improve service availability.

Monitor resource usage:
```zsh
docker stats dns-proxy
```

Inspect container logs:
```zsh
docker logs --tail 50 dns-proxy
```

Verify DNS resolution over UDP and TCP:
```zsh
dig @127.0.0.1 x.org
dig @127.0.0.1 x.org +tcp
```

The provided configuration is intended as a lightweight, hardened deployment for local workstations. Resource limits, network exposure, and split-DNS settings can be adjusted to meet infrastructure requirements.

## 💻 Configure Your Operating System to Use the Local DNS Proxy
### Why configure the system DNS resolver?
Running the container does not automatically cause the operating system or applications to use it. The host's DNS configuration must point to the locally published resolver, so that or ordinary system DNS queries are sent to `127.0.0.1:53` and then forwarded by `dnsmasq` to `dnscrypt-proxy` for encrypted upstream resolution (except domains explicitly configured for split DNS).

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

```zsh
nmcli connection show --active
```

```zsh
sudo nmcli connection modify "CONNECTION_NAME" \
  ipv4.dns "127.0.0.1" \
  ipv4.ignore-auto-dns yes \
  ipv6.ignore-auto-dns yes \
  ipv6.dns ""
```

```zsh
sudo nmcli connection up "CONNECTION_NAME"
```

### macOS
On macOS, open **System Settings → Network → your active connection → Details → DNS**, and add `127.0.0.1` as the DNS server. Remove or adjust competing servers only if appropriate for your network. The proxy must be reachable on the macOS host (for example, via Docker Desktop with the port published on loopback). VPNs, managed profiles, and scoped DNS resolvers may take precedence.

```zsh
networksetup -listallnetworkservices

sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1

networksetup -getdnsservers "Wi-Fi"

scutil --dns
```

```zsh
sudo networksetup -setdnsservers "Wi-Fi" Empty
```

### Windows 10/11
Open **Settings → Network & Internet → your active Wi-Fi/Ethernet connection → DNS server assignment → Edit** and configure manual DNS to `127.0.0.1` where supported. Ensure Docker Desktop publishes UDP port 53 to the Windows host loopback and that no other service occupies the port. Corporate policies, VPN clients, browser-specific DNS settings, and IPv6 may affect which resolver is actually used.

```powershell
Get-NetAdapter

Set-DnsClientServerAddress `
  -InterfaceAlias "Wi-Fi" `
  -ServerAddresses "127.0.0.1"

Get-DnsClientServerAddress -InterfaceAlias "Wi-Fi"

Clear-DnsClientCache
```

```powershell
Set-DnsClientServerAddress `
  -InterfaceAlias "Wi-Fi" `
  -ResetServerAddresses
```
### Verify and troubleshoot
```zsh
# Confirm Docker publishes TCP and UDP port 53
docker ps

# Query the local DNS proxy directly (requires dig)
dig @127.0.0.1 x.org
dig @127.0.0.1 x.org +tcp

# Inspect resolver ownership/configuration on Linux
ls -l /etc/resolv.conf
cat /etc/resolv.conf

sudo ss -lntup '( sport = :53 )'
```

A successful direct `dig` test confirms the proxy responds, **not** that every application uses it.  Test normal OS resolution separately (`getent hosts x.org` on Linux), and inspect `resolvectl status` where applicable.  Some browsers enable their own secure DNS; VPNs may inject dedicated resolvers;  and IPv6 DNS configuration may bypass the intended path.  Check both IPv4 and IPv6 behavior.  DNS encryption also does not prevent destination-IP or traffic-metadata observation.

**Availability:** This local setup creates a DNS dependency on the container and Docker daemon.  If they stop, ordinary DNS resolution can fail.  For production, consider health checks, restart policies, monitoring, controlled network access, and redundant resolvers.  The examples publish both UDP and TCP on port 53 to support standard DNS resolution, TCP fallback, and large DNS responses.  Verify connectivity over both protocols before configuring the operating system to use the local resolver.

## 🔐 DNS Privacy Considerations
Before configuring your operating system to use the local DNS proxy, ensure that the container is running and accessible on 127.0.0.1:53.

To minimize DNS leaks, avoid configuring external fallback DNS servers and review any DNS settings introduced by IPv6, VPN clients, network management services, or browsers with built-in encrypted DNS capabilities.

Bootstrap DNS queries may use traditional, unencrypted DNS during service initialization or when encrypted upstream resolvers are unavailable. Domains explicitly configured for split DNS are also forwarded through traditional DNS unless otherwise protected by the underlying network.

Keep in mind that encrypted DNS improves DNS privacy but does not provide full network anonymity or replace a VPN.

## ⚙️ Configuration via Parameters
Configuration is handled at runtime using command-line flags:

| Flag | Description |
| :--- | --- |
| `--server=`     | Specifies a traditional DNS upstream server, optionally for specific domains. Useful for split-DNS configurations. |
| `--doh-server=` | Specifies an encrypted upstream resolver supported by dnscrypt-proxy. |
| `--doh-route=`  | Specifies relay routes for compatible anonymized DNS configurations. |
| `--nocache`     | Disables DNS caching as implemented by the container. |
| `--cachesize=`  | Configures the DNS cache size. |
| `--anonymized`  | Enables anonymized DNS routing for compatible resolvers and relay configurations. |
| `--loglevel=`   | Sets the logging level: `DEBUG`, `INFO`, `NOTICE`, `WARNING`, `ERROR`, or `QUIET`. |

All parameters are optional and can be combined as supported by the image.

## ♻️ Default Settings
**dnsmasq**
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

## 📦 Package Inventory
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

## 📦 Architecture Support
* amd64
* arm64
* riscv64 
* ppc64le
* s390x

## ✅ Use Cases
* Lightweight encrypted DNS proxy for secure-by-default setups
* Internal DNS resolver for cloud environments (e.g. AWS, GCP)
* Drop-in replacement for public resolvers in private infrastructure
* Self-hosted DNS gateway for IoT, edge, or containerized environments

## 🔐 Security & Footprint
* DNS services run as the unprivileged nobody user.
* Minimal runtime filesystem built from Docker's scratch base image.
* Contains only the selected Debian packages and required runtime files.
* No runtime package manager, cron daemon, or unnecessary background services.
* Reduced attack surface through a minimal set of runtime dependencies.
* Supports read-only root filesystems and restricted Linux capabilities.
* Includes Debian package metadata for security auditing and vulnerability assessment.

## 🛠 Build Philosophy
* Based on official Debian Stable packages.
* Multi-stage build with a scratch runtime image.
* Uses dpkg -x to extract selected Debian packages without installing them into the runtime image.
* Includes only the runtime dependencies required by the selected components.
* Preserves package metadata for software inventory and vulnerability assessment.
* No runtime package installation or package manager.
* Focused on minimalism, security, transparency, and reproducibility.

## 📎 Links
* **GitHub:** [edelux/dns-proxy](https://github.com/edelux/dns-proxy)
* **DockerHub:** [edelux/dns-proxy](https://hub.docker.com/r/edelux/dns-proxy)

## ✨ License
This project is released under the [MIT License](LICENSE)
