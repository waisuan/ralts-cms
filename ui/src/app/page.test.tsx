import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Home from './page';

interface MockRecordsListProps {
  searchOptions?: { query: string; property: string };
  filterType?: string;
  sortBy?: string;
  onShowAll?: () => void;
  onSortChange?: (sortBy: string) => void;
}

interface MockSearchBarProps {
  onSearch: (options: { query: string; property: string }) => void;
  searchOptions: { query: string; property: string };
}

interface MockOverdueAlertProps {
  stats: {
    overdueCount: number;
    dueCount: number;
    totalCriticalCount: number;
    overdueMachines: unknown[];
    dueMachines: unknown[];
  };
  onShowOverdue: () => void;
  onShowDue: () => void;
  onDismissOverdue?: () => void;
  onDismissDue?: () => void;
  isOverdueDismissed?: boolean;
  isDueDismissed?: boolean;
}

// Mock the components since we're testing the page integration
jest.mock('../components/RecordsList', () => {
  return function MockRecordsList({ searchOptions, filterType, sortBy }: MockRecordsListProps) {
    return (
      <div data-testid="records-list">
        <div>Mock Records List</div>
        <div>Search Query: {searchOptions?.query || ''}</div>
        <div>Search Property: {searchOptions?.property || ''}</div>
        <div>Filter Type: {filterType || 'all'}</div>
        <div>Sort By: {sortBy || 'newest'}</div>
      </div>
    );
  };
});

jest.mock('../components/SearchBar', () => {
  return function MockSearchBar({ onSearch, searchOptions }: MockSearchBarProps) {
    return (
      <div data-testid="search-bar">
        <input
          data-testid="search-input"
          placeholder="Search by serial number..."
          value={searchOptions?.query || ''}
          onChange={(e) => onSearch({ ...searchOptions, query: e.target.value })}
        />
        <div>Serial Number</div>
      </div>
    );
  };
});

jest.mock('../components/OverdueAlert', () => {
  return function MockOverdueAlert({ 
    stats, 
    onShowOverdue, 
    onShowDue, 
    onDismissOverdue,
    onDismissDue,
    isOverdueDismissed = false,
    isDueDismissed = false
  }: MockOverdueAlertProps) {
    if ((stats.overdueCount === 0 || isOverdueDismissed) && (stats.dueCount === 0 || isDueDismissed)) {
      return null;
    }

    return (
      <div data-testid="overdue-alert">
        {stats.overdueCount > 0 && !isOverdueDismissed && (
          <div>
            <div>
              {stats.overdueCount} machine{stats.overdueCount !== 1 ? 's are' : ' is'} overdue for
              PPM maintenance
            </div>
            <button onClick={onShowOverdue}>View Overdue</button>
            {onDismissOverdue && (
              <button onClick={onDismissOverdue}>Dismiss Overdue</button>
            )}
          </div>
        )}
        {stats.dueCount > 0 && !isDueDismissed && (
          <div>
            <div>
              {stats.dueCount} machine{stats.dueCount !== 1 ? 's are' : ' is'} due for PPM
              maintenance today
            </div>
            <button onClick={onShowDue}>View Due</button>
            {onDismissDue && (
              <button onClick={onDismissDue}>Dismiss Due</button>
            )}
          </div>
        )}
      </div>
    );
  };
});

// Mock the useOverdueStats hook
jest.mock('../hooks/useOverdueStats', () => ({
  useOverdueStats: () => ({
    overdueCount: 1,
    dueCount: 1,
    dueSoonCount: 1,
    totalCriticalCount: 2,
    overdueMachines: [],
    dueMachines: [],
  }),
}));

describe('Home Page', () => {
  it('renders the main page with all components', () => {
    render(<Home />);

    // Check main title and description
    expect(screen.getByText('Ralts CMS')).toBeInTheDocument();
    expect(screen.getByText('Content Management System')).toBeInTheDocument();

    // Check components are present
    expect(screen.getByTestId('search-bar')).toBeInTheDocument();
    expect(screen.getByTestId('records-list')).toBeInTheDocument();
    expect(screen.getByTestId('overdue-alert')).toBeInTheDocument();
  });

  it('displays overdue alert with correct data', () => {
    render(<Home />);

    expect(screen.getByText('1 machine is overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
  });

  it('displays due alert with correct data', () => {
    render(<Home />);

    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();
  });

  it('handles View Overdue button click', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const viewOverdueButton = screen.getByText('View Overdue');
    await user.click(viewOverdueButton);

    // Should update filter type in RecordsList
    expect(screen.getByText('Filter Type: overdue')).toBeInTheDocument();
  });

  it('handles View Due button click', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const viewDueButton = screen.getByText('View Due');
    await user.click(viewDueButton);

    // Should update filter type in RecordsList
    expect(screen.getByText('Filter Type: due')).toBeInTheDocument();
  });

  it('handles dismiss overdue button click', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const dismissOverdueButton = screen.getByText('Dismiss Overdue');
    await user.click(dismissOverdueButton);

    // Should dismiss the overdue alert
    expect(screen.queryByText('1 machine is overdue for PPM maintenance')).not.toBeInTheDocument();
  });

  it('handles dismiss due button click', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const dismissDueButton = screen.getByText('Dismiss Due');
    await user.click(dismissDueButton);

    // Should dismiss the due alert
    expect(screen.queryByText('1 machine is due for PPM maintenance today')).not.toBeInTheDocument();
  });

  it('handles search functionality', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const searchInput = screen.getByTestId('search-input');
    await user.type(searchInput, 'SN-001');

    // Should update search query in RecordsList
    expect(screen.getByText('Search Query: SN-001')).toBeInTheDocument();
  });

  it('passes search options to RecordsList', () => {
    render(<Home />);

    // Should show default search options
    expect(screen.getByText('Search Property: serial_number')).toBeInTheDocument();
    expect(screen.getByText('Filter Type: all')).toBeInTheDocument();
  });

  it('integrates search and filter state management', async () => {
    const user = userEvent.setup();
    render(<Home />);

    // Start with default state
    expect(screen.getByText('Filter Type: all')).toBeInTheDocument();
    expect(screen.getByText('Search Query:')).toBeInTheDocument();

    // Apply overdue filter
    const viewOverdueButton = screen.getByText('View Overdue');
    await user.click(viewOverdueButton);
    expect(screen.getByText('Filter Type: overdue')).toBeInTheDocument();

    // Add search query
    const searchInput = screen.getByTestId('search-input');
    await user.type(searchInput, 'test');
    expect(screen.getByText('Search Query: test')).toBeInTheDocument();

    // Filter should still be applied
    expect(screen.getByText('Filter Type: overdue')).toBeInTheDocument();
  });
});
