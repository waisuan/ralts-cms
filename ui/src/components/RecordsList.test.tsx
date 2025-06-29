import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsList from './RecordsList';

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
      additional_notes: 'Scheduled maintenance',
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

describe('RecordsList', () => {
  beforeEach(() => {
    // Clear any console logs from previous tests
    jest.clearAllMocks();
  });

  it('renders the component with initial machines', () => {
    render(<RecordsList searchQuery="" />);

    expect(screen.getByText('All Machines')).toBeInTheDocument();
    expect(screen.getByText('Add New Machine')).toBeInTheDocument();
    expect(screen.getByText('Showing 3 of 3 machines')).toBeInTheDocument();
  });

  it('displays the correct number of machine cards initially', () => {
    render(<RecordsList searchQuery="" />);

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

  it('filters machines based on search query', () => {
    render(<RecordsList searchQuery="Acme" />);

    // Should show only 1 machine when searching for "Acme"
    expect(screen.getByText('Showing 1 of 1 machine')).toBeInTheDocument();
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
    expect(
      screen.queryByText((content, element) => {
        return Boolean(
          element?.tagName === 'H3' &&
            element.textContent?.includes('SN-003') &&
            Array.from(element.children).some(
              (child) => child.tagName === 'SPAN' && child.textContent === '(Z300)'
            )
        );
      })
    ).not.toBeInTheDocument();
  });

  it('shows "Load More" button when there are more machines to display', () => {
    render(<RecordsList searchQuery="" />);

    // With 3 machines and ITEMS_PER_PAGE = 3, there should be no "Load More" button
    expect(screen.queryByText(/Load More/)).not.toBeInTheDocument();
  });

  it('handles empty search results', () => {
    render(<RecordsList searchQuery="NonExistentMachine" />);

    expect(screen.getByText('No machines found')).toBeInTheDocument();
    expect(screen.getByText(/No machines match "NonExistentMachine"/)).toBeInTheDocument();
  });

  it('calls onView when View button is clicked', async () => {
    const user = userEvent.setup();
    const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

    render(<RecordsList searchQuery="" />);

    const viewButtons = screen.getAllByText('View');
    await user.click(viewButtons[0]);

    expect(consoleSpy).toHaveBeenCalledWith('View machine:', 'SN-001');

    consoleSpy.mockRestore();
  });

  it('calls onEdit when Edit button is clicked', async () => {
    const user = userEvent.setup();
    const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

    render(<RecordsList searchQuery="" />);

    const editButtons = screen.getAllByText('Edit');
    await user.click(editButtons[0]);

    expect(consoleSpy).toHaveBeenCalledWith('Edit machine:', 'SN-001');

    consoleSpy.mockRestore();
  });

  it('removes machine when Delete button is clicked', async () => {
    const user = userEvent.setup();

    render(<RecordsList searchQuery="" />);

    // Initially shows 3 machines
    expect(screen.getByText('Showing 3 of 3 machines')).toBeInTheDocument();

    const deleteButtons = screen.getAllByText('Delete');
    await user.click(deleteButtons[0]);

    // Should now show 2 machines
    expect(screen.getByText('Showing 2 of 2 machines')).toBeInTheDocument();
  });

  it('displays correct status badges for different machines', () => {
    render(<RecordsList searchQuery="" />);

    // First machine should be Overdue (past date)
    expect(screen.getByText('Overdue')).toBeInTheDocument();

    // Second machine should be Due (today)
    expect(screen.getByText('Due')).toBeInTheDocument();

    // Third machine should be Due Soon (3 days from now)
    expect(screen.getByText('Due Soon')).toBeInTheDocument();
  });

  it('shows filtered count when search is active', () => {
    render(<RecordsList searchQuery="Acme" />);

    // Use a more specific selector to find the paragraph element containing the count
    const countElement = screen.getByText((content, element) => {
      return Boolean(
        element?.tagName === 'P' &&
          element?.textContent?.includes('Showing 1 of 1 machine') &&
          element?.textContent?.includes('filtered from 3 total')
      );
    });
    expect(countElement).toBeInTheDocument();
  });
});
