import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import RegisterPage from './RegisterPage';
import { UserService } from '@/services/userService';

// Mock Next.js Link
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode; href: string }) {
    return <a href={href}>{children}</a>;
  };
});

// Mock the UserService
jest.mock('@/services/userService', () => ({
  UserService: {
    registerUser: jest.fn(),
  },
}));

// Mock the API client
jest.mock('@/utils/api', () => ({
  apiClient: {
    post: jest.fn(),
  },
  handleApiError: jest.fn(),
}));

describe('RegisterPage', () => {
  const mockRegisterUser = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock the UserService.registerUser method
    UserService.registerUser = mockRegisterUser;
  });

  it('should render the registration form with all fields', () => {
    render(<RegisterPage />);

    // Check main title and description
    expect(screen.getByText('Ralts CMS')).toBeInTheDocument();
    expect(screen.getByText('Content Management System')).toBeInTheDocument();
    expect(screen.getByText('Create your account')).toBeInTheDocument();
    expect(screen.getByText('Join the system to manage your content')).toBeInTheDocument();

    // Check form fields
    expect(screen.getByLabelText('Full Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Email address')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm Password')).toBeInTheDocument();

    // Check submit button
    expect(screen.getByRole('button', { name: 'Create Account' })).toBeInTheDocument();

    // Check link to login
    expect(screen.getByText('Sign in instead')).toBeInTheDocument();
  });

  it('should successfully register a new user', async () => {
    // Mock successful registration
    mockRegisterUser.mockResolvedValueOnce({
      data: {
        id: 1,
        name: 'John Doe',
        email: 'john@example.com',
        role: 'user',
        status: 'active',
        avatar: 'https://ui-avatars.com/api/?name=JD&background=random&color=fff&size=150',
        created_at: '2024-01-01T00:00:00.000Z',
        updated_at: '2024-01-01T00:00:00.000Z',
      },
      message: 'User registered successfully',
    });

    render(<RegisterPage />);

    // Fill out the form
    fireEvent.change(screen.getByLabelText('Full Name'), {
      target: { value: 'John Doe' },
    });
    fireEvent.change(screen.getByLabelText('Email address'), {
      target: { value: 'john@example.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'password123' },
    });
    fireEvent.change(screen.getByLabelText('Confirm Password'), {
      target: { value: 'password123' },
    });

    // Submit the form
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Create Account' }));
    });

    // Wait for the API call
    await waitFor(() => {
      expect(mockRegisterUser).toHaveBeenCalledWith({
        name: 'John Doe',
        email: 'john@example.com',
        password: 'password123',
      });
    });

    // Check success message
    await waitFor(() => {
      expect(screen.getByText('Registration Successful!')).toBeInTheDocument();
      expect(
        screen.getByText('Your account has been created successfully. You can now sign in.')
      ).toBeInTheDocument();
      expect(screen.getByText('Go to Sign In')).toBeInTheDocument();
    });
  });

  it('should show loading state during registration', async () => {
    // Create a promise that doesn't resolve immediately
    let resolveRegister: (value: { data: { id: number; name: string; email: string; role: string; status: string }; message: string }) => void;
    const registerPromise = new Promise<{ data: { id: number; name: string; email: string; role: string; status: string }; message: string }>((resolve) => {
      resolveRegister = resolve;
    });
    mockRegisterUser.mockReturnValueOnce(registerPromise);

    render(<RegisterPage />);

    // Fill out the form
    fireEvent.change(screen.getByLabelText('Full Name'), {
      target: { value: 'John Doe' },
    });
    fireEvent.change(screen.getByLabelText('Email address'), {
      target: { value: 'john@example.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'password123' },
    });
    fireEvent.change(screen.getByLabelText('Confirm Password'), {
      target: { value: 'password123' },
    });

    // Submit the form
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Create Account' }));
    });

    // Check loading state
    await waitFor(() => {
      expect(screen.getByText('Creating account...')).toBeInTheDocument();
      expect(screen.getByRole('button')).toBeDisabled();
    });

    // Resolve the registration
    await act(async () => {
      resolveRegister!({
        data: { id: 1, name: 'John Doe', email: 'john@example.com', role: 'user', status: 'active' },
        message: 'User registered successfully',
      });
    });
  });
});
