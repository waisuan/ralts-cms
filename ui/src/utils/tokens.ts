/**
 * Access + opaque refresh session storage and refresh (single in-flight request).
 */
const KEY_ACCESS = 'ralts_token';
const KEY_REFRESH = 'ralts_refresh';

let refreshInFlight: Promise<string | null> | null = null;

export function getAccessToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  return localStorage.getItem(KEY_ACCESS);
}

export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  return localStorage.getItem(KEY_REFRESH);
}

export function setSessionTokens(access: string, refresh: string | null | undefined): void {
  if (typeof window === 'undefined') {
    return;
  }
  localStorage.setItem(KEY_ACCESS, access);
  if (refresh) {
    localStorage.setItem(KEY_REFRESH, refresh);
  }
}

export function clearSessionAuthKeys(): void {
  if (typeof window === 'undefined') {
    return;
  }
  localStorage.removeItem(KEY_ACCESS);
  localStorage.removeItem(KEY_REFRESH);
}

/**
 * Public auth routes where a 401 must not trigger refresh or session redirect.
 */
export function isPublicAuthPath(endpoint: string): boolean {
  if (endpoint.includes('/users/login') || endpoint.includes('/auth/refresh')) {
    return true;
  }
  if (endpoint === '/api/v1/users' || endpoint.startsWith('/api/v1/users?')) {
    return true;
  }
  return false;
}

/**
 * Endpoints whose 401 must not end the session: either the status means
 * something other than "sign in again" (a wrong current password), or the call
 * is a background one whose failure the user should never be shown.
 */
export function shouldSuppressAuthRedirectOn401(endpoint: string): boolean {
  if (isPublicAuthPath(endpoint)) {
    return true;
  }
  if (endpoint.includes('/users/password')) {
    return true;
  }
  // The unread badge fetches on every navigation and is decorative; a stale
  // token here should not tear the page the user is reading out from under them.
  // Token refresh is still attempted, and the next real request decides.
  if (endpoint.includes('/notifications/unread-count')) {
    return true;
  }
  return false;
}

/**
 * Call POST /api/v1/auth/refresh; updates localStorage on success. Single-flight.
 */
export async function refreshAccessToken(
  getApiBase: () => string
): Promise<string | null> {
  if (refreshInFlight) {
    return refreshInFlight;
  }
  const rt = getRefreshToken();
  if (!rt) {
    return null;
  }
  refreshInFlight = (async () => {
    try {
      const base = getApiBase();
      const res = await fetch(`${base}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: rt }),
      });
      if (!res.ok) {
        return null;
      }
      const data = (await res.json()) as { token?: string; refresh_token?: string };
      if (data.token) {
        setSessionTokens(data.token, data.refresh_token ?? null);
        return data.token;
      }
      return null;
    } catch {
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}
