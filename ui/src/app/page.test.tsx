import { render, screen } from '@testing-library/react';
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
          placeholder="Search by any..."
          value={searchOptions?.query || ''}
          onChange={(e) => onSearch({ ...searchOptions, query: e.target.value })}
        />
        <div>Any</div>
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
    isDueDismissed = false,
  }: MockOverdueAlertProps) {
    // Provide default stats if undefined
    const defaultStats = {
      overdueCount: 1,
      dueCount: 1,
      totalCriticalCount: 2,
      overdueMachines: [],
      dueMachines: [],
    };
    const safeStats = stats || defaultStats;
    
    if (
      (safeStats.overdueCount === 0 || isOverdueDismissed) &&
      (safeStats.dueCount === 0 || isDueDismissed)
    ) {
      return null;
    }

    return (
      <div data-testid="overdue-alert">
        {safeStats.overdueCount > 0 && !isOverdueDismissed && (
          <div>
            <div>
              {safeStats.overdueCount} machine{safeStats.overdueCount !== 1 ? 's are' : ' is'} overdue for
              PPM maintenance
            </div>
            <button onClick={onShowOverdue}>View Overdue</button>
            {onDismissOverdue && <button onClick={onDismissOverdue}>Dismiss Overdue</button>}
          </div>
        )}
        {safeStats.dueCount > 0 && !isDueDismissed && (
          <div>
            <div>
              {safeStats.dueCount} machine{safeStats.dueCount !== 1 ? 's are' : ' is'} due for PPM
              maintenance today
            </div>
            <button onClick={onShowDue}>View Due</button>
            {onDismissDue && <button onClick={onDismissDue}>Dismiss Due</button>}
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
    almostDueCount: 1,
    totalCriticalCount: 2,
    overdueMachines: [],
    dueMachines: [],
  }),
}));

// Mock the useMachines hook to prevent API calls
jest.mock('../hooks/useMachines', () => ({
  useMachines: () => ({
    machines: [],
    total: 0,
    offset: 0,
    limit: 10,
    totalPages: 0,
    loading: false,
    error: null,
    overdueCount: 1,
    dueCount: 1,
    almostDueCount: 1,
    refetch: jest.fn(),
    setLimit: jest.fn(),
    setFilters: jest.fn(),
    loadMore: jest.fn(),
    reset: jest.fn(),
  }),
}));

describe('Home Page', () => {
  it('renders the main page with all components and basic functionality', () => {
    render(<Home />);

    // Check that all main components are present
    expect(screen.getByTestId('search-bar')).toBeInTheDocument();
    expect(screen.getByTestId('records-list')).toBeInTheDocument();
    expect(screen.getByTestId('overdue-alert')).toBeInTheDocument();

    // Check that overdue and due alerts are displayed
    expect(screen.getByText('1 machine is overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();

    // Check that search functionality is available
    expect(screen.getByTestId('search-input')).toBeInTheDocument();
    expect(screen.getByText('Any')).toBeInTheDocument();

    // Check that RecordsList shows default state
    expect(screen.getByText('Search Property: any')).toBeInTheDocument();
    expect(screen.getByText('Filter Type: all')).toBeInTheDocument();
  });
});
