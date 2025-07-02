import { render, screen } from '@testing-library/react';
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

  it('renders search bar with all basic functionality', () => {
    render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

    // Check that search input is present with correct placeholder
    expect(screen.getByPlaceholderText('Search by serial number...')).toBeInTheDocument();

    // Check that property dropdown is present with default selection
    expect(screen.getByText('Serial Number')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();

    // Check that search input has correct type and value
    const searchInput = screen.getByPlaceholderText('Search by serial number...');
    expect(searchInput).toHaveAttribute('type', 'text');
    expect(searchInput).toHaveValue('');
  });

  it('renders with different property types correctly', () => {
    // Test text property
    const textSearchOptions: SearchOptions = {
      query: 'test query',
      property: 'customer',
    };
    const { rerender } = render(
      <SearchBar searchOptions={textSearchOptions} onSearch={mockOnSearch} />
    );

    expect(screen.getByPlaceholderText('Search by customer...')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test query')).toBeInTheDocument();
    expect(screen.getByRole('button')).toHaveTextContent('Customer');

    // Test date property
    const dateSearchOptions: SearchOptions = {
      query: '2024-01-15',
      property: 'tnc_date',
    };
    rerender(<SearchBar searchOptions={dateSearchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByDisplayValue('2024-01-15')).toHaveAttribute('type', 'date');
    expect(screen.getByRole('button')).toHaveTextContent('TNC Date');

    // Test PPM status property
    const ppmSearchOptions: SearchOptions = {
      query: 'Overdue',
      property: 'ppm_status',
    };
    rerender(<SearchBar searchOptions={ppmSearchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByRole('combobox')).toBeInTheDocument();
    // Use getAllByText to handle multiple elements with same text
    expect(screen.getAllByText('Overdue').length).toBeGreaterThan(0);
    expect(screen.getByRole('button')).toHaveTextContent('PPM Status');
  });
});
