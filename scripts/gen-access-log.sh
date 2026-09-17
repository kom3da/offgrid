#!/usr/bin/env bash
# F3 などで使う、ダミーのアクセスログを標準出力に書き出す。
# 使い方: ./scripts/gen-access-log.sh [行数] > access.log
# IPアドレスは、文書や例示のために予約された範囲（RFC 5737）だけを使う。
# 乱数の種を固定しているので、同じ行数なら毎回同じ内容になる。
set -euo pipefail

lines="${1:-5000}"
if ! [[ "$lines" =~ ^[0-9]+$ ]]; then
  echo "usage: gen-access-log.sh [number of lines]" >&2
  exit 1
fi
RANDOM=42

paths=(/ /index.html /about /items /items/1 /items/2 /items/42 /login /api/items /api/items/7 /static/app.js /static/style.css /favicon.ico /admin /wp-login.php)
methods=(GET GET GET GET GET GET POST PUT DELETE)
statuses=(200 200 200 200 200 200 301 304 400 401 403 404 404 500 503)
agents=("Mozilla/5.0 (X11; Linux x86_64)" "Mozilla/5.0 (Windows NT 10.0; Win64; x64)" "curl/8.5.0" "ExampleBot/1.0")
prefixes=(192.0.2 198.51.100 203.0.113)

# 12 busy hosts. Host k is listed weights[k] times, so each one is clearly busier than the next
# and a "top 10" has an unambiguous order.
weights=(30 24 19 15 12 10 8 6 5 4 3 2)
busy=()
for ((k = 0; k < ${#weights[@]}; k++)); do
  for ((n = 0; n < weights[k]; n++)); do
    busy+=("${prefixes[k % 3]}.$((k * 7 + 10))")
  done
done

for ((i = 0; i < lines; i++)); do
  if ((RANDOM % 100 < 60)); then
    ip="${busy[RANDOM % ${#busy[@]}]}"
  else
    ip="${prefixes[RANDOM % 3]}.$((RANDOM % 100 + 150))"
  fi
  printf '%s - - [%02d/Sep/2026:%02d:%02d:%02d +0900] "%s %s HTTP/1.1" %s %d "-" "%s"\n' \
    "$ip" "$((i * 7 / lines + 1))" "$((RANDOM % 24))" "$((RANDOM % 60))" "$((RANDOM % 60))" \
    "${methods[RANDOM % ${#methods[@]}]}" "${paths[RANDOM % ${#paths[@]}]}" \
    "${statuses[RANDOM % ${#statuses[@]}]}" "$((RANDOM % 50000 + 200))" \
    "${agents[RANDOM % ${#agents[@]}]}"
done
