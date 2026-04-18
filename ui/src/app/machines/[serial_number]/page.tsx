'use client';

import { useRouter } from 'next/navigation';
import { Suspense, useEffect, useState } from 'react';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import MachineModal from '@/components/MachineModal';
import LoadingOverlay from '@/components/LoadingOverlay';
import { Machine } from '@/types/machine';
import { MachineService } from '@/services/machineService';
import { handleApiError } from '@/utils/api';
import { isAuthError } from '@/utils/auth';
import { notFound } from 'next/navigation';

interface MachinePageProps {
  params: Promise<{
    serial_number: string;
  }>;
}

function MachineContent({ machine }: { machine: Machine }) {
  const router = useRouter();
  const [isMachineModalOpen, setIsMachineModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'add' | 'edit'>('edit');
  const [machineToEdit, setMachineToEdit] = useState<Machine | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [machineToDelete, setMachineToDelete] = useState<Machine | null>(null);
  const [isDeletingMachine, setIsDeletingMachine] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);


  const handleEdit = () => {
    setMachineToEdit(machine);
    setModalMode('edit');
    setIsMachineModalOpen(true);
  };

  const handleDelete = () => {
    setMachineToDelete(machine);
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = async () => {
    if (machineToDelete) {
      try {
        setIsDeletingMachine(true);
        await MachineService.deleteMachine(machineToDelete.serial_number);
        // Navigate back to home page
        router.push('/');
      } catch (error) {
        console.error('Failed to delete machine:', error);
        alert('Failed to delete machine. Please try again.');
        setIsDeletingMachine(false);
      }
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

      const handleMachineSubmit = async (
      updatedMachine: Machine | Omit<Machine, 'created_at' | 'updated_at'>
    ) => {
      try {
        if (modalMode === 'edit' && machineToEdit) {
          await MachineService.updateMachine(machineToEdit.serial_number, updatedMachine);
        }

        // Show loading indicator immediately after successful update
        setIsUpdating(true);

        // Force a hard page reload to ensure fresh data and prevent race conditions
        window.location.href = `/machines/${machine.serial_number}`;
      } catch (error) {
        console.error('Failed to update machine:', error);
        // Don't close modal or reload - let the modal handle the error
        throw error;
      }
    };

  return (
    <>
      {/* Loading Overlay - shows when updating or deleting machine data */}
      <LoadingOverlay 
        isVisible={isUpdating || isDeletingMachine} 
        message={
          isUpdating ? "Loading machine details..." :
          isDeletingMachine ? "Deleting machine..." :
          "Loading..."
        }
      />

      <Suspense
        fallback={<LoadingOverlay isVisible message="Loading maintenance..." />}
      >
        <MaintenanceHistory
          machine={machine}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </Suspense>

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
                </div>
                <div className="flex space-x-3 mt-6">
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
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadMachine() {
      try {
        // Await the params in Next.js 15
        const resolvedParams = await params;

        // Decode the serial number from the URL
        const serialNumber = decodeURIComponent(resolvedParams.serial_number);

        // Fetch the machine from the API
        const response = await MachineService.getMachine(serialNumber);
        
        if (response.data) {
          setMachine(response.data);
        } else {
          notFound();
          return;
        }
      } catch (error) {
        console.error('Error loading machine:', error);
        const apiError = handleApiError(error);
        if (apiError.status === 404) {
          notFound();
          return;
        }
        // Don't set error if we're redirecting due to auth error
        if (!isAuthError(error)) {
          setError(apiError.message);
        }
      } finally {
        setIsLoading(false);
      }
    }

    loadMachine();
  }, [params]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <LoadingOverlay 
          isVisible={true} 
          message="Loading machine details..." 
        />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-600 text-6xl mb-4">⚠️</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">Error Loading Machine</h3>
          <p className="text-gray-500 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Try Again
          </button>
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
