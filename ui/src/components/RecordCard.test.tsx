import { render, screen, fireEvent } from '@testing-library/react';
import RecordCard from './RecordCard';
import { Machine } from '../types/machine';
import { USER_ROLE, type UserRole } from '../services/adminUserService';
import { useOptionalAuth } from '../contexts/AuthContext';

jest.mock('../contexts/AuthContext', () => ({
  useOptionalAuth: jest.fn(),
}));

const signedInAs = (id: number, role: UserRole) =>
  (useOptionalAuth as jest.Mock).mockReturnValue({
    user: { id, username: 'u', email: 'e', role, approved: true },
  });

describe('RecordCard', () => {
  const baseMachine: Machine = {
    serial_number: 'SN-TEST',
    customer: 'Test Customer',
    state: 'CA',
    account_type: 'Premium',
    model: 'TestModel',
    status: 'Active',
    brand: 'TestBrand',
    district: 'TestDistrict',
    person_in_charge: 'Test Person',
    reported_by: 'Test Reporter',
    additional_notes: 'Test notes for the machine',
    attachment: 'test_file.pdf',
    ppm_status: 'overdue',
    tnc_date: '2024-07-01',
    ppm_date: '2024-06-24',
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
    updated_by: 'admin',
    maintenance_count: 2,
  };

  const mockOnEdit = jest.fn();
  const mockOnDelete = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders collapsed card with key fields visible', () => {
    render(
      <RecordCard machine={baseMachine} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    expect(screen.getByText('SN-TEST')).toBeInTheDocument();
    expect(screen.getByText('(TestModel)')).toBeInTheDocument();
    expect(screen.getByText('Overdue')).toBeInTheDocument();

    expect(screen.getAllByText((_, el) => Boolean(el?.textContent?.includes('Test Customer'))).length).toBeGreaterThan(0);
    expect(screen.getAllByText((_, el) => Boolean(el?.textContent?.includes('CA'))).length).toBeGreaterThan(0);

    expect(screen.getAllByText((_, el) => Boolean(el?.textContent?.includes('TNC:'))).length).toBeGreaterThan(0);
    expect(screen.getAllByText((_, el) => Boolean(el?.textContent?.includes('PPM:'))).length).toBeGreaterThan(0);
  });

  it('hides detail fields and action buttons when collapsed', () => {
    render(
      <RecordCard machine={baseMachine} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    expect(screen.queryByText('View')).not.toBeInTheDocument();
    expect(screen.queryByText('Edit')).not.toBeInTheDocument();
    expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    expect(screen.queryByText('Brand')).not.toBeInTheDocument();
  });

  it('shows detail fields and action buttons when expanded', () => {
    render(
      <RecordCard machine={baseMachine} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    const toggle = screen.getByRole('button', { expanded: false });
    fireEvent.click(toggle);

    expect(screen.getByText('View')).toBeInTheDocument();
    expect(screen.getByText('Edit')).toBeInTheDocument();
    expect(screen.getByText('Delete')).toBeInTheDocument();
    expect(screen.getByText('TestBrand')).toBeInTheDocument();
    expect(screen.getByText('TestDistrict')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
    expect(screen.getByText('Premium')).toBeInTheDocument();
    expect(screen.getByText('Test Person')).toBeInTheDocument();
    expect(screen.getByText('Test Reporter')).toBeInTheDocument();
    expect(screen.getByText('Test notes for the machine')).toBeInTheDocument();
    expect(screen.getByText('test_file.pdf')).toBeInTheDocument();
    expect(screen.getByText('Updated By')).toBeInTheDocument();
    expect(screen.getByText('admin')).toBeInTheDocument();
  });

  it('omits Updated By field when not present', () => {
    const machineNoUpdatedBy = { ...baseMachine, updated_by: '' };
    render(
      <RecordCard machine={machineNoUpdatedBy} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    const toggle = screen.getByRole('button', { expanded: false });
    fireEvent.click(toggle);

    expect(screen.queryByText('Updated By')).not.toBeInTheDocument();
  });

  it('omits model when not present', () => {
    const machineNoModel = { ...baseMachine, model: '' };
    render(
      <RecordCard machine={machineNoModel} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    expect(screen.getByText('SN-TEST')).toBeInTheDocument();
    expect(screen.queryByText('()')).not.toBeInTheDocument();
  });

  it('calls action handlers with correct serial number', () => {
    render(
      <RecordCard machine={baseMachine} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );

    const toggle = screen.getByRole('button', { expanded: false });
    fireEvent.click(toggle);

    expect(screen.getByRole('link', { name: 'View' })).toHaveAttribute('href', '/machines/SN-TEST');

    fireEvent.click(screen.getByText('Edit'));
    expect(mockOnEdit).toHaveBeenCalledWith('SN-TEST');

    fireEvent.click(screen.getByText('Delete'));
    expect(mockOnDelete).toHaveBeenCalledWith('SN-TEST');
  });

  it('displays PPM status badges correctly', () => {
    const { rerender } = render(
      <RecordCard machine={{ ...baseMachine, ppm_status: 'due' }} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );
    expect(screen.getByText('Due')).toBeInTheDocument();

    rerender(
      <RecordCard machine={{ ...baseMachine, ppm_status: 'almost_due' }} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );
    expect(screen.getByText('Upcoming')).toBeInTheDocument();

    rerender(
      <RecordCard machine={{ ...baseMachine, ppm_status: '' }} onEdit={mockOnEdit} onDelete={mockOnDelete} />
    );
    expect(screen.queryByText('Overdue')).not.toBeInTheDocument();
    expect(screen.queryByText('Due')).not.toBeInTheDocument();
    expect(screen.queryByText('Upcoming')).not.toBeInTheDocument();
  });

  describe('resolving the machine\u2019s flag', () => {
    const openFlag = {
      id: 'f-1',
      machine_serial_number: 'SN-TEST',
      reason: 'missing_values' as const,
      status: 'open' as const,
      created_at: '2024-01-01T00:00:00Z',
    };
    const assignedToMe = { ...baseMachine, assigned_user_id: 7 };

    const expandCard = ({ onResolveFlag = jest.fn(), flagged = true } = {}) => {
      render(
        <RecordCard
          machine={assignedToMe}
          onEdit={mockOnEdit}
          onDelete={mockOnDelete}
          onResolveFlag={onResolveFlag}
          openFlags={flagged ? [openFlag] : []}
        />
      );
      fireEvent.click(screen.getByRole('button', { expanded: false }));
      return onResolveFlag;
    };

    it('hands the assignee the open flag to resolve', () => {
      signedInAs(7, USER_ROLE.NON_ADMIN);
      const onResolveFlag = expandCard();

      fireEvent.click(screen.getByText('Resolve'));

      expect(onResolveFlag).toHaveBeenCalledWith(openFlag);
    });

    it('withholds the action from a user the machine is not assigned to', () => {
      signedInAs(8, USER_ROLE.NON_ADMIN);
      expandCard();

      // The badge shows regardless: reading a flag is open to everyone.
      expect(screen.getByTitle(/missing_values/)).toBeInTheDocument();
      expect(screen.queryByText('Resolve')).not.toBeInTheDocument();
    });

    it('offers nothing to resolve on an unflagged machine', () => {
      signedInAs(7, USER_ROLE.NON_ADMIN);
      expandCard({ flagged: false });

      expect(screen.queryByText('Resolve')).not.toBeInTheDocument();
    });
  });
});
