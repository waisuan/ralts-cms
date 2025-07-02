import { render, screen } from '@testing-library/react';
import MachineModal from './MachineModal';
import { Machine } from '../types/machine';

// Mock the current date for consistent testing
const MOCK_CURRENT_DATE = '2024-06-29T12:00:00.000Z';

// Mock machine data for testing
const mockMachine: Machine = {
  serial_number: 'TEST-001',
  customer: 'Test Customer',
  state: 'Selangor',
  account_type: 'Premium',
  model: 'Model-X',
  status: 'Active',
  brand: 'Test Brand',
  district: 'Petaling',
  person_in_charge: 'John Doe',
  reported_by: 'Jane Smith',
  additional_notes: 'Test notes',
  attachment: 'test-file.pdf',
  ppm_status: 'Good',
  tnc_date: '2024-06-15',
  ppm_date: '2024-07-15',
  created_at: '2024-01-01T00:00:00.000Z',
  updated_at: '2024-06-01T00:00:00.000Z',
};

describe('MachineModal', () => {
  beforeEach(() => {
    // Mock the current date for consistent testing
    jest.useFakeTimers();
    jest.setSystemTime(new Date(MOCK_CURRENT_DATE));
  });

  afterEach(() => {
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it('renders correctly in add mode with form fields', () => {
    const addProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    render(<MachineModal {...addProps} />);

    // Check that the modal renders with add mode title
    expect(screen.getByRole('heading', { name: 'Add New Machine' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add machine/i })).toBeInTheDocument();

    // Check that form fields are present
    expect(screen.getByLabelText(/serial number/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/customer/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/state/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/model/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/brand/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/district/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/person in charge/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/reported by/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/additional notes/i)).toBeInTheDocument();
  });

  it('renders correctly in edit mode with pre-populated data', () => {
    const editProps = {
      isOpen: true,
      mode: 'edit' as const,
      machine: mockMachine,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    render(<MachineModal {...editProps} />);

    // Check that the modal renders with edit mode title
    expect(screen.getByRole('heading', { name: 'Edit Machine' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /update machine/i })).toBeInTheDocument();

    // Check that form fields are pre-populated with machine data
    expect(screen.getByDisplayValue('TEST-001')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Customer')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Selangor')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Premium')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Model-X')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Brand')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Petaling')).toBeInTheDocument();
    expect(screen.getByDisplayValue('John Doe')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Jane Smith')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test notes')).toBeInTheDocument();
  });
});
