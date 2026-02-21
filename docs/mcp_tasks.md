# MCP Plan Tasks Checklist

This checklist translates the MCP Integration Plan (docs/mcp_plan.md) into actionable tasks. Use it to track progress. Check items as you complete them. Non-invasive to flip-shop unless Phase 1.1 is pursued.

Legend: [ ] todo • [x] done • (!) blocked • (?) decision needed

## Phase 0: Prep
1. [x] Review docs/openapi.yaml and align expected MCP tool shapes (params, responses, examples).
   - Note: OpenAPI indicates POST /cart and mutation routes return full Cart objects. MCP plan updated to return { cart: Cart } for create/add/remove/submit to align. Params remain unchanged.
2. [x] Decide MCP server implementation language: Go.
   - [x] Confirm viable MCP SDK or plan minimal JSON-RPC+MCP framing. Decision: Proceed with Go; implement minimal JSON-RPC 2.0 + MCP framing if no mature SDK is selected.
3. [x] Define configuration contract: FLIP_SHOP_BASE_URL (default http://localhost:8001), FLIP_SHOP_TIMEOUT_MS (default 8000). Document defaults.
   - Documented in docs/mcp_plan.md (Section 2: Architecture Overview → Configuration) with defaults.

Acceptance criteria:
- [x] Decision recorded in this file and referenced in docs/mcp_plan.md.
- [x] Schemas in plan cross-checked against OpenAPI.

## Phase 1: Minimal viable MCP server (sidecar proxy)
4. [x] Scaffold Go MCP server project: cmd/flipshop-mcp/main.go and utils/mcp/ package for tool handlers.
5. [x] Add dependencies if needed: choose a Go MCP SDK or implement minimal JSON-RPC 2.0 + MCP framing; use net/http for HTTP proxying.
   - Decision: No external MCP SDK for now. Implement minimal JSON-RPC 2.0 + MCP framing in Go under utils/mcp, and use net/http for proxying to flip-shop.
6. [x] Implement config loader (env → config) with sane defaults and validation.
   - Implemented in utils/mcp/config.go; wired via cmd/flipshop-mcp/main.go and utils/mcp/server.go. Defaults: FLIP_SHOP_BASE_URL=http://localhost:8001, FLIP_SHOP_TIMEOUT_MS=8000. Validation on URL scheme/host and timeout > 0.
7. [x] Register tools and implement handlers:
   - [x] cart.create → POST /cart
   - [x] cart.purchase.add → PUT /cart/{cartID}/purchase
   - [x] cart.purchase.remove → DELETE /cart/{cartID}/purchase
   - [x] cart.submit → PUT /cart/{cartID}/status/submitted
8. [x] Define JSON Schemas for params/responses in tool declarations; include examples.
9. [x] Implement HTTP → MCP error mapping: 404→NOT_FOUND, 422→INVALID_ARGUMENT, else→INTERNAL (include server message/body when safe).
10. [x] Add basic logging (start/stop, tool invocations, latency) and graceful shutdown.

Acceptance criteria:
- [x] MCP server starts and lists tools; simple end-to-end flow works against local flip-shop.
- [x] Tools validate inputs and return structured errors.

## Phase 1.1: Optional flip-shop read endpoints (only if needed)
11. [x] Add GET /cart/{cartID} route in flip-shop. (Go)
    - [x] Handler uses repo.WithTx; returns JSON aligned with submit response.
    - [x] Tests with httptest: 200/404/422 cases.
12. [x] Add GET /items route in flip-shop (if not present).
   - [x] Handler lists seeded items; simple 200 test.

Acceptance criteria:
- [x] New endpoints documented in docs/openapi.yaml and README.
- [x] Handler-level tests pass locally and in CI.

## Phase 2: Testing and CI for MCP
13. [x] Unit tests in Go for MCP handlers using httptest and a mocked HTTP layer.
14. [x] Integration test: spin up flip-shop via `go run ./...` and the MCP server binary, then call tools via a Go test client.
    - Implemented as an in-memory integration using httptest with real routes to avoid flakiness; exercises create→add→submit via MCP.
15. [x] CI wiring:
    - [x] Extend existing CI/script to run Go tests (including MCP) and build cmd/flipshop-mcp.

Acceptance criteria:
- [x] CI green for Go jobs.
- [x] Integration test validates create→add→submit happy path.

## Phase 3: Packaging and Examples
16. [x] Provide `docs/mcp_usage.md` with setup/run instructions, env vars, and troubleshooting.
17. [x] Provide `examples/mcp/claude_desktop.json` showing client configuration.
18. [ ] Examples documenting typical tool calls: create cart, add/remove, submit.

Acceptance criteria:
- [x] Users can follow docs to run MCP server and exercise tools in a client.

## Phase 4: Hardening
19. [ ] Add HTTP client timeouts/retries with idempotency guidance (avoid retrying non-idempotent operations without safeguards).
20. [ ] Tighten schemas: no additionalProperties; stricter patterns; numeric bounds.
21. [ ] Structured logging and correlation IDs; propagate X-Request-ID if present.
22. [ ] Security notes: local-only trust model; document risks and future auth options.

Acceptance criteria:
- [ ] Negative tests for schema validation and timeout behavior.

## Phase 5: Optional enhancements
23. [ ] Tool: cart.promotions.preview (requires server-side support or simulate by dry-run if available).
24. [ ] Tool: cart.purchase.set (absolute quantity) if/when server adds endpoint.
25. [ ] Stream/notify support for long operations (not required in current scope).

## Project hygiene
26. [ ] Ensure go fmt and go vet are clean; add make targets or update ci.sh if needed.
27. [x] Link this checklist from docs/mcp_plan.md for discoverability.

Notes:
- Keep MCP server stateless; do not cache cart state beyond request/response to avoid drift.
- Prefer table-driven tests where applicable in Go.
