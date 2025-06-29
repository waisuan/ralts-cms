import { render, screen, fireEvent } from '@testing-library/react';
import RecordCard from './RecordCard';
import { Machine } from '../types/machine';

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
    ppm_status: '',
    tnc_date: '2024-07-01',
    ppm_date: '',
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
  };

  it('renders machine model and serial number', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.getByText(/TestModel/)).toBeInTheDocument();
    expect(screen.getByText(/SN-TEST/)).toBeInTheDocument();
  });

  it('displays district before state in subtitle', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    const subtitle = screen.getByText((content, element) => {
      return element?.textContent === 'TestBrand · TestDistrict, CA';
    });
    expect(subtitle).toBeInTheDocument();
  });

  it('displays account type in dedicated row', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.getByText('Account Type:')).toBeInTheDocument();
    expect(screen.getByText('Premium')).toBeInTheDocument();
  });

  it('displays reported_by in the dates section', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.getByText(/Reported By: Test Reporter/)).toBeInTheDocument();
  });

  it('shows attachment download icon when attachment exists', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    const downloadButton = screen.getByTitle('Download attachment');
    expect(downloadButton).toBeInTheDocument();
  });

  it('does not show attachment download icon when no attachment', () => {
    const machineWithoutAttachment = { ...baseMachine, attachment: '' };
    render(
      <RecordCard
        machine={machineWithoutAttachment}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.queryByTitle('Download attachment')).not.toBeInTheDocument();
  });

  it('shows notes icon when notes exist', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    const notesButton = screen.getByTitle('View additional notes');
    expect(notesButton).toBeInTheDocument();
  });

  it('does not show notes icon when no notes', () => {
    const machineWithoutNotes = { ...baseMachine, additional_notes: '' };
    render(
      <RecordCard
        machine={machineWithoutNotes}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.queryByTitle('View additional notes')).not.toBeInTheDocument();
  });

  it('opens notes modal when notes icon is clicked', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );

    const notesButton = screen.getByTitle('View additional notes');
    fireEvent.click(notesButton);

    expect(screen.getByText('Additional Notes')).toBeInTheDocument();
    expect(screen.getByText('Test notes for the machine')).toBeInTheDocument();
  });

  it('closes notes modal when close button is clicked', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );

    // Open modal
    const notesButton = screen.getByTitle('View additional notes');
    fireEvent.click(notesButton);

    // Close modal
    const closeButton = screen.getByText('Close');
    fireEvent.click(closeButton);

    expect(screen.queryByText('Additional Notes')).not.toBeInTheDocument();
  });

  it('closes notes modal when clicking backdrop', () => {
    const { container } = render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );

    // Open modal
    const notesButton = screen.getByTitle('View additional notes');
    fireEvent.click(notesButton);

    // Click backdrop (the overlay div)
    const backdrop = container.querySelector('.fixed.inset-0');
    if (backdrop) {
      fireEvent.click(backdrop);
    }

    expect(screen.queryByText('Additional Notes')).not.toBeInTheDocument();
  });

  it('displays status field correctly', () => {
    render(
      <RecordCard machine={baseMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.getByText('Status:')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('shows "Not specified" for empty reported_by field', () => {
    const machineWithoutReporter = { ...baseMachine, reported_by: '' };
    render(
      <RecordCard
        machine={machineWithoutReporter}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.getByText(/Reported By: Not specified/)).toBeInTheDocument();
  });

  it('shows "Not specified" for empty status field', () => {
    const machineWithoutStatus = { ...baseMachine, status: '' };
    render(
      <RecordCard
        machine={machineWithoutStatus}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.getByText('Status:')).toBeInTheDocument();
    expect(screen.getByText('Not specified')).toBeInTheDocument();
  });

  it('shows "Not specified" for empty account type field', () => {
    const machineWithoutAccountType = { ...baseMachine, account_type: '' };
    render(
      <RecordCard
        machine={machineWithoutAccountType}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.getByText('Account Type:')).toBeInTheDocument();
    expect(screen.getByText('Not specified')).toBeInTheDocument();
  });

  it('shows Overdue badge if ppm_date is in the past', () => {
    const overdueMachine = { ...baseMachine, ppm_date: '2000-01-01' };
    render(
      <RecordCard
        machine={overdueMachine}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.getByText('Overdue')).toBeInTheDocument();
  });

  it('shows Due badge if ppm_date is today', () => {
    const today = new Date();
    const todayStr = today.toISOString().split('T')[0];
    const dueMachine = { ...baseMachine, ppm_date: todayStr };
    render(
      <RecordCard machine={dueMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.getByText('Due')).toBeInTheDocument();
  });

  it('shows Due Soon badge if ppm_date is within 7 days', () => {
    const soon = new Date();
    soon.setDate(soon.getDate() + 3);
    const soonStr = soon.toISOString().split('T')[0];
    const dueSoonMachine = { ...baseMachine, ppm_date: soonStr };
    render(
      <RecordCard
        machine={dueSoonMachine}
        onView={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    );
    expect(screen.getByText('Due Soon')).toBeInTheDocument();
  });

  it('shows no badge if ppm_date is more than 7 days in the future', () => {
    const future = new Date();
    future.setDate(future.getDate() + 30);
    const futureStr = future.toISOString().split('T')[0];
    const futureMachine = { ...baseMachine, ppm_date: futureStr };
    render(
      <RecordCard machine={futureMachine} onView={() => {}} onEdit={() => {}} onDelete={() => {}} />
    );
    expect(screen.queryByText('Overdue')).not.toBeInTheDocument();
    expect(screen.queryByText('Due')).not.toBeInTheDocument();
    expect(screen.queryByText('Due Soon')).not.toBeInTheDocument();
  });
});
