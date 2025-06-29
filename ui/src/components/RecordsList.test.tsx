import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsList from './RecordsList';
import { SearchOptions } from './SearchBar';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

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
      ppm_date: '2024-01-01', // Past date for Overdue status
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
      ppm_date: new Date().toISOString().split('T')[0], // Today for Due status
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
      ppm_date: (() => {
        const soon = new Date();
        soon.setDate(soon.getDate() + 3);
        return soon.toISOString().split('T')[0];
      })(), // 3 days from now for Due Soon status
      created_at: '2024-03-01',
      updated_at: '2024-06-15',
    },
  ],
}));

const defaultSearchOptions: SearchOptions = {
  query: '',
  property: DEFAULT_SEARCH_PROPERTY,
};

// Helper function to find text that might be broken up by elements
const findTextAcrossElements = (text: string) => {
  return screen.getAllByText((content, element) => {
    return Boolean(element?.textContent?.includes(text));
  })[0]; // Get the first match to avoid multiple element errors
};

describe('RecordsList', () => {
  beforeEach(() => {
    // Clear any console logs from previous tests
    jest.clearAllMocks();
  });

  it('renders the component with initial machines', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} />);

    expect(screen.getByText('All Machines')).toBeInTheDocument();
    expect(screen.getByText('Add New Machine')).toBeInTheDocument();
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();
  });

  it('displays the correct number of machine cards initially', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Should show 3 machines initially (ITEMS_PER_PAGE = 3)
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-001') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(X100)'
            )
        );
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-002') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Y200)'
            )
        );
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-003') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Z300)'
            )
        );
      })
    ).toBeInTheDocument();
  });

  it('displays overdue and due statistics badges', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Should show overdue and due badges
    expect(screen.getByText('1 Overdue')).toBeInTheDocument();
    expect(screen.getByText('1 Due Today')).toBeInTheDocument();
  });

  it('filters machines based on search query', () => {
    const searchOptions: SearchOptions = {
      query: 'Acme',
      property: 'customer',
    };
    render(<RecordsList searchOptions={searchOptions} />);

    // Should show only 1 machine when searching for "Acme" in customer field
    expect(findTextAcrossElements('Showing 1 of 1 machine')).toBeInTheDocument();
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-001') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(X100)'
            )
        );
      })
    ).toBeInTheDocument();
    expect(
      screen.queryByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-002') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Y200)'
            )
        );
      })
    ).not.toBeInTheDocument();
  });

  it('filters machines by specific property', () => {
    const searchOptions: SearchOptions = {
      query: 'SN-002',
      property: 'serial_number',
    };
    render(<RecordsList searchOptions={searchOptions} />);

    // Should show only the machine with serial number SN-002
    expect(findTextAcrossElements('Showing 1 of 1 machine')).toBeInTheDocument();
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-002') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Y200)'
            )
        );
      })
    ).toBeInTheDocument();
  });

  it('shows "Load More" button when there are more machines to display', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // With 3 machines and ITEMS_PER_PAGE = 3, there should be no "Load More" button
    expect(screen.queryByText(/Load More/)).not.toBeInTheDocument();
  });

  it('handles empty search results', () => {
    const searchOptions: SearchOptions = {
      query: 'NonExistentMachine',
      property: 'serial_number',
    };
    render(<RecordsList searchOptions={searchOptions} />);

    expect(screen.getByText('No machines found')).toBeInTheDocument();
    expect(screen.getByText(/No machines match "NonExistentMachine"/)).toBeInTheDocument();
  });

  it('calls onView when View button is clicked', async () => {
    const user = userEvent.setup();
    const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    const viewButtons = screen.getAllByText('View');
    await user.click(viewButtons[0]);

    expect(consoleSpy).toHaveBeenCalledWith('View machine:', 'SN-001');

    consoleSpy.mockRestore();
  });

  it('calls onEdit when Edit button is clicked', async () => {
    const user = userEvent.setup();
    const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    const editButtons = screen.getAllByText('Edit');
    await user.click(editButtons[0]);

    expect(consoleSpy).toHaveBeenCalledWith('Edit machine:', 'SN-001');

    consoleSpy.mockRestore();
  });

  it('removes machine when Delete button is clicked', async () => {
    const user = userEvent.setup();

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Initially shows 3 machines
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();

    const deleteButtons = screen.getAllByText('Delete');
    await user.click(deleteButtons[0]);

    // Should now show 2 machines
    expect(findTextAcrossElements('Showing 2 of 2 machines')).toBeInTheDocument();
  });

  it('displays correct status badges for different machines', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // First machine should be Overdue (past date)
    expect(screen.getByText('Overdue')).toBeInTheDocument();

    // Second machine should be Due (today)
    expect(screen.getByText('Due')).toBeInTheDocument();

    // Third machine should be Due Soon (3 days from now)
    expect(screen.getByText('Due Soon')).toBeInTheDocument();
  });

  it('shows filtered count when search is active', () => {
    const searchOptions: SearchOptions = {
      query: 'Acme',
      property: 'customer',
    };
    render(<RecordsList searchOptions={searchOptions} />);

    // Should show filtered text
    expect(findTextAcrossElements('(filtered from 3 total)')).toBeInTheDocument();
  });

  it('filters machines by PPM status', () => {
    const searchOptions: SearchOptions = {
      query: 'Overdue',
      property: 'ppm_status',
    };
    render(<RecordsList searchOptions={searchOptions} />);

    // Should show only machines with Overdue status (first machine)
    expect(findTextAcrossElements('Showing 1 of 1 machine')).toBeInTheDocument();
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-001') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(X100)'
            )
        );
      })
    ).toBeInTheDocument();

    // Should show the Overdue badge
    expect(screen.getByText('Overdue')).toBeInTheDocument();
  });

  it('handles overdue filter type', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} filterType="overdue" />);

    // Should show only overdue machines
    expect(findTextAcrossElements('Showing 1 of 1 machine')).toBeInTheDocument();
    expect(findTextAcrossElements('(overdue only)')).toBeInTheDocument();

    // Should show the overdue machine
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-001') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(X100)'
            )
        );
      })
    ).toBeInTheDocument();
  });

  it('handles due filter type', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} filterType="due" />);

    // Should show only due machines
    expect(findTextAcrossElements('Showing 1 of 1 machine')).toBeInTheDocument();
    expect(findTextAcrossElements('(due today only)')).toBeInTheDocument();

    // Should show the due machine
    expect(
      screen.getByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-002') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Y200)'
            )
        );
      })
    ).toBeInTheDocument();
  });

  it('displays sort dropdown when onSortChange is provided', () => {
    const mockSortChange = jest.fn();
    render(
      <RecordsList 
        searchOptions={defaultSearchOptions} 
        sortBy="newest"
        onSortChange={mockSortChange}
      />
    );

    expect(screen.getByText('Sort By:')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Newest First')).toBeInTheDocument();
  });

  it('calls onSortChange when sort selection changes', async () => {
    const user = userEvent.setup();
    const mockSortChange = jest.fn();
    
    render(
      <RecordsList 
        searchOptions={defaultSearchOptions} 
        sortBy="newest"
        onSortChange={mockSortChange}
      />
    );

    const sortDropdown = screen.getByDisplayValue('Newest First');
    await user.selectOptions(sortDropdown, 'oldest');

    expect(mockSortChange).toHaveBeenCalledWith('oldest');
  });

  it('sorts machines by newest first by default', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} sortBy="newest" />);

    const machineCards = screen.getAllByText((content, element) => {
      return Boolean(element?.tagName === 'H3' && element.textContent?.includes('SN-'));
    });

    // With newest first, SN-003 (2024-03-01) should come first, then SN-002 (2024-02-01), then SN-001 (2024-01-01)
    expect(machineCards[0]).toHaveTextContent('SN-003');
    expect(machineCards[1]).toHaveTextContent('SN-002'); 
    expect(machineCards[2]).toHaveTextContent('SN-001');
  });

  it('sorts machines by oldest first when selected', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} sortBy="oldest" />);

    const machineCards = screen.getAllByText((content, element) => {
      return Boolean(element?.tagName === 'H3' && element.textContent?.includes('SN-'));
    });

    // With oldest first, SN-001 (2024-01-01) should come first, then SN-002 (2024-02-01), then SN-003 (2024-03-01)
    expect(machineCards[0]).toHaveTextContent('SN-001');
    expect(machineCards[1]).toHaveTextContent('SN-002');
    expect(machineCards[2]).toHaveTextContent('SN-003');
  });
});
