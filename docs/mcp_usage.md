# Flip-shop MCP Server Usage

Audience: Developers and users who want to run the Flip-shop MCP server locally and connect an MCP client (e.g., Claude Desktop) to it.

Overview
- The MCP server is a sidecar process that proxies MCP tool calls to the Flip-shop HTTP API.
- Source: cmd/flipshop-mcp (entrypoint) and utils/mcp (implementation).
- Default Flip-shop API base URL: http://localhost:8001

Prerequisites
- Go 1.16+
- This repository checked out locally

Build
- Build everything and run tests:
  - ./ci.sh
- Build only:
  - go build && ./flip-shop (Flip-shop)
  - go build -o flipshop-mcp ./cmd/flipshop-mcp

Run Flip-shop API
- Option A (from binary):
  - ./flip-shop
- Option B (without building):
  - go run ./...
- Default port is 8001. You can override via FLIPSHOP_PORT or PORT.

Run MCP Server
- After building: ./flipshop-mcp
- Environment variables (with defaults):
  - FLIP_SHOP_BASE_URL: Base URL of the Flip-shop HTTP API (default http://localhost:8001)
  - FLIP_SHOP_TIMEOUT_MS: HTTP client timeout in milliseconds (default 8000)

Tools Exposed (Summary)
- cart.create → POST /cart
- cart.purchase.add → PUT /cart/{cartID}/purchase
- cart.purchase.remove → DELETE /cart/{cartID}/purchase
- cart.submit → PUT /cart/{cartID}/status/submitted
- Optional read tools if present on server:
  - items.list → GET /items
  - cart.get → GET /cart/{cartID}

Quick Smoke Flow (via HTTP directly)
- Create cart: curl -s -X POST http://localhost:8001/cart | jq
- Add purchase: curl -s -X PUT http://localhost:8001/cart/$CARTID/purchase -d '{"sku":"120P90","qty":3}' -H 'Content-Type: application/json' | jq
- Submit: curl -s -X PUT http://localhost:8001/cart/$CARTID/status/submitted | jq

Troubleshooting
- Connection refused / timeouts:
  - Ensure Flip-shop API is running and FLIP_SHOP_BASE_URL points to it.
  - Increase FLIP_SHOP_TIMEOUT_MS if needed (e.g., 15000).
- 422 INVALID_ARGUMENT when adding items:
  - The SKU must exist and qty must be > 0. Check GET /items for available SKUs.
- 404 NOT_FOUND when submitting:
  - Validate cartID; it must be a valid UUID and must exist.

Client Example (Claude Desktop)
- See examples/mcp/claude_desktop.json
- Point Claude Desktop to your local flipshop-mcp binary.

Notes
- The MCP server is stateless; it does not cache cart state.
- Logs are written to stdout by default.
