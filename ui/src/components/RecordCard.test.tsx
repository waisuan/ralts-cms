import { render, screen } from '@testing-library/react';
import RecordCard from './RecordCard';
import { Machine } from '../types/machine';

describe('RecordCard', () => {
  const baseMachine: Machine = {
    serial_number: 'SN-TEST',
    customer: 'Test Customer',
    state: 'CA',
    account_type: 'Premium',
    model: 'TestModel',
    status: '',
    brand: 'TestBrand',
    district: 'TestDistrict',
    person_in_charge: 'Test Person',
    reported_by: 'Test Reporter',
    additional_notes: '',
    attachment: '',
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
