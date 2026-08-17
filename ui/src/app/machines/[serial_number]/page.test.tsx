import { render, screen } from '@testing-library/react';
import { notFound, useRouter } from 'next/navigation';
import MachinePage from './page';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  notFound: jest.fn(),
  useRouter: jest.fn(),
}));

// Mock the MaintenanceHistory component
jest.mock('@/components/MaintenanceHistory', () => {
  return function MockMaintenanceHistory({
    machine,
    onEdit,
    onDelete,
    onFlag,
  }: {
    machine: { serial_number: string };
    onEdit?: () => void;
    onDelete?: () => void;
    onFlag?: () => void;
  }) {
    return (
      <div data-testid="maintenance-history">
        <h1>Maintenance History for {machine?.serial_number || 'Unknown'}</h1>
        {onEdit && <button onClick={onEdit}>Edit Machine</button>}
        {onDelete && <button onClick={onDelete}>Delete Machine</button>}
        {onFlag && <button onClick={onFlag}>Flag Machine</button>}
      </div>
    );
  };
});

// Mock the machine data
jest.mock('@/data/mockMachines', () => ({
  mockMachines: [
    {
      serial_number: 'SN-001',
      customer: 'Test Customer',
      model: 'Test Model',
      brand: 'Test Brand',
      state: 'CA',
      district: 'Test District',
      person_in_charge: 'Test Person',
      reported_by: 'Test Reporter',
      additional_notes: 'Test notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-01-01',
      ppm_date: '2024-02-01',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      account_type: 'Premium',
      status: 'Active',
    },
    {
      serial_number: 'SN-002',
      customer: 'Another Customer',
      model: 'Another Model',
      brand: 'Another Brand',
      state: 'NY',
      district: 'Another District',
      person_in_charge: 'Another Person',
      reported_by: 'Another Reporter',
      additional_notes: 'Another notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-01-02',
      ppm_date: '2024-02-02',
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
      account_type: 'Standard',
      status: 'Active',
    },
    {
      serial_number: 'SN 001', // Machine with space for URL-encoded test
      customer: 'Space Customer',
      model: 'Space Model',
      brand: 'Space Brand',
      state: 'TX',
      district: 'Space District',
      person_in_charge: 'Space Person',
      reported_by: 'Space Reporter',
      additional_notes: 'Space notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-01-03',
      ppm_date: '2024-02-03',
      created_at: '2024-01-03T00:00:00Z',
      updated_at: '2024-01-03T00:00:00Z',
      account_type: 'Basic',
      status: 'Active',
    },
    {
      serial_number: 'SN-001/A', // Machine with special characters
      customer: 'Special Customer',
      model: 'Special Model',
      brand: 'Special Brand',
      state: 'FL',
      district: 'Special District',
      person_in_charge: 'Special Person',
      reported_by: 'Special Reporter',
      additional_notes: 'Special notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-01-04',
      ppm_date: '2024-02-04',
      created_at: '2024-01-04T00:00:00Z',
      updated_at: '2024-01-04T00:00:00Z',
      account_type: 'Premium',
      status: 'Active',
    },
  ],
}));

// Mock the maintenance data
jest.mock('@/data/mockMaintenance', () => ({
  mockMaintenanceRecords: [
    {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-001',
      work_order_date: '2024-06-15',
      action_taken: 'Test maintenance action',
      reported_by: 'John Doe',
      work_order_type: 'Preventive',
      attachment: '',
      created_at: '2024-06-15T09:00:00Z',
      updated_at: '2024-06-15T10:30:00Z',
    },
    {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-002',
      work_order_date: '2024-06-16',
      action_taken: 'Another test maintenance action',
      reported_by: 'Jane Smith',
      work_order_type: 'Emergency',
      attachment: '',
      created_at: '2024-06-16T09:00:00Z',
      updated_at: '2024-06-16T10:30:00Z',
    },
    {
      machine_serial_number: 'SN-002',
      work_order_number: 'WO-003',
      work_order_date: '2024-06-17',
      action_taken: 'Test action for SN-002',
      reported_by: 'Bob Wilson',
      work_order_type: 'Corrective',
      attachment: '',
      created_at: '2024-06-17T09:00:00Z',
      updated_at: '2024-06-17T10:30:00Z',
    },
  ],
}));

// Mock the API client
jest.mock('@/utils/api', () => ({
  apiClient: {
    get: jest.fn(),
  },
  handleApiError: jest.fn((error) => ({
    message: error.message || 'API Error',
    status: error.status || 500,
    details: error,
  })),
}));

// Mock the useMachine hook
jest.mock('@/hooks/useMachine', () => ({
  useMachine: () => ({
    machine: {
      serial_number: 'SN-001',
      customer: 'Test Customer',
      model: 'Test Model',
      brand: 'Test Brand',
      state: 'CA',
      district: 'Test District',
      person_in_charge: 'Test Person',
      reported_by: 'Test Reporter',
      additional_notes: 'Test notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-01-01',
      ppm_date: '2024-02-01',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      account_type: 'Premium',
      status: 'Active',
    },
    loading: false,
    error: null,
    fetchMachine: jest.fn(),
    createMachine: jest.fn(),
    updateMachine: jest.fn(),
    deleteMachine: jest.fn(),
    clearError: jest.fn(),
    reset: jest.fn(),
  }),
}));

// Mock the machine service
jest.mock('@/services/machineService', () => ({
  MachineService: {
    getMachine: jest.fn().mockResolvedValue({
      data: {
        serial_number: 'SN-001',
        customer: 'Test Customer',
        model: 'Test Model',
        brand: 'Test Brand',
        state: 'CA',
        district: 'Test District',
        person_in_charge: 'Test Person',
        reported_by: 'Test Reporter',
        additional_notes: 'Test notes',
        attachment: '',
        ppm_status: '',
        tnc_date: '2024-01-01',
        ppm_date: '2024-02-01',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        account_type: 'Premium',
        status: 'Active',
      },
    }),
  },
}));

// Mock fetch globally
global.fetch = jest.fn();

describe('MachinePage', () => {
  const mockPush = jest.fn();

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    });
    jest.mocked(notFound).mockClear();
    mockPush.mockClear();
    
    // Mock successful API response
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => ({
        data: {
          serial_number: 'SN-001',
          customer: 'Test Customer',
          model: 'Test Model',
          brand: 'Test Brand',
          state: 'CA',
          district: 'Test District',
          person_in_charge: 'Test Person',
          reported_by: 'Test Reporter',
          additional_notes: 'Test notes',
          attachment: '',
          ppm_status: '',
          tnc_date: '2024-01-01',
          ppm_date: '2024-02-01',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          account_type: 'Premium',
          status: 'Active',
        },
      }),
    });
  });

  it('renders MaintenanceHistory for valid machine serial number', async () => {
    const params = Promise.resolve({ serial_number: 'SN-001' });
    render(<MachinePage params={params} />);
    expect(await screen.findByTestId('maintenance-history')).toBeInTheDocument();
  });
});
