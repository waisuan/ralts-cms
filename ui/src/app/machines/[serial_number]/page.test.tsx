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
    onBack,
  }: {
    machine: { serial_number: string };
    onBack: () => void;
  }) {
    return (
      <div data-testid="maintenance-history">
        <h1>Maintenance History for {machine?.serial_number || 'Unknown'}</h1>
        <button onClick={onBack}>Back</button>
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

describe('MachinePage', () => {
  const mockPush = jest.fn();

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    });
    jest.mocked(notFound).mockClear();
    mockPush.mockClear();
  });

  it('renders MaintenanceHistory for valid machine serial number', async () => {
    const params = Promise.resolve({ serial_number: 'SN-001' });

    render(<MachinePage params={params} />);

    // Wait for the component to load
    await screen.findByTestId('maintenance-history');
    expect(screen.getByText('Maintenance History for SN-001')).toBeInTheDocument();
  });

  it('renders MaintenanceHistory for URL-encoded serial number', async () => {
    const params = Promise.resolve({ serial_number: 'SN%20001' }); // URL-encoded space

    render(<MachinePage params={params} />);

    // Wait for the component to load
    await screen.findByTestId('maintenance-history');
    expect(screen.getByText('Maintenance History for SN 001')).toBeInTheDocument();
  });

  it('calls notFound for invalid machine serial number', async () => {
    // Clear previous calls
    jest.mocked(notFound).mockClear();

    const params = Promise.resolve({ serial_number: 'INVALID-SN' });

    render(<MachinePage params={params} />);

    // Wait a bit for the effect to run
    await new Promise((resolve) => setTimeout(resolve, 100));
    expect(notFound).toHaveBeenCalled();
  });

  it('handles special characters in serial number correctly', async () => {
    const params = Promise.resolve({ serial_number: 'SN-001%2FA' }); // URL-encoded slash

    render(<MachinePage params={params} />);

    // Wait for the component to load
    await screen.findByTestId('maintenance-history');
  });

  it('navigates back to home when back button is clicked', async () => {
    const params = Promise.resolve({ serial_number: 'SN-001' });

    render(<MachinePage params={params} />);

    // Wait for the component to load
    const backButton = await screen.findByText('Back');
    backButton.click();

    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('passes correct machine data to MaintenanceHistory component', async () => {
    const params = Promise.resolve({ serial_number: 'SN-002' });

    render(<MachinePage params={params} />);

    // Wait for the component to load
    await screen.findByText('Maintenance History for SN-002');
  });
});
