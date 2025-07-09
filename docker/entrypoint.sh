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
    server=*)
      echo $ARG >> /etc/dnsmasq.conf
      ;;
    anonymized)
      awk '/^#ANONYMIZED/{print "[anonymized_dns]";print "routes = [\n  { server_name='\''scaleway-fr'\'', via=['\''anon-relay-ams'\'', '\''anon-relay-par'\''] }\n]";next}{print}' /etc/dnscrypt-proxy/dnscrypt-proxy.toml >.tmp && mv .tmp /etc/dnscrypt-proxy/dnscrypt-proxy.toml
      ;; 
    doh-route=*)
      awk -v route="${ARG#*=}" '{sub(/server_name='\''[^'\'']+'\''/, "server_name='\''" route "'\''");print}' /etc/dnscrypt-proxy/dnscrypt-proxy.toml >.tmp && mv .tmp /etc/dnscrypt-proxy/dnscrypt-proxy.toml
      ;;
  esac
done

dnscrypt-proxy -config /etc/dnscrypt-proxy/dnscrypt-proxy.toml &
dnsmasq --conf-file=/etc/dnsmasq.conf
