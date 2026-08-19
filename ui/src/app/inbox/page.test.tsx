import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import InboxPage from './page';
import { NotificationService, Notification } from '@/services/notificationService';
import { NOTIFICATIONS_READ_EVENT } from '@/hooks/useNotificationReadSync';
import { useAuth } from '@/contexts/AuthContext';

jest.mock('@/services/notificationService', () => ({
  NotificationService: {
    list: jest.fn(),
    markRead: jest.fn(),
    markAllRead: jest.fn(),
  },
}));

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => ({ replace: jest.fn() })),
}));

const flagNotification = (overrides: Partial<Notification> = {}): Notification => ({
  id: 'n-1',
  user_id: 7,
  type: 'flagged',
  machine_serial_number: 'SN-1',
  flag_id: 'f-1',
  flag_status: 'open',
  title: 'Machine SN-1 flagged: missing values',
  created_at: '2024-06-01T00:00:00Z',
  read_at: '2024-06-01T01:00:00Z',
  ...overrides,
});

const mockList = (notifications: Notification[]) => {
  (NotificationService.list as jest.Mock).mockResolvedValue({
    data: {
      notifications,
      count: notifications.length,
      unread_count: notifications.filter((n) => !n.read_at).length,
      limit: 25,
      offset: 0,
    },
  });
};

describe('InboxPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useAuth as jest.Mock).mockReturnValue({
      user: { id: 7, username: 'tech', email: 't@example.com', role: 'USER', approved: true },
      isLoading: false,
    });
  });

  it('points an open flag at the flagged records page instead of resolving it', async () => {
    mockList([flagNotification()]);

    render(<InboxPage />);

    const link = await screen.findByRole('link', { name: /resolve in flagged records/i });
    expect(link).toHaveAttribute('href', '/flags');
    expect(screen.queryByRole('button', { name: /resolve/i })).not.toBeInTheDocument();
  });

  it('marks an already-resolved flag as resolved without offering an action', async () => {
    mockList([flagNotification({ flag_status: 'resolved' })]);

    render(<InboxPage />);

    expect(await screen.findByText(/flag resolved/i)).toBeInTheDocument();
    expect(
      screen.queryByRole('link', { name: /resolve in flagged records/i })
    ).not.toBeInTheDocument();
  });

  it('labels a resolution notification once, carrying what the resolver said', async () => {
    mockList([
      flagNotification({
        type: 'flag_resolved',
        flag_status: 'resolved',
        title: 'Machine SN-1: flag resolved',
        body: 'Filled in the district',
        actor_username: 'tech',
      }),
    ]);

    render(<InboxPage />);

    expect(await screen.findByText('Filled in the district')).toBeInTheDocument();
    expect(screen.getByText(/by tech/i)).toBeInTheDocument();
    // The type chip says it; a second "Flag resolved" would be the old badge.
    expect(screen.getAllByText('Flag resolved')).toHaveLength(1);
    expect(
      screen.queryByRole('link', { name: /resolve in flagged records/i })
    ).not.toBeInTheDocument();
  });

  it('shows no flag actions for assignment notifications', async () => {
    mockList([
      flagNotification({
        type: 'assigned',
        flag_id: null,
        flag_status: null,
        title: 'You were assigned to machine SN-1',
      }),
    ]);

    render(<InboxPage />);

    expect(await screen.findByText(/you were assigned to machine sn-1/i)).toBeInTheDocument();
    expect(
      screen.queryByRole('link', { name: /resolve in flagged records/i })
    ).not.toBeInTheDocument();
  });

  it('hides the mark-all-read button when nothing is unread', async () => {
    mockList([flagNotification()]);

    render(<InboxPage />);

    expect(await screen.findByText(/all caught up/i)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /mark all read/i })).not.toBeInTheDocument();
  });

  it('celebrates an empty inbox without repeating itself in the header', async () => {
    mockList([]);

    render(<InboxPage />);

    expect(await screen.findByText('All caught up 🎉')).toBeInTheDocument();
    expect(screen.queryByText('All caught up')).not.toBeInTheDocument();
  });

  it('offers mark-all-read while unread notifications remain', async () => {
    const user = userEvent.setup();
    mockList([flagNotification({ read_at: null })]);
    (NotificationService.markAllRead as jest.Mock).mockResolvedValue({});

    render(<InboxPage />);

    const button = await screen.findByRole('button', { name: /mark all read/i });
    await user.click(button);

    expect(NotificationService.markAllRead).toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /mark all read/i })).not.toBeInTheDocument();
    });
  });

  it('announces its own mark-all-read so the bell can follow', async () => {
    const user = userEvent.setup();
    mockList([flagNotification({ read_at: null })]);
    (NotificationService.markAllRead as jest.Mock).mockResolvedValue({});
    const heard = jest.fn();
    window.addEventListener(NOTIFICATIONS_READ_EVENT, heard);

    try {
      render(<InboxPage />);
      await user.click(await screen.findByRole('button', { name: /mark all read/i }));

      await waitFor(() => expect(heard).toHaveBeenCalledTimes(1));
    } finally {
      window.removeEventListener(NOTIFICATIONS_READ_EVENT, heard);
    }
  });

  it('follows a mark-all-read done from the bell', async () => {
    mockList([flagNotification({ read_at: null })]);

    render(<InboxPage />);
    expect(await screen.findByRole('button', { name: /mark all read/i })).toBeInTheDocument();

    mockList([flagNotification({ read_at: '2024-06-02T00:00:00Z' })]);
    await act(async () => {
      window.dispatchEvent(new Event(NOTIFICATIONS_READ_EVENT));
    });

    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /mark all read/i })).not.toBeInTheDocument();
    });
    expect(screen.getByText('All caught up')).toBeInTheDocument();
    // The reload is quiet: the row the user was reading never left the screen
    // for a spinner.
    expect(screen.getByText(/machine sn-1 flagged/i)).toBeInTheDocument();
  });

  it('keeps the current rows when a quiet reload fails', async () => {
    mockList([flagNotification({ read_at: null })]);

    render(<InboxPage />);
    await screen.findByText(/machine sn-1 flagged/i);

    (NotificationService.list as jest.Mock).mockRejectedValue(new Error('offline'));
    await act(async () => {
      window.dispatchEvent(new Event(NOTIFICATIONS_READ_EVENT));
    });

    // Nobody asked for this fetch, so its failure is not their problem.
    expect(screen.getByText(/machine sn-1 flagged/i)).toBeInTheDocument();
    expect(screen.queryByText('Failed to load notifications')).not.toBeInTheDocument();
  });
});
