import type { Maintenance } from '@/types/maintenance';
import {
  formatDate,
  formatDateTime,
  formatMachineDateDisplay,
  getTypeColor,
  isCustomWorkOrderType,
  workOrderTypePillLabel,
} from './formatters';

describe('work order type helpers', () => {
  it('treats standard types as non-custom', () => {
    expect(isCustomWorkOrderType('Preventive')).toBe(false);
    expect(workOrderTypePillLabel('Preventive')).toBe('Preventive');
  });

  it('treats unknown strings as custom (Other pill)', () => {
    expect(isCustomWorkOrderType('Custom repair')).toBe(true);
    expect(workOrderTypePillLabel('Custom repair')).toBe('Other');
  });

  it('prefers work_order_type_is_standard from API when present', () => {
    const standardFlag: Maintenance = {
      machine_serial_number: 'SN-1',
      work_order_number: 'WO-1',
      work_order_date: '2024-01-01',
      action_taken: 'x',
      reported_by: 'y',
      work_order_type: 'Anything',
      work_order_type_is_standard: true,
      attachment: null,
      created_at: '',
      updated_at: '',
    };
    expect(isCustomWorkOrderType(standardFlag)).toBe(false);
    expect(workOrderTypePillLabel(standardFlag)).toBe('Anything');

    const customFlag: Maintenance = {
      ...standardFlag,
      work_order_type: 'Preventive',
      work_order_type_is_standard: false,
    };
    expect(isCustomWorkOrderType(customFlag)).toBe(true);
    expect(workOrderTypePillLabel(customFlag)).toBe('Other');
  });
});

describe('formatDate', () => {
  it('returns dash for empty string', () => {
    expect(formatDate('')).toBe('-');
  });
});

describe('formatMachineDateDisplay', () => {
  it('shows dash and isUnset for sentinel dates', () => {
    expect(formatMachineDateDisplay('0001-01-01T00:00:00Z')).toEqual({
      text: '-',
      isUnset: true,
    });
    expect(formatMachineDateDisplay('0001-12-31T00:00:00Z')).toEqual({
      text: '-',
      isUnset: true,
    });
  });

  it('formats real dates like formatDate', () => {
    const iso = '2024-06-15T00:00:00Z';
    expect(formatMachineDateDisplay(iso)).toEqual({
      text: formatDate(iso),
      isUnset: false,
    });
  });
});

describe('formatDateTime', () => {
  it('returns dash for empty string', () => {
    expect(formatDateTime('')).toBe('-');
  });
});

describe('getTypeColor', () => {
  it.each([
    ['Preventive', 'bg-green-100 text-green-800'],
    ['Corrective', 'bg-blue-100 text-blue-800'],
    ['Emergency', 'bg-red-100 text-red-800'],
    ['Inspection', 'bg-purple-100 text-purple-800'],
    ['Other', 'bg-gray-100 text-gray-800'],
    ['Unknown', 'bg-gray-100 text-gray-800'],
  ])('maps %s to tailwind classes', (type, expected) => {
    expect(getTypeColor(type)).toBe(expected);
  });
});
