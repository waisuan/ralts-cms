import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsTable from './RecordsTable';
import { Machine } from '@/types/machine';
import { FIXTURE_MACHINES } from '@/__fixtures__/machines';
import { USER_ROLE, type UserRole } from '../services/adminUserService';
import { useOptionalAuth } from '../contexts/AuthContext';

jest.mock('../services/attachmentService', () => ({
  AttachmentService: {
    downloadMachineAttachment: jest.fn(),
  },
}));

jest.mock('../contexts/AuthContext', () => ({
  useOptionalAuth: jest.fn(),
}));

const signedInAs = (id: number, role: UserRole) =>
  (useOptionalAuth as jest.Mock).mockReturnValue({
    user: { id, username: 'u', email: 'e', role, approved: true },
  });

function baseMachine(overrides: Partial<Machine> = {}): Machine {
  return {
    serial_number: 'SN-1',
    customer: 'Acme',
    state: 'CA',
    account_type: 'T',
    model: 'M',
    status: '',
    brand: 'B',
    district: 'D',
    person_in_charge: 'P',
    reported_by: 'R',
    additional_notes: '',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-01-01',
    ppm_date: '2024-02-01',
    created_at: '',
    updated_at: '2024-06-01',
    updated_by: '',
    ...overrides,
  };
}

const defaultProps = {
  total: 0,
  offset: 0,
  limit: 50,
  loading: false,
  sortBy: 'updated_at_desc',
  onSortChange: jest.fn(),
  onPageChange: jest.fn(),
  onPageSizeChange: jest.fn(),
  onEdit: jest.fn(),
  onDelete: jest.fn(),
};

describe('RecordsTable', () => {
  it('renders machine rows and calls onSortChange when sorting PPM Date', async () => {
    const user = userEvent.setup();
    const onSortChange = jest.fn();
    const machines = [baseMachine({ serial_number: 'A' }), baseMachine({ serial_number: 'B', customer: 'Beta' })];

    render(
      <RecordsTable
        {...defaultProps}
        machines={machines}
        total={2}
        onSortChange={onSortChange}
      />
    );

    expect(screen.getByText('Acme')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();

    await user.click(screen.getByRole('columnheader', { name: /ppm date/i }));
    expect(onSortChange).toHaveBeenCalledWith('ppm_date_asc');
  });

  describe('with realistic production data', () => {
    it('renders all fixture rows', () => {
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES}
          total={FIXTURE_MACHINES.length}
        />
      );

      for (const machine of FIXTURE_MACHINES) {
        expect(screen.getByTitle(machine.serial_number)).toBeInTheDocument();
      }
    });

    it('applies truncation classes to text cells that may overflow', () => {
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES}
          total={FIXTURE_MACHINES.length}
        />
      );

      const longCustomer = screen.getByTitle('INSTITUT PERUBATAN RESPIRATORI');
      expect(longCustomer).toHaveClass('truncate', 'block');

      const longAssignee = screen.getByTitle('Puan Nurhasyimah Bt Mohd Noor');
      expect(longAssignee).toHaveClass('truncate', 'block');

      const longSerial = screen.getByTitle('UG-10203399 [R]');
      expect(longSerial).toHaveClass('truncate', 'block');
    });

    it('colors the PPM date text by status instead of showing a pill', () => {
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES}
          total={FIXTURE_MACHINES.length}
        />
      );

      // No pill badges in the collapsed row anymore.
      expect(screen.queryByText('Overdue')).not.toBeInTheDocument();
      expect(screen.queryByText('Upcoming')).not.toBeInTheDocument();
      expect(screen.queryByText('Due')).not.toBeInTheDocument();

      const overdueDates = screen.getAllByTitle('Overdue');
      expect(overdueDates.length).toBeGreaterThan(0);
      overdueDates.forEach((el) => expect(el).toHaveClass('text-red-600'));

      const upcomingDates = screen.getAllByTitle('Upcoming');
      expect(upcomingDates).toHaveLength(1);
      upcomingDates.forEach((el) => expect(el).toHaveClass('text-yellow-700'));

      const dueDates = screen.getAllByTitle('Due');
      expect(dueDates).toHaveLength(1);
      dueDates.forEach((el) => expect(el).toHaveClass('text-orange-600'));
    });

    it('surfaces the PPM status label in the expanded row detail', async () => {
      const user = userEvent.setup();
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES}
          total={FIXTURE_MACHINES.length}
        />
      );

      // First fixture row is overdue but has a sentinel ppm_date, so the
      // collapsed cell shows "-" with no title; status should still surface
      // once expanded.
      const expandButtons = screen.getAllByLabelText('Toggle row details');
      await user.click(expandButtons[0]);

      expect(screen.getByText('PPM Status')).toBeInTheDocument();
      expect(screen.getAllByText('Overdue').length).toBeGreaterThan(0);
    });

    it('shows Updated By in the expanded row detail when present', async () => {
      const user = userEvent.setup();
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES}
          total={FIXTURE_MACHINES.length}
        />
      );

      const machineWithEditor = FIXTURE_MACHINES.find((m) => m.updated_by);
      expect(machineWithEditor).toBeDefined();

      const expandButtons = screen.getAllByLabelText('Toggle row details');
      const rowIndex = FIXTURE_MACHINES.indexOf(machineWithEditor!);
      await user.click(expandButtons[rowIndex]);

      expect(screen.getByText('Updated By')).toBeInTheDocument();
      expect(screen.getByText(machineWithEditor!.updated_by)).toBeInTheDocument();
    });

    it('does not set title attribute on cells with empty values', () => {
      const emptyMachine = baseMachine({
        serial_number: 'EMPTY-TEST',
        customer: '',
        person_in_charge: '',
        district: '',
      });

      render(
        <RecordsTable
          {...defaultProps}
          machines={[emptyMachine]}
          total={1}
        />
      );

      const dashes = screen.getAllByText('-');
      for (const dash of dashes) {
        expect(dash).not.toHaveAttribute('title');
      }
    });

    it('uses table-fixed layout', () => {
      render(
        <RecordsTable
          {...defaultProps}
          machines={FIXTURE_MACHINES.slice(0, 3)}
          total={3}
        />
      );

      const table = screen.getByRole('table');
      expect(table).toHaveClass('table-fixed');
    });
  });

  describe('resolving a flag from the row actions', () => {
    const openFlag = {
      id: 'f-1',
      machine_serial_number: 'SN-FLAGGED',
      reason: 'missing_values' as const,
      status: 'open' as const,
      created_at: '2024-01-01T00:00:00Z',
    };
    const flaggedMachine = baseMachine({ serial_number: 'SN-FLAGGED', assigned_user_id: 7 });

    const openActionsOn = async (
      user: ReturnType<typeof userEvent.setup>,
      { onResolveFlag = jest.fn(), flagged = true } = {}
    ) => {
      render(
        <RecordsTable
          {...defaultProps}
          machines={[flaggedMachine]}
          total={1}
          onResolveFlag={onResolveFlag}
          flagsBySerial={flagged ? { 'SN-FLAGGED': [openFlag] } : {}}
        />
      );
      await user.click(screen.getByTitle('Actions'));
      return onResolveFlag;
    };

    it('hands the assignee the open flag to resolve', async () => {
      const user = userEvent.setup();
      signedInAs(7, USER_ROLE.NON_ADMIN);
      const onResolveFlag = await openActionsOn(user);

      await user.click(screen.getByRole('button', { name: 'Resolve flag' }));

      expect(onResolveFlag).toHaveBeenCalledWith(openFlag);
    });

    it('offers it to an admin on a machine assigned to somebody else', async () => {
      const user = userEvent.setup();
      signedInAs(99, USER_ROLE.ADMIN);
      await openActionsOn(user);

      expect(screen.getByRole('button', { name: 'Resolve flag' })).toBeInTheDocument();
    });

    it('withholds it from a user the machine is not assigned to', async () => {
      const user = userEvent.setup();
      signedInAs(8, USER_ROLE.NON_ADMIN);
      await openActionsOn(user);

      expect(screen.queryByRole('button', { name: 'Resolve flag' })).not.toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Edit' })).toBeInTheDocument();
    });

    it('shows nothing to resolve on a machine without an open flag', async () => {
      const user = userEvent.setup();
      signedInAs(7, USER_ROLE.NON_ADMIN);
      await openActionsOn(user, { flagged: false });

      expect(screen.queryByRole('button', { name: 'Resolve flag' })).not.toBeInTheDocument();
    });
  });
});
