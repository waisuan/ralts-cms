import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RecordsTable from './RecordsTable';
import { Machine } from '@/types/machine';

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
    ...overrides,
  };
}

describe('RecordsTable', () => {
  it('renders machine rows and calls onSortChange when sorting PPM Date', async () => {
    const user = userEvent.setup();
    const onSortChange = jest.fn();
    const machines = [baseMachine({ serial_number: 'A' }), baseMachine({ serial_number: 'B', customer: 'Beta' })];

    render(
      <RecordsTable
        machines={machines}
        total={2}
        offset={0}
        limit={50}
        loading={false}
        sortBy="updated_at_desc"
        onSortChange={onSortChange}
        onPageChange={jest.fn()}
        onPageSizeChange={jest.fn()}
        onView={jest.fn()}
        onEdit={jest.fn()}
        onDelete={jest.fn()}
      />
    );

    expect(screen.getByText('Acme')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();

    await user.click(screen.getByRole('columnheader', { name: /ppm date/i }));
    expect(onSortChange).toHaveBeenCalledWith('ppm_date_asc');
  });
});
