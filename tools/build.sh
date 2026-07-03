#!/usr/bin/env bash
# Cross-compile the crypto-exchange demo/ops helper tools to linux/amd64 (pure Go,
# CGO off). These are STANDALONE programs (each its own module) — NOT part of the
# crypto-exchange app module. They are deployed to the server as systemd services:
#   simctl   :8091  start/stop/reset the 5 load generators (the FE buttons call this)
#   loadtest        the order generator; run as systemd units gen1..gen5
#   festatic :80    serves the built Vue frontend (dist/) with SPA fallback + no-cache
#   dbviewer :8090  read-only SQLite web viewer over /app/exg.db
#   wsprobe         debug: subscribes to WS channels (orderbook/user_data/sysstats) and prints
set -e
cd "$(dirname "$0")"
mkdir -p bin
for t in simctl loadtest festatic dbviewer wsprobe; do
  [ -f "$t/main.go" ] || continue
  echo "building $t -> bin/$t"
  ( cd "$t" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "../bin/$t" . )
done
echo "done. deploy: scp tools/bin/* app@<server>:/app/  (see systemd units exg/fe/dbviewer/simctl)"
