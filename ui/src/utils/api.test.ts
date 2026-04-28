import { ApiClient, ApiError, handleApiError } from './api';

jest.mock('./auth', () => ({
  redirectToLogin: jest.fn(),
  isAuthError: jest.fn(() => false),
}));

jest.mock('./tokens', () => ({
  getAccessToken: jest.fn(() => null),
  getRefreshToken: jest.fn(() => null),
  isPublicAuthPath: jest.fn(() => false),
  refreshAccessToken: jest.fn(async () => null),
  shouldSuppressAuthRedirectOn401: jest.fn(() => false),
}));

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

describe('ApiClient', () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
    jest.restoreAllMocks();
  });

  it('uses empty baseURL in browser context (window defined)', async () => {
    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ ok: true }));

    const client = new ApiClient();
    await client.get('/api/v1/machines');

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/machines',
      expect.any(Object)
    );
  });

  it('get merges query params and returns unwrapped JSON as data', async () => {
    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ hello: 'world' }));

    const client = new ApiClient('http://api.example');
    const res = await client.get('/v1/items', { a: '1', b: 'two' });

    expect(global.fetch).toHaveBeenCalled();
    const fetchOpts = (global.fetch as jest.Mock).mock.calls[0][1] as RequestInit;
    const h = fetchOpts.headers as Headers;
    expect(h.get('Content-Type')).toBe('application/json');
    expect(res.data).toEqual({ hello: 'world' });
  });

  it('returns wrapped shape when response has top-level data', async () => {
    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ data: { id: 5 }, meta: 'x' }));

    const client = new ApiClient('http://api.example');
    const res = await client.get('/v1/x');

    expect(res.data).toEqual({ id: 5 });
    expect((res as { meta?: string }).meta).toBe('x');
  });

  it('throws ApiError with default message for 404 when body has no message', async () => {
    global.fetch = jest.fn().mockResolvedValue(
      jsonResponse({}, { ok: false, status: 404 })
    );

    const client = new ApiClient('http://api.example');
    await expect(client.get('/missing')).rejects.toMatchObject({
      name: 'ApiError',
      status: 404,
      message: 'The requested resource was not found.',
    });
  });

  it('uses message from JSON error body when present', async () => {
    global.fetch = jest.fn().mockResolvedValue(
      jsonResponse({ message: 'Bad input' }, { ok: false, status: 400 })
    );

    const client = new ApiClient('http://api.example');
    await expect(client.get('/x')).rejects.toMatchObject({
      status: 400,
      message: 'Bad input',
    });
  });

  it('post sends JSON body', async () => {
    global.fetch = jest.fn().mockResolvedValue(jsonResponse({ ok: true }));

    const client = new ApiClient('http://api.example');
    await client.post('/v1/create', { name: 'n' });

    expect(global.fetch).toHaveBeenCalledWith(
      'http://api.example/v1/create',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'n' }),
      })
    );
  });

  it('delete skips JSON parse and returns undefined data on 2xx', async () => {
    const json = jest.fn();
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers: { get: () => null },
      json,
    });

    const client = new ApiClient('http://api.example');
    const res = await client.delete('/v1/a');
    expect(res.data).toBeUndefined();
    expect(json).not.toHaveBeenCalled();
  });

  it('getBlob fetches relative URL in browser context', async () => {
    const blobMock = new Blob(['binary'], { type: 'application/octet-stream' });
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      blob: async () => blobMock,
    });

    const client = new ApiClient();
    const blob = await client.getBlob('/api/v1/machines/SN1/attachments/file.pdf');

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/machines/SN1/attachments/file.pdf',
      expect.any(Object)
    );
    expect(blob).toBe(blobMock);
  });

  it('getBlob appends query params correctly', async () => {
    const blobMock = new Blob(['binary']);
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      blob: async () => blobMock,
    });

    const client = new ApiClient('http://api.example');
    await client.getBlob('/v1/download', { format: 'pdf', q: 'test' });

    expect(global.fetch).toHaveBeenCalledWith(
      'http://api.example/v1/download?format=pdf&q=test',
      expect.any(Object)
    );
  });
});

describe('handleApiError', () => {
  it('returns same ApiError instance', () => {
    const err = new ApiError('x', 500);
    expect(handleApiError(err)).toBe(err);
  });

  it('wraps unknown errors', () => {
    const out = handleApiError(new Error('boom'));
    expect(out).toBeInstanceOf(ApiError);
    expect(out.message).toBe('boom');
  });
});
