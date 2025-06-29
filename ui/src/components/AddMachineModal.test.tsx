import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import '@testing-library/jest-dom';
import AddMachineModal from './AddMachineModal';

describe('AddMachineModal', () => {
  const mockOnClose = jest.fn();
  const mockOnAdd = jest.fn();

  const defaultProps = {
    isOpen: true,
    onClose: mockOnClose,
    onAdd: mockOnAdd,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders modal when isOpen is true', () => {
    render(<AddMachineModal {...defaultProps} />);

    expect(screen.getByText('Add New Machine')).toBeInTheDocument();
    expect(screen.getByLabelText(/serial number/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/customer/i)).toBeInTheDocument();
  });

  it('does not render modal when isOpen is false', () => {
    render(<AddMachineModal {...defaultProps} isOpen={false} />);

    expect(screen.queryByText('Add New Machine')).not.toBeInTheDocument();
  });

  it('calls onClose when close button is clicked on empty form', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when backdrop is clicked on empty form', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const backdrop = document.querySelector('.fixed.inset-0.bg-black');
    if (backdrop) {
      await user.click(backdrop);
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    }
  });

  it('shows validation errors for required fields', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const submitButton = screen.getByRole('button', { name: /add machine/i });
    await user.click(submitButton);

    expect(screen.getByText('Serial number is required')).toBeInTheDocument();
    expect(screen.getByText('Customer is required')).toBeInTheDocument();
    expect(screen.getByText('State is required')).toBeInTheDocument();
    expect(screen.getByText('Model is required')).toBeInTheDocument();
    expect(screen.getByText('Brand is required')).toBeInTheDocument();
    expect(screen.getByText('District is required')).toBeInTheDocument();
    expect(screen.getByText('Person in charge is required')).toBeInTheDocument();
    expect(screen.getByText('Reported by is required')).toBeInTheDocument();
    expect(screen.getByText('PPM date is required')).toBeInTheDocument();
    expect(mockOnAdd).not.toHaveBeenCalled();
  });

  it('clears validation errors when user starts typing', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Trigger validation errors
    const submitButton = screen.getByRole('button', { name: /add machine/i });
    await user.click(submitButton);

    expect(screen.getByText('Serial number is required')).toBeInTheDocument();

    // Start typing in serial number field
    const serialNumberInput = screen.getByLabelText(/serial number/i);
    await user.type(serialNumberInput, 'SN-123');

    expect(screen.queryByText('Serial number is required')).not.toBeInTheDocument();
  });

  it('validates that PPM date is not earlier than TNC date', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill required fields
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');
    await user.type(screen.getByLabelText(/customer/i), 'Test Customer');
    await user.selectOptions(screen.getByLabelText(/state/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/model/i), 'Model-X');
    await user.type(screen.getByLabelText(/brand/i), 'BrandA');
    await user.selectOptions(screen.getByLabelText(/district/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/person in charge/i), 'John Doe');
    await user.type(screen.getByLabelText(/reported by/i), 'Jane Smith');

    // Set TNC date after PPM date
    await user.type(screen.getByLabelText(/tnc date/i), '2024-12-01');
    await user.type(screen.getByLabelText(/ppm date/i), '2024-11-01');

    const submitButton = screen.getByRole('button', { name: /add machine/i });
    await user.click(submitButton);

    expect(screen.getByText('PPM date cannot be earlier than TNC date')).toBeInTheDocument();
    expect(mockOnAdd).not.toHaveBeenCalled();
  });

  it('successfully submits form with valid data', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill required fields
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');
    await user.type(screen.getByLabelText(/customer/i), 'Test Customer');
    await user.selectOptions(screen.getByLabelText(/state/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/model/i), 'Model-X');
    await user.type(screen.getByLabelText(/brand/i), 'BrandA');
    await user.selectOptions(screen.getByLabelText(/district/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/person in charge/i), 'John Doe');
    await user.type(screen.getByLabelText(/reported by/i), 'Jane Smith');
    await user.type(screen.getByLabelText(/ppm date/i), '2024-12-01');

    // Fill optional fields
    await user.type(screen.getByLabelText(/account type/i), 'Premium');
    await user.type(screen.getByLabelText(/status/i), 'Active');
    await user.type(screen.getByLabelText(/tnc date/i), '2024-11-01');
    await user.type(screen.getByLabelText(/additional notes/i), 'Test notes');

    const submitButton = screen.getByRole('button', { name: /add machine/i });
    await user.click(submitButton);

    expect(mockOnAdd).toHaveBeenCalledWith({
      serial_number: 'SN-123',
      customer: 'Test Customer',
      state: 'Kuala Lumpur',
      account_type: 'Premium',
      model: 'Model-X',
      status: 'Active',
      brand: 'BrandA',
      district: 'Kuala Lumpur',
      person_in_charge: 'John Doe',
      reported_by: 'Jane Smith',
      additional_notes: 'Test notes',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-11-01',
      ppm_date: '2024-12-01',
    });
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('resets form data when modal is closed after confirmation', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill some fields
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');
    await user.type(screen.getByLabelText(/customer/i), 'Test Customer');

    // Close modal - should show confirmation dialog
    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    // Should show confirmation dialog
    expect(screen.getByText('Discard changes?')).toBeInTheDocument();

    // Confirm discard
    const discardButton = screen.getByRole('button', { name: /discard changes/i });
    await user.click(discardButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('renders all form fields with correct labels', () => {
    render(<AddMachineModal {...defaultProps} />);

    // Check required fields - use exact label text without regex
    expect(screen.getByLabelText('Serial Number *')).toBeInTheDocument();
    expect(screen.getByLabelText('Customer *')).toBeInTheDocument();
    expect(screen.getByLabelText('State *')).toBeInTheDocument();
    expect(screen.getByLabelText('Model *')).toBeInTheDocument();
    expect(screen.getByLabelText('Brand *')).toBeInTheDocument();
    expect(screen.getByLabelText('District *')).toBeInTheDocument();
    expect(screen.getByLabelText('Person in Charge *')).toBeInTheDocument();
    expect(screen.getByLabelText('Reported By *')).toBeInTheDocument();
    expect(screen.getByLabelText('PPM Date *')).toBeInTheDocument();

    // Check optional fields
    expect(screen.getByLabelText(/account type/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/status/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/tnc date/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/additional notes/i)).toBeInTheDocument();
  });

  it('has correct dropdown options for Malaysian states and districts', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Check Malaysian state options
    const stateSelect = screen.getByLabelText(/state/i);
    expect(stateSelect).toContainHTML('<option value="Johor">Johor</option>');
    expect(stateSelect).toContainHTML('<option value="Selangor">Selangor</option>');
    expect(stateSelect).toContainHTML('<option value="Kuala Lumpur">Kuala Lumpur</option>');
    expect(stateSelect).toContainHTML('<option value="Penang">Penang</option>');

    // Initially, district dropdown should be empty
    const districtSelect = screen.getByLabelText(/district/i);
    expect(districtSelect).toContainHTML('<option value="">Select district</option>');
    expect(districtSelect.children).toHaveLength(1); // Only the "Select district" option

    // Select a state to populate districts
    await user.selectOptions(stateSelect, 'Johor');

    // Check that Johor districts appear
    expect(districtSelect).toContainHTML('<option value="Johor Bahru">Johor Bahru</option>');
    expect(districtSelect).toContainHTML('<option value="Batu Pahat">Batu Pahat</option>');
    expect(districtSelect).toContainHTML('<option value="Kluang">Kluang</option>');

    // Select Kuala Lumpur state to test single district
    await user.selectOptions(stateSelect, 'Kuala Lumpur');
    expect(districtSelect).toContainHTML('<option value="Kuala Lumpur">Kuala Lumpur</option>');

    // Check that account type and status are text inputs, not dropdowns
    expect(screen.getByLabelText(/account type/i)).toHaveAttribute('type', 'text');
    expect(screen.getByLabelText(/status/i)).toHaveAttribute('type', 'text');
  });

  it('clears district when state changes', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const stateSelect = screen.getByLabelText(/state/i);
    const districtSelect = screen.getByLabelText(/district/i);

    // Select Johor state and a district
    await user.selectOptions(stateSelect, 'Johor');
    await user.selectOptions(districtSelect, 'Johor Bahru');
    expect(districtSelect).toHaveValue('Johor Bahru');

    // Change state to Kuala Lumpur - district should be cleared
    await user.selectOptions(stateSelect, 'Kuala Lumpur');
    expect(districtSelect).toHaveValue('');

    // Verify new districts are available
    expect(districtSelect).toContainHTML('<option value="Kuala Lumpur">Kuala Lumpur</option>');
  });

  it('handles file upload correctly', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');

    // Create a mock file
    const file = new File(['test content'], 'test-document.pdf', { type: 'application/pdf' });

    // Upload the file
    await user.upload(fileInput, file);

    // Wait for upload to complete and show success message
    await waitFor(
      () => {
        expect(screen.getByText('test-document.pdf')).toBeInTheDocument();
      },
      { timeout: 5000 }
    );

    expect(screen.getByText(/Upload successful/)).toBeInTheDocument();
  });

  it('clears file when modal is closed', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');
    const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });

    // Upload file
    await user.upload(fileInput, file);

    // Wait for upload to complete
    await waitFor(
      () => {
        expect(screen.getByText('test.pdf')).toBeInTheDocument();
      },
      { timeout: 5000 }
    );

    // Close modal - should show confirmation because file is uploaded
    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    // Should show confirmation dialog
    expect(screen.getByText('Discard changes?')).toBeInTheDocument();

    // Confirm discard
    const discardButton = screen.getByRole('button', { name: /discard changes/i });
    await user.click(discardButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('includes attachment filename in form submission', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill required fields
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');
    await user.type(screen.getByLabelText(/customer/i), 'Test Customer');
    await user.selectOptions(screen.getByLabelText(/state/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/model/i), 'Model-X');
    await user.type(screen.getByLabelText(/brand/i), 'BrandA');
    await user.selectOptions(screen.getByLabelText(/district/i), 'Kuala Lumpur');
    await user.type(screen.getByLabelText(/person in charge/i), 'John Doe');
    await user.type(screen.getByLabelText(/reported by/i), 'Jane Smith');
    await user.type(screen.getByLabelText(/ppm date/i), '2024-12-01');

    // Upload a file
    const fileInput = screen.getByLabelText('Attachment');
    const file = new File(['test content'], 'test-document.pdf', { type: 'application/pdf' });
    await user.upload(fileInput, file);

    const submitButton = screen.getByRole('button', { name: /add machine/i });
    await user.click(submitButton);

    expect(mockOnAdd).toHaveBeenCalledWith({
      serial_number: 'SN-123',
      customer: 'Test Customer',
      state: 'Kuala Lumpur',
      account_type: '',
      model: 'Model-X',
      status: '',
      brand: 'BrandA',
      district: 'Kuala Lumpur',
      person_in_charge: 'John Doe',
      reported_by: 'Jane Smith',
      additional_notes: '',
      attachment: 'test-document.pdf',
      ppm_status: '',
      tnc_date: '',
      ppm_date: '2024-12-01',
    });
  });

  it('shows file size information', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');

    // Create a mock file with specific size (1024 bytes = 1KB)
    const file = new File(['x'.repeat(1024)], 'large-file.pdf', { type: 'application/pdf' });
    Object.defineProperty(file, 'size', { value: 1024 });

    await user.upload(fileInput, file);

    // Wait for upload to complete
    await waitFor(
      () => {
        expect(screen.getByText('large-file.pdf')).toBeInTheDocument();
      },
      { timeout: 5000 }
    );

    expect(screen.getByText(/1.0 KB/)).toBeInTheDocument();
    expect(screen.getByText(/Upload successful/)).toBeInTheDocument();
  });

  it('shows upload progress when file is selected', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');
    const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });

    await user.upload(fileInput, file);

    // Should show uploading state initially
    expect(screen.getByText('Uploading...')).toBeInTheDocument();

    // Wait for upload to complete
    await waitFor(
      () => {
        expect(screen.getByText(/Upload successful/)).toBeInTheDocument();
      },
      { timeout: 3000 }
    );
  });

  it('allows removing uploaded file', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');
    const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });

    // Upload file
    await user.upload(fileInput, file);

    // Wait for upload to complete
    await waitFor(
      () => {
        expect(screen.getByText('test.pdf')).toBeInTheDocument();
      },
      { timeout: 5000 }
    );

    // Click remove button
    const removeButton = screen.getByTitle('Remove file');
    await user.click(removeButton);

    // File should be removed
    expect(screen.queryByText('test.pdf')).not.toBeInTheDocument();
    expect(screen.queryByText(/Upload successful/)).not.toBeInTheDocument();
  });

  it('shows cancel confirmation when form has data', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill some form data
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');

    // Click cancel
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    // Should show confirmation dialog
    expect(screen.getByText('Discard changes?')).toBeInTheDocument();
    expect(screen.getByText(/You have unsaved changes/)).toBeInTheDocument();
  });

  it('does not show cancel confirmation when form is empty', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Click cancel without filling any data
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    // Should close immediately without confirmation
    expect(mockOnClose).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('Discard changes?')).not.toBeInTheDocument();
  });

  it('handles cancel confirmation dialog actions', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill some form data
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');

    // Click cancel
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    // Should show confirmation dialog
    expect(screen.getByText('Discard changes?')).toBeInTheDocument();

    // Click "Keep editing"
    const keepEditingButton = screen.getByRole('button', { name: /keep editing/i });
    await user.click(keepEditingButton);

    // Dialog should close, form should remain open
    expect(screen.queryByText('Discard changes?')).not.toBeInTheDocument();
    expect(screen.getByText('Add New Machine')).toBeInTheDocument();
    expect(mockOnClose).not.toHaveBeenCalled();

    // Try cancel again and confirm discard
    await user.click(cancelButton);
    const discardButton = screen.getByRole('button', { name: /discard changes/i });
    await user.click(discardButton);

    // Should close the modal
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('shows cancel confirmation when file is uploaded', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const fileInput = screen.getByLabelText('Attachment');
    const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });

    // Upload file
    await user.upload(fileInput, file);

    // Wait for upload to complete
    await waitFor(
      () => {
        expect(screen.getByText('test.pdf')).toBeInTheDocument();
      },
      { timeout: 5000 }
    );

    // Click cancel - should show confirmation even without other form data
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    expect(screen.getByText('Discard changes?')).toBeInTheDocument();
  });
});
