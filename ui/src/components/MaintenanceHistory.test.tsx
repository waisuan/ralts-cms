import { render, screen, waitFor } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import MaintenanceHistory from './MaintenanceHistory';
import { Machine } from '../types/machine';
import { MaintenanceService } from '../services/maintenanceService';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock the maintenance service
jest.mock('../services/maintenanceService', () => ({
  MaintenanceService: {
    getMaintenanceList: jest.fn(),
    getMaintenance: jest.fn(),
    createMaintenance: jest.fn(),
    updateMaintenance: jest.fn(),
    deleteMaintenance: jest.fn(),
  },
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
    
    // Reset all mocks
    jest.clearAllMocks();
  });

  it('renders machine information correctly', async () => {
    // Mock the API response with work order type counts
    const mockApiResponse = {
      data: {
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
          },
        ],
        preventative_count: 2,
        corrective_count: 1,
        emergency_count: 3,
        inspection_count: 1,
        count: 7,
        limit: 10,
        offset: 0,
        sort: 'work_order_date_desc',
      },
    };

    (MaintenanceService.getMaintenanceList as jest.Mock).mockResolvedValue(mockApiResponse);

    render(<MaintenanceHistory machine={mockMachine} />);
    
    // Wait for the API call to complete and verify the machine information and counts are displayed
    await waitFor(() => {
      expect(screen.getByText(mockMachine.serial_number)).toBeInTheDocument();
      expect(screen.getByText(mockMachine.model)).toBeInTheDocument();
    });
    
    // Verify the counts are displayed (using more specific selectors)
    await waitFor(() => {
      expect(screen.getByText('2')).toBeInTheDocument(); // Preventative count
      expect(screen.getByText('3')).toBeInTheDocument(); // Emergency count
    });
    
    // Check that both corrective and inspection counts of 1 are present
    const countElements = screen.getAllByText('1');
    expect(countElements.length).toBeGreaterThanOrEqual(2); // At least 2 elements with "1"
  });

  it('displays work order type counts from backend', async () => {
    const mockApiResponse = {
      data: {
        maintenance: [],
        preventative_count: 5,
        corrective_count: 2,
        emergency_count: 1,
        inspection_count: 3,
        count: 11,
        limit: 10,
        offset: 0,
        sort: 'work_order_date_desc',
      },
    };

    (MaintenanceService.getMaintenanceList as jest.Mock).mockResolvedValue(mockApiResponse);

    render(<MaintenanceHistory machine={mockMachine} />);
    
    await waitFor(() => {
      // Use more specific queries that target the count display elements
      const preventativeCountElements = screen.getAllByText('5');
      const correctiveCountElements = screen.getAllByText('2');
      const emergencyCountElements = screen.getAllByText('1');
      const inspectionCountElements = screen.getAllByText('3');
      
      // Verify counts are displayed (there might be multiple elements with the same text)
      expect(preventativeCountElements.length).toBeGreaterThanOrEqual(1);
      expect(correctiveCountElements.length).toBeGreaterThanOrEqual(1);
      expect(emergencyCountElements.length).toBeGreaterThanOrEqual(1);
      expect(inspectionCountElements.length).toBeGreaterThanOrEqual(1);
      
      // Verify the category labels are also present
      expect(screen.getByText('Preventive')).toBeInTheDocument();
      expect(screen.getByText('Corrective')).toBeInTheDocument();
      expect(screen.getByText('Emergency')).toBeInTheDocument();
      expect(screen.getByText('Inspection')).toBeInTheDocument();
    });
  });
});
