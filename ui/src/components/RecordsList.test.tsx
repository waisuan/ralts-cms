import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsList from './RecordsList';
import { SearchOptions } from './SearchBar';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

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

// Helper function to find machine card by serial number
const findMachineCard = (serialNumber: string) => {
  return screen
    .getByText((content, element) => {
      return Boolean(element?.tagName === 'H3' && element.textContent?.includes(serialNumber));
    })
    .closest('div[class*="bg-white"]');
};

describe('RecordsList', () => {
  beforeEach(() => {
    // Clear any console logs from previous tests
    jest.clearAllMocks();

    // Mock the current date for consistent testing
    jest.useFakeTimers();
    jest.setSystemTime(new Date(MOCK_CURRENT_DATE));
  });

  afterEach(() => {
    jest.useRealTimers();
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

    // Should show overdue and due badges based on mock data
    // SN-001 has ppm_date: '2024-06-24' (5 days ago) -> Overdue
    // SN-002 has ppm_date: '2024-06-29' (today) -> Due
    expect(findTextAcrossElements('1 Overdue')).toBeInTheDocument();
    expect(findTextAcrossElements('1 Due Today')).toBeInTheDocument();
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
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Find the machine card for SN-003 (first machine in sorted order)
    const machineCard = findMachineCard('SN-003');
    expect(machineCard).toBeInTheDocument();

    // Find the View button within this specific machine card
    const viewButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'View'
    );

    expect(viewButton).toBeInTheDocument();
    await user.click(viewButton!);

    expect(consoleSpy).toHaveBeenCalledWith('View machine:', 'SN-003');

    consoleSpy.mockRestore();
  });

  it('opens edit modal when Edit button is clicked', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Find the machine card for SN-003 (first machine in sorted order)
    const machineCard = findMachineCard('SN-003');
    expect(machineCard).toBeInTheDocument();

    // Find the Edit button within this specific machine card
    const editButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Edit'
    );

    expect(editButton).toBeInTheDocument();
    await user.click(editButton!);

    // Should show the edit modal
    expect(screen.getByRole('heading', { name: 'Edit Machine' })).toBeInTheDocument();

    // The form should be pre-populated with the machine data
    expect(screen.getByDisplayValue('SN-003')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Gamma Inc')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Z300')).toBeInTheDocument();
  });

  it('updates machine data when edit form is submitted', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Find the machine card for SN-003 and click edit
    const machineCard = findMachineCard('SN-003');
    const editButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Edit'
    );

    await user.click(editButton!);

    // Should show the edit modal
    expect(screen.getByRole('heading', { name: 'Edit Machine' })).toBeInTheDocument();

    // Update the customer field
    const customerInput = screen.getByDisplayValue('Gamma Inc');
    await user.clear(customerInput);
    await user.type(customerInput, 'Updated Customer Name');

    // Submit the form
    const updateButton = screen.getByRole('button', { name: /update machine/i });
    await user.click(updateButton);

    // Modal should be closed
    expect(screen.queryByRole('heading', { name: 'Edit Machine' })).not.toBeInTheDocument();

    // Machine card should show updated data
    expect(screen.getByText('Updated Customer Name')).toBeInTheDocument();
    expect(screen.queryByText('Gamma Inc')).not.toBeInTheDocument();
  });

  it('shows delete confirmation modal when Delete button is clicked', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Initially shows 3 machines
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();

    // Find the machine card for SN-003 (first machine in sorted order)
    const machineCard = findMachineCard('SN-003');
    expect(machineCard).toBeInTheDocument();

    // Find the Delete button within this specific machine card
    const deleteButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Delete'
    );

    expect(deleteButton).toBeInTheDocument();
    await user.click(deleteButton!);

    // Should show confirmation modal
    expect(screen.getByRole('heading', { name: 'Delete Machine' })).toBeInTheDocument();
    expect(screen.getByText(/Are you sure you want to delete machine/)).toBeInTheDocument();
    expect(screen.getByText(/This action cannot be undone/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /delete machine/i })).toBeInTheDocument();

    // Check that the modal contains the correct machine serial number
    const modalContent = screen.getByRole('heading', { name: 'Delete Machine' }).closest('div');
    expect(modalContent).toHaveTextContent('SN-003');

    // Machine count should not change until confirmed
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();
  });

  it('cancels delete when Cancel button is clicked in confirmation modal', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Find the machine card for SN-003 and click delete
    const machineCard = findMachineCard('SN-003');
    const deleteButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Delete'
    );

    await user.click(deleteButton!);

    // Modal should be visible
    expect(screen.getByRole('heading', { name: 'Delete Machine' })).toBeInTheDocument();

    // Click cancel
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);

    // Modal should be gone
    expect(screen.queryByRole('heading', { name: 'Delete Machine' })).not.toBeInTheDocument();

    // Machine count should remain the same
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();
  });

  it('removes machine when delete is confirmed', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Initially shows 3 machines
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();

    // Find the machine card for SN-003 and click delete
    const machineCard = findMachineCard('SN-003');
    const deleteButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Delete'
    );

    await user.click(deleteButton!);

    // Should show confirmation modal
    expect(screen.getByRole('heading', { name: 'Delete Machine' })).toBeInTheDocument();

    // Click confirm delete
    const confirmButton = screen.getByRole('button', { name: /delete machine/i });
    await user.click(confirmButton);

    // Modal should be gone
    expect(screen.queryByRole('heading', { name: 'Delete Machine' })).not.toBeInTheDocument();

    // Should now show 2 machines
    expect(findTextAcrossElements('Showing 2 of 2 machines')).toBeInTheDocument();
  });

  it('closes delete confirmation modal when clicking backdrop', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    render(<RecordsList searchOptions={defaultSearchOptions} />);

    // Find the machine card for SN-003 and click delete
    const machineCard = findMachineCard('SN-003');
    const deleteButton = Array.from(machineCard?.querySelectorAll('button') || []).find(
      (btn) => btn.textContent === 'Delete'
    );

    await user.click(deleteButton!);

    // Modal should be visible
    expect(screen.getByRole('heading', { name: 'Delete Machine' })).toBeInTheDocument();

    // Click backdrop
    const backdrop = document.querySelector('.fixed.inset-0.bg-black');
    if (backdrop) {
      await user.click(backdrop);
    }

    // Modal should be gone
    expect(screen.queryByRole('heading', { name: 'Delete Machine' })).not.toBeInTheDocument();

    // Machine count should remain the same
    expect(findTextAcrossElements('Showing 3 of 3 machines')).toBeInTheDocument();
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

    // Should show only overdue machines (based on our mock data, only SN-001 should be overdue)
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

    // Should not show non-overdue machines
    expect(
      screen.queryByText((content, element) => {
        return Boolean(element?.tagName === 'H3' && element.textContent?.includes('SN-002'));
      })
    ).not.toBeInTheDocument();
  });

  it('handles due filter type', () => {
    render(<RecordsList searchOptions={defaultSearchOptions} filterType="due" />);

    // Should show only due machines
    expect(findTextAcrossElements('(due today only)')).toBeInTheDocument();

    // Should show the due machine (SN-002 with today's date)
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

    // Should not show non-due machines
    expect(
      screen.queryByText((content, element) => {
        return Boolean(element?.tagName === 'H3' && element.textContent?.includes('SN-001'));
      })
    ).not.toBeInTheDocument();
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
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
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
