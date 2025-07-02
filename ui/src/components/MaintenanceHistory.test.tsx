import { render, screen } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import MaintenanceHistory from './MaintenanceHistory';
import { Machine } from '../types/machine';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock the maintenance data
jest.mock('../data/mockMaintenance', () => ({
  mockMaintenanceRecords: [
    {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-001',
      work_order_date: '2024-06-15',
      action_taken: 'Performed routine maintenance and cleaning',
      reported_by: 'John Doe',
      worker_order_type: 'Preventive',
      attachment: 'maintenance_report.pdf',
      created_at: '2024-06-15T09:00:00Z',
      updated_at: '2024-06-15T10:30:00Z',
    },
    {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-002',
      work_order_date: '2024-06-10',
      action_taken:
        'This is a very long action description that exceeds the character limit and should be truncated when displayed in the table but shown in full when the modal is opened',
      reported_by: 'Jane Smith',
      worker_order_type: 'Emergency',
      attachment: '',
      created_at: '2024-06-10T14:00:00Z',
      updated_at: '2024-06-10T16:00:00Z',
    },
    {
      machine_serial_number: 'NO-RECORDS',
      work_order_number: 'WO-003',
      work_order_date: '2024-06-01',
      action_taken: 'Test action',
      reported_by: 'Test User',
      worker_order_type: 'Corrective',
      attachment: '',
      created_at: '2024-06-01T08:00:00Z',
      updated_at: '2024-06-01T09:00:00Z',
    },
    // Add more records to trigger pagination
    ...Array.from({ length: 15 }, (_, i) => ({
      machine_serial_number: 'SN-001',
      work_order_number: `WO-${String(i + 4).padStart(3, '0')}`,
      work_order_date: '2024-06-01',
      action_taken: `Additional maintenance record ${i + 4}`,
      reported_by: 'Test User',
      worker_order_type: 'Preventive',
      attachment: '',
      created_at: '2024-06-01T08:00:00Z',
      updated_at: '2024-06-01T09:00:00Z',
    })),
  ],
}));

describe('MaintenanceHistory', () => {
  const mockPush = jest.fn();
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
  };

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    });
    mockPush.mockClear();
  });

  it('renders machine information correctly', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);
    expect(screen.getByText(mockMachine.serial_number)).toBeInTheDocument();
    expect(screen.getByText(mockMachine.model)).toBeInTheDocument();
  });
});
