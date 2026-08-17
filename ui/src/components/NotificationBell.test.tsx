import { render, screen, waitFor, act } from '@testing-library/react';
import NotificationBell from './NotificationBell';
import { NotificationService } from '../services/notificationService';
import { useAuth } from '../contexts/AuthContext';

jest.mock('../services/notificationService', () => ({
  NotificationService: {
    unreadCount: jest.fn(),
    list: jest.fn(),
    markRead: jest.fn(),
    markAllRead: jest.fn(),
  },
}));

jest.mock('../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('next/navigation', () => ({
  usePathname: jest.fn(() => '/machines'),
}));

describe('NotificationBell', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useAuth as jest.Mock).mockReturnValue({ isAuthenticated: true });
    (NotificationService.unreadCount as jest.Mock).mockResolvedValue({
      data: { unread_count: 3 },
    });
  });

  it('shows the unread count from a single fetch on load', async () => {
    render(<NotificationBell />);

    expect(await screen.findByText('3')).toBeInTheDocument();
    expect(NotificationService.unreadCount).toHaveBeenCalledTimes(1);
  });

  it('does not poll on a timer', async () => {
    jest.useFakeTimers();
    try {
      render(<NotificationBell />);

      await waitFor(() => {
        expect(NotificationService.unreadCount).toHaveBeenCalledTimes(1);
      });

      await act(async () => {
        jest.advanceTimersByTime(5 * 60 * 1000);
      });

      expect(NotificationService.unreadCount).toHaveBeenCalledTimes(1);
    } finally {
      jest.useRealTimers();
    }
  });

  it('renders nothing when not signed in', () => {
    (useAuth as jest.Mock).mockReturnValue({ isAuthenticated: false });

    const { container } = render(<NotificationBell />);

    expect(container).toBeEmptyDOMElement();
    expect(NotificationService.unreadCount).not.toHaveBeenCalled();
  });
});
