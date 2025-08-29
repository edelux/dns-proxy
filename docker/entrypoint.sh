#!/bin/sh

mkdir -p /run/conf
cp /etc/dnsmasq.conf /etc/dnscrypt-proxy/dnscrypt-proxy.toml /run/conf

for _ARG in "$@"; do
  case "$_ARG" in
    --*) ARG="${_ARG#--}" ;;
    -*)  ARG="${_ARG#-}" ;;
    *)   ARG="$_ARG" ;;
  esac

  case "$ARG" in
    doh-server=*)
      awk -v doh="${ARG#*=}" '/server_names =/{print "server_names = [\x27" doh "\x27]";next}{print}' /run/conf/dnscrypt-proxy.toml >/run/conf/.tmp && mv /run/conf/.tmp /run/conf/dnscrypt-proxy.toml
      ;;
    nocache)
      awk '/^cache-size/{ print "no-negcache\ncache-size=0";next }{print}' /run/conf/dnsmasq.conf >/run/conf/.tmp && mv /run/conf/.tmp /run/conf/dnsmasq.conf
      ;;
    cachesize=*)
      awk -v size="${ARG#*=}" '/^cache-size/{ $0 = "cache-size=" size }{print}' /run/conf/dnsmasq.conf >/run/conf/.tmp && mv /run/conf/.tmp /run/conf/dnsmasq.conf
      ;;
    server=*)
      echo $ARG >>/run/conf/dnsmasq.conf
      ;;
    anonymized)
      awk '/^\[sources\]/{print "[anonymized_dns]";print "routes = [";print "  { server_name='\''*'\'', via=['\''anon-relay-ams'\'', '\''anon-relay-par'\'', '\''anon-cs-md'\'', '\''anon-cs-fi'\'', '\''anon-cs-nl'\'', '\''anon-cs-pt'\'', '\''anon-ibksturm'\'', '\''anon-kama'\''] }";print "]\n";print "[sources]";next}{ print }' /run/conf/dnscrypt-proxy.toml >/run/conf/.tmp && mv /run/conf/.tmp /run/conf/dnscrypt-proxy.toml
      ;;
    doh-route=*)
      awk -v route="${ARG#*=}" '{sub(/server_name='\''[^'\'']+'\''/, "server_name='\''" route "'\''");print}' /run/conf/dnscrypt-proxy.toml >/run/conf/.tmp && mv /run/conf/.tmp /run/conf/dnscrypt-proxy.toml
      ;;
  esac
done

dnscrypt-proxy -config /run/conf/dnscrypt-proxy.toml &
dnsmasq --conf-file=/run/conf/dnsmasq.conf
