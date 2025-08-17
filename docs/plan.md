# Flip-shop Improvement Plan

Source note: The file docs/requirements.md is not present in the repository. This plan synthesizes key goals and constraints from README.md (domain model, API behavior, DB semantics) and docs/tasks.md (prioritized checklist). It organizes the improvements by theme, providing rationale and intended outcomes. It honors the project’s guidelines: Go 1.16+, gorilla/mux routing, in-memory KV DB, and promotion pipeline initialization in main.go.

## 1. Architecture and API Consistency

1.1 Unify HTTP error handling
- Change: Remove legacy helpers in internal/route/route.go; use utils.AppServer ResponseErrorEntityUnproc/ResponseErrorNotfound/ResponseErrorServerErr everywhere.
- Rationale: Centralizes error mapping, ensures consistent status codes and payloads, and aligns with utils.AppServer capabilities. Reduces divergence and duplicate logic.
- Outcome: Predictable error responses (JSON, correct status), easier maintenance.

1.2 Standardize response headers
- Change: Set Content-Type: application/json on all responses (success and error). Provide a small helper for JSON responses (see 6.3).
- Rationale: Clients rely on consistent content types; prevents subtle parsing bugs.
- Outcome: Uniform API contract.

1.3 Normalize 404/422 mapping
- Change: purchase/remove map ErrCartNotFound → 404; ErrItemNotFound and domain validation errors → 422. Use AppServer helpers consistently.
- Rationale: Correct REST semantics improve client behavior and test clarity.
- Outcome: Clearer error taxonomy, better DX.

1.4 Add /health endpoint
- Change: GET /health returns 200 with uptime/version.
- Rationale: Operational readiness; enables container/orchestrator health checks.
- Outcome: Easier observability and automation hooks.

1.5 API versioning scaffold
- Change: Optionally introduce /v1 prefix while keeping existing routes for back-compat initially.
- Rationale: Future-proof API evolution without breaking clients.
- Outcome: Smoother iteration.

## 2. Domain and Data Layer (memdb and transactions)

2.1 Rename memdb constructor
- Change: NewDMemoryKVDatabase → NewMemoryKVDatabase; keep deprecated alias for compatibility.
- Rationale: Naming clarity reduces confusion.
- Outcome: Cleaner code without breaking current callers.

2.2 Document isolation and rollback behavior
- Change: Expand utils/memdb package comment: serializable isolation via mutex; no rollback.
- Rationale: Sets correct expectations for contributors and test authors.
- Outcome: Fewer transactional misunderstandings.

2.3 Optional: Copy-on-write per-transaction snapshot
- Change: Implement snapshot isolation to enable rollback on handler error.
- Rationale: Current no-rollback requires careful handler design; snapshots improve safety and testability.
- Outcome: Atomicity guarantees on error. This is larger in scope; can be phased.

2.4 Repository tests covering WithTx
- Change: Add unit tests validating read/store paths, ErrValueNotFound mappings, and transaction usage patterns.
- Rationale: Prevent regressions in data access semantics.
- Outcome: Confidence in transactional boundaries.

## 3. Handlers and Input Validation

3.1 Validate payloads for purchase/remove
- Change: Ensure sku non-empty and matches known format; qty > 0; use json.Decoder with DisallowUnknownFields.
- Rationale: Early validation produces clearer 422 errors and guards invariants.
- Outcome: Robustness and clearer client feedback.

3.2 Switch from ioutil.ReadAll to json.Decoder
- Change: Stream decode; detect malformed JSON early.
- Rationale: Modernize for Go 1.16+; reduce memory and improve error signaling.
- Outcome: Cleaner code, better error paths.

3.3 Early-return error flow and header discipline
- Change: Ensure headers written before body; no duplicate WriteHeader; extract JSON write helper.
- Rationale: Avoids subtle HTTP bugs and duplicated code.
- Outcome: Simpler, correct handlers.

## 4. Promotions Pipeline

4.1 Strengthen promotion tests
- Change: Add edge cases for free item multiples, percentage discount rounding/precision, and interactions across promotions affecting same SKU.
- Rationale: Promotions are order-sensitive; tests capture invariants and idempotency.
- Outcome: Stable, predictable discounts.

4.2 Submit handler short-circuit on Apply errors
- Change: Stop processing on first failing promotion; do not partially apply changes.
- Rationale: Ensures atomicity semantics for promotions; avoids mixed states.
- Outcome: Clear error behavior and simpler reasoning.

4.3 Purity of promotions where possible
- Change: Prefer pure calculation functions that return intended cart deltas; apply state mutation separately within the transaction.
- Rationale: Improves testability and composability.
- Outcome: Cleaner separation of concerns.

## 5. Testing Strategy

5.1 Route handler tests with httptest and utils.NewServer
- Cases: POST /cart happy path; PUT purchase happy path and 422 invalid qty/unavailable item; DELETE remove happy path and 422/404; PUT submit with promotions applied.
- Rationale: Verify end-to-end handler behavior without real network.
- Outcome: Confidence in HTTP layer and mappings.

5.2 Promotion unit tests (table-driven)
- Change: Expand existing tests as per 4.1; emphasize determinism and rounding rules.
- Rationale: Business rules coverage.
- Outcome: Prevent regressions.

5.3 Repository tests via WithTx
- Change: New tests constructing fresh memdb for isolation.
- Rationale: Validate persistence semantics.
- Outcome: Data integrity under test.

5.4 Benchmarks for critical operations
- Change: Benchmarks for reservation, purchase, submit.
- Rationale: Track performance trends.
- Outcome: Early detection of regressions.

## 6. Observability and Logging

6.1 Request-scoped logging with request ID
- Change: Generate/propagate X-Request-ID; log method, path, duration in AppServer.requestInterceptor.
- Rationale: Correlate logs and measure latency.
- Outcome: Easier debugging and tracing.

6.2 Structured logging
- Change: Minimal interface with logfmt or JSON; default to std log for simplicity.
- Rationale: Machine-parsable logs improve analysis and CI logs.
- Outcome: Better production readiness.

6.3 JSON response helper
- Change: utils helper to set content type, write status, marshal, and handle marshal errors consistently.
- Rationale: Reduce duplication and header mistakes.
- Outcome: Cleaner handlers and uniform responses.

## 7. Reliability and Server Lifecycle

7.1 HTTP server timeouts
- Change: Configure read/write timeouts on http.Server; consider context timeouts for handler operations that may block.
- Rationale: Prevent resource exhaustion and slowloris.
- Outcome: More resilient server.

7.2 Graceful shutdown and background cancellation
- Change: AppServer.Start to handle context cancellation for future background tasks; already supports SIGINT/SIGTERM graceful shutdown per README.
- Rationale: Future-proof lifecycle management.
- Outcome: Clean shutdown semantics.

7.3 Basic rate limiting middleware (optional, default off)
- Change: Add simple token bucket or leaky-bucket per-process option.
- Rationale: Guard against naive DoS.
- Outcome: Improved resilience.

## 8. Tooling, CI, and Quality Gates

8.1 CI Workflow (GitHub Actions)
- Change: Mirror ./ci.sh with go test ./..., race detector, coverage artifact upload, and go build.
- Rationale: Automated validation across PRs.
- Outcome: Consistent build/test signals.

8.2 Static analysis
- Change: Add go vet and staticcheck/golangci-lint; fix reported issues.
- Rationale: Catch common bugs early.
- Outcome: Higher code quality.

8.3 Coverage targets
- Change: Script/Makefile target for coverage with HTML; consider threshold gate.
- Rationale: Maintain test discipline.
- Outcome: Prevents coverage backsliding.

## 9. Documentation and Developer Experience

9.1 README enhancements
- Change: Expand API docs with examples, promotions list with examples, and error responses/codes.
- Rationale: Self-service onboarding for users and contributors.
- Outcome: Reduced support burden.

9.2 OpenAPI specification
- Change: Provide docs/openapi.yaml and link from README.
- Rationale: Contract clarity; enables tooling and client generation.
- Outcome: Stronger API governance.

9.3 Examples directory
- Change: curl scripts or Postman/Insomnia collection for local exercise.
- Rationale: Faster validation and demos.
- Outcome: Improved DX.

9.4 GoDoc coverage
- Change: Ensure public types/functions, especially exported error values, have comments.
- Rationale: Better IDE and pkg.go.dev documentation.
- Outcome: Easier navigation and usage.

## 10. Consistency, Naming, and Style

10.1 Naming harmonization
- Change: Review names across packages (e.g., CartStatus constants, repo interfaces). Replace magic store strings with typed constants from repos.
- Rationale: Reduce drift and confusion.
- Outcome: Consistent codebase.

10.2 Error wrapping
- Change: Use fmt.Errorf("...: %w", err) for wrapping.
- Rationale: Preserves error chains for diagnostics.
- Outcome: Better troubleshooting.

10.3 Early returns and dead code removal
- Change: Consistent early-return style; remove unreachable returns.
- Rationale: Readability and correctness.
- Outcome: Cleaner code.

## 11. Money and Correctness

11.1 Money representation
- Change: Clarify int64 cents usage; consider a Money type alias or small struct with helpers (add, multiply, format) and overflow checks.
- Rationale: Monetary correctness and safety.
- Outcome: Fewer arithmetic bugs and clearer intent.

11.2 Defensive promotion application
- Change: Ensure promotions don’t exceed reserved quantities; maintain invariants when promotions add items.
- Rationale: Prevents invalid cart states.
- Outcome: Safe application of promotions.

11.3 Submit invariants
- Change: Atomic application of promotions within a single transaction; combined with 4.2 short-circuiting.
- Rationale: No partial state commits (especially if rollback remains absent).
- Outcome: Stronger guarantees for finalization.

## 12. Configuration and Operability

12.1 Configurable server port and initial inventory
- Change: Read from env vars or config struct with defaults to preserve current behavior.
- Rationale: Easier deployment across environments.
- Outcome: Flexible operations without code changes.

12.2 Seed loader (optional)
- Change: JSON/CSV seed loader gated by env var; fallback to hardcoded defaults.
- Rationale: Repeatable environment setup for demos/tests.
- Outcome: Faster bootstrap.

## 13. Future-proofing and Concurrency Notes

13.1 Concurrency documentation
- Change: Document threading assumptions; add read-only getters with mutex if carts were shared across goroutines in future.
- Rationale: Preempts misuse as features evolve.
- Outcome: Safer concurrent use.

13.2 Router registration layering (optional)
- Change: Introduce RouterRegistrar to keep SetRoutes minimal.
- Rationale: Encapsulation and separation of concerns.
- Outcome: Easier route management and testing.

---

Implementation Phasing Recommendation:
- Phase 1 (low risk/high ROI): 1.1–1.3, 3.1–3.3, 6.3, 8.1, 8.2, 9.1, 10.2.
- Phase 2: 5.1–5.3, 4.1–4.2, 11.1–11.3, 12.1.
- Phase 3: 2.1–2.2, 6.1–6.2, 7.1, 7.3, 9.2–9.3, 10.1.
- Phase 4 (larger scope): 2.3 (snapshots/rollback), 4.3 (purity refactor), 12.2, 13.2, 5.4 benchmarks.

Risks and Mitigations:
- Changing error mappings may break clients: mitigate by documenting in README and OpenAPI; consider dual behavior behind a flag during transition.
- Introducing structured logging can affect log parsers: keep default simple logger; make structured format opt-in.
- Snapshot/rollback implementation complexity: gate behind build tag or feature flag; add comprehensive tests before enabling by default.
