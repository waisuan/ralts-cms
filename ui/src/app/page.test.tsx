import { act, fireEvent, render, screen } from '@testing-library/react';
import { withNuqsTestingAdapter, type UrlUpdateEvent } from 'nuqs/adapters/testing';
import HomeContent from './HomeContent';

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
    reset: jest.fn(),
  }),
}));

describe('Home Page', () => {
  it('renders the main page with all components and basic functionality', () => {
    render(<HomeContent />, { wrapper: withNuqsTestingAdapter({ searchParams: '' }) });

    expect(screen.getByTestId('search-bar')).toBeInTheDocument();
    expect(screen.getByTestId('records-list')).toBeInTheDocument();
    expect(screen.getByTestId('overdue-alert')).toBeInTheDocument();

    expect(screen.getByText('1 machine is overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();

    expect(screen.getByTestId('search-input')).toBeInTheDocument();
    expect(screen.getByText('Any')).toBeInTheDocument();

    expect(screen.getByText('Search Property: any')).toBeInTheDocument();
    expect(screen.getByText('Filter Type: all')).toBeInTheDocument();
  });

  it('derives initial state from URL search params', () => {
    render(<HomeContent />, {
      wrapper: withNuqsTestingAdapter({
        searchParams: '?q=klinik&search_by=any&filter=overdue&sort=ppm_date_asc',
      }),
    });

    expect(screen.getByText('Search Query: klinik')).toBeInTheDocument();
    expect(screen.getByText('Search Property: any')).toBeInTheDocument();
    expect(screen.getByText('Filter Type: overdue')).toBeInTheDocument();
    expect(screen.getByText('Sort By: ppm_date_asc')).toBeInTheDocument();
  });

  it('writes the expected params when the user clicks "View Overdue"', async () => {
    const onUrlUpdate = jest.fn<void, [UrlUpdateEvent]>();

    render(<HomeContent />, {
      wrapper: withNuqsTestingAdapter({ searchParams: '', onUrlUpdate }),
    });

    await act(async () => {
      fireEvent.click(screen.getByText('View Overdue'));
      await new Promise((resolve) => setTimeout(resolve, 0));
    });

    expect(onUrlUpdate).toHaveBeenCalled();
    const last = onUrlUpdate.mock.calls.at(-1)?.[0];
    expect(last).toBeDefined();
    expect(last?.queryString).toContain('q=overdue');
    expect(last?.queryString).toContain('search_by=ppm_status');
    expect(last?.queryString).toContain('filter=overdue');
  });

  it('falls back to defaults when URL has invalid values', () => {
    render(<HomeContent />, {
      wrapper: withNuqsTestingAdapter({
        searchParams: '?filter=bogus&sort=nope&limit=999',
      }),
    });

    expect(screen.getByText('Filter Type: all')).toBeInTheDocument();
    expect(screen.getByText('Sort By: newest')).toBeInTheDocument();
  });
});
