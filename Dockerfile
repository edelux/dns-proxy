# syntax=docker/dockerfile:1
# vim: set filetype=dockerfile
#----------------------------------------------------------------------------------------------------
FROM alpine

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
