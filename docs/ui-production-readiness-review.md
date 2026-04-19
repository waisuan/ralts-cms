# UI production readiness review

**Scope:** `ui/` (Next.js 16, React 19, TypeScript).  
**Purpose:** Surface risks and gaps before production, prioritize follow-up work, and separate **must-fix-now** items from **curated tech debt**.

This is a snapshot review (static analysis + targeted file reads), not a full penetration test or performance audit.

**See also:** [backend-production-readiness-review.md](./backend-production-readiness-review.md) and [backend-deployment.md](./backend-deployment.md) for the Go API.

**Completed (critical):** Legacy Next.js handlers `GET /api/users` and `POST /api/register` (file-based demo auth) were **removed** — they were unused by the app (real auth is `/api/v1/users` → Go backend). See §1.

**Completed (quick wins):** [ui-deployment.md](./ui-deployment.md) documents `NEXT_PUBLIC_API_BASE_URL` and rewrites; unused `Toast.tsx` (including `useToast` stub) **removed**; `MachineDetailLink` unused `stopRowClick` prop **removed**.

---

## Executive summary

- ~~**Highest priority:**~~ **Resolved:** The insecure file-based routes were removed; only `GET /api/health` remains under `ui/src/app/api/`.
- **Auth model:** Authorization is enforced in the UI by client state and **must** be enforced by the backend on every API call. No Next.js `middleware.ts` exists; treat the UI as a thin client only.
- **Consistency:** Machine detail URLs are centralized in `machineDetailHref` (including post-edit reload on the machine page). Remaining URL construction should follow the same helper.
- **UX / polish:** Widespread `alert()` for errors and verbose `console.log` in services are acceptable for an internal tool short-term but are poor fit for production UX and noisy in browser consoles.

---

## 1. Address before or at release (prioritized)

Items are ordered by **severity**, then **usefulness**. **Complexity** is noted (L = low, M = medium, H = high).

| # | Item | Severity | Complexity | Usefulness |
|---|------|----------|------------|------------|
| ~~1~~ | ~~**Legacy Next API routes (`/api/users`, `/api/register`)** — removed; handlers and their tests deleted.~~ | ~~**Critical**~~ | — | **Done** |
| 2 | **401 handling in `ApiClient`** — `isLoginEndpoint` treats any path **containing** `/users` as non-redirecting. That may be too broad (e.g. admin user APIs under `/api/v1/.../users/...`). **Risk:** stale session / confusing errors instead of clean logout. | Medium | M | High — narrow to explicit login/register paths; add tests. |
| ~~3~~ | ~~**Production environment** — document `NEXT_PUBLIC_API_BASE_URL` and rewrites.~~ See [ui-deployment.md](./ui-deployment.md). | ~~Medium~~ | — | **Done** |
| 4 | **`alert()` for failures** (CSV export, attachment download, session expiry) — blocks UI, inconsistent tone, poor a11y. | Low–medium | M | Medium — replace with inline error + optional toast system. |
| 5 | **Service-layer `console.log`** — many API wrappers log request/response metadata; can leak structure in prod and add noise. | Low | L | Medium — gate with `NODE_ENV === 'development'` or a tiny logger. |
| ~~6~~ | ~~**`useToast` / `Toast.tsx`** — dead code removed (nothing imported it).~~ | ~~Low~~ | — | **Done** |

**Already improved in-tree:** machine page hard reload after edit uses `machineDetailHref` so encoded serials match list links.

---

## 2. Security and trust boundaries (short)

| Topic | Notes |
|--------|--------|
| **JWT in `localStorage`** | Standard SPA pattern; vulnerable if XSS exists. No `dangerouslySetInnerHTML` found in `ui/src`. Keep CSP and dependency hygiene on the roadmap. |
| **Admin pages** | `/admin/users` and `/admin/events` redirect non-admins client-side; **backend must return 403/401** for admin APIs regardless of UI. |
| **CORS / cookies** | Not audited here; confirm Go API CORS and auth headers for production origins. |
| **`poweredByHeader: false`** | Already disabled in `next.config.ts` — good. |

---

## 3. Behavioral gaps and edge cases

| Gap | Notes |
|-----|--------|
| **Empty or weird serial numbers** | `machineDetailHref('')` yields `/machines/`; backend + `notFound()` should define behavior. Rare but worth an API contract note. |
| **Expanded table row vs cards** | Maintenance count in expanded row is plain text; cards link to machine detail — minor UX inconsistency. |
| **Session expiry** | Multiple modals use `alert` for “session expired”; centralized handling would reduce duplication. |
| **No Next middleware** | Deep links load client bundle; unauthenticated users see `LoginPage` after auth bootstrap — acceptable; crawlers and no-JS are out of scope unless you add SSR guards. |

---

## 4. Curated tech debt (later)

Work is **not** blocking release if the backend is authoritative. Sorted by theme.

### Simplification and duplication

- **Attachment download** — Similar try/catch + `alert` flows in `RecordCard`, `RecordsTable`, `MaintenanceTable`, `MaintenanceCard`, `MachineInfoCard`. Extract a small `useAttachmentDownload` or shared handler.
- **`ApiClient`** — Repeated error parsing between `request`, `postFormData`, `putFormData`; consider one internal `parseErrorResponse(response)` helper.
### Refactoring and types

- **`eslint-disable` for `any`** in `RecordsTable` / `MaintenanceTable` column defs — replace with typed column meta where feasible (TanStack Table patterns).
- **`eslint-disable` for `react-hooks/exhaustive-deps`** in admin pages and `DateRangePicker` — revisit dependencies or document why stale closures are safe.

### Observability and UX

- Structured logging (or silent prod) for API layer.
- Toast/notification system used consistently instead of `alert` (no `Toast` component in repo until you add one).
- Optional: OpenTelemetry / Web Vitals for production (if product requirements demand it).

### Testing

- **E2E:** Playwright smoke exists; expand critical paths: login, machine list, machine detail, maintenance CRUD (even one happy path).
- **Unit:** Good coverage in places (`machineRoutes`, hooks, filters); consider tests for 401 redirect behavior after §1 item 2 changes.

### Dependencies and tooling

- Stay on supported Next/React versions; run `npm audit` in CI for prod builds.
- Consider **security headers** (CSP, `X-Frame-Options`, etc.) via `next.config` or reverse proxy — not present in repo at review time.

---

## 5. Deployment checklist (minimal)

- [ ] `NEXT_PUBLIC_API_BASE_URL` points at the real Go API for each environment.
- [ ] Rewrites in `next.config.ts` behave as expected (no accidental routing to unintended Next handlers).
- [x] **Legacy `ui/src/app/api/register` and `ui/src/app/api/users` removed** (only `health` remains under `ui/src/app/api/`).
- [ ] Health: `GET /api/health` (or platform health) wired for load balancers — `ui/src/app/api/health/route.ts` exists.
- [ ] Smoke test after deploy: login, list machines, open machine detail, one maintenance action.

---

## 6. Suggested scoring for the backlog

When grooming tech debt, use a simple score:

**Score = usefulness / (complexity × risk_if_ignored)**

Examples:

- ~~Removing legacy file-based API routes~~ **done** — high score justified (high usefulness, low complexity, high risk if ignored).
- Replacing every `alert`: **medium score** (medium usefulness, medium work).
- Full toast framework: **lower priority** unless users complain.

---

*Document generated as part of pre-release UI review. Update as items are completed or superseded.*
