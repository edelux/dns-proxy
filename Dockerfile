# syntax=docker/dockerfile:1
#----------------------------------------------------------------------------------------------------
FROM alpine



LABEL org.opencontainers.image.title="Lightweight DNS proxy with dnsmasq and DNSCrypt"
LABEL org.opencontainers.image.description="A minimal Alpine-based container that runs dnsmasq with DNSCrypt-proxy as a local resolver for secure DNS queries over HTTPS (DoH)."
LABEL org.opencontainers.image.architecture="${TARGETARCH}"
LABEL org.opencontainers.image.supported.architectures="amd64,arm64,ppc64le,s390x,mips64le,riscv64,arm32v6,arm32v7,arm64v8,arm32v5,i386,windows-amd64"
LABEL org.opencontainers.image.platform="linux/${TARGETARCH}"
LABEL org.opencontainers.image.source="https://github.com/edelux/dnsmasq"
LABEL org.opencontainers.image.url="https://github.com/edelux/dnsmasq"
LABEL org.opencontainers.image.version="0.0"
LABEL org.opencontainers.image.authors="Ernie D'lux (edelux) EDH <edeluquez@hotmail.com>"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.created="2025-06-17T21:55:56Z"
LABEL org.opencontainers.image.documentation="https://github.com/edelux/dnsmasq#readme"
LABEL org.opencontainers.image.vendor="edelux"
LABEL org.opencontainers.image.ref.name="edelux/dnsmasq:0.0"



RUN apk --update --no-cache upgrade && apk --no-cache add dnscrypt-proxy dnsmasq && \
    printf '%s\n' "listen_addresses = ['127.0.0.1:5300']" \
    "server_names = ['cloudflare']" "user_name = 'nobody'" \
    "" "[sources]" "  [sources.'public-resolvers']" \
    "  url = 'https://download.dnscrypt.info/resolvers-list/v2/public-resolvers.md'" \
    "  cache_file = '/var/cache/dnscrypt-proxy/public-resolvers.md'" \
    "  minisign_key = 'RWQf6LRCGA9i53mlYecO4IzT51TGPpvWucNSCh1CBM0QTaLn73Y7GFO3'" \
    "  refresh_delay = 72" >/etc/dnscrypt-proxy/dnscrypt-proxy.toml && \
    printf '%s\n' '#!/bin/sh' '' 'ARGS=$@' '' \
    '/usr/bin/dnscrypt-proxy -config /etc/dnscrypt-proxy/dnscrypt-proxy.toml &' '' \
    '/usr/sbin/dnsmasq --keep-in-foreground --cache-size=1024 --no-poll --no-resolv --server=127.0.0.1#5300 $ARGS' \
    >/entrypoint.sh



ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
CMD ["--user=nobody"]
#----------------------------------------------------------------------------------------------------
# vim: set filetype=dockerfile tabstop=4 shiftwidth=4 expandtab
