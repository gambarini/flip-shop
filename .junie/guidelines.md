Flip-shop: Development Guidelines for Contributors

Audience: Advanced Go developers contributing to this repository. This document captures project-specific practices, build/test workflows, and architectural notes that will help you work effectively in this codebase.

1. Build and Configuration
- Language/Toolchain
  - Go module: github.com/gambarini/flip-shop
  - Minimum Go version: 1.16 (see go.mod). Using a newer Go version typically works; keep module compatibility in mind.
  - External deps: gorilla/mux for routing; gofrs/uuid for IDs.
- Build
  - Fast path: go build (from repository root). Produces flip-shop binary in the project directory (see ci.sh).
  - CI script: ./ci.sh runs go test ./... followed by go build. Prefer reproducing CI locally with this script when changing build/test wiring.
- Run (local server)
  - The app seeds an in-memory KV database at startup (see main.go init()) and starts an HTTP server on port 8001.
  - Run: ./flip-shop (after build) or go run ./...
  - Logging is stdout; shutdown is graceful on SIGINT/SIGTERM.
- HTTP API (quick reference)
  - POST /cart → create a cart.
  - PUT /cart/{cartID}/purchase → add item to cart. Body: {"sku":"120P90","qty":3}
  - DELETE /cart/{cartID}/purchase → remove item from cart. Body mirrors PUT.
  - PUT /cart/{cartID}/status/submitted → apply promotions and finalize cart.
  - Router: gorilla/mux via utils.AppServer.

2. Testing
- Running tests
  - All packages: go test ./...
  - Specific package: go test ./internal/model/promotion -run TestItemQtyPriceDiscountPromotion_Apply
  - With race detector: go test -race ./...
  - With coverage (text): go test -cover ./...
  - Coverage profile + HTML: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
- Project testing conventions
  - Table-driven tests are used extensively for business rules (see internal/model/promotion/*_test.go). Prefer this style for clarity and future expansion of cases.
  - Keep tests package-scoped to the code under test (e.g., package promotion) to exercise exported behavior; use helpers minimally.
  - Determinism: The in-memory DB does not inject randomness; tests in domain layers are deterministic. Avoid spinning up the HTTP server in unit tests; prefer testing domain logic directly.
- Adding new tests
  - File naming: *_test.go alongside the code.
  - Structure
    - For rule-like logic (promotions, cart operations), use table-driven tests with clear args/want/wantErr.
    - For repository logic, test through the repository with a memdb.Tx; use WithTx to ensure correct isolation. If you need a fresh DB, construct a new memdb.MemoryKVDatabase per test case.
    - For route handlers, prefer handler-level tests using httptest and utils.NewServer to register routes in-memory (do not start a real listener). Assert JSON payloads and status codes. Use AppServer.requestInterceptor behavior if relevant to logging/metrics, but don’t assert on logs.
  - Transactions: When testing code paths that require database writes, wrap operations in repo.WithTx to mirror production behavior. The memdb ensures serializable isolation by mutex; there’s no rollback, so design tests to cleanly create state per test.
- Demonstrated test flow (validated)
  - Baseline: go test ./... passes after fixing the invalid init signature in main.go (init must have no return values).
  - Example temporary smoke test that was validated locally:
    - File (temporary): temp_smoke_test.go
      - Content:
        - package main
        - func Test_NoOp_Smoke(t *testing.T) { t.Log("smoke ok") }
    - Execution: go test ./...
    - Cleanup: file removed after validation. Do not keep such smoke tests in the repo.

3. Architecture and Development Notes
- High-level layout
  - /utils: cross-cutting utilities: AppServer (HTTP lifecycle), KV DB interface, memdb implementation.
  - /internal/model: domain entities and business rules.
    - cart: cart lifecycle, purchases, totals, discounts; state machine: Available → Submitted.
    - item: item availability, reservation/release/remove semantics.
    - promotion: promotion interfaces and implementations (free item, price discount, qty-based freebies/discounts).
  - /internal/repo: repositories backed by utils.KVDatabase. Repos accept a Tx in methods that mutate state; prefer repo.WithTx to group operations atomically.
  - /internal/route: HTTP handlers. Each handler composes repo operations and domain logic within a transaction boundary and uses utils.AppServer response helpers for consistent error responses.
- In-memory DB (utils/memdb)
  - Isolation: Serializable via a mutex guarding the store; no partial interleavings.
  - No rollback: On handler errors inside a WithTx, business invariants are preserved by returning errors before committing writes. Keep this in mind: write ordering in tests/handlers should only persist after success.
  - Stores: logical namespaces identified by utils.StoreName; typical stores include "Items" and cart collections.
- Transaction usage pattern (production and tests)
  - repo.WithTx(func(tx utils.Tx) error { ... repo.Store(tx, model) ... })
  - Use the same tx to read/update multiple models; commit is implicit after the handler returns nil.
- HTTP error mapping
  - Prefer utils.AppServer.ResponseErrorEntityUnproc for domain/user input errors (422), ResponseErrorNotfound for missing resources, and ResponseErrorServerErr for unexpected issues (500). Some legacy helpers in internal/route/route.go mirror this signature; prefer the utils.AppServer methods in new code for consistency.
- Promotions pipeline (main.go)
  - Promotions are assembled at startup and injected into the submit route. If adding new promotions:
    - Implement promotion.Promotion interface.
    - Inject in main.go’s initializeFunc in the desired order. Consider idempotency and ordering if promotions interact.
- Code style
  - Go fmt/go vet should be clean. Avoid returning values from init() (invalid per language spec) and keep init minimal; prefer explicit initialization paths when feasible.
  - Error-first, early-return style in handlers; do not leak partial state.
  - Use domain errors for user-facing validations (e.g., item.ErrItemNotAvailableReservation) to map to 422.
- Local debugging tips
  - Run a server: go run ./... and exercise endpoints via curl/Postman.
  - Extract cartID from POST /cart response to compose subsequent calls.
  - Logging prints request method/URI per request via AppServer interceptor.

4. Common Pitfalls and Gotchas
- init() signature must be func init() with no args/returns. A prior invalid signature broke go test ./....
- Ensure all cart/item mutations happen inside a single WithTx call per request to avoid committing partial state as there’s no rollback.
- When adding routes, remember to register both method and path with AppServer.AddRoute, and set correct Content-Type on responses. Handlers should always write headers before body.
- Promotions in submit handler may call back into item reservations and cart purchases; ensure these are also wrapped in the transaction and persisted via itemRepo.Store.

5. Quick Commands Cheat Sheet
- Run tests (all): go test ./...
- Run tests with race: go test -race ./...
- Coverage (HTML): go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
- Build: go build && ./flip-shop
- Run without building: go run ./...
- CI-local: chmod +x ci.sh && ./ci.sh

If you extend the domain or HTTP surface, mirror patterns used in existing packages (table-driven tests, repo.WithTx, AppServer helpers) and update this document if the workflow materially changes.
