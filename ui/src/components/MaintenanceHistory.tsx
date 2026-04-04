'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Machine } from '../types/machine';
import { Maintenance, MaintenanceOrderType } from '../types/maintenance';
import { MaintenanceService, CreateMaintenanceRequest, UpdateMaintenanceRequest } from '../services/maintenanceService';
import { MachineService } from '../services/machineService';
import { AttachmentService } from '../services/attachmentService';
import { handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import { backendDateToHtmlDate } from '../utils/dateUtils';
import { useMaintenance } from '../hooks/useMaintenance';
import MaintenanceTable from './MaintenanceTable';
import FullPageLoader from './FullPageLoader';
import LoadingOverlay from './LoadingOverlay';

interface MaintenanceHistoryProps {
  machine: Machine;
  onBack: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

export default function MaintenanceHistory({
  machine,
  onBack,
  onEdit,
  onDelete,
}: MaintenanceHistoryProps) {
  // ---------- Data hook ----------
  const maint = useMaintenance({ serialNumber: machine.serial_number });

  // ---------- CSV export ----------
  const [csvExporting, setCsvExporting] = useState(false);

  // ---------- Add record modal ----------
  const [isAddRecordModalOpen, setIsAddRecordModalOpen] = useState(false);
  const [newRecordForm, setNewRecordForm] = useState({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    work_order_type: 'Preventive' as MaintenanceOrderType,
    custom_work_order_type: '',
    attachment: '',
  });
  const [newRecordErrors, setNewRecordErrors] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [createError, setCreateError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  // ---------- Edit record modal ----------
  const [isEditRecordModalOpen, setIsEditRecordModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Maintenance | null>(null);
  const [editRecordForm, setEditRecordForm] = useState({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    work_order_type: 'Preventive' as MaintenanceOrderType,
    custom_work_order_type: '',
    attachment: '',
  });
  const [editRecordErrors, setEditRecordErrors] = useState<Record<string, string>>({});
  const [editSelectedFile, setEditSelectedFile] = useState<File | null>(null);
  const [editIsUploading, setEditIsUploading] = useState(false);
  const [editUploadProgress, setEditUploadProgress] = useState(0);
  const [showEditCancelConfirm, setShowEditCancelConfirm] = useState(false);
  const [originalEditFormData, setOriginalEditFormData] = useState({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    work_order_type: 'Preventive' as MaintenanceOrderType,
    custom_work_order_type: '',
    attachment: '',
  });
  const [updateError, setUpdateError] = useState<string | null>(null);
  const [isUpdating, setIsUpdating] = useState(false);

  // ---------- Delete confirmation ----------
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [recordToDelete, setRecordToDelete] = useState<Maintenance | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // ---------- Upload interval refs (cleanup on unmount) ----------
  const uploadIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const editUploadIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    return () => {
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
      if (editUploadIntervalRef.current) clearInterval(editUploadIntervalRef.current);
    };
  }, []);

  // ---------- Navigation ----------
  const [isNavigatingBack, setIsNavigatingBack] = useState(false);

  // ---------- Handlers ----------
  const handleBackNavigation = () => {
    setIsNavigatingBack(true);
    onBack();
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatDateTime = (dateString: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Machine attachment download
  const handleDownloadMachineAttachment = async (attachment: string) => {
    if (!attachment) return;
    try {
      const blob = await AttachmentService.downloadMachineAttachment(
        machine.serial_number,
        attachment
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = attachment;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download machine attachment:', error);
      alert(`Failed to download attachment: ${attachment}`);
    }
  };

  // CSV export
  const handleExportMaintenanceCSV = async () => {
    setCsvExporting(true);
    try {
      await MachineService.exportMaintenanceCSV(machine.serial_number, maint.debouncedSearchQuery || undefined);
    } catch (err) {
      console.error('Maintenance CSV export failed:', err);
      alert('Failed to export maintenance CSV. Please try again.');
    } finally {
      setCsvExporting(false);
    }
  };

  // ======================== Add Record ========================

  const emptyForm = () => ({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    work_order_type: 'Preventive' as MaintenanceOrderType,
    custom_work_order_type: '',
    attachment: '',
  });

  const openAddRecordModal = () => {
    setIsAddRecordModalOpen(true);
    setNewRecordForm(emptyForm());
    setNewRecordErrors({});
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
    setCreateError(null);
  };

  const closeAddRecordModal = () => {
    setIsAddRecordModalOpen(false);
    setNewRecordForm(emptyForm());
    setNewRecordErrors({});
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
    setCreateError(null);
  };

  const handleNewRecordInputChange = (field: string, value: string) => {
    setNewRecordForm((prev) => ({ ...prev, [field]: value }));
    if (newRecordErrors[field]) {
      setNewRecordErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  const handleNewRecordFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setIsUploading(true);
      setUploadProgress(0);
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
      uploadIntervalRef.current = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 100) {
            if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
            uploadIntervalRef.current = null;
            setIsUploading(false);
            return 100;
          }
          return Math.min(prev + Math.random() * 25 + 10, 100);
        });
      }, 150);
      setNewRecordForm((prev) => ({ ...prev, attachment: file.name }));
    } else {
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
      uploadIntervalRef.current = null;
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setNewRecordForm((prev) => ({ ...prev, attachment: '' }));
    }
  };

  const validateNewRecordForm = () => {
    const errors: Record<string, string> = {};
    if (!newRecordForm.work_order_number.trim()) errors.work_order_number = 'Work order number is required';
    else if (newRecordForm.work_order_number.length > 50) errors.work_order_number = 'Work order number must not exceed 50 characters';
    if (!newRecordForm.work_order_date) errors.work_order_date = 'Work order date is required';
    if (!newRecordForm.action_taken.trim()) errors.action_taken = 'Action taken is required';
    else if (newRecordForm.action_taken.length > 500) errors.action_taken = 'Action taken must not exceed 500 characters';
    if (!newRecordForm.reported_by.trim()) errors.reported_by = 'Reported by is required';
    else if (newRecordForm.reported_by.length > 50) errors.reported_by = 'Reported by must not exceed 50 characters';
    if (newRecordForm.work_order_type === 'Other' && !newRecordForm.custom_work_order_type.trim())
      errors.custom_work_order_type = 'Custom maintenance type is required when "Other" is selected';
    else if (newRecordForm.work_order_type === 'Other' && newRecordForm.custom_work_order_type.length > 50)
      errors.custom_work_order_type = 'Custom maintenance type must not exceed 50 characters';

    const existing = maint.records?.find((r) => r.work_order_number === newRecordForm.work_order_number.trim());
    if (existing) errors.work_order_number = 'Work order number already exists';

    setNewRecordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleAddNewRecord = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateNewRecordForm()) return;

    try {
      setIsCreating(true);
      setCreateError(null);
      const createData: CreateMaintenanceRequest = {
        work_order_number: newRecordForm.work_order_number.trim(),
        work_order_date: newRecordForm.work_order_date,
        action_taken: newRecordForm.action_taken.trim(),
        reported_by: newRecordForm.reported_by.trim(),
        work_order_type: newRecordForm.work_order_type === 'Other' ? newRecordForm.custom_work_order_type.trim() : newRecordForm.work_order_type,
        attachment: newRecordForm.attachment || undefined,
      };

      await MaintenanceService.createMaintenance(machine.serial_number, createData);

      if (selectedFile) {
        try {
          await AttachmentService.uploadMaintenanceAttachment(
            machine.serial_number,
            newRecordForm.work_order_number.trim(),
            selectedFile
          );
        } catch (attachmentError) {
          if (isAuthError(attachmentError)) {
            alert('Your session has expired. Please log in again.');
            window.location.href = '/login';
            return;
          }
          const errorMessage = handleApiError(attachmentError);
          alert(`Maintenance record created but attachment upload failed: ${errorMessage}`);
        }
      }

      maint.refetch();
      closeAddRecordModal();
    } catch (error) {
      if (isAuthError(error)) {
        alert('Your session has expired. Please log in again.');
        window.location.href = '/login';
        return;
      }
      const errorMessage = handleApiError(error);
      setCreateError(`An error occurred while creating the maintenance record: ${errorMessage}`);
    } finally {
      setIsCreating(false);
    }
  };

  // ======================== Edit Record ========================

  const openEditRecordModal = (record: Maintenance) => {
    const isStandardType = ['Preventive', 'Corrective', 'Emergency', 'Inspection'].includes(record.work_order_type);
    const formData = {
      work_order_number: record.work_order_number,
      work_order_date: backendDateToHtmlDate(record.work_order_date),
      action_taken: record.action_taken,
      reported_by: record.reported_by,
      work_order_type: isStandardType ? (record.work_order_type as MaintenanceOrderType) : 'Other',
      custom_work_order_type: isStandardType ? '' : record.work_order_type,
      attachment: record.attachment || '',
    };
    setEditingRecord(record);
    setEditRecordForm(formData);
    setOriginalEditFormData(formData);
    setEditRecordErrors({});
    setEditSelectedFile(null);
    setEditUploadProgress(0);
    setEditIsUploading(false);
    setUpdateError(null);
    setIsEditRecordModalOpen(true);
  };

  const closeEditRecordModal = () => {
    setIsEditRecordModalOpen(false);
    setEditingRecord(null);
    setEditRecordForm(emptyForm());
    setOriginalEditFormData(emptyForm());
    setEditRecordErrors({});
    setEditSelectedFile(null);
    setEditUploadProgress(0);
    setEditIsUploading(false);
    setShowEditCancelConfirm(false);
    setUpdateError(null);
  };

  const handleEditCancelClick = () => {
    const hasChanges =
      Object.keys(editRecordForm).some(
        (key) => editRecordForm[key as keyof typeof editRecordForm] !== originalEditFormData[key as keyof typeof originalEditFormData]
      ) || editSelectedFile;

    if (hasChanges) setShowEditCancelConfirm(true);
    else closeEditRecordModal();
  };

  const handleEditRecordInputChange = (field: string, value: string) => {
    setEditRecordForm((prev) => ({ ...prev, [field]: value }));
    if (editRecordErrors[field]) setEditRecordErrors((prev) => ({ ...prev, [field]: '' }));
  };

  const handleEditRecordFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setEditSelectedFile(file);
      setEditIsUploading(true);
      setEditUploadProgress(0);
      if (editUploadIntervalRef.current) clearInterval(editUploadIntervalRef.current);
      editUploadIntervalRef.current = setInterval(() => {
        setEditUploadProgress((prev) => {
          if (prev >= 100) {
            if (editUploadIntervalRef.current) clearInterval(editUploadIntervalRef.current);
            editUploadIntervalRef.current = null;
            setEditIsUploading(false);
            return 100;
          }
          return Math.min(prev + Math.random() * 25 + 10, 100);
        });
      }, 150);
      setEditRecordForm((prev) => ({ ...prev, attachment: file.name }));
    } else {
      if (editUploadIntervalRef.current) clearInterval(editUploadIntervalRef.current);
      editUploadIntervalRef.current = null;
      setEditSelectedFile(null);
      setEditUploadProgress(0);
      setEditIsUploading(false);
      setEditRecordForm((prev) => ({ ...prev, attachment: '' }));
    }
  };

  const validateEditRecordForm = () => {
    const errors: Record<string, string> = {};
    if (!editRecordForm.work_order_number.trim()) errors.work_order_number = 'Work order number is required';
    else if (editRecordForm.work_order_number.length > 50) errors.work_order_number = 'Work order number must not exceed 50 characters';
    if (!editRecordForm.work_order_date) errors.work_order_date = 'Work order date is required';
    if (!editRecordForm.action_taken.trim()) errors.action_taken = 'Action taken is required';
    else if (editRecordForm.action_taken.length > 500) errors.action_taken = 'Action taken must not exceed 500 characters';
    if (!editRecordForm.reported_by.trim()) errors.reported_by = 'Reported by is required';
    else if (editRecordForm.reported_by.length > 50) errors.reported_by = 'Reported by must not exceed 50 characters';
    if (editRecordForm.work_order_type === 'Other' && !editRecordForm.custom_work_order_type.trim())
      errors.custom_work_order_type = 'Custom maintenance type is required when "Other" is selected';
    else if (editRecordForm.work_order_type === 'Other' && editRecordForm.custom_work_order_type.length > 50)
      errors.custom_work_order_type = 'Custom maintenance type must not exceed 50 characters';

    const existing = maint.records?.find(
      (r) => r.work_order_number === editRecordForm.work_order_number.trim() && r.work_order_number !== editingRecord?.work_order_number
    );
    if (existing) errors.work_order_number = 'Work order number already exists';

    setEditRecordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleEditRecord = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateEditRecordForm() || !editingRecord) return;

    try {
      setIsUpdating(true);
      setUpdateError(null);
      const updateData: UpdateMaintenanceRequest = {
        work_order_number: editRecordForm.work_order_number.trim(),
        work_order_date: editRecordForm.work_order_date,
        action_taken: editRecordForm.action_taken.trim(),
        reported_by: editRecordForm.reported_by.trim(),
        work_order_type: editRecordForm.work_order_type === 'Other' ? editRecordForm.custom_work_order_type.trim() : editRecordForm.work_order_type,
        attachment: editRecordForm.attachment || undefined,
      };

      await MaintenanceService.updateMaintenance(machine.serial_number, editingRecord.work_order_number, updateData);

      if (editSelectedFile) {
        try {
          if (originalEditFormData.attachment) {
            await AttachmentService.replaceMaintenanceAttachment(
              machine.serial_number, editingRecord.work_order_number, originalEditFormData.attachment, editSelectedFile
            );
          } else {
            await AttachmentService.uploadMaintenanceAttachment(
              machine.serial_number, editingRecord.work_order_number, editSelectedFile
            );
          }
        } catch (attachmentError) {
          if (isAuthError(attachmentError)) {
            alert('Your session has expired. Please log in again.');
            window.location.href = '/login';
            return;
          }
          const errorMessage = handleApiError(attachmentError);
          alert(`Maintenance record updated but attachment operation failed: ${errorMessage}`);
        }
      } else if (originalEditFormData.attachment && !editRecordForm.attachment) {
        try {
          await AttachmentService.deleteMaintenanceAttachment(
            machine.serial_number, editingRecord.work_order_number, originalEditFormData.attachment
          );
        } catch (attachmentError) {
          if (isAuthError(attachmentError)) {
            alert('Your session has expired. Please log in again.');
            window.location.href = '/login';
            return;
          }
          const errorMessage = handleApiError(attachmentError);
          alert(`Maintenance record updated but attachment deletion failed: ${errorMessage}`);
        }
      }

      maint.refetch();
      closeEditRecordModal();
    } catch (error) {
      if (isAuthError(error)) {
        alert('Your session has expired. Please log in again.');
        window.location.href = '/login';
        return;
      }
      const errorMessage = handleApiError(error);
      setUpdateError(`An error occurred while updating the maintenance record: ${errorMessage}`);
    } finally {
      setIsUpdating(false);
    }
  };

  // ======================== Delete Record ========================

  const openDeleteConfirm = (record: Maintenance) => {
    setRecordToDelete(record);
    setDeleteError(null);
    setShowDeleteConfirm(true);
  };

  const closeDeleteConfirm = () => {
    setRecordToDelete(null);
    setDeleteError(null);
    setShowDeleteConfirm(false);
  };

  const handleDeleteRecord = async () => {
    if (!recordToDelete) return;
    try {
      setIsDeleting(true);
      setDeleteError(null);

      await MaintenanceService.deleteMaintenance(machine.serial_number, recordToDelete.work_order_number);

      if (recordToDelete.attachment) {
        try {
          await AttachmentService.deleteMaintenanceAttachment(
            machine.serial_number, recordToDelete.work_order_number, recordToDelete.attachment
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
      maint.refetch();
      closeDeleteConfirm();
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

  // ======================== Render: Loading / Error ========================

  if (maint.isInitialLoading && maint.records.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading maintenance records...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (maint.error && maint.records.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="text-red-600 text-6xl mb-4">⚠️</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">Error Loading Maintenance Records</h3>
              <p className="text-gray-500 mb-4">{maint.error}</p>
              <button
                onClick={() => maint.refetch()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // ======================== Render: Main ========================

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <LoadingOverlay
        isVisible={isCreating || isUpdating || isDeleting}
        message={
          isCreating ? 'Creating maintenance record...' :
          isUpdating ? 'Updating maintenance record...' :
          isDeleting ? 'Deleting maintenance record...' :
          'Loading...'
        }
      />

      <div className="container mx-auto px-4">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-4 mb-4">
            <button
              onClick={handleBackNavigation}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back to Machines
            </button>
          </div>

          {/* Machine Info Card */}
          <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-4">Machine Information</h1>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <div className="space-y-2">
                <div>
                  <span className="text-sm text-gray-500">Serial Number:</span>
                  <div className="font-medium text-gray-900">{machine.serial_number}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Model:</span>
                  <div className="font-medium text-gray-900">{machine.model}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Brand:</span>
                  <div className="font-medium text-gray-900">{machine.brand}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Status:</span>
                  <div className="font-medium text-gray-900">{machine.status || 'Not specified'}</div>
                </div>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-sm text-gray-500">Customer:</span>
                  <div className="font-medium text-gray-900">{machine.customer}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Account Type:</span>
                  <div className="font-medium text-gray-900">{machine.account_type || 'Not specified'}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Location:</span>
                  <div className="font-medium text-gray-900">{machine.district}, {machine.state}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Person in Charge:</span>
                  <div className="font-medium text-gray-900">{machine.person_in_charge}</div>
                </div>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-sm text-gray-500">TNC Date:</span>
                  <div className="font-medium text-gray-900">{machine.tnc_date ? formatDate(machine.tnc_date) : 'Not set'}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">PPM Date:</span>
                  <div className="font-medium text-gray-900">{machine.ppm_date ? formatDate(machine.ppm_date) : 'Not set'}</div>
                </div>
                <div>
                  <span className="text-sm text-gray-500">Reported By:</span>
                  <div className="font-medium text-gray-900">{machine.reported_by || 'Not specified'}</div>
                </div>
              </div>
            </div>

            {/* Timestamps */}
            <div className="mt-6 pt-4 border-t border-gray-200">
              <div className="flex flex-wrap items-center justify-center gap-8 text-sm">
                <div className="flex items-center gap-2">
                  <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                  <span className="text-gray-500">Created:</span>
                  <span className="font-medium text-gray-900">{formatDateTime(machine.created_at)}</span>
                </div>
                <div className="hidden sm:block text-gray-300">•</div>
                <div className="flex items-center gap-2">
                  <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                  <span className="text-gray-500">Last Updated:</span>
                  <span className="font-medium text-gray-900">{formatDateTime(machine.updated_at)}</span>
                </div>
              </div>
            </div>

            {/* Notes & Attachment */}
            {(machine.additional_notes || machine.attachment) && (
              <div className="mt-6 pt-6 border-t border-gray-200">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {machine.additional_notes && (
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-2">Additional Notes</h4>
                      <div className="text-sm text-gray-700 bg-gray-100 rounded-lg p-3">{machine.additional_notes}</div>
                    </div>
                  )}
                  {machine.attachment && (
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-2">Attachment</h4>
                      <button
                        onClick={() => handleDownloadMachineAttachment(machine.attachment)}
                        className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-800 transition-colors"
                        title="Download machine attachment"
                      >
                        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <span>{machine.attachment}</span>
                      </button>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Summary Statistics */}
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-900">{maint.total}</div>
              <div className="text-sm text-gray-600">Total Records</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-green-600">{maint.preventativeCount}</div>
              <div className="text-sm text-gray-600">Preventive</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-red-600">{maint.emergencyCount}</div>
              <div className="text-sm text-gray-600">Emergency</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-blue-600">{maint.correctiveCount}</div>
              <div className="text-sm text-gray-600">Corrective</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-purple-600">{maint.inspectionCount}</div>
              <div className="text-sm text-gray-600">Inspection</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-600">{maint.otherCount}</div>
              <div className="text-sm text-gray-600">Other</div>
            </div>
          </div>

          {/* Title + Machine Actions */}
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-gray-900">Maintenance History</h2>
            <div className="flex items-center gap-3">
              {onEdit && (
                <button
                  onClick={onEdit}
                  className="bg-yellow-600 hover:bg-yellow-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                  Edit Machine
                </button>
              )}
              {onDelete && (
                <button
                  onClick={onDelete}
                  className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  Delete Machine
                </button>
              )}
              <button
                onClick={openAddRecordModal}
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Add New Record
              </button>
            </div>
          </div>
        </div>

        {/* Search Bar */}
        <div className="mb-4">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <input
                type="text"
                placeholder="Search maintenance records..."
                value={maint.searchQuery}
                onChange={(e) => maint.setSearchQuery(e.target.value)}
                className="block w-full pl-10 pr-10 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 bg-white"
              />
              {maint.hasActiveSearch && (
                <button
                  onClick={maint.clearSearch}
                  className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                  title="Clear search"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              )}
            </div>
            {maint.hasActiveSearch && (
              <div className="text-sm text-gray-500 whitespace-nowrap">
                {maint.total} result{maint.total !== 1 ? 's' : ''} found
              </div>
            )}
          </div>
        </div>

        {/* CSV Export — right-aligned, between search and table */}
        <div className="flex justify-end mb-3">
          <button
            onClick={handleExportMaintenanceCSV}
            disabled={csvExporting}
            className="inline-flex items-center gap-1.5 px-3 py-2 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 text-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="Export maintenance to CSV"
          >
            {csvExporting ? (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600" />
            ) : (
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            )}
            <span className="hidden sm:inline">Export CSV</span>
          </button>
        </div>

        {/* Inline error banner (visible even when records exist) */}
        {maint.error && maint.records.length > 0 && (
          <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center justify-between">
            <div className="flex items-center gap-2">
              <svg className="h-5 w-5 text-red-400 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <span className="text-sm font-medium">{maint.error}</span>
            </div>
            <button onClick={() => maint.refetch()} className="text-sm text-red-600 hover:text-red-800 font-medium underline">Retry</button>
          </div>
        )}

        {/* Maintenance Table */}
        <MaintenanceTable
          machineSerialNumber={machine.serial_number}
          records={maint.records}
          total={maint.total}
          currentPage={maint.currentPage}
          limit={maint.limit}
          loading={maint.isInitialLoading || maint.isSearchLoading || maint.isPaginationLoading}
          sortBy={maint.sort}
          onSortChange={maint.setSort}
          onPageChange={maint.goToPage}
          onPageSizeChange={maint.setLimit}
          onEdit={openEditRecordModal}
          onDelete={openDeleteConfirm}
        />
      </div>

      {/* Add New Record Modal */}
      {isAddRecordModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50 transition-opacity" onClick={closeAddRecordModal} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl">
              <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">Add New Maintenance Record</h2>
                <button onClick={closeAddRecordModal} className="text-gray-400 hover:text-gray-600 transition-colors">
                  <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <form onSubmit={handleAddNewRecord} className="px-6 py-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label htmlFor="work_order_number" className="block text-sm font-medium text-gray-700 mb-2">Work Order Number <span className="text-red-500">*</span></label>
                    <input id="work_order_number" type="text" value={newRecordForm.work_order_number} onChange={(e) => handleNewRecordInputChange('work_order_number', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${newRecordErrors.work_order_number ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter work order number" />
                    <p className="mt-1 text-sm text-gray-500">{newRecordForm.work_order_number.length}/50 characters</p>
                    {newRecordErrors.work_order_number && <p className="mt-1 text-sm text-red-600">{newRecordErrors.work_order_number}</p>}
                  </div>
                  <div>
                    <label htmlFor="work_order_date" className="block text-sm font-medium text-gray-700 mb-2">Work Order Date <span className="text-red-500">*</span></label>
                    <input id="work_order_date" type="date" value={newRecordForm.work_order_date} onChange={(e) => handleNewRecordInputChange('work_order_date', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${newRecordErrors.work_order_date ? 'border-red-500' : 'border-gray-300'}`} />
                    {newRecordErrors.work_order_date && <p className="mt-1 text-sm text-red-600">{newRecordErrors.work_order_date}</p>}
                  </div>
                  <div>
                    <label htmlFor="work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Maintenance Type <span className="text-red-500">*</span></label>
                    <select id="work_order_type" value={newRecordForm.work_order_type} onChange={(e) => handleNewRecordInputChange('work_order_type', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900">
                      <option value="Preventive">Preventive</option>
                      <option value="Corrective">Corrective</option>
                      <option value="Emergency">Emergency</option>
                      <option value="Inspection">Inspection</option>
                      <option value="Other">Other</option>
                    </select>
                    {newRecordForm.work_order_type === 'Other' && (
                      <div className="mt-3">
                        <label htmlFor="custom_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Custom Maintenance Type <span className="text-red-500">*</span></label>
                        <input id="custom_work_order_type" type="text" value={newRecordForm.custom_work_order_type} onChange={(e) => handleNewRecordInputChange('custom_work_order_type', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${newRecordErrors.custom_work_order_type ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter custom maintenance type" />
                        <p className="mt-1 text-sm text-gray-500">{newRecordForm.custom_work_order_type.length}/50 characters</p>
                        {newRecordErrors.custom_work_order_type && <p className="mt-1 text-sm text-red-600">{newRecordErrors.custom_work_order_type}</p>}
                      </div>
                    )}
                  </div>
                  <div>
                    <label htmlFor="reported_by" className="block text-sm font-medium text-gray-700 mb-2">Reported By <span className="text-red-500">*</span></label>
                    <input id="reported_by" type="text" value={newRecordForm.reported_by} onChange={(e) => handleNewRecordInputChange('reported_by', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${newRecordErrors.reported_by ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter technician name" />
                    <p className="mt-1 text-sm text-gray-500">{newRecordForm.reported_by.length}/50 characters</p>
                    {newRecordErrors.reported_by && <p className="mt-1 text-sm text-red-600">{newRecordErrors.reported_by}</p>}
                  </div>
                  <div className="md:col-span-2">
                    <label htmlFor="attachment" className="block text-sm font-medium text-gray-700 mb-2">Attachment</label>
                    <div className="space-y-3">
                      <input id="attachment" type="file" onChange={handleNewRecordFileChange} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100" accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt" disabled={isUploading} />
                      {isUploading && (
                        <div className="space-y-2">
                          <div className="flex justify-between text-sm text-gray-600"><span>Uploading...</span><span>{Math.round(uploadProgress)}%</span></div>
                          <div className="w-full bg-gray-200 rounded-full h-2"><div className="bg-blue-600 h-2 rounded-full transition-all duration-300 ease-out" style={{ width: `${uploadProgress}%` }} /></div>
                        </div>
                      )}
                      {selectedFile && !isUploading && (
                        <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2">
                            <svg className="h-5 w-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                            <div>
                              <p className="text-sm font-medium text-green-800">{selectedFile.name}</p>
                              <p className="text-xs text-green-600">{(selectedFile.size / 1024).toFixed(1)} KB - File selected</p>
                            </div>
                          </div>
                          <button type="button" onClick={() => { setSelectedFile(null); setNewRecordForm((prev) => ({ ...prev, attachment: '' })); const fi = document.getElementById('attachment') as HTMLInputElement; if (fi) fi.value = ''; }} className="text-red-500 hover:text-red-700 transition-colors" title="Remove file">
                            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                <div className="mt-6">
                  <label htmlFor="action_taken" className="block text-sm font-medium text-gray-700 mb-2">Action Taken / Description <span className="text-red-500">*</span></label>
                  <textarea id="action_taken" value={newRecordForm.action_taken} onChange={(e) => handleNewRecordInputChange('action_taken', e.target.value)} rows={4} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${newRecordErrors.action_taken ? 'border-red-500' : 'border-gray-300'}`} placeholder="Describe the maintenance action performed, parts replaced, issues found, etc." />
                  <p className="mt-1 text-sm text-gray-500">{newRecordForm.action_taken.length}/500 characters</p>
                  {newRecordErrors.action_taken && <p className="mt-1 text-sm text-red-600">{newRecordErrors.action_taken}</p>}
                </div>
                {createError && (
                  <div className="mt-3 mb-3 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
                    <div className="flex">
                      <svg className="h-5 w-5 text-red-400 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" /></svg>
                      <p className="ml-3 text-sm font-medium">{createError}</p>
                    </div>
                  </div>
                )}
                <div className="flex justify-end gap-3 mt-6">
                  <button type="button" onClick={closeAddRecordModal} className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                  <button type="submit" disabled={isCreating} className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-blue-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                    {isCreating ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Adding...</>) : 'Add Record'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Edit Record Modal */}
      {isEditRecordModalOpen && editingRecord && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50 transition-opacity" onClick={handleEditCancelClick} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl">
              <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">Edit Maintenance Record</h2>
                <button onClick={handleEditCancelClick} className="text-gray-400 hover:text-gray-600 transition-colors">
                  <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                </button>
              </div>
              <form onSubmit={handleEditRecord} className="px-6 py-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label htmlFor="edit_work_order_number" className="block text-sm font-medium text-gray-700 mb-2">Work Order Number <span className="text-red-500">*</span></label>
                    <input id="edit_work_order_number" type="text" value={editRecordForm.work_order_number} disabled className="w-full px-3 py-2 border border-gray-300 rounded-lg text-gray-900 bg-gray-100 cursor-not-allowed" />
                    {editRecordErrors.work_order_number && <p className="mt-1 text-sm text-red-600">{editRecordErrors.work_order_number}</p>}
                  </div>
                  <div>
                    <label htmlFor="edit_work_order_date" className="block text-sm font-medium text-gray-700 mb-2">Work Order Date <span className="text-red-500">*</span></label>
                    <input id="edit_work_order_date" type="date" value={editRecordForm.work_order_date} onChange={(e) => handleEditRecordInputChange('work_order_date', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${editRecordErrors.work_order_date ? 'border-red-500' : 'border-gray-300'}`} />
                    {editRecordErrors.work_order_date && <p className="mt-1 text-sm text-red-600">{editRecordErrors.work_order_date}</p>}
                  </div>
                  <div>
                    <label htmlFor="edit_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Maintenance Type <span className="text-red-500">*</span></label>
                    <select id="edit_work_order_type" value={editRecordForm.work_order_type} onChange={(e) => handleEditRecordInputChange('work_order_type', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900">
                      <option value="Preventive">Preventive</option>
                      <option value="Corrective">Corrective</option>
                      <option value="Emergency">Emergency</option>
                      <option value="Inspection">Inspection</option>
                      <option value="Other">Other</option>
                    </select>
                    {editRecordForm.work_order_type === 'Other' && (
                      <div className="mt-3">
                        <label htmlFor="edit_custom_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Custom Maintenance Type <span className="text-red-500">*</span></label>
                        <input id="edit_custom_work_order_type" type="text" value={editRecordForm.custom_work_order_type} onChange={(e) => handleEditRecordInputChange('custom_work_order_type', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${editRecordErrors.custom_work_order_type ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter custom maintenance type" />
                        <p className="mt-1 text-sm text-gray-500">{editRecordForm.custom_work_order_type.length}/50 characters</p>
                        {editRecordErrors.custom_work_order_type && <p className="mt-1 text-sm text-red-600">{editRecordErrors.custom_work_order_type}</p>}
                      </div>
                    )}
                  </div>
                  <div>
                    <label htmlFor="edit_reported_by" className="block text-sm font-medium text-gray-700 mb-2">Reported By <span className="text-red-500">*</span></label>
                    <input id="edit_reported_by" type="text" value={editRecordForm.reported_by} onChange={(e) => handleEditRecordInputChange('reported_by', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${editRecordErrors.reported_by ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter technician name" />
                    <p className="mt-1 text-sm text-gray-500">{editRecordForm.reported_by.length}/50 characters</p>
                    {editRecordErrors.reported_by && <p className="mt-1 text-sm text-red-600">{editRecordErrors.reported_by}</p>}
                  </div>
                  <div className="md:col-span-2">
                    <label htmlFor="edit_attachment" className="block text-sm font-medium text-gray-700 mb-2">Attachment</label>
                    <div className="space-y-3">
                      <input id="edit_attachment" type="file" onChange={handleEditRecordFileChange} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100" accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt" disabled={editIsUploading} />
                      {editIsUploading && (
                        <div className="space-y-2">
                          <div className="flex justify-between text-sm text-gray-600"><span>Uploading...</span><span>{Math.round(editUploadProgress)}%</span></div>
                          <div className="w-full bg-gray-200 rounded-full h-2"><div className="bg-blue-600 h-2 rounded-full transition-all duration-300 ease-out" style={{ width: `${editUploadProgress}%` }} /></div>
                        </div>
                      )}
                      {editSelectedFile && !editIsUploading && (
                        <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2">
                            <svg className="h-5 w-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                            <div>
                              <p className="text-sm font-medium text-green-800">{editSelectedFile.name}</p>
                              <p className="text-xs text-green-600">{(editSelectedFile.size / 1024).toFixed(1)} KB - File selected</p>
                            </div>
                          </div>
                          <button type="button" onClick={() => { setEditSelectedFile(null); setEditRecordForm((prev) => ({ ...prev, attachment: '' })); const fi = document.getElementById('edit_attachment') as HTMLInputElement; if (fi) fi.value = ''; }} className="text-red-500 hover:text-red-700 transition-colors" title="Remove file">
                            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                          </button>
                        </div>
                      )}
                      {!editSelectedFile && !editIsUploading && editRecordForm.attachment && (
                        <div className="flex items-center justify-between bg-blue-50 border border-blue-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2">
                            <svg className="h-5 w-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                            <div>
                              <p className="text-sm font-medium text-blue-800">{editRecordForm.attachment}</p>
                              <p className="text-xs text-blue-600">Current attachment</p>
                            </div>
                          </div>
                          <button type="button" onClick={() => handleEditRecordInputChange('attachment', '')} className="text-red-500 hover:text-red-700 transition-colors" title="Remove attachment">
                            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                <div className="mt-6">
                  <label htmlFor="edit_action_taken" className="block text-sm font-medium text-gray-700 mb-2">Action Taken / Description <span className="text-red-500">*</span></label>
                  <textarea id="edit_action_taken" value={editRecordForm.action_taken} onChange={(e) => handleEditRecordInputChange('action_taken', e.target.value)} rows={4} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${editRecordErrors.action_taken ? 'border-red-500' : 'border-gray-300'}`} placeholder="Describe the maintenance action performed..." />
                  <p className="mt-1 text-sm text-gray-500">{editRecordForm.action_taken.length}/500 characters</p>
                  {editRecordErrors.action_taken && <p className="mt-1 text-sm text-red-600">{editRecordErrors.action_taken}</p>}
                </div>
                {updateError && (
                  <div className="mt-3 mb-3 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
                    <div className="flex">
                      <svg className="h-5 w-5 text-red-400 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" /></svg>
                      <p className="ml-3 text-sm font-medium">{updateError}</p>
                    </div>
                  </div>
                )}
                <div className="flex justify-end gap-3 mt-6">
                  <button type="button" onClick={handleEditCancelClick} className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                  <button type="submit" disabled={isUpdating} className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-blue-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                    {isUpdating ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Updating...</>) : 'Update Record'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && recordToDelete && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50" onClick={closeDeleteConfirm} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full">
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
                  <p className="text-sm text-gray-500 mb-2">Are you sure you want to delete maintenance record <strong>{recordToDelete.work_order_number}</strong>?</p>
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
                  <button type="button" onClick={closeDeleteConfirm} className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                  <button type="button" onClick={handleDeleteRecord} disabled={isDeleting} className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:bg-red-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                    {isDeleting ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Deleting...</>) : 'Delete Record'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Cancel Confirmation */}
      {showEditCancelConfirm && (
        <div className="fixed inset-0 z-[60] overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50" />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex items-center mb-4">
                  <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100">
                    <svg className="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                  </div>
                </div>
                <div className="text-center">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Discard changes?</h3>
                  <p className="text-sm text-gray-500 mb-6">You have unsaved changes to this maintenance record. Are you sure you want to discard them and close the form?</p>
                </div>
                <div className="flex space-x-3">
                  <button type="button" onClick={() => setShowEditCancelConfirm(false)} className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Keep editing</button>
                  <button type="button" onClick={() => { setShowEditCancelConfirm(false); closeEditRecordModal(); }} className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">Discard changes</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      <FullPageLoader isVisible={isNavigatingBack} message="Loading machines list..." />
    </div>
  );
}
