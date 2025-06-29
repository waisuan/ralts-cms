'use client';

import { useState, useMemo, useEffect } from 'react';
import RecordCard from './RecordCard';
import { mockMachines } from '../data/mockMachines';

interface RecordsListProps {
  searchQuery: string;
}

const ITEMS_PER_PAGE = 3; // Show 3 machines initially, then load more

export default function RecordsList({ searchQuery }: RecordsListProps) {
  const [machines, setMachines] = useState(mockMachines);
  const [displayedCount, setDisplayedCount] = useState(ITEMS_PER_PAGE);

  // Filter machines based on search query
  const filteredMachines = useMemo(() => {
    if (!searchQuery.trim()) {
      return machines;
    }
    const query = searchQuery.toLowerCase();
    return machines.filter(
      (machine) =>
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

  // Get machines to display (limited by displayedCount)
  const displayedMachines = filteredMachines.slice(0, displayedCount);
  const hasMoreMachines = displayedCount < filteredMachines.length;

  // Reset displayed count when search changes
  useEffect(() => {
    setDisplayedCount(ITEMS_PER_PAGE);
  }, [searchQuery]);

  const handleLoadMore = () => {
    setDisplayedCount((prev) => Math.min(prev + ITEMS_PER_PAGE, filteredMachines.length));
  };

  const handleView = (serial_number: string) => {
    console.log('View machine:', serial_number);
  };

  const handleEdit = (serial_number: string) => {
    console.log('Edit machine:', serial_number);
  };

  const handleDelete = (serial_number: string) => {
    setMachines(machines.filter((machine) => machine.serial_number !== serial_number));
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">All Machines</h2>
          <p className="text-sm text-gray-500 mt-1">
            Showing {displayedMachines.length} of {filteredMachines.length} machine
            {filteredMachines.length !== 1 ? 's' : ''}
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
        {displayedMachines.map((machine) => (
          <RecordCard
            key={machine.serial_number}
            machine={machine}
            onView={handleView}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        ))}
      </div>

      {/* Load More Button */}
      {hasMoreMachines && (
        <div className="flex justify-center pt-4">
          <button
            onClick={handleLoadMore}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-6 py-3 rounded-lg font-medium transition-colors"
          >
            Load More ({filteredMachines.length - displayedCount} remaining)
          </button>
        </div>
      )}

      {filteredMachines.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-400 text-6xl mb-4">📄</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {searchQuery ? 'No machines found' : 'No machines found'}
          </h3>
          <p className="text-gray-500">
            {searchQuery
              ? `No machines match "${searchQuery}". Try a different search term.`
              : 'Get started by creating your first machine.'}
          </p>
        </div>
      )}
    </div>
  );
}
