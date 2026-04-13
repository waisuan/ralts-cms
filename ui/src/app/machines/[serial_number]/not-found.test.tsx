import { render, screen } from '@testing-library/react';
import NotFound from './not-found';

// Mock next/link
jest.mock('next/link', () => {
  return function MockLink({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    [key: string]: unknown;
  }) {
    return (
      <a href={href} {...props}>
        {children}
      </a>
    );
  };
});

describe('NotFound', () => {
  it('renders the 404 page with correct content', () => {
    render(<NotFound />);

    expect(screen.getByText('Machine Not Found')).toBeInTheDocument();
    expect(
      screen.getByText('The machine with this serial number could not be found.')
    ).toBeInTheDocument();
  });

  it('displays the maintenance wrench emoji', () => {
    render(<NotFound />);

    expect(screen.getByText('🔧')).toBeInTheDocument();
  });

  it('has a link to the machines list', () => {
    render(<NotFound />);

    const homeLink = screen.getByText('Go to Machines');
    expect(homeLink).toBeInTheDocument();
    expect(homeLink.closest('a')).toHaveAttribute('href', '/');
  });

  it('has proper button styling on the home link', () => {
    render(<NotFound />);

    const homeLink = screen.getByText('Go to Machines');
    expect(homeLink).toHaveClass('bg-blue-600', 'hover:bg-blue-700', 'text-white');
  });

  it('has centered layout', () => {
    const { container } = render(<NotFound />);

    const mainDiv = container.firstChild;
    expect(mainDiv).toHaveClass(
      'min-h-screen',
      'bg-gray-50',
      'flex',
      'items-center',
      'justify-center'
    );
  });
});
