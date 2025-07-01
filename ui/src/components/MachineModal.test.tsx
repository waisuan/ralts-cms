import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { act } from 'react';
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

  describe('Add Mode', () => {
    const defaultAddProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    it('renders with add mode title', () => {
      render(<MachineModal {...defaultAddProps} />);

      expect(screen.getByRole('heading', { name: 'Add New Machine' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /add machine/i })).toBeInTheDocument();
    });

    it('does not render when isOpen is false', () => {
      render(<MachineModal {...defaultAddProps} isOpen={false} />);

      expect(screen.queryByRole('heading', { name: 'Add New Machine' })).not.toBeInTheDocument();
    });

    it('renders all form fields with empty values', () => {
      render(<MachineModal {...defaultAddProps} />);

      expect(screen.getByLabelText(/serial number/i)).toHaveValue('');
      expect(screen.getByLabelText(/customer/i)).toHaveValue('');
      expect(screen.getByLabelText(/state/i)).toHaveValue('');
      expect(screen.getByLabelText(/account type/i)).toHaveValue('');
      expect(screen.getByLabelText(/model/i)).toHaveValue('');
      expect(screen.getByLabelText(/brand/i)).toHaveValue('');
      expect(screen.getByLabelText(/district/i)).toHaveValue('');
      expect(screen.getByLabelText(/person in charge/i)).toHaveValue('');
      expect(screen.getByLabelText(/reported by/i)).toHaveValue('');
      expect(screen.getByLabelText(/additional notes/i)).toHaveValue('');
    });

    it('calls onSubmit with form data when submitted with valid data', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultAddProps} />);

      // Fill required fields
      await user.type(screen.getByLabelText(/serial number/i), 'NEW-001');
      await user.type(screen.getByLabelText(/customer/i), 'New Customer');
      await user.selectOptions(screen.getByLabelText(/state/i), 'Selangor');
      await user.type(screen.getByLabelText(/model/i), 'Model-Y');
      await user.type(screen.getByLabelText(/brand/i), 'New Brand');
      await user.selectOptions(screen.getByLabelText(/district/i), 'Petaling');
      await user.type(screen.getByLabelText(/person in charge/i), 'John Smith');
      await user.type(screen.getByLabelText(/reported by/i), 'Jane Doe');
      await user.type(screen.getByLabelText(/ppm date/i), '2024-07-01');

      await user.click(screen.getByRole('button', { name: /add machine/i }));

      expect(defaultAddProps.onSubmit).toHaveBeenCalledWith({
        serial_number: 'NEW-001',
        customer: 'New Customer',
        state: 'Selangor',
        account_type: '',
        model: 'Model-Y',
        status: '',
        brand: 'New Brand',
        district: 'Petaling',
        person_in_charge: 'John Smith',
        reported_by: 'Jane Doe',
        additional_notes: '',
        attachment: '',
        ppm_status: '',
        tnc_date: '',
        ppm_date: '2024-07-01',
      });
    });

    it('shows validation errors for required fields', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultAddProps} />);

      await user.click(screen.getByRole('button', { name: /add machine/i }));

      expect(screen.getByText('Serial number is required')).toBeInTheDocument();
      expect(screen.getByText('Customer is required')).toBeInTheDocument();
      expect(screen.getByText('State is required')).toBeInTheDocument();
      expect(screen.getByText('Model is required')).toBeInTheDocument();
      expect(screen.getByText('Brand is required')).toBeInTheDocument();
      expect(screen.getByText('District is required')).toBeInTheDocument();
      expect(screen.getByText('Person in charge is required')).toBeInTheDocument();
      expect(screen.getByText('Reported by is required')).toBeInTheDocument();
      expect(screen.getByText('PPM date is required')).toBeInTheDocument();
    });

    it('calls onClose when cancel button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultAddProps} />);

      await user.click(screen.getByRole('button', { name: /cancel/i }));

      expect(defaultAddProps.onClose).toHaveBeenCalled();
    });

    it('shows confirmation dialog when closing with unsaved changes', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultAddProps} />);

      // Make some changes
      await user.type(screen.getByLabelText(/serial number/i), 'TEST');

      // Try to close
      await user.click(screen.getByRole('button', { name: /cancel/i }));

      expect(screen.getByText('Discard changes?')).toBeInTheDocument();
      expect(screen.getByText(/You have unsaved changes/)).toBeInTheDocument();
    });
  });

  describe('Edit Mode', () => {
    const defaultEditProps = {
      isOpen: true,
      mode: 'edit' as const,
      machine: mockMachine,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    it('renders with edit mode title', () => {
      render(<MachineModal {...defaultEditProps} />);

      expect(screen.getByRole('heading', { name: 'Edit Machine' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /update machine/i })).toBeInTheDocument();
    });

    it('does not render when machine is null', () => {
      render(<MachineModal {...defaultEditProps} machine={null} />);

      expect(screen.queryByRole('heading', { name: 'Edit Machine' })).not.toBeInTheDocument();
    });

    it('pre-populates form fields with machine data', () => {
      render(<MachineModal {...defaultEditProps} />);

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

    it('shows existing attachment', () => {
      render(<MachineModal {...defaultEditProps} />);

      expect(screen.getByText('test-file.pdf')).toBeInTheDocument();
      expect(screen.getByText('Current attachment')).toBeInTheDocument();
    });

    it('calls onSubmit with updated machine data when submitted', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultEditProps} />);

      // Update customer field
      const customerInput = screen.getByDisplayValue('Test Customer');
      await user.clear(customerInput);
      await user.type(customerInput, 'Updated Customer');

      await user.click(screen.getByRole('button', { name: /update machine/i }));

      expect(defaultEditProps.onSubmit).toHaveBeenCalledWith({
        ...mockMachine,
        customer: 'Updated Customer',
        updated_at: expect.any(String),
      });
    });

    it('updates districts when state changes', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultEditProps} />);

      // Change state
      await user.selectOptions(screen.getByLabelText(/state/i), 'Johor');

      // District should be cleared and new options should be available
      expect(screen.getByLabelText(/district/i)).toHaveValue('');

      // Check that Johor districts are now available (example: Johor Bahru)
      await user.selectOptions(screen.getByLabelText(/district/i), 'Johor Bahru');
      expect(screen.getByLabelText(/district/i)).toHaveValue('Johor Bahru');
    });

    it('removes existing attachment when remove button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultEditProps} />);

      expect(screen.getByText('test-file.pdf')).toBeInTheDocument();

      // Find and click the remove button for the attachment
      const removeButtons = screen.getAllByTitle('Remove attachment');
      await user.click(removeButtons[0]);

      expect(screen.queryByText('test-file.pdf')).not.toBeInTheDocument();
    });
  });

  describe('File Upload', () => {
    const defaultProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    it('handles file upload', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = screen.getByLabelText(/attachment/i);

      await user.upload(fileInput, file);

      // Should show upload progress initially
      expect(screen.getByText('Uploading...')).toBeInTheDocument();

      // Advance timers to complete the upload simulation with act()
      await act(async () => {
        jest.advanceTimersByTime(1000);
      });

      // Wait for upload to complete - look for the file size and success text
      await waitFor(() => {
        expect(screen.getByText(/Upload successful/)).toBeInTheDocument();
      });

      expect(screen.getByText('test.pdf')).toBeInTheDocument();
    });

    it('allows removing uploaded file', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });
      const fileInput = screen.getByLabelText(/attachment/i);

      await user.upload(fileInput, file);

      // Advance timers to complete the upload simulation with act()
      await act(async () => {
        jest.advanceTimersByTime(1000);
      });

      // Wait for upload to complete - look for the file size and success text
      await waitFor(() => {
        expect(screen.getByText(/Upload successful/)).toBeInTheDocument();
      });

      // Remove the file
      const removeButton = screen.getByTitle('Remove file');
      await user.click(removeButton);

      expect(screen.queryByText('test.pdf')).not.toBeInTheDocument();
    });
  });

  describe('Validation', () => {
    const defaultProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    it('validates PPM date is not earlier than TNC date', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      // Fill required fields
      await user.type(screen.getByLabelText(/serial number/i), 'TEST-001');
      await user.type(screen.getByLabelText(/customer/i), 'Test Customer');
      await user.selectOptions(screen.getByLabelText(/state/i), 'Selangor');
      await user.type(screen.getByLabelText(/model/i), 'Model-X');
      await user.type(screen.getByLabelText(/brand/i), 'Test Brand');
      await user.selectOptions(screen.getByLabelText(/district/i), 'Petaling');
      await user.type(screen.getByLabelText(/person in charge/i), 'John Doe');
      await user.type(screen.getByLabelText(/reported by/i), 'Jane Smith');

      // Set TNC date after PPM date
      await user.type(screen.getByLabelText(/tnc date/i), '2024-07-15');
      await user.type(screen.getByLabelText(/ppm date/i), '2024-07-01');

      await user.click(screen.getByRole('button', { name: /add machine/i }));

      expect(screen.getByText('PPM date cannot be earlier than TNC date')).toBeInTheDocument();
    });

    it('clears field errors when user starts typing', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      // Submit form to trigger errors
      await user.click(screen.getByRole('button', { name: /add machine/i }));
      expect(screen.getByText('Serial number is required')).toBeInTheDocument();

      // Start typing in the field
      await user.type(screen.getByLabelText(/serial number/i), 'T');

      // Error should be cleared
      expect(screen.queryByText('Serial number is required')).not.toBeInTheDocument();
    });
  });

  describe('Modal Behavior', () => {
    const defaultProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    it('closes modal when backdrop is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      const backdrop = document.querySelector('.fixed.inset-0.bg-black');
      if (backdrop) {
        await user.click(backdrop);
      }

      expect(defaultProps.onClose).toHaveBeenCalled();
    });

    it('closes modal when X button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      render(<MachineModal {...defaultProps} />);

      const closeButton = screen.getByLabelText('Close');
      await user.click(closeButton);

      expect(defaultProps.onClose).toHaveBeenCalled();
    });

    it('resets form when modal is closed and reopened in add mode', () => {
      const { rerender } = render(<MachineModal {...defaultProps} />);

      // Fill a field
      fireEvent.change(screen.getByLabelText(/serial number/i), { target: { value: 'TEST' } });
      expect(screen.getByLabelText(/serial number/i)).toHaveValue('TEST');

      // Close and reopen modal
      rerender(<MachineModal {...defaultProps} isOpen={false} />);
      rerender(<MachineModal {...defaultProps} isOpen={true} />);

      // Field should be cleared
      expect(screen.getByLabelText(/serial number/i)).toHaveValue('');
    });
  });
});
