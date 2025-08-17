#!/usr/bin/env bash
set -euo pipefail
HOST=${HOST:-localhost}
PORT=${PORT:-8001}

curl -i "http://${HOST}:${PORT}/health"
