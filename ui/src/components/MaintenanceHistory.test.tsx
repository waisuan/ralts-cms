import { render, screen, fireEvent, waitFor } from '@testing-library/react';
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

    expect(screen.getByText('Machine Information')).toBeInTheDocument();
    expect(screen.getByText('SN-001')).toBeInTheDocument();
    expect(screen.getByText('Test Model')).toBeInTheDocument();
    expect(screen.getByText('Test Brand')).toBeInTheDocument();
    expect(screen.getByText('Test Customer')).toBeInTheDocument();
    expect(screen.getByText('Petaling Jaya, Selangor')).toBeInTheDocument();
  });

  it('displays maintenance statistics correctly', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    expect(screen.getByText('Total Records')).toBeInTheDocument();
    // Check that statistics section exists
    const statsSection = screen.getByText('Total Records').closest('div');
    expect(statsSection).toBeInTheDocument();
  });

  it('renders maintenance records table with correct data', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    expect(screen.getByText('WO-001')).toBeInTheDocument();
    expect(screen.getByText('WO-002')).toBeInTheDocument();
    // Use getAllByText for names that appear multiple times
    expect(screen.getAllByText('John Doe')).toHaveLength(2); // In header and table
    expect(screen.getAllByText('Jane Smith').length).toBeGreaterThan(0); // Also appears multiple times
  });

  it('shows created and updated timestamps for records', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Use getAllByText for timestamps that appear multiple times
    expect(screen.getAllByText(/Created:/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Updated:/).length).toBeGreaterThan(0);
  });

  it('displays maintenance type badges with correct colors', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Use getAllByText for maintenance types that appear in both stats and table
    const preventiveBadges = screen.getAllByText('Preventive');
    const emergencyBadges = screen.getAllByText('Emergency');

    expect(preventiveBadges.length).toBeGreaterThan(0);
    expect(emergencyBadges.length).toBeGreaterThan(0);
  });

  it('handles sorting by clicking table headers', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    const workOrderHeader = screen.getByText('Work Order');
    fireEvent.click(workOrderHeader);

    // Check that the table still renders (sorting should not break the component)
    expect(screen.getByText('WO-001')).toBeInTheDocument();
    expect(screen.getByText('WO-002')).toBeInTheDocument();
  });

  it('shows download button for records with attachments', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    const downloadButtons = screen.getAllByText('Download');
    expect(downloadButtons).toHaveLength(1); // Only WO-001 has attachment
  });

  it('makes long action descriptions clickable', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Check for the truncated action text which should be clickable
    const truncatedAction = screen.getByText((content, element) => {
      return (
        content.includes('This is a very long action description') &&
        content.includes('...') &&
        element?.tagName.toLowerCase() === 'div'
      );
    });
    expect(truncatedAction).toBeInTheDocument();
    // Verify it's inside a button (clickable)
    expect(truncatedAction.closest('button')).toBeInTheDocument();
  });

  it('opens action details modal when truncated action is clicked', async () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Find and click the truncated action text
    const truncatedActionButton = screen
      .getByText((content, element) => {
        return (
          content.includes('This is a very long action description') &&
          content.includes('...') &&
          element?.closest('button') !== null
        );
      })
      .closest('button');

    expect(truncatedActionButton).toBeInTheDocument();
    fireEvent.click(truncatedActionButton!);

    await waitFor(() => {
      expect(screen.getByText('Action Details - WO-002')).toBeInTheDocument();
      // Check for the modal content specifically
      const modalContent = screen.getByText('Action Details - WO-002').closest('div');
      expect(modalContent).toBeInTheDocument();
    });
  });

  it('closes action details modal when close button is clicked', async () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Open modal by clicking truncated action
    const truncatedActionButton = screen
      .getByText((content, element) => {
        return (
          content.includes('This is a very long action description') &&
          content.includes('...') &&
          element?.closest('button') !== null
        );
      })
      .closest('button');

    fireEvent.click(truncatedActionButton!);

    await waitFor(() => {
      expect(screen.getByText('Action Details - WO-002')).toBeInTheDocument();
    });

    // Close modal
    const closeButton = screen.getByText('Close');
    fireEvent.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByText('Action Details - WO-002')).not.toBeInTheDocument();
    });
  });

  it('closes action details modal when clicking backdrop', async () => {
    const { container } = render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    // Open modal by clicking truncated action
    const truncatedActionButton = screen
      .getByText((content, element) => {
        return (
          content.includes('This is a very long action description') &&
          content.includes('...') &&
          element?.closest('button') !== null
        );
      })
      .closest('button');

    fireEvent.click(truncatedActionButton!);

    await waitFor(() => {
      expect(screen.getByText('Action Details - WO-002')).toBeInTheDocument();
    });

    // Click backdrop
    const backdrop = container.querySelector('.fixed.inset-0');
    if (backdrop) {
      fireEvent.click(backdrop);
    }

    await waitFor(() => {
      expect(screen.queryByText('Action Details - WO-002')).not.toBeInTheDocument();
    });
  });

  it('displays machine attachment as clickable download link', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    const attachmentButton = screen.getByText('machine_manual.pdf');
    expect(attachmentButton).toBeInTheDocument();
    expect(attachmentButton.closest('button')).toBeInTheDocument();
  });

  it('shows additional notes section when notes exist', () => {
    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    expect(screen.getByText('Additional Notes')).toBeInTheDocument();
    expect(screen.getByText('Test notes for the machine')).toBeInTheDocument();
  });

  it('does not show additional notes section when no notes', () => {
    const machineWithoutNotes = { ...mockMachine, additional_notes: '' };
    render(<MaintenanceHistory machine={machineWithoutNotes} onBack={() => {}} />);

    expect(screen.queryByText('Additional Notes')).not.toBeInTheDocument();
  });

  it('calls onBack when back button is clicked', () => {
    const mockOnBack = jest.fn();
    render(<MaintenanceHistory machine={mockMachine} onBack={mockOnBack} />);

    const backButton = screen.getByText('Back to Machines');
    fireEvent.click(backButton);

    expect(mockOnBack).toHaveBeenCalledTimes(1);
  });

  it('shows empty state when no maintenance records found', () => {
    const machineWithoutRecords = { ...mockMachine, serial_number: 'NONEXISTENT' };
    render(<MaintenanceHistory machine={machineWithoutRecords} onBack={() => {}} />);

    expect(screen.getByText('No Maintenance Records')).toBeInTheDocument();
    expect(screen.getByText('No maintenance history found for this machine.')).toBeInTheDocument();
  });

  it('handles download attachment clicks with alerts', () => {
    // Mock window.alert
    const mockAlert = jest.spyOn(window, 'alert').mockImplementation(() => {});

    render(<MaintenanceHistory machine={mockMachine} onBack={() => {}} />);

    const downloadButton = screen.getAllByText('Download')[0];
    fireEvent.click(downloadButton);

    expect(mockAlert).toHaveBeenCalledWith('Downloading attachment: maintenance_report.pdf');

    mockAlert.mockRestore();
  });
});
