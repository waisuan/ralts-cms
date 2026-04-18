'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Maintenance, MaintenanceOrderType } from '../types/maintenance';
import { MaintenanceService, UpdateMaintenanceRequest } from '../services/maintenanceService';
import { AttachmentService } from '../services/attachmentService';
import { handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import { backendDateToHtmlDate } from '../utils/dateUtils';
import LoadingOverlay from './LoadingOverlay';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

interface EditMaintenanceModalProps {
  machineSerialNumber: string;
  record: Maintenance | null;
  existingWorkOrderNumbers: string[];
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

type FormData = {
  work_order_number: string;
  work_order_date: string;
  action_taken: string;
  reported_by: string;
  work_order_type: MaintenanceOrderType;
  custom_work_order_type: string;
  attachment: string;
};

const emptyForm = (): FormData => ({
  work_order_number: '',
  work_order_date: '',
  action_taken: '',
  reported_by: '',
  work_order_type: 'Preventive',
  custom_work_order_type: '',
  attachment: '',
});

export default function EditMaintenanceModal({
  machineSerialNumber,
  record,
  existingWorkOrderNumbers,
  isOpen,
  onClose,
  onSuccess,
}: EditMaintenanceModalProps) {
  useLockBodyScroll(isOpen);

  const [form, setForm] = useState<FormData>(emptyForm);
  const [originalForm, setOriginalForm] = useState<FormData>(emptyForm);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const [isUpdating, setIsUpdating] = useState(false);
  const [showDiscardConfirm, setShowDiscardConfirm] = useState(false);

  const uploadIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const firstInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    return () => {
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
    };
  }, []);

  useEffect(() => {
    if (isOpen && record) {
      const isStandardType = ['Preventive', 'Corrective', 'Emergency', 'Inspection'].includes(record.work_order_type);
      const formData: FormData = {
        work_order_number: record.work_order_number,
        work_order_date: backendDateToHtmlDate(record.work_order_date),
        action_taken: record.action_taken,
        reported_by: record.reported_by,
        work_order_type: isStandardType ? (record.work_order_type as MaintenanceOrderType) : 'Other',
        custom_work_order_type: isStandardType ? '' : record.work_order_type,
        attachment: record.attachment || '',
      };
      setForm(formData);
      setOriginalForm(formData);
      setErrors({});
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setUpdateError(null);
      setShowDiscardConfirm(false);
      setTimeout(() => firstInputRef.current?.focus(), 50);
    }
  }, [isOpen, record]);

  if (!isOpen || !record) return null;

  const handleInputChange = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: '' }));
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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
      setForm((prev) => ({ ...prev, attachment: file.name }));
    } else {
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
      uploadIntervalRef.current = null;
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setForm((prev) => ({ ...prev, attachment: '' }));
    }
  };

  const hasChanges = () =>
    Object.keys(form).some(
      (key) => form[key as keyof FormData] !== originalForm[key as keyof FormData]
    ) || !!selectedFile;

  const handleCancelClick = () => {
    if (hasChanges()) setShowDiscardConfirm(true);
    else handleClose();
  };

  const handleClose = () => {
    setShowDiscardConfirm(false);
    onClose();
  };

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!form.work_order_number.trim()) errs.work_order_number = 'Work order number is required';
    else if (form.work_order_number.length > 50) errs.work_order_number = 'Work order number must not exceed 50 characters';
    if (!form.work_order_date) errs.work_order_date = 'Work order date is required';
    if (!form.action_taken.trim()) errs.action_taken = 'Action taken is required';
    else if (form.action_taken.length > 500) errs.action_taken = 'Action taken must not exceed 500 characters';
    if (!form.reported_by.trim()) errs.reported_by = 'Reported by is required';
    else if (form.reported_by.length > 50) errs.reported_by = 'Reported by must not exceed 50 characters';
    if (form.work_order_type === 'Other' && !form.custom_work_order_type.trim())
      errs.custom_work_order_type = 'Custom maintenance type is required when "Other" is selected';
    else if (form.work_order_type === 'Other' && form.custom_work_order_type.length > 50)
      errs.custom_work_order_type = 'Custom maintenance type must not exceed 50 characters';

    const duplicate = existingWorkOrderNumbers.find(
      (wo) => wo === form.work_order_number.trim() && wo !== record.work_order_number
    );
    if (duplicate) errs.work_order_number = 'Work order number already exists';

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      setIsUpdating(true);
      setUpdateError(null);
      const updateData: UpdateMaintenanceRequest = {
        work_order_number: form.work_order_number.trim(),
        work_order_date: form.work_order_date,
        action_taken: form.action_taken.trim(),
        reported_by: form.reported_by.trim(),
        work_order_type: form.work_order_type === 'Other' ? form.custom_work_order_type.trim() : form.work_order_type,
        attachment: form.attachment || undefined,
      };

      await MaintenanceService.updateMaintenance(machineSerialNumber, record.work_order_number, updateData);

      if (selectedFile) {
        try {
          if (originalForm.attachment) {
            await AttachmentService.replaceMaintenanceAttachment(
              machineSerialNumber, record.work_order_number, originalForm.attachment, selectedFile
            );
          } else {
            await AttachmentService.uploadMaintenanceAttachment(
              machineSerialNumber, record.work_order_number, selectedFile
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
      } else if (originalForm.attachment && !form.attachment) {
        try {
          await AttachmentService.deleteMaintenanceAttachment(
            machineSerialNumber, record.work_order_number, originalForm.attachment
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

      onSuccess();
      handleClose();
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

  return (
    <>
      <LoadingOverlay isVisible={isUpdating} message="Updating maintenance record..." />

      <div
        className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
        onClick={handleCancelClick}
      >
        <div className="flex min-h-full items-center justify-center p-4">
          <div
            className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Edit Maintenance Record</h2>
              <button onClick={handleCancelClick} className="text-gray-400 hover:text-gray-600 transition-colors">
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <form onSubmit={handleSubmit} className="px-6 py-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="edit_work_order_number" className="block text-sm font-medium text-gray-700 mb-2">Work Order Number <span className="text-red-500">*</span></label>
                  <input id="edit_work_order_number" type="text" value={form.work_order_number} disabled className="w-full px-3 py-2 border border-gray-300 rounded-lg text-gray-900 bg-gray-100 cursor-not-allowed" />
                  {errors.work_order_number && <p className="mt-1 text-sm text-red-600">{errors.work_order_number}</p>}
                </div>
                <div>
                  <label htmlFor="edit_work_order_date" className="block text-sm font-medium text-gray-700 mb-2">Work Order Date <span className="text-red-500">*</span></label>
                  <input ref={firstInputRef} id="edit_work_order_date" type="date" value={form.work_order_date} onChange={(e) => handleInputChange('work_order_date', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${errors.work_order_date ? 'border-red-500' : 'border-gray-300'}`} />
                  {errors.work_order_date && <p className="mt-1 text-sm text-red-600">{errors.work_order_date}</p>}
                </div>
                <div>
                  <label htmlFor="edit_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Maintenance Type <span className="text-red-500">*</span></label>
                  <select id="edit_work_order_type" value={form.work_order_type} onChange={(e) => handleInputChange('work_order_type', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900">
                    <option value="Preventive">Preventive</option>
                    <option value="Corrective">Corrective</option>
                    <option value="Emergency">Emergency</option>
                    <option value="Inspection">Inspection</option>
                    <option value="Other">Other</option>
                  </select>
                  {form.work_order_type === 'Other' && (
                    <div className="mt-3">
                      <label htmlFor="edit_custom_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Custom Maintenance Type <span className="text-red-500">*</span></label>
                      <input id="edit_custom_work_order_type" type="text" value={form.custom_work_order_type} onChange={(e) => handleInputChange('custom_work_order_type', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.custom_work_order_type ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter custom maintenance type" />
                      <p className="mt-1 text-sm text-gray-500">{form.custom_work_order_type.length}/50 characters</p>
                      {errors.custom_work_order_type && <p className="mt-1 text-sm text-red-600">{errors.custom_work_order_type}</p>}
                    </div>
                  )}
                </div>
                <div>
                  <label htmlFor="edit_reported_by" className="block text-sm font-medium text-gray-700 mb-2">Reported By <span className="text-red-500">*</span></label>
                  <input id="edit_reported_by" type="text" value={form.reported_by} onChange={(e) => handleInputChange('reported_by', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.reported_by ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter technician name" />
                  <p className="mt-1 text-sm text-gray-500">{form.reported_by.length}/50 characters</p>
                  {errors.reported_by && <p className="mt-1 text-sm text-red-600">{errors.reported_by}</p>}
                </div>
                <div className="md:col-span-2">
                  <label htmlFor="edit_attachment" className="block text-sm font-medium text-gray-700 mb-2">Attachment</label>
                  <div className="space-y-3">
                    <input id="edit_attachment" type="file" onChange={handleFileChange} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100" accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt" disabled={isUploading} />
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
                        <button type="button" onClick={() => { setSelectedFile(null); setForm((prev) => ({ ...prev, attachment: '' })); const fi = document.getElementById('edit_attachment') as HTMLInputElement; if (fi) fi.value = ''; }} className="text-red-500 hover:text-red-700 transition-colors" title="Remove file">
                          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                        </button>
                      </div>
                    )}
                    {!selectedFile && !isUploading && form.attachment && (
                      <div className="flex items-center justify-between bg-blue-50 border border-blue-200 rounded-lg p-3">
                        <div className="flex items-center space-x-2">
                          <svg className="h-5 w-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                          <div>
                            <p className="text-sm font-medium text-blue-800">{form.attachment}</p>
                            <p className="text-xs text-blue-600">Current attachment</p>
                          </div>
                        </div>
                        <button type="button" onClick={() => handleInputChange('attachment', '')} className="text-red-500 hover:text-red-700 transition-colors" title="Remove attachment">
                          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
              <div className="mt-6">
                <label htmlFor="edit_action_taken" className="block text-sm font-medium text-gray-700 mb-2">Action Taken / Description <span className="text-red-500">*</span></label>
                <textarea id="edit_action_taken" value={form.action_taken} onChange={(e) => handleInputChange('action_taken', e.target.value)} rows={4} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.action_taken ? 'border-red-500' : 'border-gray-300'}`} placeholder="Describe the maintenance action performed..." />
                <p className="mt-1 text-sm text-gray-500">{form.action_taken.length}/500 characters</p>
                {errors.action_taken && <p className="mt-1 text-sm text-red-600">{errors.action_taken}</p>}
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
                <button type="button" onClick={handleCancelClick} className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                <button type="submit" disabled={isUpdating} className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-blue-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                  {isUpdating ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Updating...</>) : 'Update Record'}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      {/* Discard changes confirmation */}
      {showDiscardConfirm && (
        <div
          className="fixed inset-0 z-[60] overflow-y-auto bg-black bg-opacity-50"
          onClick={(e) => {
            e.stopPropagation();
            setShowDiscardConfirm(false);
          }}
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
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Discard changes?</h3>
                  <p className="text-sm text-gray-500 mb-6">You have unsaved changes to this maintenance record. Are you sure you want to discard them and close the form?</p>
                </div>
                <div className="flex space-x-3">
                  <button type="button" onClick={() => setShowDiscardConfirm(false)} className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Keep editing</button>
                  <button type="button" onClick={handleClose} className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">Discard changes</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
