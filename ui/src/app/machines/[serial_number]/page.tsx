'use client';

import { notFound, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import MachineModal from '@/components/MachineModal';
import { mockMachines } from '@/data/mockMachines';
import { mockMaintenanceRecords } from '@/data/mockMaintenance';
import { Machine } from '@/types/machine';

interface MachinePageProps {
  params: Promise<{
    serial_number: string;
  }>;
}

function MachineContent({ machine }: { machine: Machine }) {
  const router = useRouter();
  const [machines, setMachines] = useState(mockMachines);
  const [isMachineModalOpen, setIsMachineModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'add' | 'edit'>('edit');
  const [machineToEdit, setMachineToEdit] = useState<Machine | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [machineToDelete, setMachineToDelete] = useState<Machine | null>(null);

  // Count maintenance records for this machine
  const maintenanceCount = mockMaintenanceRecords.filter(
    (record) => record.machine_serial_number === machine.serial_number
  ).length;

  const handleBack = () => {
    router.push('/');
  };

  const handleEdit = () => {
    setMachineToEdit(machine);
    setModalMode('edit');
    setIsMachineModalOpen(true);
  };

  const handleDelete = () => {
    setMachineToDelete(machine);
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = () => {
    if (machineToDelete) {
      // Remove the machine from the list
      const updatedMachines = machines.filter(
        (m) => m.serial_number !== machineToDelete.serial_number
      );
      setMachines(updatedMachines);

      // Navigate back to home page
      router.push('/');
    }
    setShowDeleteConfirm(false);
  };

  const handleCancelDelete = () => {
    setMachineToDelete(null);
    setShowDeleteConfirm(false);
  };

  const handleCloseMachineModal = () => {
    setMachineToEdit(null);
    setIsMachineModalOpen(false);
  };

  const handleMachineSubmit = (
    updatedMachine: Machine | Omit<Machine, 'created_at' | 'updated_at'>
  ) => {
    // Update the machine in the list
    const updatedMachines = machines.map((m) =>
      m.serial_number === updatedMachine.serial_number ? (updatedMachine as Machine) : m
    );
    setMachines(updatedMachines);

    // Update the current machine state
    setMachineToEdit(null);
    setIsMachineModalOpen(false);

    // Refresh the page to show updated data
    window.location.reload();
  };

  return (
    <>
      <MaintenanceHistory
        machine={machine}
        onBack={handleBack}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />

      {/* Machine Modal (Edit) */}
      <MachineModal
        isOpen={isMachineModalOpen}
        mode={modalMode}
        machine={machineToEdit}
        onClose={handleCloseMachineModal}
        onSubmit={handleMachineSubmit}
      />

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && machineToDelete && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50" onClick={handleCancelDelete} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex items-center mb-4">
                  <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100">
                    <svg
                      className="h-6 w-6 text-red-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </div>
                </div>
                <div>
                  <h3 className="text-lg font-medium text-gray-900 mb-2 text-center">
                    Delete Machine
                  </h3>
                  <p className="text-sm text-gray-500 mb-2 text-center">
                    Are you sure you want to delete machine{' '}
                    <strong>{machineToDelete.serial_number}</strong>?
                  </p>
                  <p className="text-sm text-gray-500 mb-2 text-center">
                    This action cannot be undone. All data associated with this machine will be
                    permanently removed.
                  </p>
                  {maintenanceCount > 0 && (
                    <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3 mb-4">
                      <div>
                        <h4 className="text-sm font-bold text-red-800 mb-1 text-center">Warning</h4>
                        <p className="text-sm text-red-800">
                          This machine has {maintenanceCount} maintenance record
                          {maintenanceCount !== 1 ? 's' : ''} that will also be deleted.
                        </p>
                      </div>
                    </div>
                  )}
                </div>
                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={handleCancelDelete}
                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleConfirmDelete}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
                  >
                    Delete Machine
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

export default function MachinePage({ params }: MachinePageProps) {
  const [machine, setMachine] = useState<Machine | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadMachine() {
      try {
        // Await the params in Next.js 15
        const resolvedParams = await params;

        // Decode the serial number from the URL
        const serialNumber = decodeURIComponent(resolvedParams.serial_number);

        // Find the machine by serial number
        const foundMachine = mockMachines.find((m) => m.serial_number === serialNumber);

        if (!foundMachine) {
          notFound();
          return;
        }

        setMachine(foundMachine);
      } catch (error) {
        console.error('Error loading machine:', error);
        notFound();
      } finally {
        setIsLoading(false);
      }
    }

    loadMachine();
  }, [params]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading machine information...</p>
        </div>
      </div>
    );
  }

  if (!machine) {
    notFound();
    return null;
  }

  return <MachineContent machine={machine} />;
}
