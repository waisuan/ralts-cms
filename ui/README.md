# Ralts CMS — UI

Next.js (App Router) frontend for machine and maintenance management. It talks to the Go API (`NEXT_PUBLIC_API_BASE_URL`, default `http://localhost:8080`). See the [root README](../README.md) for backend setup.

## Prerequisites

- **Node.js 20+** (matches CI)
- **npm** (this repo uses `npm` / `package-lock.json` in `ui/`)

## Setup

```bash
cd ui
npm ci   # or npm install
```

### Environment

Create `ui/.env.local` if the API is not on the default host:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

JWT is stored in the browser as `ralts_token` after login (see `AuthContext` / `utils/api.ts`); there is no separate `NEXT_PUBLIC_*` key for the token name today.

## Run

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). Run the Go API on port 8080 (or set `NEXT_PUBLIC_API_BASE_URL`) so list/detail actions work.

## Scripts

| Command | Purpose |
| -------- | ------- |
| `npm run dev` | Dev server |
| `npm run build` | Production build |
| `npm start` | Serve production build |
| `npm run lint` | ESLint |
| `npm run format` / `npm run format:check` | Prettier |
| `npm test` | Jest + Testing Library (unit/component; `e2e/` is excluded) |
| `npm run test:e2e` | Playwright (needs `npm run build` first locally, or rely on `playwright.config` `webServer`) |

E2E uses Chromium by default; first run: `npx playwright install` (CI installs browsers in the workflow). Specs under `e2e/` seed `localStorage` for `ralts_user` and mock `**/api/v1/machines**` where needed so `/` is not stuck on `LoginPage`.

## API surface (reference)

Paths are relative to `NEXT_PUBLIC_API_BASE_URL`. The app uses `src/config/api.ts` and `src/services/*` (e.g. `machineService.ts`, `maintenanceService.ts`).

Examples:

- Machines: `/api/v1/machines` (list/create/update/delete by serial)
- Maintenance: `/api/v1/maintenance` (scoped by machine serial in the service layer)
- Users/auth: `/api/v1/users` and login/register flows

Exact routes match the Go handlers; prefer reading `*_service.ts` over duplicating paths here.

## Source layout

```
ui/src/
├── app/           # App Router: pages, layouts, route handlers
├── components/    # UI (`admin/`, `recordsList/` subfolders)
├── services/      # API clients (fetch + types)
├── hooks/
├── contexts/
├── config/
├── types/
└── utils/           # Includes `machineListFilters`, `bannerSearchLock`, `ppmUtils`, `formatters`
```

## Stack

- Next.js 16, React 19, TypeScript
- Tailwind CSS 4
- Jest + Testing Library for unit/component tests
- Playwright for E2E

## Troubleshooting

| Issue | What to try |
| ----- | ----------- |
| API errors / empty data | Confirm backend is up and `NEXT_PUBLIC_API_BASE_URL` matches |
| CORS | Configure the Go server to allow the UI origin |
| Auth | Inspect `ralts_token` / `ralts_user` in devtools; log in again if expired |
| Bad build | `rm -rf .next && npm run build` |

## Links

- [Next.js](https://nextjs.org/docs)
- [Testing Library](https://testing-library.com/)
- [Playwright](https://playwright.dev/)
