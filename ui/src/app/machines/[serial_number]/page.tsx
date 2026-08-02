'use client';

import { notFound, useRouter } from 'next/navigation';
import { Suspense, useEffect, useState } from 'react';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import MachineModal from '@/components/MachineModal';
import LoadingOverlay from '@/components/LoadingOverlay';
import RecordsListDeleteModal from '@/components/recordsList/RecordsListDeleteModal';
import { Machine } from '@/types/machine';
import { MachineService } from '@/services/machineService';
import { handleApiError } from '@/utils/api';
import { isAuthError } from '@/utils/auth';
import { machineDetailHref } from '@/utils/machineRoutes';

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
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isUpdating, setIsUpdating] = useState(false);


  const handleEdit = () => {
    setMachineToEdit(machine);
    setModalMode('edit');
    setIsMachineModalOpen(true);
  };

  const handleDelete = () => {
    setMachineToDelete(machine);
    setDeleteError(null);
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = async () => {
    if (!machineToDelete) return;
    try {
      setIsDeletingMachine(true);
      setDeleteError(null);
      await MachineService.deleteMachine(machineToDelete.serial_number);
      setShowDeleteConfirm(false);
      router.push('/');
    } catch (error) {
      console.error('Failed to delete machine:', error);
      setDeleteError(handleApiError(error).message);
    } finally {
      setIsDeletingMachine(false);
    }
  };

  const handleCancelDelete = () => {
    setMachineToDelete(null);
    setDeleteError(null);
    setShowDeleteConfirm(false);
  };

  const handleCloseMachineModal = () => {
    setMachineToEdit(null);
    setIsMachineModalOpen(false);
  };

  const handleMachineSubmit = async (
    updatedMachine: Machine | Omit<Machine, 'created_at' | 'updated_at' | 'updated_by'>
  ) => {
    if (modalMode !== 'edit' || !machineToEdit) {
      const err = new Error('Cannot update machine: missing edit context.');
      console.error(err.message);
      throw err;
    }
    try {
      await MachineService.updateMachine(machineToEdit.serial_number, updatedMachine);
      setIsUpdating(true);
      window.location.href = machineDetailHref(machine.serial_number);
    } catch (error) {
      console.error('Failed to update machine:', error);
      throw error;
    }
  };

  return (
    <>
      <LoadingOverlay isVisible={isUpdating} message="Loading machine details..." />

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

      {showDeleteConfirm && machineToDelete && (
        <RecordsListDeleteModal
          machine={machineToDelete}
          deleteError={deleteError}
          isDeleting={isDeletingMachine}
          onCancel={handleCancelDelete}
          onConfirm={handleConfirmDelete}
        />
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
