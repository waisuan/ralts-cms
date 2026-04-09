/**
 * @jest-environment node
 */

describe('API_CONFIG (SSR)', () => {
  beforeEach(() => {
    jest.resetModules();
  });

  afterEach(() => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
  });

  it('BASE_URL uses env var when window is undefined', async () => {
    process.env.NEXT_PUBLIC_API_BASE_URL = 'http://internal:8080';

    const { API_CONFIG } = await import('./api');
    expect(API_CONFIG.BASE_URL).toBe('http://internal:8080');
  });

  it('BASE_URL falls back to localhost when env is unset', async () => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;

    const { API_CONFIG } = await import('./api');
    expect(API_CONFIG.BASE_URL).toBe('http://localhost:8080');
  });

  it('buildApiUrl prepends full base URL in SSR context', async () => {
    process.env.NEXT_PUBLIC_API_BASE_URL = 'http://internal:8080';

    const { buildApiUrl } = await import('./api');
    expect(buildApiUrl('/api/v1/machines')).toBe('http://internal:8080/api/v1/machines');
  });
});
