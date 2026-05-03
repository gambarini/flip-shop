#!/usr/bin/env bash
set -euo pipefail
path_local=$(pwd)

echo "Running go vet..."
go vet ./...

echo "Running Tests..."
go test -coverprofile=coverage.out ./...

# Enforce a minimal coverage threshold (e.g., 70%)
threshold=30.0
coverage=$(go tool cover -func=coverage.out | awk '/total:/ {print substr($3, 1, length($3)-1)}')
# Convert to float comparison using awk
ok=$(awk -v c=$coverage -v t=$threshold 'BEGIN {print (c+0 >= t+0) ? "yes" : "no"}')
echo "Total coverage: ${coverage}% (threshold ${threshold}%)"
if [ "$ok" != "yes" ]; then
  echo "Coverage threshold not met: ${coverage}% < ${threshold}%"
  exit 1
fi

echo "Building flip-shop binary..."
go build -o flip-shop ./cmd/flip-shop

echo "Building flipshop-mcp binary..."
go build -o flipshop-mcp ./cmd/flipshop-mcp

echo "Binaries: ${path_local}/flip-shop and ${path_local}/flipshop-mcp"