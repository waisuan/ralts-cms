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

## Changelog and announcement banner

The in-app changelog is the route **`/changelog`** ([`src/app/changelog/page.tsx`](src/app/changelog/page.tsx)). Entries come from [`src/data/changelog.ts`](src/data/changelog.ts). The footer, user menu, and the optional top banner link there; the banner’s **What’s new?** control points at the same route.

### Temporary top banner

1. Set **`NEXT_PUBLIC_ANNOUNCEMENT_UNTIL`** to an **ISO 8601 UTC** instant (e.g. `2027-04-19T23:59:59.999Z`) in `ui/.env.local` for local dev, or in your host’s environment (e.g. Railway) for production.
2. While the current time is **before** that instant, a dismissible blue banner appears at the top of the app. Implementation: [`src/components/AnnouncementBanner.tsx`](src/components/AnnouncementBanner.tsx).
3. **Dismiss** is stored in `sessionStorage` keyed to the exact env string. Change the date (a new “campaign”) and users who dismissed the old one will see the banner again.
4. **Hide** the banner by unsetting the variable or using a time in the past.

Banner **wording** is in `AnnouncementBanner.tsx` today; update that component when the message should change for a release.

### Adding a changelog entry (“What’s new?”)

1. Open [`src/data/changelog.ts`](src/data/changelog.ts).
2. Add a new object at the **top** of the `changelogEntries` array (newest first; the page also sorts by `date`).
3. Use the existing entries as a template: `date` as `YYYY-MM-DD`, `title`, optional `summary`, and `sections` with `heading` + `items` (short bullets).

The file header comment in `changelog.ts` describes the types and intent.

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
