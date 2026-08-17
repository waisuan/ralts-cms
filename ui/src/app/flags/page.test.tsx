import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import FlagsPage from './page';
import { Flag, FlagScope, FlagService } from '@/services/flagService';
import { formatDateTime } from '@/utils/formatters';
import { useAuth } from '@/contexts/AuthContext';

jest.mock('@/services/flagService', () => ({
  FlagService: {
    listAll: jest.fn(),
    resolve: jest.fn(),
    reasonLabel: (reason: string) => (reason === 'missing_values' ? 'Missing values' : 'Other'),
  },
  MAX_FLAG_NOTE_LENGTH: 500,
}));

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

const replace = jest.fn();
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace }),
}));

const flag = (overrides: Partial<Flag> = {}): Flag => ({
  id: 'f-1',
  machine_serial_number: 'SN-1',
  reason: 'missing_values',
  note: 'PPM date is blank',
  status: 'open',
  created_by_username: 'admin',
  created_at: '2024-06-01T00:00:00Z',
  ...overrides,
});

const mockList = (flags: Flag[], scope: FlagScope = 'all') => {
  (FlagService.listAll as jest.Mock).mockResolvedValue({
    data: { flags, count: flags.length, limit: 50, offset: 0, scope },
  });
};

const signInAs = (role: 'ADMIN' | 'USER') => {
  (useAuth as jest.Mock).mockReturnValue({
    user: { id: 1, username: 'u', email: 'u@example.com', role, approved: true },
    isLoading: false,
  });
};

describe('FlagsPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    signInAs('ADMIN');
  });

  it('lists open flags by default', async () => {
    mockList([flag()]);

    render(<FlagsPage />);

    expect(await screen.findByText('SN-1')).toBeInTheDocument();
    expect(screen.getByText('PPM date is blank')).toBeInTheDocument();
    expect(FlagService.listAll).toHaveBeenCalledWith({ status: 'open', limit: 50, offset: 0 });
  });

  it('asks for an optional resolution note before resolving, then reloads', async () => {
    const user = userEvent.setup();
    mockList([flag()]);
    (FlagService.resolve as jest.Mock).mockResolvedValue({});

    render(<FlagsPage />);

    await user.click(await screen.findByRole('button', { name: /mark resolved/i }));

    // Nothing is resolved until the prompt is confirmed.
    const dialog = await screen.findByRole('dialog');
    expect(FlagService.resolve).not.toHaveBeenCalled();

    await user.type(screen.getByLabelText(/resolution note/i), 'Filled in the PPM date');
    await user.click(within(dialog).getByRole('button', { name: /resolve flag/i }));

    expect(FlagService.resolve).toHaveBeenCalledWith('SN-1', 'f-1', 'Filled in the PPM date');
    await waitFor(() => {
      expect(FlagService.listAll).toHaveBeenCalledTimes(2);
    });
  });

  it('resolves with no note when the prompt is left blank', async () => {
    const user = userEvent.setup();
    mockList([flag()]);
    (FlagService.resolve as jest.Mock).mockResolvedValue({});

    render(<FlagsPage />);

    await user.click(await screen.findByRole('button', { name: /mark resolved/i }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: /resolve flag/i }));

    expect(FlagService.resolve).toHaveBeenCalledWith('SN-1', 'f-1', '');
  });

  it('leaves the flag alone when the prompt is cancelled', async () => {
    const user = userEvent.setup();
    mockList([flag()]);

    render(<FlagsPage />);

    await user.click(await screen.findByRole('button', { name: /mark resolved/i }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: /cancel/i }));

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
    expect(FlagService.resolve).not.toHaveBeenCalled();
  });

  it('shows the resolution note on a resolved flag', async () => {
    mockList([
      flag({
        status: 'resolved',
        resolved_by_username: 'admin',
        resolution_note: 'Replaced the sensor',
      }),
    ]);

    render(<FlagsPage />);

    expect(await screen.findByText(/Replaced the sensor/)).toBeInTheDocument();
    expect(screen.getByText(/Resolution:/)).toBeInTheDocument();
  });

  it('dates an open flag by when it was raised', async () => {
    mockList([flag()]);

    render(<FlagsPage />);

    const raised = await screen.findByText(/^Flagged .+ by /);
    expect(raised).toHaveTextContent(formatDateTime('2024-06-01T00:00:00Z'));
    expect(raised).toHaveTextContent('by admin');
    expect(screen.queryByText(/^Resolved .+ by /)).not.toBeInTheDocument();
  });

  it('dates a resolved flag by both events, not just the raise', async () => {
    mockList([
      flag({
        status: 'resolved',
        resolved_by_username: 'tech',
        resolved_at: '2024-06-03T08:30:00Z',
      }),
    ]);

    render(<FlagsPage />);

    const closed = await screen.findByText(/^Resolved .+ by /);
    expect(closed).toHaveTextContent(formatDateTime('2024-06-03T08:30:00Z'));
    expect(closed).toHaveTextContent('by tech');
    expect(screen.getByText(/^Flagged .+ by /)).toHaveTextContent(
      formatDateTime('2024-06-01T00:00:00Z')
    );
  });

  it('switches the status filter', async () => {
    const user = userEvent.setup();
    mockList([]);

    render(<FlagsPage />);

    await user.click(await screen.findByRole('button', { name: 'Resolved' }));

    await waitFor(() => {
      expect(FlagService.listAll).toHaveBeenLastCalledWith({
        status: 'resolved',
        limit: 50,
        offset: 0,
      });
    });
  });

  it('shows non-admins their own assigned flags, which they can resolve', async () => {
    const user = userEvent.setup();
    signInAs('USER');
    mockList([flag({ machine_serial_number: 'SN-MINE' })], 'assigned');
    (FlagService.resolve as jest.Mock).mockResolvedValue({});

    render(<FlagsPage />);

    expect(await screen.findByText('SN-MINE')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'My flagged records' })).toBeInTheDocument();
    expect(screen.getByText(/machines assigned to you/i)).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /mark resolved/i }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: /resolve flag/i }));
    expect(FlagService.resolve).toHaveBeenCalledWith('SN-MINE', 'f-1', '');
  });

  it('tells non-admins why their resolved list is narrow, but not on the open tab', async () => {
    const user = userEvent.setup();
    signInAs('USER');
    mockList([flag({ machine_serial_number: 'SN-MINE' })], 'assigned');

    render(<FlagsPage />);

    await screen.findByText('SN-MINE');
    expect(screen.queryByText(/limited to ones you closed yourself/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Resolved' }));
    expect(await screen.findByText(/limited to ones you closed yourself/i)).toBeInTheDocument();
  });

  it('does not narrow the resolved view for admins', async () => {
    const user = userEvent.setup();
    mockList([flag()], 'all');

    render(<FlagsPage />);

    await screen.findByText('SN-1');
    await user.click(screen.getByRole('button', { name: 'Resolved' }));

    await waitFor(() => {
      expect(FlagService.listAll).toHaveBeenCalledTimes(2);
    });
    expect(screen.queryByText(/limited to ones you closed yourself/i)).not.toBeInTheDocument();
  });

  it('sends signed-out visitors away without requesting any flags', async () => {
    (useAuth as jest.Mock).mockReturnValue({ user: null, isLoading: false });
    mockList([flag()]);

    render(<FlagsPage />);

    await waitFor(() => {
      expect(replace).toHaveBeenCalledWith('/');
    });
    expect(FlagService.listAll).not.toHaveBeenCalled();
  });
});
