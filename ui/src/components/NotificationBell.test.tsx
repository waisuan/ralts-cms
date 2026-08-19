import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NotificationBell from './NotificationBell';
import { NotificationService } from '../services/notificationService';
import { NOTIFICATIONS_READ_EVENT } from '../hooks/useNotificationReadSync';
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

  it('re-reads the count when the inbox page marks something read', async () => {
    render(<NotificationBell />);
    expect(await screen.findByText('3')).toBeInTheDocument();

    (NotificationService.unreadCount as jest.Mock).mockResolvedValue({
      data: { unread_count: 0 },
    });
    await act(async () => {
      window.dispatchEvent(new Event(NOTIFICATIONS_READ_EVENT));
    });

    // The badge is gone entirely once nothing is unread.
    await waitFor(() => expect(screen.queryByText('3')).not.toBeInTheDocument());
    expect(NotificationService.unreadCount).toHaveBeenCalledTimes(2);
  });

  it('announces its own mark-all-read so the inbox page can follow', async () => {
    const user = userEvent.setup();
    (NotificationService.list as jest.Mock).mockResolvedValue({
      data: {
        notifications: [
          {
            id: 'n-1',
            user_id: 7,
            type: 'assigned',
            title: 'You were assigned to machine SN-1',
            created_at: '2024-06-01T00:00:00Z',
            read_at: null,
          },
        ],
        count: 1,
        unread_count: 1,
        limit: 10,
        offset: 0,
      },
    });
    (NotificationService.markAllRead as jest.Mock).mockResolvedValue({});
    const heard = jest.fn();
    window.addEventListener(NOTIFICATIONS_READ_EVENT, heard);

    try {
      render(<NotificationBell />);
      await user.click(screen.getByRole('button', { name: 'Notifications' }));
      await user.click(await screen.findByRole('button', { name: 'Mark all read' }));

      await waitFor(() => expect(heard).toHaveBeenCalledTimes(1));
    } finally {
      window.removeEventListener(NOTIFICATIONS_READ_EVENT, heard);
    }
  });

  it('renders nothing when not signed in', () => {
    (useAuth as jest.Mock).mockReturnValue({ isAuthenticated: false });

    const { container } = render(<NotificationBell />);

    expect(container).toBeEmptyDOMElement();
    expect(NotificationService.unreadCount).not.toHaveBeenCalled();
  });
});
