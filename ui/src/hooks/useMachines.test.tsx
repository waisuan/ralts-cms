import { renderHook, waitFor } from '@testing-library/react';
import { useMachines } from './useMachines';
import { MachineService } from '../services/machineService';

jest.mock('../services/machineService', () => ({
  MachineService: {
    getMachines: jest.fn(),
  },
}));

const getMachines = MachineService.getMachines as jest.MockedFunction<typeof MachineService.getMachines>;

const mockApiResponse = (overrides: Record<string, unknown> = {}) => ({
  data: {
    machines: [
      {
        serial_number: 'SN-1',
        customer: 'A',
        state: 'S',
        account_type: 'T',
        model: 'M',
        status: '',
        brand: 'B',
        district: 'D',
        person_in_charge: 'P',
        reported_by: 'R',
        additional_notes: '',
        attachment: '',
        ppm_status: '',
        tnc_date: '2024-01-01',
        ppm_date: '2024-01-01',
        created_at: '',
        updated_at: '',
      },
    ],
    count: 42,
    offset: 0,
    limit: 50,
    sort: 'updated_at_desc',
    overdue_count: 0,
    due_count: 0,
    almost_due_count: 0,
    ...overrides,
  },
});

describe('useMachines', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('sets total from API count', async () => {
    getMachines.mockResolvedValue(mockApiResponse());

    const { result } = renderHook(() =>
      useMachines({ page: 0, limit: 50, filters: {}, autoFetch: true })
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.total).toBe(42);
    expect(result.current.machines).toHaveLength(1);
  });

  it('uses page offset on fetch', async () => {
    getMachines.mockResolvedValue(mockApiResponse({ offset: 100 }));

    renderHook(() =>
      useMachines({ page: 2, limit: 50, filters: {}, autoFetch: true })
    );

    await waitFor(() => {
      expect(getMachines).toHaveBeenCalled();
    });
    // page=2, limit=50 -> offset=100 -> API page = 100/50 + 1 = 3
    expect(getMachines).toHaveBeenCalledWith(3, 50, {});
  });

  it('fetches from page 1 when page is 0', async () => {
    getMachines.mockResolvedValue(mockApiResponse());

    renderHook(() =>
      useMachines({ page: 0, limit: 50, filters: {}, autoFetch: true })
    );

    await waitFor(() => {
      expect(getMachines).toHaveBeenCalled();
    });
    // page=0 -> offset=0 -> API page = 0/50 + 1 = 1
    expect(getMachines).toHaveBeenCalledWith(1, 50, {});
  });
});
