import { renderHook, waitFor, act } from '@testing-library/react';
import { useMaintenance } from './useMaintenance';
import { MaintenanceService } from '../services/maintenanceService';

jest.mock('./useDebounce', () => ({
  useDebounce: <T,>(v: T) => v,
}));

jest.mock('../services/maintenanceService', () => ({
  MaintenanceService: {
    getMaintenanceList: jest.fn(),
  },
}));

const listPayload = {
  maintenance: [
    {
      machine_serial_number: 'M1',
      work_order_number: 'WO-1',
      work_order_date: '2024-01-01',
      action_taken: 'x',
      reported_by: 'y',
      work_order_type: 'Preventive',
      attachment: null,
      created_at: '',
      updated_at: '',
    },
  ],
  count: 1,
  preventative_count: 1,
  corrective_count: 0,
  emergency_count: 0,
  inspection_count: 0,
  other_count: 0,
  limit: 50,
  offset: 0,
  sort: 'updated_at_desc',
};

const getList = MaintenanceService.getMaintenanceList as jest.MockedFunction<
  typeof MaintenanceService.getMaintenanceList
>;

describe('useMaintenance', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    getList.mockResolvedValue({ data: listPayload });
  });

  it('loads maintenance on mount with sort filter', async () => {
    const { result } = renderHook(() =>
      useMaintenance({ serialNumber: 'M1', limit: 50, autoFetch: true })
    );

    await waitFor(() => expect(result.current.isInitialLoading).toBe(false));

    expect(getList).toHaveBeenCalledWith('M1', 1, 50, { sort: 'updated_at_desc' });
    expect(result.current.records).toHaveLength(1);
    expect(result.current.total).toBe(1);
    expect(result.current.preventativeCount).toBe(1);
  });

  it('refetches when setSort changes after initial load', async () => {
    const { result } = renderHook(() =>
      useMaintenance({ serialNumber: 'M1', limit: 50, autoFetch: true })
    );

    await waitFor(() => expect(result.current.isInitialLoading).toBe(false));
    const callsAfterInitial = getList.mock.calls.length;

    act(() => {
      result.current.setSort('work_order_date_desc');
    });

    await waitFor(() =>
      expect(getList.mock.calls.length).toBeGreaterThan(callsAfterInitial)
    );

    const lastCall = getList.mock.calls[getList.mock.calls.length - 1];
    expect(lastCall[3]).toMatchObject({ sort: 'work_order_date_desc' });
  });
});
