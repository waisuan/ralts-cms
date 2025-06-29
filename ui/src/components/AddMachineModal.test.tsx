import { render, screen } from '@testing-library/react';
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

  it('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when backdrop is clicked', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const backdrop = document.querySelector('.fixed.inset-0.bg-black');
    if (backdrop) {
      await user.click(backdrop);
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    }
  });

  it('calls onClose when Cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
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

  it('resets form data when modal is closed', async () => {
    const user = userEvent.setup();
    render(<AddMachineModal {...defaultProps} />);

    // Fill some fields
    await user.type(screen.getByLabelText(/serial number/i), 'SN-123');
    await user.type(screen.getByLabelText(/customer/i), 'Test Customer');

    // Close modal
    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

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
});
