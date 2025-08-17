#!/usr/bin/env bash
set -euo pipefail
HOST=${HOST:-localhost}
PORT=${PORT:-8001}

# Create a new cart
RESP=$(curl -sS -i -X POST "http://${HOST}:${PORT}/cart" -H 'Content-Type: application/json')
echo "$RESP"

# Try to extract cartID from Location header or response body
CART_ID=$(echo "$RESP" | awk '/Location: \/cart\//{print $2}' | sed -E 's#.*/cart/([^[:space:]]+)#\1#' | tr -d '\r')
if [[ -z "${CART_ID}" ]]; then
  CART_ID=$(echo "$RESP" | sed -n 's/.*"id"\s*:\s*"\([^"]\+\)".*/\1/p' | head -n1)
fi

if [[ -n "${CART_ID}" ]]; then
  echo "\nExtracted CART_ID=${CART_ID}"
  echo "export CART_ID=${CART_ID}"
else
  echo "\nCould not auto-extract CART_ID. Please inspect the response above."
fi
