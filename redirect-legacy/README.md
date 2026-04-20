# redirect-legacy

Thin Express app: deprecation notice + countdown, then redirect to the current Ralts CMS origin (`TARGET_ORIGIN`). Used when legacy DNS (`clown-cms.com` / `www`) is pointed at Railway instead of Heroku.

## Requirements

- Node.js 20+

## Run locally

```bash
cd redirect-legacy
npm install
cp .env.example .env   # optional; defaults match production target
PORT=3099 npm start
```

Environment variables are read from `.env` in this directory (via [dotenv](https://github.com/motdotla/dotenv)). You can skip `.env` and pass vars inline instead.

Open [http://localhost:3099/](http://localhost:3099/) — use **`localhost`**, not `127.0.0.1`, especially on WSL / Windows. You should see the countdown page, then a redirect to `https://ralts-cms.app/` (or whatever you set in `TARGET_ORIGIN`).

**Health check:**

```bash
curl -sS http://localhost:3099/health
```

**Immediate 301 (no HTML):** set `REDIRECT_DELAY=0`:

```bash
REDIRECT_DELAY=0 PORT=3099 npm start
curl -sSI -o /dev/null -w '%{http_code} %{redirect_url}\n' 'http://localhost:3099/old/path?a=1'
```

**Deep link preservation:** open `http://localhost:3099/login?next=%2F` — the redirect target should include `/login` and the query string.

## Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `PORT` | `8080` | Listen port |
| `HOST` | `0.0.0.0` | Bind address (`0.0.0.0` = reachable from Windows when Node runs in WSL2) |
| `TARGET_ORIGIN` | `https://ralts-cms.app` | Scheme + host only (no path) |
| `REDIRECT_DELAY` | `10` | Seconds before redirect. `0` = HTTP 301 immediately; else HTML + meta refresh + JS |
| `LEGACY_SITE_LABEL` | `clown-cms.com` | Copy only |
| `NEW_SITE_LABEL` | `ralts-cms.app` | Copy only |

Invalid `TARGET_ORIGIN` fails fast at process start.

### Security

Redirect targets are built from the request path and query under **`TARGET_ORIGIN`**. If a request path would resolve to **another origin** (e.g. scheme-relative `//example.com/...`), the service redirects to **`/`** on `TARGET_ORIGIN` instead, avoiding open redirects.

## From repo root (Makefile)

```bash
make legacy-redirect-dev
```

Override port: `make legacy-redirect-dev PORT=4000`.

Docker (optional):

```bash
docker build -t redirect-legacy -f redirect-legacy/Dockerfile redirect-legacy
docker run --rm -p 3099:8080 -e PORT=8080 redirect-legacy
```

## Railway

Set **Root Directory** to `redirect-legacy`, use [`railway.toml`](./railway.toml), add custom domains for the legacy hostnames, and set env vars in the dashboard. Health check path: `/health`.

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| **Connection refused** / browser can’t connect | Server not running, or it exited on boot. Run `node server.js` in a terminal and read the first lines. |
| **`redirect-legacy: invalid TARGET_ORIGIN`** | Fix `.env`: value must be a full URL with **no path**, e.g. `https://ralts-cms.app` (not `ralts-cms.app` alone). |
| **`port … is already in use`** | Another process (often an old `node server.js`) is bound to that port. Stop it or use `PORT=3100 npm start`. |
| **Windows browser can’t reach Node running in WSL2** | This server **listens on `0.0.0.0` by default** so Windows can forward to WSL. In the browser use **`http://localhost:<port>/`**. If it still fails, see Microsoft’s [Accessing network applications with WSL](https://learn.microsoft.com/en-us/windows/wsl/networking) (e.g. **mirrored** networking mode on Windows 11, or ensure the app binds to `0.0.0.0`, not only loopback). |
| **Cursor terminal opens `127.0.0.1` and it fails** | Known quirk: terminal link detection may prefer the numeric loopback. **Paste `http://localhost:<port>/` into Edge/Chrome manually**, or open the URL from the log line. |

Why not only change docs? WSL2 uses a NAT; binding **`127.0.0.1`** only accepts connections *inside* Linux, so **Windows → `localhost:3099` never reaches** your server. Binding **`0.0.0.0`** fixes that. Production (Railway/Docker) also expects an all-interfaces listen.
