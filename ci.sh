#!/usr/bin/env bash
set -euo pipefail
path_local=$(pwd)

echo "Running go vet..."
go vet ./...

# Prefer golangci-lint if installed; fall back to staticcheck if available
if command -v golangci-lint >/dev/null 2>&1; then
  echo "Running golangci-lint..."
  golangci-lint run ./...
elif command -v staticcheck >/dev/null 2>&1; then
  echo "Running staticcheck..."
  staticcheck ./...
else
  echo "Skipping golangci-lint/staticcheck (not installed)."
fi

echo "Running Tests..."
go test ./...

echo "Running tests with coverage..."
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

echo "Building binary on ${path_local}..."
go build

echo "binary on ${path_local}/flip-shop"