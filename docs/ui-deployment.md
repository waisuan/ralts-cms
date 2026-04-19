# UI (Next.js) deployment notes

## Environment variables

| Variable | Required | Purpose |
|----------|----------|---------|
| `NEXT_PUBLIC_API_BASE_URL` | **Yes** in staging/production | Base URL of the **Go API** (scheme + host + port), no trailing slash. Used by `next.config.ts` to rewrite `/api/*` to the backend. |

Example:

```bash
NEXT_PUBLIC_API_BASE_URL=https://api.example.com
```

Local development defaults to `http://localhost:8080` when unset (see `ui/next.config.ts`).

## How API traffic reaches the backend

1. **Browser:** The client uses relative URLs such as `/api/v1/machines`. Next.js serves the app on the same origin, so requests go to the **UI origin** `/api/...`.

2. **Rewrites:** `next.config.ts` maps:

   ` /api/:path*  →  ${NEXT_PUBLIC_API_BASE_URL}/api/:path*`

   So `/api/v1/users/login` is proxied to the Go service, not handled by Next (except routes under `ui/src/app/api/`, currently only **`GET /api/health`**).

3. **Server-side / build:** Anything that runs during `next build` or RSC with an empty browser `baseURL` in `ApiClient` still relies on the same rewrite rules when requests originate from the Next server—ensure `NEXT_PUBLIC_API_BASE_URL` is set in CI and production so rewrites target the correct API.

## Health checks

- **UI process:** `GET /api/health` → `ui/src/app/api/health/route.ts` (served by Next).

## Related

- Broader risks and follow-ups: [ui-production-readiness-review.md](./ui-production-readiness-review.md)
