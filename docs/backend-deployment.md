# Backend (Go API) deployment notes

## Environment variables

Configuration is loaded from the process environment (`github.com/caarlos0/env`). When `APP_ENV` is set, `.env.<APP_ENV>` is loaded from the module root (see `internal/deps/config.go`).

| Variable | Notes |
|----------|--------|
| `APP_ENV` | `development` and `test` allow the default JWT placeholder; **production/staging must set a strong `JWT_SECRET`**. |
| `JWT_SECRET` | Required to be non-empty and **not** the dev default when `APP_ENV` is neither `development` nor `test`. |
| `DATABASE_URL` | Required for PostgreSQL (`internal/deps/pg.go` fails fast if empty). Use TLS query params for managed databases in production. |
| `PORT` | Listen address port (default `8080`). The binary serves **plain HTTP**; terminate TLS at a load balancer or reverse proxy. |
| `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` | Server timeouts (defaults in `Config`). |
| `AWS_*`, `AWS_S3_BUCKET_NAME` | S3-compatible storage for attachments; `AWS_ENDPOINT_URL` + `AWS_S3_FORCE_PATH_STYLE` for LocalStack-style endpoints. |
| `DEFAULT_MACHINE_LIMIT`, `MAX_MACHINE_LIMIT`, `DEFAULT_MAINTENANCE_LIMIT`, `MAX_MAINTENANCE_LIMIT` | Pagination caps. |
| `AUDIT_RETENTION_DAYS`, `AUDIT_CLEANUP_INTERVAL` | Audit log retention and background cleanup. |

See also: `env.example` at the repo root.

## Health and readiness

- **`GET /health`** (`internal/handlers/health.go`) returns JSON **without** checking PostgreSQL or S3. Use it as a **liveness** probe; add a separate **readiness** check (or extend `/health`) if orchestrators must wait for DB/S3 before traffic.

## CORS

- Global middleware sets `Access-Control-Allow-Origin: *` (`internal/middlewares/cors.go`). The API uses **Bearer JWT** (not cookies), but a public `*` origin is still broad—tighten to an allowlist if the API is exposed directly to browsers from fixed origins.

## Integration tests

- Heavy tests live under `internal/integration/` behind `-tags=integration` (see `internal/integration/doc.go`). Run in CI with Docker for pre-release confidence.

## Related

- [backend-production-readiness-review.md](./backend-production-readiness-review.md)
