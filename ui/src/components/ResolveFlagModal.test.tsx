import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ResolveFlagModal from './ResolveFlagModal';
import { Flag, FlagService } from '../services/flagService';
import { formatDateTime } from '../utils/formatters';

jest.mock('../services/flagService', () => ({
  FlagService: {
    resolve: jest.fn(),
    reasonLabel: (reason: string) => (reason === 'missing_values' ? 'Missing values' : 'Other'),
  },
  MAX_FLAG_NOTE_LENGTH: 500,
}));

const flag: Flag = {
  id: 'f-1',
  machine_serial_number: 'SN-1',
  reason: 'missing_values',
  note: 'PPM date is blank',
  status: 'open',
  created_at: '2024-06-01T00:00:00Z',
};

const renderModal = (overrides: Partial<React.ComponentProps<typeof ResolveFlagModal>> = {}) => {
  const props = {
    flag,
    onClose: jest.fn(),
    onResolved: jest.fn(),
    ...overrides,
  };
  render(<ResolveFlagModal {...props} />);
  return props;
};

describe('ResolveFlagModal', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (FlagService.resolve as jest.Mock).mockResolvedValue({});
  });

  it('shows why the machine was flagged, so the resolver has the context', () => {
    renderModal();

    expect(screen.getByText('SN-1 · Missing values')).toBeInTheDocument();
    expect(screen.getByText('PPM date is blank')).toBeInTheDocument();
  });

  it('says when the flag was raised, so the resolver knows how old it is', () => {
    renderModal({ flag: { ...flag, created_by_username: 'admin' } });

    const raised = screen.getByText(/^Flagged .+ by /);
    expect(raised).toHaveTextContent(formatDateTime('2024-06-01T00:00:00Z'));
    expect(raised).toHaveTextContent('by admin');
  });

  it('resolves with the note and reports back', async () => {
    const user = userEvent.setup();
    const props = renderModal();

    await user.type(screen.getByLabelText(/resolution note/i), 'Replaced the sensor');
    await user.click(screen.getByRole('button', { name: /resolve flag/i }));

    expect(FlagService.resolve).toHaveBeenCalledWith('SN-1', 'f-1', 'Replaced the sensor');
    await waitFor(() => {
      expect(props.onResolved).toHaveBeenCalled();
    });
    expect(props.onClose).toHaveBeenCalled();
  });

  it('treats the note as optional', async () => {
    const user = userEvent.setup();
    renderModal();

    await user.click(screen.getByRole('button', { name: /resolve flag/i }));

    expect(FlagService.resolve).toHaveBeenCalledWith('SN-1', 'f-1', '');
  });

  it('counts the note against the limit and stops at it', async () => {
    const user = userEvent.setup();
    renderModal();

    const note = screen.getByLabelText(/resolution note/i);
    expect(screen.getByText('0/500')).toBeInTheDocument();

    await user.click(note);
    await user.paste('x'.repeat(600));
    expect(note).toHaveValue('x'.repeat(500));
    expect(screen.getByText('500/500')).toBeInTheDocument();
  });

  it('keeps the modal open and shows the error when resolving fails', async () => {
    const user = userEvent.setup();
    (FlagService.resolve as jest.Mock).mockRejectedValue(new Error('Flag not found'));
    const props = renderModal();

    await user.click(screen.getByRole('button', { name: /resolve flag/i }));

    expect(await screen.findByText('Flag not found')).toBeInTheDocument();
    expect(props.onClose).not.toHaveBeenCalled();
    expect(props.onResolved).not.toHaveBeenCalled();
  });
});
