describe('API_CONFIG (browser)', () => {
  beforeEach(() => {
    jest.resetModules();
  });

  it('BASE_URL is empty string in browser context', async () => {
    const { API_CONFIG } = await import('./api');
    expect(API_CONFIG.BASE_URL).toBe('');
  });

  it('buildApiUrl returns relative path in browser context', async () => {
    const { buildApiUrl } = await import('./api');
    expect(buildApiUrl('/api/v1/machines')).toBe('/api/v1/machines');
  });
});
