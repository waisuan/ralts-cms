import { renderHook } from '@testing-library/react';
import { useOverdueStats } from '../useOverdueStats';
import { Machine } from '../../types/machine';

describe('useOverdueStats', () => {
  const createMachine = (serial: string, ppmDate: string): Machine => ({
    serial_number: serial,
    customer: 'Test Customer',
    state: 'CA',
    account_type: 'Premium',
    model: 'TestModel',
    status: 'Active',
    brand: 'TestBrand',
    district: 'TestDistrict',
    person_in_charge: 'Test Person',
    reported_by: 'Test Reporter',
    additional_notes: 'Test notes',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-01',
    ppm_date: ppmDate,
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
  });

  it('returns zero counts for empty machine list', () => {
    const { result } = renderHook(() => useOverdueStats([]));

    expect(result.current).toEqual({
      overdueCount: 0,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 0,
      overdueMachines: [],
      dueMachines: [],
    });
  });

  it('calculates overdue machines correctly', () => {
    const overdueDate = new Date();
    overdueDate.setDate(overdueDate.getDate() - 5); // 5 days ago

    const machines = [
      createMachine('SN-001', overdueDate.toISOString().split('T')[0]),
      createMachine('SN-002', overdueDate.toISOString().split('T')[0]),
    ];

    const { result } = renderHook(() => useOverdueStats(machines));

    expect(result.current.overdueCount).toBe(2);
    expect(result.current.overdueMachines).toHaveLength(2);
    expect(result.current.overdueMachines[0].serial_number).toBe('SN-001');
    expect(result.current.overdueMachines[1].serial_number).toBe('SN-002');
    expect(result.current.totalCriticalCount).toBe(2);
  });

  it('calculates due machines correctly', () => {
    const today = new Date().toISOString().split('T')[0];

    const machines = [createMachine('SN-001', today), createMachine('SN-002', today)];

    const { result } = renderHook(() => useOverdueStats(machines));

    expect(result.current.dueCount).toBe(2);
    expect(result.current.dueMachines).toHaveLength(2);
    expect(result.current.dueMachines[0].serial_number).toBe('SN-001');
    expect(result.current.dueMachines[1].serial_number).toBe('SN-002');
    expect(result.current.totalCriticalCount).toBe(2);
  });

  it('calculates due soon machines correctly', () => {
    const dueSoonDate = new Date();
    dueSoonDate.setDate(dueSoonDate.getDate() + 3); // 3 days from now

    const machines = [createMachine('SN-001', dueSoonDate.toISOString().split('T')[0])];

    const { result } = renderHook(() => useOverdueStats(machines));

    expect(result.current.dueSoonCount).toBe(1);
    expect(result.current.totalCriticalCount).toBe(0); // Due soon doesn't count as critical
  });

  it('handles mixed machine statuses correctly', () => {
    const overdueDate = new Date();
    overdueDate.setDate(overdueDate.getDate() - 5);

    const dueDate = new Date().toISOString().split('T')[0];

    const dueSoonDate = new Date();
    dueSoonDate.setDate(dueSoonDate.getDate() + 3);

    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + 30);

    const machines = [
      createMachine('SN-OVERDUE', overdueDate.toISOString().split('T')[0]),
      createMachine('SN-DUE', dueDate),
      createMachine('SN-DUE-SOON', dueSoonDate.toISOString().split('T')[0]),
      createMachine('SN-FUTURE', futureDate.toISOString().split('T')[0]),
    ];

    const { result } = renderHook(() => useOverdueStats(machines));

    expect(result.current.overdueCount).toBe(1);
    expect(result.current.dueCount).toBe(1);
    expect(result.current.dueSoonCount).toBe(1);
    expect(result.current.totalCriticalCount).toBe(2); // overdue + due
    expect(result.current.overdueMachines[0].serial_number).toBe('SN-OVERDUE');
    expect(result.current.dueMachines[0].serial_number).toBe('SN-DUE');
  });

  it('handles machines with empty ppm_date', () => {
    const machines = [createMachine('SN-001', ''), createMachine('SN-002', '')];

    const { result } = renderHook(() => useOverdueStats(machines));

    expect(result.current).toEqual({
      overdueCount: 0,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 0,
      overdueMachines: [],
      dueMachines: [],
    });
  });

  it('updates when machine data changes', () => {
    const overdueDate = new Date();
    overdueDate.setDate(overdueDate.getDate() - 5);

    const initialMachines = [createMachine('SN-001', overdueDate.toISOString().split('T')[0])];

    const { result, rerender } = renderHook(({ machines }) => useOverdueStats(machines), {
      initialProps: { machines: initialMachines },
    });

    expect(result.current.overdueCount).toBe(1);

    // Update with no machines
    rerender({ machines: [] });

    expect(result.current.overdueCount).toBe(0);
    expect(result.current.overdueMachines).toHaveLength(0);
  });
});
