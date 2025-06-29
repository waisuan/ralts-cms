import { getPPMStatus, getPPMStatusLabel } from '../ppmUtils';
import { PPM_STATUSES } from '../constants';

describe('PPM Utils', () => {
  // Mock the current date to ensure consistent test results
  const mockDate = new Date('2024-06-29T12:00:00Z');

  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(mockDate);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('getPPMStatus', () => {
    it('should return null for empty date', () => {
      expect(getPPMStatus('')).toBeNull();
      expect(getPPMStatus(null as unknown as string)).toBeNull();
      expect(getPPMStatus(undefined as unknown as string)).toBeNull();
    });

    it('should return Overdue status for past dates', () => {
      const pastDate = '2024-06-01';
      const result = getPPMStatus(pastDate);

      expect(result).not.toBeNull();
      expect(result!.label).toBe(PPM_STATUSES.OVERDUE);
      expect(result!.color).toBe('bg-red-100 text-red-800');
    });

    it('should return Due status for today', () => {
      const todayStr = '2024-06-29';
      const result = getPPMStatus(todayStr);

      expect(result).not.toBeNull();
      expect(result!.label).toBe(PPM_STATUSES.DUE);
      expect(result!.color).toBe('bg-orange-100 text-orange-800');
    });

    it('should return Due Soon status for dates within 7 days', () => {
      const soonDate = '2024-07-03'; // 4 days from mock date
      const result = getPPMStatus(soonDate);

      expect(result).not.toBeNull();
      expect(result!.label).toBe(PPM_STATUSES.DUE_SOON);
      expect(result!.color).toBe('bg-yellow-100 text-yellow-800');
    });

    it('should return null for dates more than 7 days in the future', () => {
      const futureDate = '2024-07-30'; // More than 7 days from mock date
      const result = getPPMStatus(futureDate);

      expect(result).toBeNull();
    });

    it('should handle edge case of exactly 7 days', () => {
      const sevenDaysDate = '2024-07-06'; // Exactly 7 days from mock date
      const result = getPPMStatus(sevenDaysDate);

      expect(result).not.toBeNull();
      expect(result!.label).toBe(PPM_STATUSES.DUE_SOON);
    });

    it('should handle edge case of exactly 8 days', () => {
      const eightDaysDate = '2024-07-07'; // Exactly 8 days from mock date
      const result = getPPMStatus(eightDaysDate);

      expect(result).toBeNull();
    });
  });

  describe('getPPMStatusLabel', () => {
    it('should return null for empty date', () => {
      expect(getPPMStatusLabel('')).toBeNull();
    });

    it('should return only the label for valid dates', () => {
      const pastDate = '2024-06-01';
      const label = getPPMStatusLabel(pastDate);

      expect(label).toBe(PPM_STATUSES.OVERDUE);
    });

    it('should return Due label for today', () => {
      const todayStr = '2024-06-29';
      const label = getPPMStatusLabel(todayStr);

      expect(label).toBe(PPM_STATUSES.DUE);
    });

    it('should return Due Soon label for dates within 7 days', () => {
      const soonDate = '2024-07-03';
      const label = getPPMStatusLabel(soonDate);

      expect(label).toBe(PPM_STATUSES.DUE_SOON);
    });

    it('should return null for dates more than 7 days in the future', () => {
      const futureDate = '2024-07-30';
      const label = getPPMStatusLabel(futureDate);

      expect(label).toBeNull();
    });
  });
});
