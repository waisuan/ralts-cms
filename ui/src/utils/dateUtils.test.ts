import { backendDateToHtmlDate, htmlDateToBackendDate, isMachineDateUnset } from './dateUtils';

describe('isMachineDateUnset', () => {
  it('treats empty and whitespace as unset', () => {
    expect(isMachineDateUnset('')).toBe(true);
    expect(isMachineDateUnset('   ')).toBe(true);
    expect(isMachineDateUnset(null)).toBe(true);
    expect(isMachineDateUnset(undefined)).toBe(true);
  });

  it('treats Go zero / year-1 RFC3339 as unset', () => {
    expect(isMachineDateUnset('0001-01-01T00:00:00Z')).toBe(true);
    expect(isMachineDateUnset('0001-12-31T00:00:00Z')).toBe(true);
  });

  it('treats real calendar dates as set', () => {
    expect(isMachineDateUnset('2024-06-15T00:00:00Z')).toBe(false);
    expect(isMachineDateUnset('2007-12-04T00:00:00Z')).toBe(false);
  });
});

describe('backendDateToHtmlDate', () => {
  it('returns YYYY-MM-DD for RFC3339 input', () => {
    expect(backendDateToHtmlDate('2025-09-02T00:00:00Z')).toBe('2025-09-02');
  });

  it('returns empty for sentinel dates', () => {
    expect(backendDateToHtmlDate('0001-01-01T00:00:00Z')).toBe('');
    expect(backendDateToHtmlDate('0001-12-31T00:00:00Z')).toBe('');
  });

  it('returns empty string for empty or invalid input', () => {
    expect(backendDateToHtmlDate('')).toBe('');
    expect(backendDateToHtmlDate('not-a-date')).toBe('');
  });
});

describe('htmlDateToBackendDate', () => {
  it('returns ISO string for valid HTML date', () => {
    const out = htmlDateToBackendDate('2025-09-02');
    expect(out).toMatch(/^2025-09-02T00:00:00\.000Z$/);
  });

  it('returns empty string for empty or invalid input', () => {
    expect(htmlDateToBackendDate('')).toBe('');
    expect(htmlDateToBackendDate('99-99-99')).toBe('');
  });
});
