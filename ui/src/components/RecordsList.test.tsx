import { render, screen } from '@testing-library/react';
import RecordsList from './RecordsList';
import { SearchOptions } from './SearchBar';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

// Mock next/navigation
const mockPush = jest.fn();
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => ({
    push: mockPush,
  })),
}));

// Mock the current date for consistent testing
const MOCK_CURRENT_DATE = '2024-06-29T12:00:00.000Z';

// Mock the mockMachines to have predictable data for testing
jest.mock('../data/mockMachines', () => ({
  mockMachines: [
    {
      serial_number: 'SN-001',
      customer: 'Acme Corp',
      state: 'CA',
      account_type: 'Premium',
      model: 'X100',
      status: '',
      brand: 'BrandA',
      district: 'North',
      person_in_charge: 'Alice Johnson',
      reported_by: 'Bob Smith',
      additional_notes: 'Needs inspection',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-07-01',
      ppm_date: '2024-06-24', // 5 days ago from mock date for Overdue status
      created_at: '2024-01-01',
      updated_at: '2024-06-01',
    },
    {
      serial_number: 'SN-002',
      customer: 'Beta LLC',
      state: 'NY',
      account_type: 'Standard',
      model: 'Y200',
      status: '',
      brand: 'BrandB',
      district: 'East',
      person_in_charge: 'Charlie Brown',
      reported_by: 'Dana White',
      additional_notes: 'Regular maintenance',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-07-10',
      ppm_date: '2024-06-29', // Today for Due status
      created_at: '2024-02-01',
      updated_at: '2024-06-10',
    },
    {
      serial_number: 'SN-003',
      customer: 'Gamma Inc',
      state: 'TX',
      account_type: 'Basic',
      model: 'Z300',
      status: '',
      brand: 'BrandC',
      district: 'South',
      person_in_charge: 'Eve Wilson',
      reported_by: 'Frank Miller',
      additional_notes: 'Due soon maintenance',
      attachment: '',
      ppm_status: '',
      tnc_date: '2024-06-01',
      ppm_date: '2024-07-02', // 3 days from mock date for Due Soon status
      created_at: '2024-03-01',
      updated_at: '2024-06-15',
    },
  ],
}));

// Mock the useMachines hook
jest.mock('../hooks/useMachines', () => ({
  useMachines: () => ({
    machines: [
      {
        serial_number: 'SN-001',
        customer: 'Acme Corp',
        state: 'CA',
        account_type: 'Premium',
        model: 'X100',
        status: '',
        brand: 'BrandA',
        district: 'North',
        person_in_charge: 'Alice Johnson',
        reported_by: 'Bob Smith',
        additional_notes: 'Needs inspection',
        attachment: '',
        ppm_status: '',
        tnc_date: '2024-07-01',
        ppm_date: '2024-06-24',
        created_at: '2024-01-01',
        updated_at: '2024-06-01',
        maintenance_count: 5,
      },
      {
        serial_number: 'SN-002',
        customer: 'Beta LLC',
        state: 'NY',
        account_type: 'Standard',
        model: 'Y200',
        status: '',
        brand: 'BrandB',
        district: 'East',
        person_in_charge: 'Charlie Brown',
        reported_by: 'Dana White',
        additional_notes: 'Regular maintenance',
        attachment: '',
        ppm_status: '',
        tnc_date: '2024-07-10',
        ppm_date: '2024-06-29',
        created_at: '2024-02-01',
        updated_at: '2024-06-10',
        maintenance_count: 3,
      },
      {
        serial_number: 'SN-003',
        customer: 'Gamma Inc',
        state: 'TX',
        account_type: 'Basic',
        model: 'Z300',
        status: '',
        brand: 'BrandC',
        district: 'South',
        person_in_charge: 'Eve Wilson',
        reported_by: 'Frank Miller',
        additional_notes: 'Due soon maintenance',
        attachment: '',
        ppm_status: '',
        tnc_date: '2024-06-01',
        ppm_date: '2024-07-02',
        created_at: '2024-03-01',
        updated_at: '2024-06-15',
        maintenance_count: 1,
      },
    ],
    total: 3,
    offset: 0,
    limit: 50,
    totalPages: 1,
    loading: false,
    error: null,
    overdueCount: 1,
    dueCount: 1,
    almostDueCount: 1,
    refetch: jest.fn(),
    reset: jest.fn(),
  }),
}));

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

const defaultSearchOptions: SearchOptions = {
  query: '',
  property: DEFAULT_SEARCH_PROPERTY,
};

// Helper function to find text that might be broken up by elements
const findTextAcrossElements = (text: string) => {
  try {
    return screen.getByText((content, element) => {
      return Boolean(element?.textContent?.includes(text));
    });
  } catch (error) {
    // If getByText fails, try getAllByText and return the first match
    const elements = screen.queryAllByText((content, element) => {
      return Boolean(element?.textContent?.includes(text));
    });
    if (elements.length > 0) {
      return elements[0];
    }
    throw error;
  }
};

describe('RecordsList', () => {
  beforeEach(() => {
    // Clear any console logs from previous tests
    jest.clearAllMocks();
    mockPush.mockClear();

    // Mock the current date for consistent testing
    jest.useFakeTimers();
    jest.setSystemTime(new Date(MOCK_CURRENT_DATE).getTime());
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const defaultDateRange = { from: undefined, to: undefined };

  it('renders the component with machines and basic functionality', () => {
    render(
      <RecordsList
        searchOptions={defaultSearchOptions}
        page={1}
        limit={50}
        ppmDateRange={defaultDateRange}
        tncDateRange={defaultDateRange}
        onPageChange={jest.fn()}
        onLimitChange={jest.fn()}
        onPpmDateRangeChange={jest.fn()}
        onTncDateRangeChange={jest.fn()}
        onClearDateRanges={jest.fn()}
      />
    );

    // Check that the component renders with basic elements
    // Title is "Machines" (was "All Machines"); desktop header also shows "Add New Machine" text
    expect(screen.getAllByText('Machines').length).toBeGreaterThan(0);
    // Mobile header uses icon-only add button; desktop header has text label
    expect(screen.getAllByLabelText('Add New Machine').length).toBeGreaterThan(0);

    // Check that machine cards are displayed
    expect(screen.getByText('SN-001')).toBeInTheDocument();
    expect(screen.getByText('SN-002')).toBeInTheDocument();
    expect(screen.getByText('SN-003')).toBeInTheDocument();

    // Cards start collapsed — View/Edit/Delete are hidden until expanded
    // Verify the status badges are displayed in the header
    expect(findTextAcrossElements('Overdue')).toBeInTheDocument();
    expect(findTextAcrossElements('due')).toBeInTheDocument();
  });
});
