# Makefile for flip-shop
# Fast path targets mirror README and ci.sh

PROJECT_NAME := flip-shop
BIN          := $(PROJECT_NAME)
GO           ?= go

.PHONY: all help build run test race cover cover-html vet lint fmt tidy ci clean

all: build

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary (./$(BIN))"
	@echo "  run          - Build and run the server"
	@echo "  test         - Run unit tests"
	@echo "  race         - Run tests with race detector"
	@echo "  cover        - Run tests with coverage (text)"
	@echo "  cover-html   - Generate coverage.out and open HTML report"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint or staticcheck if available"
	@echo "  fmt          - go fmt all packages"
	@echo "  tidy         - go mod tidy"
	@echo "  ci           - Run repository CI script (./ci.sh)"
	@echo "  clean        - Remove built artifacts"

build:
	$(GO) build -o $(BIN) .

run: build
	./$(BIN)

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

cover:
	$(GO) test -cover ./...

cover-html:
	$(GO) test -coverprofile=coverage.out ./... && $(GO) tool cover -html=coverage.out

vet:
	$(GO) vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; golangci-lint run ./...; \
	elif command -v staticcheck >/dev/null 2>&1; then \
		echo "Running staticcheck..."; staticcheck ./...; \
	else \
		echo "No linter installed (golangci-lint/staticcheck). Skipping."; \
	fi

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

ci:
	chmod +x ./ci.sh && ./ci.sh

clean:
	rm -f $(BIN) coverage.out
