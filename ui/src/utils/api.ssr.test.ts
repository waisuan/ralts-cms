/**
 * @jest-environment node
 */

function jsonResponse(
  body: unknown,
  init: { ok?: boolean; status?: number; contentType?: string | null } = {}
) {
  const { ok = true, status = 200, contentType = 'application/json' } = init;
  return {
    ok,
    status,
    headers: {
      get: (name: string) => {
        if (name === 'content-type') return contentType;
        if (name === 'content-length') return null;
        return null;
      },
    },
    json: async () => body,
    text: async () => (typeof body === 'string' ? body : JSON.stringify(body)),
  };
}

describe('ApiClient (SSR)', () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    jest.resetModules();
  });

  afterEach(() => {
    global.fetch = originalFetch;
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
  });

  it('uses NEXT_PUBLIC_API_BASE_URL when window is undefined', async () => {
    process.env.NEXT_PUBLIC_API_BASE_URL = 'http://internal:8080';

    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ ok: true }));

    const { ApiClient } = await import('./api');
    const client = new ApiClient();
    await client.get('/api/v1/machines');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://internal:8080/api/v1/machines',
      expect.any(Object)
    );
  });

  it('falls back to localhost when env is unset', async () => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;

    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ ok: true }));

    const { ApiClient } = await import('./api');
    const client = new ApiClient();
    await client.get('/api/v1/machines');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/machines',
      expect.any(Object)
    );
  });

  it('getBlob prepends full base URL in SSR context', async () => {
    process.env.NEXT_PUBLIC_API_BASE_URL = 'http://internal:8080';

    const blobMock = { size: 42 };
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      blob: async () => blobMock,
    });

    const { ApiClient } = await import('./api');
    const client = new ApiClient();
    await client.getBlob('/api/v1/machines/SN1/attachments/file.pdf');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://internal:8080/api/v1/machines/SN1/attachments/file.pdf',
      expect.any(Object)
    );
  });
});
