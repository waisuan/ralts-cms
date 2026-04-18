'use client';

import { useState } from 'react';
import { Maintenance } from '../types/maintenance';
import { MaintenanceService } from '../services/maintenanceService';
import { AttachmentService } from '../services/attachmentService';
import { handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import LoadingOverlay from './LoadingOverlay';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

interface DeleteMaintenanceConfirmProps {
  machineSerialNumber: string;
  record: Maintenance | null;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function DeleteMaintenanceConfirm({
  machineSerialNumber,
  record,
  isOpen,
  onClose,
  onSuccess,
}: DeleteMaintenanceConfirmProps) {
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  useLockBodyScroll(isOpen && record !== null);

  if (!isOpen || !record) return null;

  const handleDelete = async () => {
    try {
      setIsDeleting(true);
      setDeleteError(null);

      await MaintenanceService.deleteMaintenance(machineSerialNumber, record.work_order_number);

      if (record.attachment) {
        try {
          await AttachmentService.deleteMaintenanceAttachment(
            machineSerialNumber, record.work_order_number, record.attachment
          );
        } catch (attachmentError) {
          if (isAuthError(attachmentError)) {
            alert('Your session has expired. Please log in again.');
            window.location.href = '/login';
            return;
          }
          console.warn('Record deleted but attachment cleanup failed — orphan file may remain');
        }
      }

      onSuccess();
      handleClose();
    } catch (error) {
      if (isAuthError(error)) {
        alert('Your session has expired. Please log in again.');
        window.location.href = '/login';
        return;
      }
      const errorMessage = handleApiError(error);
      setDeleteError(`Failed to delete maintenance record: ${errorMessage}`);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleClose = () => {
    setDeleteError(null);
    onClose();
  };

  return (
    <>
      <LoadingOverlay isVisible={isDeleting} message="Deleting maintenance record..." />
      <div
        className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
        onClick={handleClose}
      >
        <div className="flex min-h-full items-center justify-center p-4">
          <div
            className="relative bg-white rounded-lg shadow-xl max-w-md w-full"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="p-6">
              <div className="flex items-center mb-4">
                <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100">
                  <svg className="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                  </svg>
                </div>
              </div>
              <div className="text-center">
                <h3 className="text-lg font-medium text-gray-900 mb-2">Delete Maintenance Record</h3>
                <p className="text-sm text-gray-500 mb-2">Are you sure you want to delete maintenance record <strong>{record.work_order_number}</strong>?</p>
                <p className="text-sm text-gray-500 mb-6">This action cannot be undone. All data associated with this maintenance record will be permanently removed.</p>
              </div>
              {deleteError && (
                <div className="mt-4 mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
                  <div className="flex">
                    <svg className="h-5 w-5 text-red-400 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" /></svg>
                    <p className="ml-3 text-sm font-medium">{deleteError}</p>
                  </div>
                </div>
              )}
              <div className="flex space-x-3 mt-6">
                <button type="button" onClick={handleClose} className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                <button type="button" onClick={handleDelete} disabled={isDeleting} className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:bg-red-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                  {isDeleting ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Deleting...</>) : 'Delete Record'}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
