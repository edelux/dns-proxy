#!/bin/sh

for _ARG in "$@"; do
  case "$_ARG" in
    --*) ARG="${_ARG#--}" ;;
    -*)  ARG="${_ARG#-}" ;;
    *)   ARG="$_ARG" ;;
  esac

  case "$ARG" in
    doh-server=*)
      awk -v doh="${ARG#*=}" '/server_names =/{print "server_names = [\x27" doh "\x27]";next}{print}' /etc/dnscrypt-proxy/dnscrypt-proxy.toml >.tmp && mv .tmp /etc/dnscrypt-proxy/dnscrypt-proxy.toml
      ;;
    nocache)
      awk '/^cache-size/{ print "no-negcache\ncache-size=0";next }{print}' /etc/dnsmasq.conf >.tmp && mv .tmp /etc/dnsmasq.conf
      ;;
    cachesize=*)
      awk -v size="${ARG#*=}" '/^cache-size/{ $0 = "cache-size=" size }{print}' /etc/dnsmasq.conf >.tmp && mv .tmp /etc/dnsmasq.conf
      ;;
    server=*)
      echo $ARG >> /etc/dnsmasq.conf
      ;;
    anonymized)
      awk '/^\[sources\]/{print "[anonymized_dns]";print "routes = [";print "  { server_name='\''*'\'', via=['\''anon-relay-ams'\'', '\''anon-relay-par'\'', '\''anon-cs-md'\'', '\''anon-cs-fi'\'', '\''anon-cs-nl'\'', '\''anon-cs-pt'\'', '\''anon-ibksturm'\'', '\''anon-kama'\''] }";print "]\n";print "[sources]";next}{ print }' /etc/dnscrypt-proxy/dnscrypt-proxy.toml >.tmp && mv .tmp /etc/dnscrypt-proxy/dnscrypt-proxy.toml
      ;; 
    doh-route=*)
      awk -v route="${ARG#*=}" '{sub(/server_name='\''[^'\'']+'\''/, "server_name='\''" route "'\''");print}' /etc/dnscrypt-proxy/dnscrypt-proxy.toml >.tmp && mv .tmp /etc/dnscrypt-proxy/dnscrypt-proxy.toml
      ;;
  esac
done

dnscrypt-proxy -config /etc/dnscrypt-proxy/dnscrypt-proxy.toml &
dnsmasq --conf-file=/etc/dnsmasq.conf
