#!/usr/bin/env bash
set -euo pipefail
HOST=${HOST:-localhost}
PORT=${PORT:-8001}
CART_ID=${CART_ID:-}
SKU=${SKU:-120P90}
QTY=${QTY:-1}

if [[ -z "${CART_ID}" ]]; then
  echo "CART_ID is required. Use export CART_ID=... or run create_cart.sh first."
  exit 1
fi

curl -i -X PUT "http://${HOST}:${PORT}/cart/${CART_ID}/purchase" \
  -H 'Content-Type: application/json' \
  -d "{\"sku\":\"${SKU}\",\"qty\":${QTY}}"
