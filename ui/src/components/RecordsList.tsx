'use client';

import { useState, useMemo, useEffect } from 'react';
import RecordCard from './RecordCard';
import { mockMachines } from '../data/mockMachines';
import { SearchOptions } from './SearchBar';
import { isDateProperty } from '@/utils/constants';
import { getPPMStatusLabel } from '@/utils/ppmUtils';

interface RecordsListProps {
  searchOptions: SearchOptions;
}

const ITEMS_PER_PAGE = 3; // Show 3 machines initially, then load more

export default function RecordsList({ searchOptions }: RecordsListProps) {
  const [machines, setMachines] = useState(mockMachines);
  const [displayedCount, setDisplayedCount] = useState(ITEMS_PER_PAGE);

  // Filter machines based on search options
  const filteredMachines = useMemo(() => {
    if (!searchOptions.query.trim()) {
      return machines;
    }

    const query = searchOptions.query.toLowerCase();
    const { property } = searchOptions;

    return machines.filter((machine) => {
      // Handle date properties differently
      if (isDateProperty(property)) {
        const fieldValue = machine[property as keyof typeof machine];
        if (!fieldValue) return false;

        // Convert both dates to YYYY-MM-DD format for comparison
        const machineDate = fieldValue.toString().split('T')[0]; // Extract date part from ISO string
        const searchDate = searchOptions.query; // Already in YYYY-MM-DD format from date input

        return machineDate === searchDate;
      } else if (property === 'ppm_status') {
        // Handle PPM status search by calculating status from ppm_date
        const calculatedStatus = getPPMStatusLabel(machine.ppm_date);
        if (!calculatedStatus) return false;

        return calculatedStatus.toLowerCase().includes(query);
      } else {
        // Search in specific text property
        const fieldValue = machine[property as keyof typeof machine];
        return fieldValue && fieldValue.toString().toLowerCase().includes(query);
      }
    });
  }, [machines, searchOptions]);

  // Get machines to display (limited by displayedCount)
  const displayedMachines = filteredMachines.slice(0, displayedCount);
  const hasMoreMachines = displayedCount < filteredMachines.length;

  // Reset displayed count when search changes
  useEffect(() => {
    setDisplayedCount(ITEMS_PER_PAGE);
  }, [searchOptions]);

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
            {searchOptions.query && filteredMachines.length !== machines.length && (
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
            {searchOptions.query ? 'No machines found' : 'No machines found'}
          </h3>
          <p className="text-gray-500">
            {searchOptions.query
              ? `No machines match "${searchOptions.query}". Try a different search term.`
              : 'Get started by creating your first machine.'}
          </p>
        </div>
      )}
    </div>
  );
}
