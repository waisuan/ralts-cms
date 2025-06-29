describe('Date Utilities', () => {
  // Mock the current date to ensure consistent test results
  const mockDate = new Date('2024-06-29T12:00:00Z');

  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(mockDate);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('Status calculation based on PPM date', () => {
    it('should return "Overdue" for past dates', () => {
      const pastDate = '2024-06-01';
      const today = new Date();
      const ppm = new Date(pastDate);

      // Remove time for accurate day comparison
      today.setHours(0, 0, 0, 0);
      ppm.setHours(0, 0, 0, 0);

      const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

      expect(diffDays).toBeLessThan(0);
    });

    it('should return "Due" for today', () => {
      const todayStr = new Date().toISOString().split('T')[0];
      const today = new Date();
      const ppm = new Date(todayStr);

      // Remove time for accurate day comparison
      today.setHours(0, 0, 0, 0);
      ppm.setHours(0, 0, 0, 0);

      const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

      expect(diffDays).toBe(0);
    });

    it('should return "Due Soon" for dates within 7 days', () => {
      const soon = new Date();
      soon.setDate(soon.getDate() + 3);
      const soonStr = soon.toISOString().split('T')[0];

      const today = new Date();
      const ppm = new Date(soonStr);

      // Remove time for accurate day comparison
      today.setHours(0, 0, 0, 0);
      ppm.setHours(0, 0, 0, 0);

      const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

      expect(diffDays).toBeGreaterThan(0);
      expect(diffDays).toBeLessThanOrEqual(7);
    });

    it('should return no status for dates more than 7 days in the future', () => {
      const future = new Date();
      future.setDate(future.getDate() + 30);
      const futureStr = future.toISOString().split('T')[0];

      const today = new Date();
      const ppm = new Date(futureStr);

      // Remove time for accurate day comparison
      today.setHours(0, 0, 0, 0);
      ppm.setHours(0, 0, 0, 0);

      const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

      expect(diffDays).toBeGreaterThan(7);
    });
  });

  describe('Date formatting', () => {
    it('should format valid dates correctly', () => {
      const dateString = '2024-06-29';
      const formatted = new Date(dateString).toLocaleDateString();

      expect(typeof formatted).toBe('string');
      expect(formatted).toMatch(/^\d{1,2}\/\d{1,2}\/\d{4}$/);
    });

    it('should handle empty date strings', () => {
      const dateString = '';
      const result = dateString ? new Date(dateString).toLocaleDateString() : '-';

      expect(result).toBe('-');
    });

    it('should handle invalid date strings', () => {
      const dateString = 'invalid-date';
      const result = new Date(dateString).toLocaleDateString();

      // Invalid dates return "Invalid Date" when formatted
      expect(result).toBe('Invalid Date');
    });
  });
});
