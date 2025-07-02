import { render, screen } from '@testing-library/react';
import RecordCard from './RecordCard';
import { Machine } from '../types/machine';

// Mock maintenance data for RecordCard tests
jest.mock('../data/mockMaintenance', () => ({
  mockMaintenanceRecords: [
    {
      machine_serial_number: 'SN-TEST',
      work_order_number: 'WO-TEST-001',
      work_order_date: '2024-06-15',
      action_taken: 'Test maintenance action',
      reported_by: 'Test Tech',
      worker_order_type: 'Preventive',
      attachment: 'test_report.pdf',
      created_at: '2024-06-15T09:00:00Z',
      updated_at: '2024-06-15T10:30:00Z',
    },
    {
      machine_serial_number: 'SN-TEST',
      work_order_number: 'WO-TEST-002',
      work_order_date: '2024-06-10',
      action_taken: 'Another test maintenance action',
      reported_by: 'Test Tech 2',
      worker_order_type: 'Emergency',
      attachment: '',
      created_at: '2024-06-10T14:00:00Z',
      updated_at: '2024-06-10T16:00:00Z',
    },
  ],
}));

// Mock the current date for consistent testing
const MOCK_CURRENT_DATE = '2024-06-29T12:00:00.000Z';

describe('RecordCard', () => {
  const baseMachine: Machine = {
    serial_number: 'SN-TEST',
    customer: 'Test Customer',
    state: 'CA',
    account_type: 'Premium',
    model: 'TestModel',
    status: 'Active',
    brand: 'TestBrand',
    district: 'TestDistrict',
    person_in_charge: 'Test Person',
    reported_by: 'Test Reporter',
    additional_notes: 'Test notes for the machine',
    attachment: 'test_file.pdf',
    ppm_status: '',
    tnc_date: '2024-07-01',
    ppm_date: '2024-06-24', // Overdue date for testing
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
  };

  beforeEach(() => {
    // Mock the current date for consistent testing
    jest.useFakeTimers();
    jest.setSystemTime(new Date(MOCK_CURRENT_DATE));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders machine card with all basic information and functionality', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );

    // Check that basic machine information is displayed
    expect(
      screen.getAllByText((content, element) => {
        return Boolean(element?.textContent?.includes('TestModel'));
      }).length
    ).toBeGreaterThan(0);
    expect(screen.getByText('SN-TEST')).toBeInTheDocument();
    expect(screen.getByText('TestBrand · TestDistrict, CA')).toBeInTheDocument();
    expect(screen.getByText('Test Customer')).toBeInTheDocument();

    // Check that account type and status are displayed
    expect(screen.getByText('Account Type:')).toBeInTheDocument();
    expect(screen.getByText('Premium')).toBeInTheDocument();
    expect(screen.getByText('Status:')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();

    // Check that reported by information is displayed
    expect(screen.getByText('Reported By:')).toBeInTheDocument();
    expect(screen.getByText('Test Reporter')).toBeInTheDocument();

    // Check that action buttons are present
    expect(screen.getByText('View')).toBeInTheDocument();
    expect(screen.getByText('Edit')).toBeInTheDocument();
    expect(screen.getByText('Delete')).toBeInTheDocument();

    // Check that status badge is displayed (Overdue in this case)
    expect(screen.getByText('Overdue')).toBeInTheDocument();

    // Check that maintenance count badge is displayed
    expect(screen.getByTitle('2 maintenance records available')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();

    // Check that attachment and notes icons are present
    expect(screen.getByTitle('Download attachment')).toBeInTheDocument();
    expect(screen.getByTitle('View additional notes')).toBeInTheDocument();
  });
});
