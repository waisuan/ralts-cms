import { getPPMStatus, getPPMStatusLabel } from '../ppmUtils';
import { PPM_STATUSES } from '../constants';

describe('getPPMStatus', () => {
  it('should return null for empty date', () => {
    const result = getPPMStatus('');
    expect(result).toBeNull();
  });

  it('should return Overdue status for past dates', () => {
    const pastDate = new Date();
    pastDate.setDate(pastDate.getDate() - 1);
    const result = getPPMStatus(pastDate.toISOString().split('T')[0]);
    expect(result!.label).toBe(PPM_STATUSES.OVERDUE);
  });

  it('should return Due status for today', () => {
    const today = new Date().toISOString().split('T')[0];
    const result = getPPMStatus(today);
    expect(result!.label).toBe(PPM_STATUSES.DUE);
  });

  it('should return Almost Due status for dates within 2 weeks', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 7);
    const result = getPPMStatus(futureDate.toISOString().split('T')[0]);
    expect(result!.label).toBe(PPM_STATUSES.ALMOST_DUE);
  });

  it('should return null for dates more than 2 weeks in the future', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 15);
    const result = getPPMStatus(futureDate.toISOString().split('T')[0]);
    expect(result).toBeNull();
  });

  it('should return Almost Due status for dates exactly 2 weeks in the future', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 14);
    const result = getPPMStatus(futureDate.toISOString().split('T')[0]);
    expect(result!.label).toBe(PPM_STATUSES.ALMOST_DUE);
  });

  it('should return null for dates more than 2 weeks in the future', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 15);
    const result = getPPMStatus(futureDate.toISOString().split('T')[0]);
    expect(result).toBeNull();
  });
});

describe('getPPMStatusLabel', () => {
  it('should return null for empty date', () => {
    const label = getPPMStatusLabel('');
    expect(label).toBeNull();
  });

  it('should return Overdue label for past dates', () => {
    const pastDate = new Date();
    pastDate.setDate(pastDate.getDate() - 1);
    const label = getPPMStatusLabel(pastDate.toISOString().split('T')[0]);
    expect(label).toBe(PPM_STATUSES.OVERDUE);
  });

  it('should return Due label for today', () => {
    const today = new Date().toISOString().split('T')[0];
    const label = getPPMStatusLabel(today);
    expect(label).toBe(PPM_STATUSES.DUE);
  });

  it('should return Almost Due label for dates within 2 weeks', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 7);
    const label = getPPMStatusLabel(futureDate.toISOString().split('T')[0]);
    expect(label).toBe(PPM_STATUSES.ALMOST_DUE);
  });

  it('should return null for dates more than 2 weeks in the future', () => {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 15);
    const label = getPPMStatusLabel(futureDate.toISOString().split('T')[0]);
    expect(label).toBeNull();
  });
});
