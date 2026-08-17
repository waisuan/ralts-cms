import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import FlagMachineModal from './FlagMachineModal';
import { Flag, FlagService } from '../services/flagService';
import { formatDateTime } from '../utils/formatters';

jest.mock('../services/flagService', () => ({
  FlagService: {
    create: jest.fn(),
    listForMachine: jest.fn(),
  },
  MAX_FLAG_NOTE_LENGTH: 500,
}));

const mockOpenFlag = (flag?: Partial<Flag>) => {
  (FlagService.listForMachine as jest.Mock).mockResolvedValue({
    data: {
      flags: flag
        ? [
            {
              id: 'existing-1',
              machine_serial_number: 'SN-1',
              reason: 'missing_values',
              status: 'open',
              created_at: '2024-06-01T00:00:00Z',
              ...flag,
            },
          ]
        : [],
    },
  });
};

/**
 * Renders the modal and waits for the existing-flag lookup to settle, which is
 * what unblocks the submit button.
 */
const renderModal = async (props: Partial<React.ComponentProps<typeof FlagMachineModal>> = {}) => {
  const onClose = jest.fn();
  const onCreated = jest.fn();
  render(
    <FlagMachineModal
      isOpen
      serialNumber="SN-1"
      onClose={onClose}
      onCreated={onCreated}
      {...props}
    />
  );
  await waitFor(() => {
    expect(screen.getByRole('button', { name: /flag|updat/i })).toBeEnabled();
  });
  return { onClose, onCreated };
};

describe('FlagMachineModal', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockOpenFlag();
    (FlagService.create as jest.Mock).mockResolvedValue({
      data: { id: 'f-1', machine_serial_number: 'SN-1', reason: 'missing_values', status: 'open' },
    });
  });

  it('renders nothing while closed, and looks up nothing', () => {
    render(<FlagMachineModal isOpen={false} serialNumber="SN-1" onClose={jest.fn()} />);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(FlagService.listForMachine).not.toHaveBeenCalled();
  });

  it('creates a missing-values flag and reports it back to the caller', async () => {
    const user = userEvent.setup();
    const { onClose, onCreated } = await renderModal();

    expect(screen.getByRole('heading', { name: 'Flag Machine' })).toBeInTheDocument();
    await user.type(screen.getByLabelText(/note/i), 'PPM date is blank');
    await user.click(screen.getByRole('button', { name: /flag machine/i }));

    await waitFor(() => {
      expect(FlagService.create).toHaveBeenCalledWith('SN-1', {
        reason: 'missing_values',
        note: 'PPM date is blank',
      });
    });
    expect(onCreated).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'f-1', machine_serial_number: 'SN-1' })
    );
    expect(onClose).toHaveBeenCalled();
  });

  it('opens on the existing flag when the machine is already flagged', async () => {
    const user = userEvent.setup();
    mockOpenFlag({ reason: 'other', note: 'Customer disputes the reading', created_by_username: 'admin' });

    await renderModal();

    expect(FlagService.listForMachine).toHaveBeenCalledWith('SN-1');
    expect(screen.getByRole('heading', { name: 'Update Flag' })).toBeInTheDocument();
    // Says when the flag being replaced was raised, and by whom.
    const banner = screen.getByText(/already flagged/i);
    expect(banner).toHaveTextContent(formatDateTime('2024-06-01T00:00:00Z'));
    expect(banner).toHaveTextContent('by admin');
    expect(screen.getByLabelText(/reason/i)).toHaveValue('other');
    expect(screen.getByLabelText(/note/i)).toHaveValue('Customer disputes the reading');

    // Saving replaces that flag rather than adding a second one.
    await user.clear(screen.getByLabelText(/note/i));
    await user.type(screen.getByLabelText(/note/i), 'Still wrong');
    await user.click(screen.getByRole('button', { name: /update flag/i }));

    await waitFor(() => {
      expect(FlagService.create).toHaveBeenCalledWith('SN-1', {
        reason: 'other',
        note: 'Still wrong',
      });
    });
  });

  it('falls back to a blank form when the lookup fails', async () => {
    (FlagService.listForMachine as jest.Mock).mockRejectedValue(new Error('offline'));

    await renderModal();

    expect(screen.getByRole('heading', { name: 'Flag Machine' })).toBeInTheDocument();
    expect(screen.getByLabelText(/note/i)).toHaveValue('');
  });

  it('shows how much of the note allowance is left and stops at the limit', async () => {
    const user = userEvent.setup();
    await renderModal();

    const note = screen.getByLabelText(/note/i);
    expect(screen.getByText('0/500')).toBeInTheDocument();

    await user.type(note, 'Meter reading looks wrong');
    expect(screen.getByText('25/500')).toBeInTheDocument();

    // The field refuses further input rather than silently truncating on submit.
    await user.clear(note);
    await user.paste('x'.repeat(600));
    expect(note).toHaveValue('x'.repeat(500));
    expect(screen.getByText('500/500')).toBeInTheDocument();
  });

  it('requires a note when the reason is Other', async () => {
    const user = userEvent.setup();
    const { onClose } = await renderModal();

    await user.selectOptions(screen.getByLabelText(/reason/i), 'other');
    await user.click(screen.getByRole('button', { name: /flag machine/i }));

    expect(await screen.findByText(/a note is required/i)).toBeInTheDocument();
    expect(FlagService.create).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it('keeps the modal open and shows the error when the API rejects', async () => {
    const user = userEvent.setup();
    (FlagService.create as jest.Mock).mockRejectedValue(new Error('Forbidden'));
    const { onClose } = await renderModal();

    await user.click(screen.getByRole('button', { name: /flag machine/i }));

    expect(await screen.findByText(/forbidden/i)).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });
});
