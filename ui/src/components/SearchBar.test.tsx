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
    expect(screen.getByPlaceholderText('Search by any...')).toBeInTheDocument();

    // Check that property dropdown is present with default selection
    expect(screen.getByText('Any')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();

    // Check that search input has correct type and value
    const searchInput = screen.getByPlaceholderText('Search by any...');
    expect(searchInput).toHaveAttribute('type', 'text');
    expect(searchInput).toHaveValue('');
  });

  it('renders with different property types correctly', () => {
    // Test PPM status property
    const ppmSearchOptions: SearchOptions = {
      query: 'overdue',
      property: 'ppm_status',
    };
    const { rerender } = render(
      <SearchBar searchOptions={ppmSearchOptions} onSearch={mockOnSearch} />
    );

    expect(screen.getByRole('combobox')).toBeInTheDocument();
    // Use getAllByText to handle multiple elements with same text
    expect(screen.getAllByText('Overdue').length).toBeGreaterThan(0);
    // Check the dropdown button specifically
    const dropdownButton = screen.getByRole('button', { name: /PPM Status/ });
    expect(dropdownButton).toBeInTheDocument();

    // Test with any property and a query
    const anySearchOptions: SearchOptions = {
      query: 'test query',
      property: 'any',
    };
    rerender(<SearchBar searchOptions={anySearchOptions} onSearch={mockOnSearch} />);

    expect(screen.getByPlaceholderText('Search by any...')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test query')).toBeInTheDocument();
    const anyDropdownButton = screen.getByRole('button', { name: /Any/ });
    expect(anyDropdownButton).toBeInTheDocument();
  });
});
