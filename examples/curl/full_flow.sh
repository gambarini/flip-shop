#!/usr/bin/env bash
set -euo pipefail
HOST=${HOST:-localhost}
PORT=${PORT:-8001}
SKU1=${SKU1:-120P90}    # Google Home
SKU2=${SKU2:-43N23P}    # MacBook Pro (example)

echo "== Health =="
./examples/curl/health.sh || true

echo "\n== Create cart =="
CREATE_OUT=$(./examples/curl/create_cart.sh)
echo "$CREATE_OUT"
CART_ID=$(echo "$CREATE_OUT" | sed -n 's/.*CART_ID=\([^\ ]\+\).*/\1/p' | tail -n1)
if [[ -z "${CART_ID}" ]]; then
  echo "Failed to extract CART_ID. Aborting."
  exit 1
fi
export CART_ID

echo "\n== Purchase items =="
SKU=${SKU1} QTY=3 ./examples/curl/purchase.sh
SKU=${SKU2} QTY=1 ./examples/curl/purchase.sh || true

echo "\n== Remove item (optional) =="
SKU=${SKU1} QTY=1 ./examples/curl/remove.sh || true

echo "\n== Submit cart =="
./examples/curl/submit.sh
