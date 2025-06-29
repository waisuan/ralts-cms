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

  it('shows search info when query is present', () => {
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
