import { getPPMStatus } from '../ppmUtils';
import { htmlDateToBackendDate, backendDateToHtmlDate } from '../dateUtils';

describe('Date Utilities', () => {
  describe('htmlDateToBackendDate', () => {
    it('should convert HTML date to ISO format', () => {
      const htmlDate = '2026-02-24';
      const result = htmlDateToBackendDate(htmlDate);
      expect(result).toBe('2026-02-24T00:00:00.000Z');
    });

    it('should handle empty string', () => {
      const result = htmlDateToBackendDate('');
      expect(result).toBe('');
    });

    it('should handle invalid date', () => {
      const result = htmlDateToBackendDate('invalid-date');
      expect(result).toBe('');
    });
  });

  describe('backendDateToHtmlDate', () => {
    it('should convert ISO date to HTML format', () => {
      const isoDate = '2026-02-24T00:00:00.000Z';
      const result = backendDateToHtmlDate(isoDate);
      expect(result).toBe('2026-02-24');
    });

    it('should handle empty string', () => {
      const result = backendDateToHtmlDate('');
      expect(result).toBe('');
    });

    it('should handle invalid date', () => {
      const result = backendDateToHtmlDate('invalid-date');
      expect(result).toBe('');
    });
  });
});

describe('PPM Status calculation based on PPM date', () => {
  const today = new Date();
  const todayString = today.toISOString().split('T')[0]; // YYYY-MM-DD format

  it('should return "Overdue" for past dates', () => {
    const pastDate = new Date(today);
    pastDate.setDate(today.getDate() - 7);
    const status = getPPMStatus(pastDate.toISOString());
    expect(status?.label).toBe('Overdue');
  });

  it('should return "Due" for today', () => {
    const status = getPPMStatus(today.toISOString());
    expect(status?.label).toBe('Due');
  });

  it('should return "Due Soon" for dates within 7 days', () => {
    const futureDate = new Date(today);
    futureDate.setDate(today.getDate() + 3);
    const status = getPPMStatus(futureDate.toISOString());
    expect(status?.label).toBe('Due Soon');
  });

  it('should return no status for dates more than 7 days in the future', () => {
    const farFutureDate = new Date(today);
    farFutureDate.setDate(today.getDate() + 10);
    const status = getPPMStatus(farFutureDate.toISOString());
    expect(status).toBeNull();
  });
});

describe('Date formatting', () => {
  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return '-';
    }
  };

  it('should format valid dates correctly', () => {
    const dateString = '2024-01-15T10:30:00Z';
    const formatted = formatDate(dateString);
    expect(formatted).toBe('Jan 15, 2024');
  });

  it('should handle empty date strings', () => {
    const formatted = formatDate('');
    expect(formatted).toBe('-');
  });

  it('should handle invalid date strings', () => {
    const formatted = formatDate('invalid-date');
    expect(formatted).toBe('-');
  });
});
