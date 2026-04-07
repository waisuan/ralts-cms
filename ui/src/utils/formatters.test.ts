import type { Maintenance } from '@/types/maintenance';
import { isCustomWorkOrderType, workOrderTypePillLabel } from './formatters';

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
