# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build flip-shop binary
make run            # Build and run the server (port 8001)
make test           # go test ./...
make race           # go test -race ./...
make cover          # go test -cover ./...
make cover-html     # Generate coverage.out and open HTML report
make vet            # go vet ./...
make fmt            # go fmt ./...
make ci             # Run ./ci.sh (vet + tests + build — mirrors CI)

# Run a specific test
go test ./internal/route/... -run TestHandlerName
go test ./internal/model/promotion/... -run TestItemQtyPriceDiscountPromotion_Apply
```

## Architecture

### Two binaries

- `flip-shop` — HTTP REST server (entry: `cmd/flip-shop/main.go`)
- `flipshop-mcp` — MCP server that proxies to the HTTP API (entry: `cmd/flipshop-mcp/main.go`, implementation: `utils/mcp/`)

### Package layout

- `utils/` — cross-cutting: `AppServer` (gorilla/mux HTTP lifecycle + response helpers), `KVDatabase` interface, `memdb` implementation, money formatting, logger
- `internal/model/` — domain entities and business logic only (no I/O): `cart`, `item`, `promotion`
- `internal/repo/` — repositories backed by `utils.KVDatabase`; mutations require a `utils.Tx`
- `internal/route/` — HTTP handlers; each handler composes repo calls and domain logic within a single transaction
- `static/` — frontend (served at `/` and `/static/`)

### In-memory database

`utils/memdb.MemoryKVDatabase` provides serializable isolation via a mutex. **There is no rollback.** Handlers must not persist partial state — write only after all validation passes. Logical namespaces are identified by `utils.StoreName` (e.g., `repo.ItemStoreName`).

### Transaction pattern

```go
repo.WithTx(func(tx utils.Tx) error {
    // read, mutate, then store — all under the same tx
    return nil
})
```

### Promotion pipeline

Promotions implement `promotion.Promotion` and are assembled in `cmd/flip-shop/main.go`'s `initializeFunc`, then injected into the submit route. Ordering matters — promotions run sequentially.

### HTTP error helpers (use these in handlers)

- `srv.ResponseErrorEntityUnproc` → 422 (domain/validation errors)
- `srv.ResponseErrorNotfound` → 404
- `srv.ResponseErrorServerErr` → 500

### Testing conventions

- Table-driven tests for business rules (promotions, cart state machine)
- Route handler tests use `httptest` + `utils.NewServer` — no real listener
- Repository tests construct a fresh `memdb.MemoryKVDatabase` per test case and call through `repo.WithTx`
- Race detector (`make race`) is required to pass in CI

## Configuration (env vars)

| Variable | Default | Description |
|---|---|---|
| `FLIPSHOP_PORT` / `PORT` | `8001` | Server port |
| `FLIPSHOP_VERSION` | `dev` | Version string in `/health` |
| `FLIPSHOP_CONFIG_FILE` | *(unset)* | Path to YAML file defining items and promotions (see `config.yaml`) |
| `FLIPSHOP_INVENTORY_JSON` | *(default items)* | JSON array to seed inventory; ignored when `FLIPSHOP_CONFIG_FILE` provides items |