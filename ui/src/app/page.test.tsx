import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Home from './page';

// Mock the RecordsList component to avoid complex dependencies
jest.mock('../components/RecordsList', () => {
  return function MockRecordsList({ searchQuery }: { searchQuery: string }) {
    return (
      <div data-testid="records-list">
        <div>Mock Records List</div>
        <div>Search Query: {searchQuery}</div>
      </div>
    );
  };
});

describe('Home Page', () => {
  it('renders the main page with title and search bar', () => {
    render(<Home />);

    expect(screen.getByText('Ralts CMS')).toBeInTheDocument();
    expect(screen.getByText('Content Management System')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search records...')).toBeInTheDocument();
  });

  it('has a search input with proper attributes', () => {
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search records...');
    expect(searchInput).toHaveAttribute('type', 'text');
    expect(searchInput).toBeInTheDocument();
  });

  it('has a search icon in the search bar', () => {
    render(<Home />);

    // The search icon should be present (it's an SVG)
    const searchIcon = document.querySelector('svg');
    expect(searchIcon).toBeInTheDocument();
  });

  it('renders the RecordsList component', () => {
    render(<Home />);

    expect(screen.getByTestId('records-list')).toBeInTheDocument();
    expect(screen.getByText('Mock Records List')).toBeInTheDocument();
  });

  it('passes empty search query to RecordsList initially', () => {
    render(<Home />);

    expect(screen.getByText('Search Query:')).toBeInTheDocument();
  });

  it('updates search query when user types in search bar', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search records...');
    await user.type(searchInput, 'test search');

    expect(searchInput).toHaveValue('test search');
    expect(screen.getByText('Search Query: test search')).toBeInTheDocument();
  });

  it('clears search query when user clears the input', async () => {
    const user = userEvent.setup();
    render(<Home />);

    const searchInput = screen.getByPlaceholderText('Search records...');
    await user.type(searchInput, 'test search');
    await user.clear(searchInput);

    expect(searchInput).toHaveValue('');
    expect(screen.getByText('Search Query:')).toBeInTheDocument();
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
