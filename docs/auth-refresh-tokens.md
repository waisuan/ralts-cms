# Authentication: access and refresh tokens

Ralts CMS uses a **short-lived access token** (JWT, sent as `Authorization: Bearer`) plus an **opaque refresh token** stored only on the client and as a **hash** in PostgreSQL. Together they support a **sliding session**: active users stay signed in without keeping a single long-lived JWT for a full day.

## Concepts

| Piece | What it is | Where it lives |
|--------|------------|----------------|
| **Access token** | JWT signed with `JWT_SECRET`; carries `entity_id` and `role`. | Browser: `localStorage` key `ralts_token`. Sent on every API request. |
| **Refresh token** | Random string (not a JWT); proves the browser may request new access tokens. | Browser: `localStorage` key `ralts_refresh`. Sent only to `POST /api/v1/auth/refresh`. |
| **Refresh row** | DB row: `user_id`, `token_hash`, `expires_at`, optional `revoked_at`, rotation metadata. | Table `refresh_tokens` (migration `000021`). |

The server **never stores** the raw refresh string—only `SHA-256(raw + separator + pepper)` (see `pkg/auth/refresh_token.go`).

## API

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `POST` | `/api/v1/users/login` | None | Returns `user`, `token` (access), `refresh_token` (opaque). |
| `POST` | `/api/v1/auth/refresh` | None (body carries refresh) | Validates refresh hash, rotates refresh, returns new `token` + `refresh_token`. |
| `POST` | `/api/v1/auth/logout` | None | Optional body `{ "refresh_token": "..." }`; revokes that row if still active. |
| All other `/api/v1/...` routes | | Bearer access JWT | Unchanged for callers. |

**Rotation:** Each successful refresh inserts a new row and revokes the previous one. Reusing an old refresh returns `401`.

**Password change:** After a successful `PUT /api/v1/users/password`, the server revokes **all** active refresh rows for that user so existing refresh values stop working (see `internal/handlers/users.go`).

## Environment variables

Loaded via `internal/deps/config.go` (`github.com/caarlos0/env`). Durations use Go-style strings (e.g. `1h`, `168h`).

| Variable | Default | Meaning |
|----------|---------|---------|
| `ACCESS_TOKEN_LIFETIME` | `1h` | Lifetime of the **access JWT**. When it expires, the UI may call refresh if a refresh token is present. |
| `REFRESH_TOKEN_LIFETIME` | `168h` (7 days) | Each login or successful refresh sets `expires_at` to **now + this duration** on the stored row (sliding window). If the user does not refresh before expiry, they must log in again. |
| `REFRESH_TOKEN_PEPPER` | *empty* | Secret material mixed into hashing the raw refresh token. If empty, the app uses `JWT_SECRET + ":ralts-refresh"`. Prefer a **dedicated** value in production so refresh hashing can be rotated independently of JWT signing. |
| `JWT_SECRET` | *(see config)* | Signs access JWTs only; unchanged semantically. Still required to be strong in production (see `validateConfig`). |

**Refresh handler guard:** If `RefreshPepper()` were ever empty (misconfiguration), `POST /auth/refresh` returns `500` (“Server configuration error”). In normal setups, the derived pepper from `JWT_SECRET` is non-empty.

## Database

- Migration: `db/migrations/000021_add_refresh_tokens_table.up.sql`
- Repository: `internal/refreshtokens/repository.go` (create, lookup by hash, rotate in a transaction, revoke, revoke all for user)

## Frontend behavior

- **`ui/src/utils/tokens.ts`** — stores keys, `refreshAccessToken()` (single in-flight request), path helpers so login/refresh/register and password-change errors do not mis-trigger global redirect.
- **`ui/src/utils/api.ts`** — `fetchWithAuth`: on `401`, attempt refresh once, then retry; shared error parsing for JSON and form uploads.
- **Logout** — reads refresh from storage, then clears tokens and user, and best-effort `POST /auth/logout`.

## Operations

- Apply migrations before enabling in production (`make migrate-up` or your pipeline).
- Existing users with only an old long-lived JWT in `localStorage` must **log in again** once to receive a refresh token.
- Consider rate limits on `/auth/refresh` and `/users/login` at the edge if the API is public.

## Related code

- Handlers: `internal/handlers/auth.go`, login in `internal/handlers/users.go`
- JWT helpers: `pkg/auth/auth.go`, `pkg/auth/refresh_token.go`
- Router: `internal/router/router.go` (public `auth/*` routes)
- Tests: `internal/handlers/auth_test.go`, `internal/refreshtokens/repository_test.go`, `internal/integration/auth_refresh_test.go` (build tag `integration`)

## See also

- [Backend deployment](./backend-deployment.md) — env var table
- [UI deployment](./ui-deployment.md)
