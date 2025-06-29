import { render, screen, fireEvent } from '@testing-library/react';
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

  it('renders nothing when there are no overdue or due machines', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 0,
      overdueMachines: [],
      dueMachines: [],
    };

    const { container } = render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    expect(container.firstChild).toBeNull();
  });

  it('renders overdue alert when there are overdue machines', () => {
    const stats: OverdueStats = {
      overdueCount: 3,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 3,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    expect(screen.getByText('3 machines are overdue for PPM maintenance')).toBeInTheDocument();
    expect(
      screen.getByText('Immediate attention required to avoid compliance issues')
    ).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
  });

  it('renders singular overdue alert when there is one overdue machine', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    expect(screen.getByText('1 machine is overdue for PPM maintenance')).toBeInTheDocument();
  });

  it('renders due alert when there are due machines', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 2,
      dueSoonCount: 0,
      totalCriticalCount: 2,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    expect(screen.getByText('2 machines are due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.getByText('Schedule maintenance to avoid becoming overdue')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();
  });

  it('renders singular due alert when there is one due machine', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
  });

  it('renders both overdue and due alerts when both exist', () => {
    const stats: OverdueStats = {
      overdueCount: 2,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 3,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    // Should show both alerts
    expect(screen.getByText('2 machines are overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();
  });

  it('calls onShowOverdue when View Overdue button is clicked', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const viewOverdueButton = screen.getByText('View Overdue');
    fireEvent.click(viewOverdueButton);

    expect(mockOnShowOverdue).toHaveBeenCalledTimes(1);
    expect(mockOnShowDue).not.toHaveBeenCalled();
  });

  it('calls onShowDue when View Due button is clicked', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const viewDueButton = screen.getByText('View Due');
    fireEvent.click(viewDueButton);

    expect(mockOnShowDue).toHaveBeenCalledTimes(1);
    expect(mockOnShowOverdue).not.toHaveBeenCalled();
  });

  it('shows dismiss button for overdue alert when onDismissOverdue is provided', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const dismissButton = screen.getByTitle('Dismiss alert');
    expect(dismissButton).toBeInTheDocument();
  });

  it('shows dismiss button for due alert when onDismissDue is provided', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const dismissButton = screen.getByTitle('Dismiss alert');
    expect(dismissButton).toBeInTheDocument();
  });

  it('calls onDismissOverdue when overdue dismiss button is clicked', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const dismissButton = screen.getByTitle('Dismiss alert');
    fireEvent.click(dismissButton);

    expect(mockOnDismissOverdue).toHaveBeenCalledTimes(1);
    expect(mockOnDismissDue).not.toHaveBeenCalled();
  });

  it('calls onDismissDue when due dismiss button is clicked', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    const dismissButton = screen.getByTitle('Dismiss alert');
    fireEvent.click(dismissButton);

    expect(mockOnDismissDue).toHaveBeenCalledTimes(1);
    expect(mockOnDismissOverdue).not.toHaveBeenCalled();
  });

  it('does not render overdue alert when isOverdueDismissed is true', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} isOverdueDismissed={true} />
    );

    expect(screen.queryByText('1 machine is overdue for PPM maintenance')).not.toBeInTheDocument();
  });

  it('does not render due alert when isDueDismissed is true', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert stats={stats} {...defaultProps} isDueDismissed={true} />
    );

    expect(screen.queryByText('1 machine is due for PPM maintenance today')).not.toBeInTheDocument();
  });

  it('renders nothing when both alerts are dismissed', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 2,
      overdueMachines: [],
      dueMachines: [],
    };

    const { container } = render(
      <OverdueAlert stats={stats} {...defaultProps} isOverdueDismissed={true} isDueDismissed={true} />
    );

    expect(container.firstChild).toBeNull();
  });

  it('has correct styling for overdue alert', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 0,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    const { container } = render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    // Find the red alert container
    const overdueAlert = container.querySelector('.bg-red-50.border-red-400');
    expect(overdueAlert).toBeInTheDocument();
    expect(overdueAlert).toHaveClass('bg-red-50', 'border-l-4', 'border-red-400');
  });

  it('has correct styling for due alert', () => {
    const stats: OverdueStats = {
      overdueCount: 0,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 1,
      overdueMachines: [],
      dueMachines: [],
    };

    const { container } = render(
      <OverdueAlert stats={stats} {...defaultProps} />
    );

    // Find the orange alert container
    const dueAlert = container.querySelector('.bg-orange-50.border-orange-400');
    expect(dueAlert).toBeInTheDocument();
    expect(dueAlert).toHaveClass('bg-orange-50', 'border-l-4', 'border-orange-400');
  });

  it('works without dismiss callbacks (backward compatibility)', () => {
    const stats: OverdueStats = {
      overdueCount: 1,
      dueCount: 1,
      dueSoonCount: 0,
      totalCriticalCount: 2,
      overdueMachines: [],
      dueMachines: [],
    };

    render(
      <OverdueAlert 
        stats={stats} 
        onShowOverdue={mockOnShowOverdue} 
        onShowDue={mockOnShowDue} 
      />
    );

    // Should render alerts but no dismiss buttons
    expect(screen.getByText('1 machine is overdue for PPM maintenance')).toBeInTheDocument();
    expect(screen.getByText('1 machine is due for PPM maintenance today')).toBeInTheDocument();
    expect(screen.queryByTitle('Dismiss alert')).not.toBeInTheDocument();
  });
});
