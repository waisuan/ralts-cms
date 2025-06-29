import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Home from './page';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

// Mock the RecordsList component to avoid rendering complexity in page tests
jest.mock('@/components/RecordsList', () => {
  return function MockRecordsList({
    searchOptions,
  }: {
    searchOptions: { query: string; property: string };
  }) {
    return (
      <div data-testid="records-list">
        <div>Mock Records List</div>
        <div>Search Query: {searchOptions.query}</div>
        <div>Search Property: {searchOptions.property}</div>
      </div>
    );
  };
});

describe('Home Page', () => {
  it('renders the main page with title and search bar', () => {
    render(<Home />);

    expect(screen.getByText('Ralts CMS')).toBeInTheDocument();
    expect(screen.getByText('Content Management System')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search by serial number...')).toBeInTheDocument();
  });

  it('has a search input with proper attributes', () => {
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search by serial number...');
    expect(searchInput).toHaveAttribute('type', 'text');
    expect(searchInput).toBeInTheDocument();
  });

  it('has initial search state', () => {
    render(<Home />);

    expect(screen.getByText('Search Query:')).toBeInTheDocument();
    expect(screen.getByText(`Search Property: ${DEFAULT_SEARCH_PROPERTY}`)).toBeInTheDocument();
  });

  it('renders RecordsList component', () => {
    render(<Home />);

    expect(screen.getByTestId('records-list')).toBeInTheDocument();
    expect(screen.getByText('Mock Records List')).toBeInTheDocument();
  });

  it('updates search query when user types in search bar', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search by serial number...');
    await user.type(searchInput, 'test search');

    expect(searchInput).toHaveValue('test search');
    // Should show the search query in the mocked component
    expect(screen.getByText('Search Query: test search')).toBeInTheDocument();
  });

  it('clears search query when user clears the input', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search by serial number...');
    await user.type(searchInput, 'test search');
    await user.clear(searchInput);

    expect(searchInput).toHaveValue('');
    expect(screen.getByText('Search Query:')).toBeInTheDocument();
  });

  it('has dropdown functionality for search properties', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    // Check for options that are not the currently selected one (Serial Number)
    expect(screen.getByText('Customer')).toBeInTheDocument();
    expect(screen.getByText('Model')).toBeInTheDocument();
    expect(screen.getByText('Brand')).toBeInTheDocument();

    // Verify Serial Number appears twice (button + dropdown)
    expect(screen.getAllByText('Serial Number')).toHaveLength(2);
  });

  it('updates search property when dropdown option is selected', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const dropdownButton = screen.getByRole('button');
    await user.click(dropdownButton);

    const customerOption = screen.getByText('Customer');
    await user.click(customerOption);

    expect(screen.getByText('Search Property: customer')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search by customer...')).toBeInTheDocument();
  });

  it('has proper page structure with main container', () => {
    render(<Home />);

    const mainElement = screen.getByRole('main');
    expect(mainElement).toBeInTheDocument();
    expect(mainElement).toHaveClass('min-h-screen', 'bg-gray-50');
  });

  it('has responsive container with proper padding', () => {
    render(<Home />);

    // Find the container div that has the proper classes
    const container = screen.getByText('Ralts CMS').closest('div')?.parentElement;
    expect(container).toHaveClass('container', 'mx-auto', 'px-4');
  });
});
