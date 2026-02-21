# MCP Integration Plan for Flip-shop API

Audience: Contributors implementing a Model Context Protocol (MCP) server that exposes flip-shop capabilities to AI clients (e.g., Claude Desktop, IDE agents). This plan outlines scope, design, and steps to deliver an MCP-compliant tool/service surface over the existing HTTP API without changing domain logic.

Status: Planning; no server code present yet.

Task checklist: see docs/mcp_tasks.md for actionable steps and acceptance criteria.


## 1. Goals and Scope

- Primary goal: Provide an MCP server exposing safe, well-typed tools that let AI clients create carts, add/remove items, submit (apply promotions), and retrieve cart state.
- Non-goals (phase 1):
  - Changing domain logic or HTTP surface of flip-shop.
  - Persistent storage beyond in-memory memdb.
  - Auth/session management (local, trusted development context).

- Success criteria:
  - MCP server starts and announces tools that mirror core flip-shop operations.
  - Tools include schemas, examples, and clear error mapping.
  - Round-trip tests exercise typical flows end-to-end through MCP → HTTP → flip-shop.


## 2. Architecture Overview

- Topology: Sidecar MCP server that proxies to a locally running flip-shop HTTP server (default http://localhost:8001).
- Language: Go. Decision recorded: prefer single-language repo and simplicity over TS tooling. See Appendix for prior TS notes.
- Process model: MCP server is a separate process invoked by MCP clients; it maintains no state beyond connection metadata and base URL config.
- Configuration: Base URL and timeouts via env vars.
  - FLIP_SHOP_BASE_URL (default http://localhost:8001)
  - FLIP_SHOP_TIMEOUT_MS (default 8000)


## 3. MCP Surface (Tools and Resources)

Expose the following MCP tools. For each, include JSON Schema for params and well-typed responses.

- tool: cart.create
  - POST /cart
  - params: {}
  - returns: { cart: Cart }  # Aligned with OpenAPI: server returns full Cart

- tool: cart.get
  - GET /cart/{cartID} (If not present in API, create a handler or simulate by composing repo reads; otherwise, provide cart.summary via submit handler output as a fallback.)
  - params: { cartID: string }
  - returns: { cart: Cart }
  - Note: If a GET endpoint does not exist now, Phase 1 can omit; Phase 1.1 adds a read handler in flip-shop.

- tool: cart.purchase.add
  - PUT /cart/{cartID}/purchase
  - params: { cartID: string, sku: string, qty: number }
  - returns: { cart: Cart }  # Aligned with OpenAPI

- tool: cart.purchase.remove
  - DELETE /cart/{cartID}/purchase
  - params: { cartID: string, sku: string, qty: number }
  - returns: { cart: Cart }  # Aligned with OpenAPI

- tool: cart.submit
  - PUT /cart/{cartID}/status/submitted
  - params: { cartID: string }
  - returns: { cart: Cart }  # Aligned with OpenAPI

- tool: items.list (optional if API has an items route; otherwise, add Phase 1.1 in flip-shop)
  - GET /items
  - params: {}
  - returns: { items: Item[] }

Error mapping: Map HTTP 404 to MCP tool error code NOT_FOUND; 422 to INVALID_ARGUMENT; 500/other to INTERNAL.


## 4. Schemas (Sketch)

Define JSON Schemas embedded in tool declarations:

- CartIDParam: {
  type: "object",
  required: ["cartID"],
  properties: { cartID: { type: "string", pattern: "^[0-9a-fA-F-]{36}$" } }
}
- PurchaseParam: {
  type: "object",
  required: ["cartID","sku","qty"],
  properties: {
    cartID: { $ref: "#/$defs/CartID" },
    sku: { type: "string", minLength: 1 },
    qty: { type: "integer", minimum: 1 }
  }
}
- SubmitParam: same as CartIDParam.
- Responses mirror the OpenAPI types where available (see docs/openapi.yaml). Reuse/align with that spec to avoid drift.


## 5. Implementation Plan

Phase 0: Prep
- Review docs/openapi.yaml to align tool shapes and examples.
- Decision: Go for the MCP server (see docs/mcp_tasks.md). Implement minimal JSON-RPC 2.0 + MCP framing or adopt a suitable Go MCP SDK.

Phase 1: Minimal viable MCP server
- Create new Go project under cmd/flipshop-mcp/ with utils/mcp/ for tool handlers.
  - main.go bootstraps an MCP server (JSON-RPC 2.0 + MCP framing) and registers tools.
  - Use net/http for proxying to flip-shop; standard library or a lightweight HTTP client.
- Implement tool registration and handlers for:
  - cart.create, cart.purchase.add, cart.purchase.remove, cart.submit
- Add configuration via env vars and simple validation.
- Map HTTP errors to MCP error categories and return structured error messages.
- Add examples metadata for each tool demonstrating typical params.

Phase 1.1: Optional HTTP read endpoints in flip-shop
- If not present, add GET /cart/{cartID} and GET /items and wire repos accordingly.
- Keep handlers small, use utils.AppServer helpers and repo.WithTx per project guidelines.

Phase 2: Testing and CI
- Add unit tests in Go for the MCP tool handlers using httptest and mocked HTTP responses.
- Add integration tests that spin up flip-shop via go run ./... and the MCP server binary, then call tools with a simple Go test MCP client.
- Update ci.sh to run MCP-related Go tests and build cmd/flipshop-mcp; keep CI Go-only.

Phase 3: Packaging and Examples
- Provide a claude_desktop.json example configuration referencing the local MCP server.
- Provide usage docs with end-to-end flow:
  - create → add purchase → submit → get totals
- Add troubleshooting section (connection refused, invalid schema, 422 mapping).

Phase 4: Hardening
- Timeouts/retries on HTTP calls with idempotency notes.
- Input sanitation and stricter JSON Schemas (DisallowUnknownFields equivalent at MCP boundary).
- Structured logging with correlation IDs (propagate X-Request-ID if flip-shop emits/accepts it later).

Phase 5: Optional Enhancements
- Tool: cart.promotions.preview to simulate submit without committing (requires domain support; otherwise skip).
- Tool: cart.purchase.set to set absolute qty (server currently supports add/remove only).
- Streaming progress tokens for long operations (not needed now).


## 6. Project Changes Required in flip-shop (only if Phase 1.1 pursued)

- Route: GET /cart/{cartID}
  - Handler: read cart from repo inside repo.WithTx; return JSON model consistent with submit response.
  - Tests: handler-level with httptest; 404 for missing, 422 for invalid ID format.

- Route: GET /items
  - Handler: list items from repo; already seeded in init().
  - Tests: simple 200 with array response.

No business logic modifications; only read endpoints for observability.


## 7. Security and Trust Model

- Local development only, no authentication in Phase 1.
- Warn users in docs that MCP tools can mutate state; ensure examples emphasize safe usage.
- Future: add API key or loopback-only binding if/when flip-shop exposes network beyond localhost.


## 8. Timeline and Milestones

- Week 1: Phase 0 + Phase 1 (minimal tools) ✓
- Week 2: Phase 2 (tests, CI) and Phase 3 (examples, docs)
- Week 3: Phase 4 hardening; Phase 1.1 read endpoints if needed


## 9. Deliverables

- mcp/ project with runnable MCP server
- Documentation: docs/mcp_usage.md (how to run, configure, and connect with a client)
- CI passing for Go tests (including MCP)
- Examples: sample tool calls and expected outputs


## 10. Appendix: Go-based MCP server notes

Go-based implementation notes:
- Use an MCP-compatible Go SDK (if available) or implement the minimal JSON-RPC 2.0 + MCP protocol framing.
- Keep the server in cmd/flipshop-mcp/ with utils/mcp/ for tool handlers.
- Testing mirrors the TS plan; use net/http for proxying to flip-shop.
