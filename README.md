# Ralts CMS

Go HTTP API and Next.js UI for machines, maintenance records, and attachments (PostgreSQL; S3-compatible storage).

## Prerequisites

- **Go** — same minor version as in `go.mod`
- **Docker** + Docker Compose — local PostgreSQL and optional LocalStack
- **Node.js** — for `ui/`
- **`migrate`** on your PATH — [golang-migrate CLI](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md) (`make migrate-up` uses it)

## Local development

```bash
make dev          # copies env.example → .env.development & .env.test if missing; starts dev + test Postgres + LocalStack
make migrate-up   # apply migrations to both local databases
make run          # API → http://localhost:8080
```

Web UI (separate terminal):

```bash
make ui-dev
```

(`make client` runs only `npm run dev` in `ui/` — use that after dependencies are installed.)

The UI calls the API at `http://localhost:8080` by default. Override with `NEXT_PUBLIC_API_BASE_URL` if needed.

**Optional**

- `make seed-dev` — sample data (machines + users)
- First admin user:  
  `APP_ENV=development go run cmd/cli/main.go -type admin -username <user> -password '<password>'`  
  (`cmd/cli` also supports audit queries and other helpers — see flags in `cmd/cli/main.go`.)

## Tests

```bash
make test-with-db
```

## Makefile quick reference

| Target | What it does |
|--------|----------------|
| `make run` / `make server` | Start API (`APP_ENV=development`) |
| `make ui-dev` | `npm install` in `ui/` if needed, then `npm run dev` |
| `make client` | `npm run dev` in `ui/` only |
| `make dev-with-db` | Wait for dev DB, then `make run` |
| `make db-dev-up` / `db-dev-down` | Dev Postgres only |
| `make migrate-up` / `make migrate-down` | Migrations on dev + test DBs |
| `make db-clean` | Remove dev/test Compose volumes |

Config is driven by `.env.development` / `.env.test` (see `env.example`).
