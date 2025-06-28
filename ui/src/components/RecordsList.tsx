'use client';

import { useState, useMemo } from 'react';
import RecordCard from './RecordCard';
import { Machine } from '../types/machine';

interface RecordsListProps {
  searchQuery: string;
}

// Helper function to get dates for status examples
const getDateForStatus = (status: 'overdue' | 'due' | 'due_soon' | 'future') => {
  const today = new Date();
  const todayStr = today.toISOString().split('T')[0]; // YYYY-MM-DD format
  
  switch (status) {
    case 'overdue':
      // 5 days ago
      const overdue = new Date(today);
      overdue.setDate(today.getDate() - 5);
      return overdue.toISOString().split('T')[0];
    case 'due':
      // Today
      return todayStr;
    case 'due_soon':
      // 3 days from now
      const dueSoon = new Date(today);
      dueSoon.setDate(today.getDate() + 3);
      return dueSoon.toISOString().split('T')[0];
    case 'future':
      // 30 days from now
      const future = new Date(today);
      future.setDate(today.getDate() + 30);
      return future.toISOString().split('T')[0];
  }
};

// Placeholder Machine data with different status examples
const placeholderMachines: Machine[] = [
  {
    serial_number: 'SN-001',
    customer: 'Acme Corp',
    state: 'CA',
    account_type: 'Premium',
    model: 'X100',
    status: '', // Will be calculated from ppm_date
    brand: 'BrandA',
    district: 'North',
    person_in_charge: 'Alice Johnson',
    reported_by: 'Bob Smith',
    additional_notes: 'Needs inspection',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-01',
    ppm_date: getDateForStatus('overdue'), // Will show "Overdue" (Red)
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
  },
  {
    serial_number: 'SN-002',
    customer: 'Beta LLC',
    state: 'NY',
    account_type: 'Standard',
    model: 'Y200',
    status: '',
    brand: 'BrandB',
    district: 'East',
    person_in_charge: 'Charlie Brown',
    reported_by: 'Dana White',
    additional_notes: 'Regular maintenance',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-10',
    ppm_date: getDateForStatus('due'), // Will show "Due" (Orange)
    created_at: '2024-02-01',
    updated_at: '2024-06-10',
  },
  {
    serial_number: 'SN-003',
    customer: 'Gamma Inc',
    state: 'TX',
    account_type: 'Basic',
    model: 'Z300',
    status: '',
    brand: 'BrandC',
    district: 'South',
    person_in_charge: 'Eve Wilson',
    reported_by: 'Frank Miller',
    additional_notes: 'Scheduled maintenance',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-06-01',
    ppm_date: getDateForStatus('due_soon'), // Will show "Due Soon" (Yellow)
    created_at: '2024-03-01',
    updated_at: '2024-06-15',
  },
  {
    serial_number: 'SN-004',
    customer: 'Delta Co',
    state: 'FL',
    account_type: 'Premium',
    model: 'A400',
    status: '',
    brand: 'BrandD',
    district: 'West',
    person_in_charge: 'Grace Lee',
    reported_by: 'Heidi Klum',
    additional_notes: 'Recently serviced',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-20',
    ppm_date: getDateForStatus('future'), // Will show no status badge
    created_at: '2024-04-01',
    updated_at: '2024-06-20',
  },
  {
    serial_number: 'SN-005',
    customer: 'Echo Systems',
    state: 'WA',
    account_type: 'Enterprise',
    model: 'B500',
    status: '',
    brand: 'BrandE',
    district: 'Northwest',
    person_in_charge: 'Ian Cooper',
    reported_by: 'Jane Doe',
    additional_notes: 'High priority equipment',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-08-01',
    ppm_date: getDateForStatus('overdue'), // Another "Overdue" example
    created_at: '2024-05-01',
    updated_at: '2024-06-25',
  },
  {
    serial_number: 'SN-006',
    customer: 'Foxtrot Industries',
    state: 'CO',
    account_type: 'Standard',
    model: 'C600',
    status: '',
    brand: 'BrandF',
    district: 'Mountain',
    person_in_charge: 'Kate Martinez',
    reported_by: 'Liam O\'Connor',
    additional_notes: 'Remote location',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-15',
    ppm_date: getDateForStatus('due_soon'), // Another "Due Soon" example
    created_at: '2024-06-01',
    updated_at: '2024-06-28',
  },
];

export default function RecordsList({ searchQuery }: RecordsListProps) {
  const [machines, setMachines] = useState(placeholderMachines);

  // Filter machines based on search query
  const filteredMachines = useMemo(() => {
    if (!searchQuery.trim()) {
      return machines;
    }
    const query = searchQuery.toLowerCase();
    return machines.filter(machine =>
      machine.serial_number.toLowerCase().includes(query) ||
      machine.customer.toLowerCase().includes(query) ||
      machine.model.toLowerCase().includes(query) ||
      machine.brand.toLowerCase().includes(query) ||
      machine.state.toLowerCase().includes(query) ||
      machine.person_in_charge.toLowerCase().includes(query) ||
      machine.tnc_date.toLowerCase().includes(query) ||
      machine.ppm_date.toLowerCase().includes(query)
    );
  }, [machines, searchQuery]);

  const handleView = (serial_number: string) => {
    console.log('View machine:', serial_number);
  };

  const handleEdit = (serial_number: string) => {
    console.log('Edit machine:', serial_number);
  };

  const handleDelete = (serial_number: string) => {
    setMachines(machines.filter(machine => machine.serial_number !== serial_number));
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">All Machines</h2>
          <p className="text-sm text-gray-500 mt-1">
            {filteredMachines.length} machine{filteredMachines.length !== 1 ? 's' : ''} found
            {searchQuery && filteredMachines.length !== machines.length && (
              <span className="ml-1">(filtered from {machines.length} total)</span>
            )}
          </p>
        </div>
        <button className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors">
          Add New Machine
        </button>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredMachines.map((machine) => (
          <RecordCard
            key={machine.serial_number}
            machine={machine}
            onView={handleView}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        ))}
      </div>
      {filteredMachines.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-400 text-6xl mb-4">📄</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {searchQuery ? 'No machines found' : 'No machines found'}
          </h3>
          <p className="text-gray-500">
            {searchQuery
              ? `No machines match "${searchQuery}". Try a different search term.`
              : 'Get started by creating your first machine.'
            }
          </p>
        </div>
      )}
    </div>
  );
} 