import { getPPMStatus, getPPMStatusLabel } from './ppmUtils';
import { PPM_STATUSES } from './constants';

describe('getPPMStatus', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it.each([
    ['2024-06-20', PPM_STATUSES.OVERDUE],
    ['2024-06-25', PPM_STATUSES.DUE],
    ['2024-06-26', PPM_STATUSES.ALMOST_DUE],
    ['2024-07-25', null],
  ] as const)('for today 2024-06-25, %s → %s', (ppmDate, expectedLabel) => {
    jest.setSystemTime(new Date('2024-06-25T12:00:00.000Z').getTime());
    const info = getPPMStatus(ppmDate);
    if (expectedLabel === null) {
      expect(info).toBeNull();
    } else {
      expect(info?.label).toBe(expectedLabel);
    }
  });

  it('returns null for empty ppm_date', () => {
    jest.setSystemTime(new Date('2024-06-25T12:00:00.000Z').getTime());
    expect(getPPMStatus('')).toBeNull();
  });

  it('returns null for Go zero / sentinel ppm_date', () => {
    jest.setSystemTime(new Date('2024-06-25T12:00:00.000Z').getTime());
    expect(getPPMStatus('0001-01-01T00:00:00Z')).toBeNull();
    expect(getPPMStatus('0001-12-31T00:00:00Z')).toBeNull();
  });
});

describe('getPPMStatusLabel', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2024-06-25T00:00:00.000Z').getTime());
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('returns label only', () => {
    expect(getPPMStatusLabel('2024-06-25')).toBe(PPM_STATUSES.DUE);
  });
});
