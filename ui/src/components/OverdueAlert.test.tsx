import { render, screen } from '@testing-library/react';
import OverdueAlert from './OverdueAlert';
import { OverdueStats } from '../hooks/useOverdueStats';

describe('OverdueAlert', () => {
  const mockOnShowOverdue = jest.fn();
  const mockOnShowDue = jest.fn();
  const mockOnDismissOverdue = jest.fn();
  const mockOnDismissDue = jest.fn();

  const defaultProps = {
    onShowOverdue: mockOnShowOverdue,
    onShowDue: mockOnShowDue,
    onDismissOverdue: mockOnDismissOverdue,
    onDismissDue: mockOnDismissDue,
    isOverdueDismissed: false,
    isDueDismissed: false,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders overdue and due alerts with correct content and actions', () => {
    const stats: OverdueStats = {
      overdueCount: 2,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 3,
      overdueMachines: [],
      dueMachines: [],
    };

    render(<OverdueAlert stats={stats} {...defaultProps} />);

    // Check that both alerts are rendered
    expect(screen.getByText('2 machines are overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();

    // Check that action buttons are present
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();

    // Check that dismiss buttons are present
    expect(screen.getAllByTitle('Dismiss alert')).toHaveLength(2);

    // Check that descriptive text is present
    expect(
      screen.getByText('Immediate attention required to avoid compliance issues')
    ).toBeInTheDocument();
    expect(screen.getByText('Schedule maintenance to avoid becoming overdue')).toBeInTheDocument();
  });
});
