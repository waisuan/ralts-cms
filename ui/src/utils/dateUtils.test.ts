import { backendDateToHtmlDate, htmlDateToBackendDate } from './dateUtils';

describe('dateUtils', () => {
  describe('backendDateToHtmlDate', () => {
    it('returns YYYY-MM-DD for RFC3339 input', () => {
      expect(backendDateToHtmlDate('2025-09-02T00:00:00Z')).toBe('2025-09-02');
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
});
