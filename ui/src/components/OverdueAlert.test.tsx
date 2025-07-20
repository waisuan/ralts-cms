import { render, screen, fireEvent } from '@testing-library/react';
import OverdueAlert from './OverdueAlert';

describe('OverdueAlert', () => {
  const mockOnShowOverdue = jest.fn();
  const mockOnShowDue = jest.fn();
  const mockOnDismissOverdue = jest.fn();
  const mockOnDismissDue = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should not render when no alerts to show', () => {
    render(
      <OverdueAlert
        overdueCount={0}
        dueCount={0}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    expect(screen.queryByText(/overdue/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/due/i)).not.toBeInTheDocument();
  });

  it('should render overdue alert when overdue machines exist', () => {
    render(
      <OverdueAlert
        overdueCount={3}
        dueCount={0}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    expect(screen.getByText(/3 machines are overdue/i)).toBeInTheDocument();
    expect(screen.getByText('View Overdue')).toBeInTheDocument();
  });

  it('should render due alert when due machines exist', () => {
    render(
      <OverdueAlert
        overdueCount={0}
        dueCount={2}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    expect(screen.getByText(/2 machines are due/i)).toBeInTheDocument();
    expect(screen.getByText('View Due')).toBeInTheDocument();
  });

  it('should render both alerts when both overdue and due machines exist', () => {
    render(
      <OverdueAlert
        overdueCount={1}
        dueCount={1}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    expect(screen.getByText(/1 machine is overdue/i)).toBeInTheDocument();
    expect(screen.getByText(/1 machine is due/i)).toBeInTheDocument();
  });

  it('should call onShowOverdue when View Overdue button is clicked', () => {
    render(
      <OverdueAlert
        overdueCount={1}
        dueCount={0}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    fireEvent.click(screen.getByText('View Overdue'));
    expect(mockOnShowOverdue).toHaveBeenCalledTimes(1);
  });

  it('should call onShowDue when View Due button is clicked', () => {
    render(
      <OverdueAlert
        overdueCount={0}
        dueCount={1}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
      />
    );

    fireEvent.click(screen.getByText('View Due'));
    expect(mockOnShowDue).toHaveBeenCalledTimes(1);
  });

  it('should not render overdue alert when dismissed', () => {
    render(
      <OverdueAlert
        overdueCount={1}
        dueCount={0}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
        onDismissOverdue={mockOnDismissOverdue}
        isOverdueDismissed={true}
      />
    );

    expect(screen.queryByText(/overdue/i)).not.toBeInTheDocument();
  });

  it('should not render due alert when dismissed', () => {
    render(
      <OverdueAlert
        overdueCount={0}
        dueCount={1}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
        onDismissDue={mockOnDismissDue}
        isDueDismissed={true}
      />
    );

    expect(screen.queryByText(/due/i)).not.toBeInTheDocument();
  });

  it('should call dismiss handlers when dismiss buttons are clicked', () => {
    render(
      <OverdueAlert
        overdueCount={1}
        dueCount={1}
        onShowOverdue={mockOnShowOverdue}
        onShowDue={mockOnShowDue}
        onDismissOverdue={mockOnDismissOverdue}
        onDismissDue={mockOnDismissDue}
      />
    );

    const dismissButtons = screen.getAllByTitle('Dismiss alert');
    fireEvent.click(dismissButtons[0]); // Dismiss overdue
    fireEvent.click(dismissButtons[1]); // Dismiss due

    expect(mockOnDismissOverdue).toHaveBeenCalledTimes(1);
    expect(mockOnDismissDue).toHaveBeenCalledTimes(1);
  });
});
