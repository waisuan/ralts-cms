'use strict';

require('dotenv').config();

const express = require('express');

const PORT = Number(process.env.PORT) || 8080;
// Default 0.0.0.0 so the server accepts connections from the Windows host when Node runs
// inside WSL2 (NAT). See https://learn.microsoft.com/en-us/windows/wsl/networking
const HOST = process.env.HOST || '0.0.0.0';
const TARGET_ORIGIN = (process.env.TARGET_ORIGIN || 'https://ralts-cms.app').replace(
  /\/$/,
  '',
);
const REDIRECT_DELAY = clampInt(process.env.REDIRECT_DELAY, 10, 0, 300);
const LEGACY_LABEL = process.env.LEGACY_SITE_LABEL || 'clown-cms.com';
const NEW_LABEL = process.env.NEW_SITE_LABEL || 'ralts-cms.app';

function assertValidTargetOrigin(origin) {
  let u;
  try {
    u = new URL(origin);
  } catch {
    console.error('redirect-legacy: invalid TARGET_ORIGIN (not a URL):', origin);
    process.exit(1);
  }
  if (u.protocol !== 'http:' && u.protocol !== 'https:') {
    console.error('redirect-legacy: TARGET_ORIGIN must be http or https');
    process.exit(1);
  }
  if (u.search || u.hash) {
    console.error('redirect-legacy: TARGET_ORIGIN must not include query or fragment');
    process.exit(1);
  }
  if (u.pathname !== '/' && u.pathname !== '') {
    console.error('redirect-legacy: TARGET_ORIGIN must be origin only (no path)');
    process.exit(1);
  }
}

assertValidTargetOrigin(TARGET_ORIGIN);

const allowedOrigin = new URL(TARGET_ORIGIN).origin;

function clampInt(raw, defaultVal, min, max) {
  const n = parseInt(String(raw), 10);
  if (Number.isNaN(n)) {
    return defaultVal;
  }
  return Math.min(max, Math.max(min, n));
}

/** Same-origin only — rejects scheme-relative and absolute URLs that would redirect off-site. */
function buildDestination(req) {
  const href = new URL(req.originalUrl || '/', `${TARGET_ORIGIN}/`).href;
  if (new URL(href).origin !== allowedOrigin) {
    return `${TARGET_ORIGIN}/`;
  }
  return href;
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

const app = express();
app.disable('x-powered-by');
app.set('trust proxy', 1);

app.get('/health', (_req, res) => {
  res.json({ ok: true, service: 'redirect-legacy' });
});

app.use((req, res, next) => {
  if (req.method !== 'GET' && req.method !== 'HEAD') {
    res.setHeader('Allow', 'GET, HEAD');
    res.status(405).send('Method Not Allowed');
    return;
  }
  next();
});

app.use((req, res) => {
  const dest = buildDestination(req);

  if (REDIRECT_DELAY === 0) {
    res.redirect(301, dest);
    return;
  }

  if (req.method === 'HEAD') {
    res.status(200).end();
    return;
  }

  const destEsc = escapeHtml(dest);
  const legacyEsc = escapeHtml(LEGACY_LABEL);
  const newEsc = escapeHtml(NEW_LABEL);
  const seconds = REDIRECT_DELAY;

  res.type('html').send(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Moving to ${newEsc}</title>
  <meta http-equiv="refresh" content="${seconds}; url=${destEsc}" />
  <style>
    :root { color-scheme: light dark; }
    body { font-family: system-ui, sans-serif; max-width: 36rem; margin: 2rem auto; padding: 0 1rem; line-height: 1.5; }
    a { color: inherit; }
    .btn { display: inline-block; margin-top: 1rem; font-weight: 600; }
    #count { font-variant-numeric: tabular-nums; font-weight: 700; }
  </style>
</head>
<body>
  <h1>${legacyEsc} is being retired</h1>
  <p>This site is moving. Please bookmark <strong>${newEsc}</strong> — that is the address for Ralts CMS going forward.</p>
  <p>You will be redirected in <span id="count">${seconds}</span> seconds.</p>
  <p><a class="btn" href="${destEsc}">Continue to ${newEsc} now</a></p>
  <script>
    (function () {
      var dest = ${JSON.stringify(dest)};
      var s = ${seconds};
      var el = document.getElementById('count');
      var t = window.setInterval(function () {
        s -= 1;
        if (s <= 0) {
          window.clearInterval(t);
          window.location.replace(dest);
          return;
        }
        el.textContent = String(s);
      }, 1000);
    })();
  </script>
</body>
</html>`);
});

const server = app.listen(PORT, HOST, () => {
  // eslint-disable-next-line no-console
  console.log(
    `redirect-legacy http://${HOST}:${PORT} → ${TARGET_ORIGIN} delay=${REDIRECT_DELAY}s | browser: http://localhost:${PORT}/`,
  );
});

server.on('error', (err) => {
  if (err.code === 'EADDRINUSE') {
    console.error(
      `redirect-legacy: port ${PORT} is already in use. Stop the other process or run PORT=3100 npm start`,
    );
  } else {
    console.error('redirect-legacy: listen error:', err);
  }
  process.exit(1);
});
