# Backend production readiness review

**Scope:** Go HTTP API (`cmd/web`, `internal/`, `pkg/`), PostgreSQL, S3-compatible attachments.  
**Purpose:** Same intent as [ui-production-readiness-review.md](./ui-production-readiness-review.md): surface risks before production, prioritize fixes, and separate **near-term** work from **curated tech debt**.

This is a snapshot review (static analysis + targeted reads of routing, config, middleware, auth, and handlers), not a full penetration test, load test, or formal threat model.

---

## Executive summary

- **Authz model:** Public routes are explicit (`POST /api/v1/users`, `POST /api/v1/users/login`). Everything under `/api/v1/` (except those) sits on a subrouter protected by **`AuthenticationMiddleware`**; **`/api/v1/admin/`** adds **`AdminOnlyMiddleware`**. Router tests assert 401 without a token and 403 for non-admin on admin routes—**do not rely on the UI alone for authorization**; this backend design matches that expectation.
- **Secrets:** `JWT_SECRET` is validated for non-development/non-test environments (`internal/deps/config.go`); the dev default is rejected in production-like `APP_ENV`.
- **Observability:** Request logging uses `log/slog` with JSON in non-development (`internal/deps/deps.go`); paths and query strings are logged—be mindful of sensitive query params in production log pipelines.
- **Gaps:** **CORS** allows any origin (`*`); **health** does not probe dependencies; **rate limiting** is absent on auth endpoints; **TLS** is assumed at the edge.

Operational detail: [backend-deployment.md](./backend-deployment.md).

### Pre-release: what to address **now** vs **later**

There is **no Tier‑1 “drop everything” code fix** in this review (unlike the UI’s removed insecure routes). **Before the first production deploy, you must still close the loop on operations and policy**, not necessarily new Go features.

**Address now (before go-live)** — mostly checklist + decisions; little or no code:

| Action | Why |
|--------|-----|
| Complete **§5 Deployment checklist** | `APP_ENV`, strong **`JWT_SECRET`**, **`DATABASE_URL`** (incl. TLS to DB), **S3** creds/IAM, **TLS** termination at the edge, smoke test. |
| **Decide where brute-force / abuse is handled** | App has no login rate limit. For production, either **document** WAF / API gateway / reverse-proxy limits in front of `POST .../login`, or schedule in-app throttling (§1 #3). |
| **Decide liveness vs readiness** | **`GET /health`** does not ping DB/S3. If you use Kubernetes or similar, define whether the current endpoint is enough for **liveness** only and whether you need a **readiness** probe (§1 #4) — can be platform config first, code second. |

**Short sessions (optional same sprint)** — small, contained follow-ups:

| Item | Effort |
|------|--------|
| **Duplicate CORS** in `router.go` vs `CORSMiddleware` — consolidate so OPTIONS and headers stay in sync (§4). | Low |
| **Extend [backend-deployment.md](./backend-deployment.md)** with your actual probe URLs and “rate limiting at edge” owner. | Low |

**Defer unless product/security requires it** — higher effort or can be covered outside the binary first:

| Item | Notes |
|------|--------|
| **CORS origin allowlist in code** (§1 #2) | Often acceptable to ship with `*` + Bearer token if origins are trusted and abuse is limited; allowlist when the API is directly exposed to browsers from fixed URLs. |
| **In-app rate limiting** | Prefer edge limits first; add `golang.org/x/time/rate` or similar if you need defense in depth. |
| **`/ready` + DB/S3 checks** | Required when orchestrator needs dependency-aware readiness; not required for single-node + manual ops. |
| **JSON error envelope, JWT TTL env, request IDs, metrics** | §4 tech debt. |

---

## 1. Address before or at release (prioritized)

Items are ordered by **severity**, then **usefulness**. **Complexity:** L = low, M = medium, H = high.

| # | Item | Severity | Complexity | Usefulness |
|---|------|----------|------------|------------|
| 1 | **Production config** — Confirm `APP_ENV`, `JWT_SECRET`, `DATABASE_URL`, and S3 settings for each environment. Fail-fast on DB is already in place; JWT validation for non-dev is already in code. Document who sets secrets (see [backend-deployment.md](./backend-deployment.md)). | Medium | L | **Very high** |
| 2 | **CORS `Access-Control-Allow-Origin: *`** — Acceptable for many Bearer-token SPAs, but any website can read responses from the browser **if** a user’s token is somehow used from that context. Prefer **origin allowlist** when the API is browser-facing and origins are known. | Medium | M | High |
| 3 | **Brute force / abuse on `POST /api/v1/users/login`** — No rate limiting or lockout in router/middleware. Pair with **WAF / API gateway limits** or application-level throttling for production. | Medium | M | High |
| 4 | **`GET /health` does not check DB or S3** — Fine for liveness; for Kubernetes-style **readiness**, add checks or a separate `/ready` endpoint so traffic is not sent to a process that cannot serve data. | Medium | M | High in orchestrated deploys |
| 5 | **Plain HTTP listener** — `http.Server` on `PORT` without TLS. **Expected** behind a reverse proxy; document TLS termination and forwarded headers for your platform. | Low (if proxied) | L | High (ops clarity) |
| 6 | **Error bodies** — Many paths use `http.Error` with plain text. Consistent but not a structured JSON error envelope; clients must not parse errors as JSON unless handlers set `Content-Type: application/json` (verify per handler where the UI depends on JSON errors). | Low | M | Medium for API consistency |

Nothing in this pass matched the UI’s former **critical** tier (e.g. unauthenticated credential dump). Re-run this review after large auth or routing changes.

---

## 2. Security and trust boundaries (short)

| Topic | Notes |
|--------|--------|
| **JWT** | HS256, secret from config; claims include `entity_id` and `role`. Expiry **24h** hardcoded in `pkg/auth/auth.go` — consider env-tunable TTL for policy/compliance. |
| **Passwords** | bcrypt with a custom salt + SHA-256 pre-hash to avoid bcrypt’s 72-byte limit; legacy roti migration path in `VerifyPassword`. |
| **Admin APIs** | Enforced server-side via `AdminOnlyMiddleware` after auth. |
| **SQL** | Repositories use parameterized queries via `pgx` patterns (review continues to apply on new queries). |
| **Uploads** | Attachments: size cap (`MaxFileSize`), extension allowlist (`handlers/attachments.go`) — good baseline; keep aligned with product limits. |
| **Audit** | Background audit service with configurable retention (`AUDIT_RETENTION_DAYS`, `AUDIT_CLEANUP_INTERVAL`). |

---

## 3. Behavioral gaps and edge cases

| Gap | Notes |
|-----|--------|
| **Idempotency** | Uploads and mutating operations are not generally idempotent keys—retries may duplicate work unless clients handle it. |
| **Pagination bounds** | `MAX_MACHINE_LIMIT` / `MAX_MAINTENANCE_LIMIT` cap abuse; ensure defaults are applied consistently on all list endpoints. |
| **Clock skew** | JWT `exp`/`iat` rely on server time; NTP on hosts is assumed. |
| **S3 / LocalStack** | `NewS3Client` supports static creds and custom endpoint—verify IAM or instance roles in real AWS vs dev. |

---

## 4. Curated tech debt (later)

Work is **not** automatically blocking release if operations compensate (edge TLS, rate limits, monitoring).

### Hardening and consistency

- **Rate limiting** — Login (and optionally registration) per IP / account.
- **CORS allowlist** — Config-driven allowed origins instead of `*`.
- **Health/readiness** — DB ping + optional S3 head in `/health` or `/ready`.
- **JSON error envelope** — Standard `{ "message": "...", "code": "..." }` for API consistency with the UI’s `handleApiError` expectations (verify current handler behavior before changing).
- **JWT TTL / refresh** — Short-lived access tokens + refresh flow if product requires.

### Code quality and operations

- **Duplicate CORS logic** — `CORSMiddleware` and the global `OPTIONS` handler in `router.go` both set permissive CORS headers—consolidate to one path to avoid drift.
- **`gorilla/mux` middleware order** — Auth middleware is applied with `api.Use` after route registration; tests prove 401 behavior—when adding routes, keep the same pattern and extend `router_test.go`.
- **Integration tests** — `go test -tags=integration` in CI for release candidates (Docker-dependent).

### Observability

- **Request ID** middleware for correlating logs with UI or gateway IDs.
- **Metrics** — Prometheus/OpenTelemetry if SLIs are required.

---

## 5. Deployment checklist (minimal)

- [ ] `APP_ENV` set appropriately; **`JWT_SECRET`** strong and unique per environment (not `your-jwt-secret-key`).
- [ ] **`DATABASE_URL`** points at production DB with appropriate `sslmode` / pool params.
- [ ] **S3** bucket and IAM (or compatible keys) scoped to least privilege for object CRUD used by the app.
- [ ] **TLS** terminated at load balancer / ingress; backend may stay HTTP on private network.
- [ ] **Health** probes: decide liveness vs readiness (`/health` vs extended check).
- [ ] **Smoke:** login, list machines, one authenticated mutating call, one attachment path if used.

---

## 6. Suggested scoring for the backlog

Same heuristic as the UI doc:

**Score = usefulness / (complexity × risk_if_ignored)**

Examples:

- Readiness check with DB: **high score** when running on Kubernetes.
- Origin-restricted CORS: **medium–high** if the API hostname is public.
- JSON error standardization: **medium** unless the UI already tolerates current responses.

---

*Snapshot review; update when auth, routing, or deployment assumptions change.*
