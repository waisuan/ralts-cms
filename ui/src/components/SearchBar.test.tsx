import { act, fireEvent, render, screen } from '@testing-library/react';
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

  describe('debounced text input', () => {
    beforeEach(() => {
      jest.useFakeTimers();
    });

    afterEach(() => {
      act(() => {
        jest.runOnlyPendingTimers();
      });
      jest.useRealTimers();
    });

    it('does not call onSearch until 300ms after the user stops typing', () => {
      render(<SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />);

      const input = screen.getByPlaceholderText('Search by any...');

      fireEvent.change(input, { target: { value: 'a' } });
      fireEvent.change(input, { target: { value: 'ab' } });
      fireEvent.change(input, { target: { value: 'abc' } });

      expect(mockOnSearch).not.toHaveBeenCalled();

      act(() => {
        jest.advanceTimersByTime(299);
      });
      expect(mockOnSearch).not.toHaveBeenCalled();

      act(() => {
        jest.advanceTimersByTime(1);
      });

      expect(mockOnSearch).toHaveBeenCalledTimes(1);
      expect(mockOnSearch).toHaveBeenCalledWith({
        query: 'abc',
        property: DEFAULT_SEARCH_PROPERTY,
      });
    });

    it('commits the clear action immediately', () => {
      render(
        <SearchBar
          searchOptions={{ query: 'hello', property: DEFAULT_SEARCH_PROPERTY }}
          onSearch={mockOnSearch}
        />,
      );

      const clearBtn = screen.getByTitle('Clear search');
      fireEvent.click(clearBtn);

      expect(mockOnSearch).toHaveBeenCalledTimes(1);
      expect(mockOnSearch).toHaveBeenCalledWith({
        query: '',
        property: DEFAULT_SEARCH_PROPERTY,
      });
    });

    it('overrides in-flight typing when an external query change arrives', () => {
      const { rerender } = render(
        <SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />,
      );

      const input = screen.getByPlaceholderText('Search by any...') as HTMLInputElement;

      fireEvent.change(input, { target: { value: 'klin' } });
      expect(input.value).toBe('klin');

      // Parent commits a new query before the debounce timer fires
      // (e.g. URL change, dismiss banner, programmatic update).
      rerender(
        <SearchBar
          searchOptions={{ query: 'externally-set', property: DEFAULT_SEARCH_PROPERTY }}
          onSearch={mockOnSearch}
        />,
      );

      expect(input.value).toBe('externally-set');

      act(() => {
        jest.advanceTimersByTime(500);
      });

      // The pending debounce must NOT clobber the externally-applied query
      // by firing onSearch with the stale 'klin' value.
      expect(mockOnSearch).not.toHaveBeenCalled();
    });

    it('drops a pending text commit when the parent flips property mid-debounce', () => {
      const { rerender } = render(
        <SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />,
      );

      const input = screen.getByPlaceholderText('Search by any...') as HTMLInputElement;

      fireEvent.change(input, { target: { value: 'klin' } });

      // Parent switches property without changing query (e.g. dropdown reset)
      rerender(
        <SearchBar
          searchOptions={{ query: '', property: 'ppm_status' }}
          onSearch={mockOnSearch}
        />,
      );

      act(() => {
        jest.advanceTimersByTime(500);
      });

      // Pending debounce must not write the stale 'klin' text against the new
      // property — that would commit garbage as a PPM status.
      expect(mockOnSearch).not.toHaveBeenCalled();
    });

    it('still debounces user typing after an external override', () => {
      const { rerender } = render(
        <SearchBar searchOptions={defaultSearchOptions} onSearch={mockOnSearch} />,
      );

      const input = screen.getByPlaceholderText('Search by any...') as HTMLInputElement;

      fireEvent.change(input, { target: { value: 'klin' } });
      rerender(
        <SearchBar
          searchOptions={{ query: 'externally-set', property: DEFAULT_SEARCH_PROPERTY }}
          onSearch={mockOnSearch}
        />,
      );

      // User edits the now-overwritten input value.
      fireEvent.change(input, { target: { value: 'externally-set y' } });

      act(() => {
        jest.advanceTimersByTime(300);
      });

      expect(mockOnSearch).toHaveBeenCalledTimes(1);
      expect(mockOnSearch).toHaveBeenCalledWith({
        query: 'externally-set y',
        property: DEFAULT_SEARCH_PROPERTY,
      });
    });
  });
});
