import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SearchBar, { SearchOptions } from './SearchBar';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

describe('SearchBar', () => {
  const mockOnSearch = jest.fn();
  const defaultSearchOptions: SearchOptions = {
    query: '',
    property: DEFAULT_SEARCH_PROPERTY,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders with default values', () => {
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByPlaceholderText('Search by serial number...')).toBeInTheDocument();
    expect(screen.getByText('Serial Number')).toBeInTheDocument();
  });

  it('calls onSearch when query changes', async () => {
    const user = userEvent.setup();
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    const input = screen.getByPlaceholderText('Search by serial number...');
    await user.type(input, 't');

    // Should be called when typing
    expect(mockOnSearch).toHaveBeenCalledWith({
      query: 't',
      property: DEFAULT_SEARCH_PROPERTY,
    });
  });

  it('opens dropdown when clicked', async () => {
    const user = userEvent.setup();
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    // Check for options that are not the currently selected one (Serial Number)
    expect(screen.getByText('Customer')).toBeInTheDocument();
    expect(screen.getByText('State')).toBeInTheDocument();
    expect(screen.getByText('Account Type')).toBeInTheDocument();
    expect(screen.getByText('Model')).toBeInTheDocument();
    expect(screen.getByText('Brand')).toBeInTheDocument();
    expect(screen.getByText('District')).toBeInTheDocument();
    expect(screen.getByText('Person in Charge')).toBeInTheDocument();
    expect(screen.getByText('Reported By')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('TNC Date')).toBeInTheDocument();
    expect(screen.getByText('PPM Date')).toBeInTheDocument();

    // Verify Serial Number appears twice (button + dropdown)
    expect(screen.getAllByText('Serial Number')).toHaveLength(2);
  });

  it('calls onSearch when property is selected', async () => {
    const user = userEvent.setup();
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    const customerOption = screen.getByText('Customer');
    await user.click(customerOption);

    expect(mockOnSearch).toHaveBeenCalledWith({
      query: '',
      property: 'customer',
    });
  });

  it('clears query when switching to a different property', async () => {
    const user = userEvent.setup();
    const searchOptions: SearchOptions = {
      query: 'some search',
      property: 'customer',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    const modelOption = screen.getByText('Model');
    await user.click(modelOption);

    // Should clear the query when switching properties
    expect(mockOnSearch).toHaveBeenCalledWith({
      query: '',
      property: 'model',
    });
  });

  it('updates placeholder text based on selected property', () => {
    const searchOptions: SearchOptions = {
      query: '',
      property: 'customer',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByPlaceholderText('Search by customer...')).toBeInTheDocument();
    // Use more specific selector to avoid text collision
    expect(screen.getByRole('button')).toHaveTextContent('Customer');
  });

  it('shows text input for text properties', () => {
    const searchOptions: SearchOptions = {
      query: '',
      property: 'customer',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByDisplayValue('')).toHaveAttribute('type', 'text');
    expect(screen.getByPlaceholderText('Search by customer...')).toBeInTheDocument();
  });

  it('shows date input for date properties', () => {
    const searchOptions: SearchOptions = {
      query: '',
      property: 'tnc_date',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByDisplayValue('')).toHaveAttribute('type', 'date');
    expect(screen.queryByPlaceholderText(/search by/i)).not.toBeInTheDocument();
  });

  it('shows calendar icon for date properties', () => {
    const searchOptions: SearchOptions = {
      query: '',
      property: 'ppm_date',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    // Check for calendar icon path
    const calendarIcon = document.querySelector(
      'path[d*="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"]'
    );
    expect(calendarIcon).toBeInTheDocument();
  });

  it('shows different search info for date properties', () => {
    const searchOptions: SearchOptions = {
      query: '2024-01-15',
      property: 'tnc_date',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    // Verify the date input shows the correct value
    expect(screen.getByDisplayValue('2024-01-15')).toBeInTheDocument();
    expect(screen.getByDisplayValue('2024-01-15')).toHaveAttribute('type', 'date');

    // Verify TNC Date label is shown in the dropdown button
    expect(screen.getByRole('button')).toHaveTextContent('TNC Date');
  });

  it('shows search info when query is present for text properties', () => {
    const searchOptions: SearchOptions = {
      query: 'test search',
      property: 'model',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByText(/Searching for.*test search.*in/)).toBeInTheDocument();
    // Check for the search info specifically in the info section
    const searchInfo = screen.getByText(/Searching for.*test search.*in/).parentElement;
    expect(searchInfo).toHaveTextContent('Model');
  });

  it('closes dropdown when property is selected', async () => {
    const user = userEvent.setup();
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    // Verify dropdown is open by checking for Customer option
    expect(screen.getByText('Customer')).toBeInTheDocument();

    const customerOption = screen.getByText('Customer');
    await user.click(customerOption);

    // After selection, dropdown should close - Customer option should not be visible
    expect(screen.queryByText('Customer')).not.toBeInTheDocument();
  });

  it('highlights selected property in dropdown', async () => {
    const user = userEvent.setup();
    const searchOptions: SearchOptions = {
      query: '',
      property: 'brand',
    };
    render(<SearchBar searchOptions={searchOptions} onSearch={mockOnSearch} />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    // Find the brand option in the dropdown specifically
    const dropdownOptions = screen.getAllByText('Brand');
    const brandOptionInDropdown = dropdownOptions.find((option) =>
      option.closest('div')?.classList.contains('absolute')
    );

    expect(brandOptionInDropdown).toHaveClass('bg-blue-50', 'text-blue-600', 'font-medium');
  });
});
