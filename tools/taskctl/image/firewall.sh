#!/bin/bash
# 暫定のファイアウォール: 起動時に許可ドメインを解決し、そのIPへの443番だけを許可する。
# DNS問い合わせは制限していないため、DNS経由の情報流出は防げない。
# dnsmasq によるDNSフィルタリングは #304 で実装し、このファイルを置き換える（ADR-035 決定2）。
set -euo pipefail

ALLOWED_DOMAINS=(
  api.anthropic.com
  registry.npmjs.org
  proxy.golang.org
  sum.golang.org
)
IPSET=taskctl-allowed

ipset create "$IPSET" hash:ip family inet

for domain in "${ALLOWED_DOMAINS[@]}"; do
  ips="$(getent ahostsv4 "$domain" | awk '{ print $1 }' | sort -u)"
  if [ -z "$ips" ]; then
    echo "firewall: failed to resolve $domain" >&2
    exit 1
  fi
  for ip in $ips; do
    ipset add -exist "$IPSET" "$ip"
  done
done

nameservers="$(awk '$1 == "nameserver" && $2 ~ /^[0-9.]+$/ { print $2 }' /etc/resolv.conf)"
if [ -z "$nameservers" ]; then
  echo "firewall: no IPv4 nameserver in /etc/resolv.conf" >&2
  exit 1
fi

iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT DROP

iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# ループバックは 127.0.0.1 宛てだけに限る（127.0.0.11 の Docker 内蔵DNSを経由させない）
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -d 127.0.0.1 -j ACCEPT

for ns in $nameservers; do
  iptables -A OUTPUT -p udp -d "$ns" --dport 53 -j ACCEPT
  iptables -A OUTPUT -p tcp -d "$ns" --dport 53 -j ACCEPT
done

iptables -A OUTPUT -p tcp --dport 443 -m set --match-set "$IPSET" dst -j ACCEPT

# 既定ポリシーの DROP だと、許可外の通信はタイムアウトまで固まる。CDN の IP が起動時の
# 解決結果から変わったときなどに原因が分かるよう、その場で接続エラーにする
iptables -A OUTPUT -p tcp -j REJECT --reject-with tcp-reset
iptables -A OUTPUT -j REJECT --reject-with icmp-port-unreachable

# ipset は IPv4 のみを扱うため、IPv6 はすべて拒否する
ip6tables -P INPUT DROP
ip6tables -P FORWARD DROP
ip6tables -P OUTPUT DROP
