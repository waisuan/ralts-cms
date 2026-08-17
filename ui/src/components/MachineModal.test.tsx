import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MachineModal from './MachineModal';
import { Machine } from '../types/machine';
import { UserDirectoryService } from '../services/userDirectoryService';

// Mock the current date for consistent testing
const MOCK_CURRENT_DATE = '2024-06-29T12:00:00.000Z';

// The MachineModal fetches the user directory on open; stub it so tests are
// deterministic and don't hit the real API. John Doe corresponds to the
// assignee on mockMachine so the edit-mode assertions can verify it renders.
jest.mock('../services/userDirectoryService', () => ({
  UserDirectoryService: {
    list: jest.fn(),
  },
}));

// Mock machine data for testing
const mockMachine: Machine = {
  serial_number: 'TEST-001',
  customer: 'Test Customer',
  state: 'Selangor',
  account_type: 'Premium',
  model: 'Model-X',
  status: 'Active',
  brand: 'Test Brand',
  district: 'Petaling',
  person_in_charge: 'John Doe',
  assigned_user_id: 7,
  assigned_user: { id: 7, username: 'John Doe', email: 'john@example.com' },
  reported_by: 'Jane Smith',
  additional_notes: 'Test notes',
  attachment: 'test-file.pdf',
  ppm_status: 'Good',
  tnc_date: '2024-06-15',
  ppm_date: '2024-07-15',
  created_at: '2024-01-01T00:00:00.000Z',
  updated_at: '2024-06-01T00:00:00.000Z',
  updated_by: 'admin',
};

describe('MachineModal', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date(MOCK_CURRENT_DATE).getTime());
    (UserDirectoryService.list as jest.Mock).mockResolvedValue({
      data: {
        users: [
          { id: 7, username: 'John Doe', email: 'john@example.com' },
          { id: 8, username: 'jane.smith', email: 'jane@example.com' },
        ],
      },
    });
  });

  afterEach(() => {
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it('renders correctly in add mode with form fields', async () => {
    const addProps = {
      isOpen: true,
      mode: 'add' as const,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    render(<MachineModal {...addProps} />);

    expect(screen.getByRole('heading', { name: 'Add New Machine' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add machine/i })).toBeInTheDocument();

    expect(screen.getByLabelText(/serial number/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/customer/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/state/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/model/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/brand/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/district/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/assignee/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/reported by/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/additional notes/i)).toBeInTheDocument();
  });

  it('renders correctly in edit mode with pre-populated data', async () => {
    const editProps = {
      isOpen: true,
      mode: 'edit' as const,
      machine: mockMachine,
      onClose: jest.fn(),
      onSubmit: jest.fn(),
    };

    render(<MachineModal {...editProps} />);

    expect(screen.getByRole('heading', { name: 'Edit Machine' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /update machine/i })).toBeInTheDocument();

    expect(screen.getByDisplayValue('TEST-001')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Customer')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Selangor')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Premium')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Model-X')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Brand')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Petaling')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Jane Smith')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test notes')).toBeInTheDocument();

    // Assignee dropdown pre-selects the linked user.
    await waitFor(() => {
      const assignee = screen.getByLabelText(/assignee/i) as HTMLSelectElement;
      expect(assignee.value).toBe('7');
    });
    expect(screen.queryByLabelText('Assignee name')).not.toBeInTheDocument();
  });

  describe('assignee field', () => {
    const renderAddMode = () =>
      render(
        <MachineModal isOpen mode="add" onClose={jest.fn()} onSubmit={jest.fn()} />,
      );

    const assigneeSelect = () =>
      screen.getByRole('combobox', { name: /assignee/i }) as HTMLSelectElement;

    /** Directory users only become selectable once the fetch resolves. */
    const waitForDirectory = () =>
      waitFor(() => {
        expect(Array.from(assigneeSelect().options).map((o) => o.value)).toContain('8');
      });

    it('is a dropdown of directory users with no free-text input by default', async () => {
      renderAddMode();

      expect(assigneeSelect().tagName).toBe('SELECT');

      await waitFor(() => {
        const options = Array.from(assigneeSelect().options).map((o) => o.value);
        expect(options).toEqual(['', '7', '8', '__free_text__']);
      });
      expect(screen.queryByLabelText('Assignee name')).not.toBeInTheDocument();
    });

    it('explains that the selected user will be notified', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      renderAddMode();

      await waitForDirectory();

      await user.selectOptions(assigneeSelect(), '8');

      expect(
        await screen.findByText(/notifications will go to jane\.smith/i),
      ).toBeInTheDocument();
    });

    it('only shows free text when the user explicitly asks for it', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      renderAddMode();

      await waitForDirectory();

      await user.selectOptions(assigneeSelect(), '__free_text__');

      const freeText = await screen.findByLabelText('Assignee name');
      await user.type(freeText, 'Brand New Starter');

      expect((freeText as HTMLInputElement).value).toBe('Brand New Starter');
      expect(
        await screen.findByText(/will not receive notifications/i),
      ).toBeInTheDocument();
    });

    it('switching back to a user hides the free-text input', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
      renderAddMode();

      await waitForDirectory();

      await user.selectOptions(assigneeSelect(), '__free_text__');
      await user.type(await screen.findByLabelText('Assignee name'), 'Ghost');
      await user.selectOptions(assigneeSelect(), '7');

      expect(screen.queryByLabelText('Assignee name')).not.toBeInTheDocument();
      expect(
        await screen.findByText(/notifications will go to John Doe/i),
      ).toBeInTheDocument();
    });
  });
});
