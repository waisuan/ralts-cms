import { renderHook, waitFor } from '@testing-library/react';
import { useMachines } from './useMachines';
import { MachineService } from '../services/machineService';

jest.mock('./useDebounce', () => ({
  useDebounce: <T,>(v: T) => v,
}));

jest.mock('../services/machineService', () => ({
  MachineService: {
    getMachines: jest.fn(),
  },
}));

const getMachines = MachineService.getMachines as jest.MockedFunction<typeof MachineService.getMachines>;

describe('useMachines', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('sets total from API count', async () => {
    getMachines.mockResolvedValue({
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
      },
    });

    const { result } = renderHook(() =>
      useMachines({ page: 1, limit: 50, filters: {}, autoFetch: true })
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.total).toBe(42);
    expect(result.current.machines).toHaveLength(1);
  });
});
