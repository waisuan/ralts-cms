import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { withNuqsTestingAdapter, type UrlUpdateEvent } from 'nuqs/adapters/testing';
import MaintenanceHistory from './MaintenanceHistory';
import { machineInfoExpandedSessionKey } from './MachineInfoCard';
import { Machine } from '../types/machine';
import { MaintenanceService } from '../services/maintenanceService';
import { useIsMobile } from '../hooks/useMediaQuery';

jest.mock('../services/maintenanceService', () => ({
  MaintenanceService: {
    getMaintenanceList: jest.fn(),
    getMaintenance: jest.fn(),
    createMaintenance: jest.fn(),
    updateMaintenance: jest.fn(),
    deleteMaintenance: jest.fn(),
  },
}));

jest.mock('../hooks/useMediaQuery', () => ({
  useIsMobile: jest.fn(() => false),
}));

const getList = MaintenanceService.getMaintenanceList as jest.MockedFunction<
  typeof MaintenanceService.getMaintenanceList
>;
const useIsMobileMock = useIsMobile as jest.MockedFunction<typeof useIsMobile>;

const mockMachine: Machine = {
  serial_number: 'SN-001',
  customer: 'Test Customer',
  state: 'Selangor',
  account_type: 'Premium',
  model: 'Test Model',
  status: 'Active',
  brand: 'Test Brand',
  district: 'Petaling Jaya',
  person_in_charge: 'John Doe',
  reported_by: 'Jane Smith',
  additional_notes: 'Test notes for the machine',
  attachment: 'machine_manual.pdf',
  ppm_status: '',
  tnc_date: '2024-01-15',
  ppm_date: '2024-07-15',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-06-01T00:00:00Z',
  updated_by: 'admin',
};

const baseResponse = {
  maintenance: [
    {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-001',
      work_order_date: '2024-06-15',
      action_taken: 'Performed routine maintenance and cleaning',
      reported_by: 'John Doe',
      work_order_type: 'Preventive',
      attachment: 'maintenance_report.pdf',
      created_at: '2024-06-15T09:00:00Z',
      updated_at: '2024-06-15T10:30:00Z',
      updated_by: 'admin',
    },
  ],
  preventative_count: 2,
  corrective_count: 1,
  emergency_count: 3,
  inspection_count: 1,
  other_count: 0,
  count: 7,
  limit: 50,
  offset: 0,
  sort: 'updated_at_desc',
};

describe('MaintenanceHistory', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useIsMobileMock.mockReturnValue(false);
    getList.mockResolvedValue({ data: baseResponse });
    sessionStorage.removeItem(machineInfoExpandedSessionKey(mockMachine.serial_number));
  });

  it('renders machine information and counts on initial load', async () => {
    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({ searchParams: '' }),
    });

    const expandMachineInfo = await screen.findByRole('button', { name: /show machine details/i });
    fireEvent.click(expandMachineInfo);

    await waitFor(() => {
      expect(screen.getByText(mockMachine.model)).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByText('2')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });

    expect(getList).toHaveBeenCalledWith('SN-001', 1, 50, { sort: 'updated_at_desc' });
  });

  it('hydrates initial fetch from URL params', async () => {
    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({
        searchParams: '?q=filter&sort=work_order_date_asc&page=2&limit=100',
      }),
    });

    await waitFor(() => {
      expect(getList).toHaveBeenCalledWith('SN-001', 2, 100, {
        q: 'filter',
        sort: 'work_order_date_asc',
      });
    });
  });

  it('falls back to defaults when URL has invalid values', async () => {
    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({
        searchParams: '?sort=bogus&limit=999',
      }),
    });

    await waitFor(() => {
      expect(getList).toHaveBeenCalledWith('SN-001', 1, 50, { sort: 'updated_at_desc' });
    });
  });

  it('auto-corrects an out-of-range `page` param after fetch returns empty records with total>0', async () => {
    // First fetch (page=99): service returns empty list but count>0, simulating
    // a stale deep link. Second fetch (page=1 after auto-correct) returns data.
    getList
      .mockResolvedValueOnce({
        data: { ...baseResponse, maintenance: [], count: 7 },
      })
      .mockResolvedValueOnce({ data: baseResponse });

    const onUrlUpdate = jest.fn<void, [UrlUpdateEvent]>();

    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({ searchParams: '?page=99', onUrlUpdate }),
    });

    // Wait for both fetches: the stale page=99, then the auto-corrected page=1.
    await waitFor(() => expect(getList).toHaveBeenCalledTimes(2));

    expect(getList.mock.calls[0][1]).toBe(99);
    expect(getList.mock.calls[1][1]).toBe(1);

    // URL was rewritten: page=1 is the default so it's elided from the string.
    expect(onUrlUpdate).toHaveBeenCalled();
    const last = onUrlUpdate.mock.calls.at(-1)?.[0];
    expect(last?.queryString ?? '').not.toMatch(/page=\d+/);
  });

  it('does not loop when records and total are both 0 (legitimate empty state)', async () => {
    getList.mockResolvedValue({
      data: { ...baseResponse, maintenance: [], count: 0 },
    });

    const onUrlUpdate = jest.fn<void, [UrlUpdateEvent]>();

    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({ searchParams: '?page=1', onUrlUpdate }),
    });

    await waitFor(() => expect(getList).toHaveBeenCalledTimes(1));

    // Give any would-be auto-correct effect a chance to run.
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0));
    });

    expect(getList).toHaveBeenCalledTimes(1);
    expect(onUrlUpdate).not.toHaveBeenCalled();
  });

  it('writes sort to URL and resets page when sort changes via the mobile menu', async () => {
    useIsMobileMock.mockReturnValue(true);
    const onUrlUpdate = jest.fn<void, [UrlUpdateEvent]>();

    render(<MaintenanceHistory machine={mockMachine} />, {
      wrapper: withNuqsTestingAdapter({ searchParams: '?page=3', onUrlUpdate }),
    });

    await waitFor(() => expect(getList).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: /sort records/i }));

    await act(async () => {
      fireEvent.click(
        screen.getByRole('button', { name: /work order date \(latest\)/i }),
      );
      await new Promise((resolve) => setTimeout(resolve, 0));
    });

    expect(onUrlUpdate).toHaveBeenCalled();
    const last = onUrlUpdate.mock.calls.at(-1)?.[0];
    expect(last?.queryString).toContain('sort=work_order_date_desc');
    expect(last?.queryString).not.toContain('page=3');
  });
});
