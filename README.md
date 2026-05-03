# Flip-shop

[![CI](https://github.com/gambarini/flip-shop/actions/workflows/ci.yml/badge.svg)](https://github.com/gambarini/flip-shop/actions/workflows/ci.yml)

A shopping cart REST API with a configurable promotion pipeline and an MCP server that exposes cart operations as AI tools.

## Quick start

```bash
make run          # build and start server at http://localhost:8001
make test         # run all tests
make race         # run tests with race detector
make ci           # full CI check: vet + tests + coverage + build
```

Or without Make:

```bash
go build -o flip-shop ./cmd/flip-shop && FLIPSHOP_CONFIG_FILE=./config.yaml ./flip-shop
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `FLIPSHOP_PORT` / `PORT` | `8001` | Server port |
| `FLIPSHOP_VERSION` | `dev` | Version string in `/health` |
| `FLIPSHOP_CONFIG_FILE` | *(unset)* | Path to YAML config file (items + promotions) |
| `FLIPSHOP_INVENTORY_JSON` | *(default items)* | JSON array to seed inventory; ignored when `FLIPSHOP_CONFIG_FILE` provides items |

A `config.yaml` is included with a full inventory and promotion set. Start with:

```bash
FLIPSHOP_CONFIG_FILE=./config.yaml ./flip-shop
```

## REST API

OpenAPI specification: `docs/openapi.yaml`

All responses are JSON (`Content-Type: application/json`).

### GET /health

```json
{"status":"ok","uptime_seconds":12,"version":"dev"}
```

### GET /items

Returns all available items.

### POST /cart

Create a new cart.

```bash
curl -s -X POST http://localhost:8001/cart
```

```json
{
  "CartID": "d6870d29-eb07-4a31-9469-abe898183a1c",
  "Purchases": {},
  "CartStatus": "Available",
  "Total": 0
}
```

### PUT /cart/{cartID}/purchase

Add an item to the cart.

```bash
curl -s -X PUT http://localhost:8001/cart/{cartID}/purchase \
  -H 'Content-Type: application/json' \
  -d '{"sku":"120P90","qty":3}'
```

### DELETE /cart/{cartID}/purchase

Remove a purchased item from the cart.

```bash
curl -s -X DELETE http://localhost:8001/cart/{cartID}/purchase \
  -H 'Content-Type: application/json' \
  -d '{"sku":"120P90","qty":1}'
```

### GET /cart/{cartID}

Fetch a cart by ID.

### PUT /cart/{cartID}/status/submitted

Submit the cart: applies all promotions and calculates the final total. Submitted carts are immutable.

```bash
curl -s -X PUT http://localhost:8001/cart/{cartID}/status/submitted
```

### Error responses

| Status | Meaning |
|---|---|
| `404` | Resource not found |
| `422` | Validation or domain error (invalid qty, item unavailable, etc.) |
| `500` | Unexpected server error |

All errors return `{"error":"<message>"}`.

## Promotions

Promotions are applied sequentially when a cart is submitted. They are configured via `config.yaml` or wired in code (`cmd/flip-shop/main.go`).

| Type | Config key | Description |
|---|---|---|
| Free item | `free_item` | Buy item A, get item B for free (once per unit of A purchased) |
| Qty price free | `qty_price_free` | Buy N of item A, get 1 free (every Nth unit is free) |
| Qty % discount | `qty_discount_percentage` | Buy ≥ N of item A, get X% off all units |
| Qty fixed discount | `qty_discount_fixed` | Buy ≥ N of item A, get a fixed amount off |
| Half price | `qty_half_price` | Buy N of item A, every Nth unit is half price |
| Tiered % discount | `qty_tiered_discount` | Quantity tiers, each with its own % discount |
| Bundle discount | `bundle_discount` | Buy all items in a set, get X% off each of them |
| Spend threshold | `cart_spend_threshold` | Cart total exceeds threshold → X% off entire cart |
| Cheapest item free | `cheapest_item_free` | Buy ≥ N distinct items, cheapest one is free |

## MCP server

`flipshop-mcp` is a second binary that exposes cart operations as MCP tools, proxying to the HTTP API.

```bash
go build -o flipshop-mcp ./cmd/flipshop-mcp
FLIPSHOP_MCP_BASE_URL=http://localhost:8001 ./flipshop-mcp
```

Available tools: `cart.create`, `cart.purchase.add`, `cart.purchase.remove`, `cart.submit`.

See `examples/mcp/claude_desktop.json` for Claude Desktop integration.

## Project layout

```
cmd/
  flip-shop/        # HTTP server binary (main + YAML config loader)
  flipshop-mcp/     # MCP server binary
internal/
  model/
    cart/           # Cart domain model and state machine
    item/           # Item domain model
    promotion/      # All promotion types (9 implementations)
  repo/             # KV-backed repositories (cart, item)
  route/            # HTTP handlers
utils/
  memdb/            # In-memory KV database (serializable isolation via mutex)
  mcp/              # MCP server implementation
  server.go         # HTTP server lifecycle + response helpers
  db.go             # KVDatabase interface and transaction types
static/             # Frontend (served at / and /static/)
config.yaml         # Example inventory and promotion configuration
```

## Architecture notes

**In-memory database** — `utils/memdb.MemoryKVDatabase` provides serializable isolation via a mutex. There is no rollback; handlers must write only after all validation passes.

**Promotion pipeline** — Promotions implement `promotion.Promotion` and run sequentially at submit time. Order matters; earlier promotions can affect what later ones see.

**Transaction pattern** — All mutations go through `repo.WithTx(func(tx utils.Tx) error { ... })`.

## Testing

```bash
make test         # go test ./...
make race         # go test -race ./...
make cover        # go test -cover ./...
make cover-html   # generate coverage.out and open HTML report
```

Tests use table-driven style for business rules, `httptest` for route handlers, and a fresh in-memory DB per repository test case.
