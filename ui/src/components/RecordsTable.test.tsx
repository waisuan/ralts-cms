import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsTable from './RecordsTable';
import { Machine } from '@/types/machine';
import { FIXTURE_MACHINES } from '@/__fixtures__/machines';

jest.mock('../services/attachmentService', () => ({
  AttachmentService: {
    downloadMachineAttachment: jest.fn(),
  },
}));

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
});
