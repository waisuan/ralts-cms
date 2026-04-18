import { act, renderHook, waitFor } from '@testing-library/react';
import { useMaintenance } from './useMaintenance';
import { MaintenanceService, MaintenanceFilters } from '../services/maintenanceService';
import { ApiError } from '../utils/api';

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

// Module-scoped stable refs — the hook uses `filters` in its effect deps, so
// like the production consumer (which memoises filters via useMemo), tests
// must pass the same object reference across renders.
const stableSortDesc: MaintenanceFilters = { sort: 'updated_at_desc' };
const stableSortAsc: MaintenanceFilters = { sort: 'work_order_date_desc' };
const stableCompound: MaintenanceFilters = { q: 'hello', sort: 'updated_at_asc' };

describe('useMaintenance', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    getList.mockResolvedValue({ data: listPayload });
  });

  it('fetches on mount with the provided page/limit/filters', async () => {
    const { result } = renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 1,
        limit: 50,
        filters: stableSortDesc,
      }),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(getList).toHaveBeenCalledTimes(1);
    expect(getList).toHaveBeenCalledWith('M1', 1, 50, stableSortDesc);
    expect(result.current.records).toHaveLength(1);
    expect(result.current.total).toBe(1);
    expect(result.current.preventativeCount).toBe(1);
    expect(result.current.totalPages).toBe(1);
  });

  it('refetches when filters change', async () => {
    const { result, rerender } = renderHook(
      (props: { filters: MaintenanceFilters }) =>
        useMaintenance({
          serialNumber: 'M1',
          page: 1,
          limit: 50,
          filters: props.filters,
        }),
      { initialProps: { filters: stableSortDesc } },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getList).toHaveBeenCalledTimes(1);

    rerender({ filters: stableSortAsc });

    await waitFor(() => expect(getList).toHaveBeenCalledTimes(2));
    expect(getList.mock.calls[1][3]).toBe(stableSortAsc);
  });

  it('refetches when page changes', async () => {
    const { result, rerender } = renderHook(
      (props: { page: number }) =>
        useMaintenance({
          serialNumber: 'M1',
          page: props.page,
          limit: 50,
          filters: stableSortDesc,
        }),
      { initialProps: { page: 1 } },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getList).toHaveBeenCalledTimes(1);
    expect(getList.mock.calls[0][1]).toBe(1);

    rerender({ page: 3 });

    await waitFor(() => expect(getList).toHaveBeenCalledTimes(2));
    expect(getList.mock.calls[1][1]).toBe(3);
  });

  it('refetches when limit changes', async () => {
    const { result, rerender } = renderHook(
      (props: { limit: number }) =>
        useMaintenance({
          serialNumber: 'M1',
          page: 1,
          limit: props.limit,
          filters: stableSortDesc,
        }),
      { initialProps: { limit: 50 } },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getList).toHaveBeenCalledTimes(1);
    expect(getList.mock.calls[0][2]).toBe(50);

    rerender({ limit: 100 });

    await waitFor(() => expect(getList).toHaveBeenCalledTimes(2));
    expect(getList.mock.calls[1][2]).toBe(100);
  });

  it('refetch() re-invokes the service with current props', async () => {
    const { result } = renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 2,
        limit: 100,
        filters: stableCompound,
      }),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(getList).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.refetch();
    });

    expect(getList).toHaveBeenCalledTimes(2);
    expect(getList.mock.calls[1]).toEqual(['M1', 2, 100, stableCompound]);
  });

  it('does not fetch when autoFetch is false', async () => {
    renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 1,
        limit: 50,
        filters: stableSortDesc,
        autoFetch: false,
      }),
    );

    await act(async () => {
      await Promise.resolve();
    });

    expect(getList).not.toHaveBeenCalled();
  });

  it('resets state to empty when response has no data', async () => {
    getList.mockResolvedValueOnce({ data: undefined });

    const { result } = renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 1,
        limit: 50,
        filters: stableSortDesc,
      }),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.records).toEqual([]);
    expect(result.current.total).toBe(0);
    expect(result.current.preventativeCount).toBe(0);
    expect(result.current.correctiveCount).toBe(0);
    expect(result.current.emergencyCount).toBe(0);
    expect(result.current.inspectionCount).toBe(0);
    expect(result.current.otherCount).toBe(0);
    expect(result.current.error).toBeNull();
  });

  it('surfaces non-auth errors via the error state', async () => {
    getList.mockRejectedValueOnce(new ApiError('Something broke', 500));

    const { result } = renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 1,
        limit: 50,
        filters: stableSortDesc,
      }),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Something broke');
    expect(result.current.records).toEqual([]);
  });

  it('swallows 401 auth errors and leaves error state null', async () => {
    getList.mockRejectedValueOnce(new ApiError('Unauthorized', 401));

    const { result } = renderHook(() =>
      useMaintenance({
        serialNumber: 'M1',
        page: 1,
        limit: 50,
        filters: stableSortDesc,
      }),
    );

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBeNull();
    expect(result.current.records).toEqual([]);
  });

  it('ignores stale in-flight responses when a newer request has started', async () => {
    // Resolver captured refs so we can control which request settles first.
    let resolveFirst: (value: { data: typeof listPayload }) => void = () => {};
    const firstPromise = new Promise<{ data: typeof listPayload }>((resolve) => {
      resolveFirst = resolve;
    });
    const secondPayload = {
      ...listPayload,
      maintenance: [
        { ...listPayload.maintenance[0], work_order_number: 'WO-2' },
      ],
      count: 99,
    };

    getList
      .mockImplementationOnce(() => firstPromise)
      .mockResolvedValueOnce({ data: secondPayload });

    const { result, rerender } = renderHook(
      (props: { page: number }) =>
        useMaintenance({
          serialNumber: 'M1',
          page: props.page,
          limit: 50,
          filters: stableSortDesc,
        }),
      { initialProps: { page: 1 } },
    );

    // Kick off a second fetch before the first resolves.
    rerender({ page: 2 });
    await waitFor(() => expect(getList).toHaveBeenCalledTimes(2));

    // Resolve the (now stale) first request last. Its result must be ignored.
    await act(async () => {
      resolveFirst({ data: listPayload });
      await Promise.resolve();
    });

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.total).toBe(99);
    expect(result.current.records[0].work_order_number).toBe('WO-2');
  });
});
