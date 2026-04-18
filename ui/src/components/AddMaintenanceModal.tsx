'use client';

import React, { useState, useRef, useEffect } from 'react';
import { MaintenanceOrderType } from '../types/maintenance';
import { MaintenanceService, CreateMaintenanceRequest } from '../services/maintenanceService';
import { AttachmentService } from '../services/attachmentService';
import { handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import LoadingOverlay from './LoadingOverlay';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

interface AddMaintenanceModalProps {
  machineSerialNumber: string;
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

export default function AddMaintenanceModal({
  machineSerialNumber,
  existingWorkOrderNumbers,
  isOpen,
  onClose,
  onSuccess,
}: AddMaintenanceModalProps) {
  useLockBodyScroll(isOpen);

  const [form, setForm] = useState<FormData>(emptyForm);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [createError, setCreateError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const uploadIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const firstInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    return () => {
      if (uploadIntervalRef.current) clearInterval(uploadIntervalRef.current);
    };
  }, []);

  useEffect(() => {
    if (isOpen) {
      setForm(emptyForm());
      setErrors({});
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setCreateError(null);
      setTimeout(() => firstInputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleInputChange = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
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

    if (existingWorkOrderNumbers.includes(form.work_order_number.trim()))
      errs.work_order_number = 'Work order number already exists';

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      setIsCreating(true);
      setCreateError(null);
      const createData: CreateMaintenanceRequest = {
        work_order_number: form.work_order_number.trim(),
        work_order_date: form.work_order_date,
        action_taken: form.action_taken.trim(),
        reported_by: form.reported_by.trim(),
        work_order_type: form.work_order_type === 'Other' ? form.custom_work_order_type.trim() : form.work_order_type,
        attachment: form.attachment || undefined,
      };

      await MaintenanceService.createMaintenance(machineSerialNumber, createData);

      if (selectedFile) {
        try {
          await AttachmentService.uploadMaintenanceAttachment(
            machineSerialNumber,
            form.work_order_number.trim(),
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

      onSuccess();
      onClose();
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

  return (
    <>
      <LoadingOverlay isVisible={isCreating} message="Creating maintenance record..." />
      <div
        className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
        onClick={onClose}
      >
        <div className="flex min-h-full items-center justify-center p-4">
          <div
            className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Add New Maintenance Record</h2>
              <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors">
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleSubmit} className="px-6 py-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="add_work_order_number" className="block text-sm font-medium text-gray-700 mb-2">Work Order Number <span className="text-red-500">*</span></label>
                  <input ref={firstInputRef} id="add_work_order_number" type="text" value={form.work_order_number} onChange={(e) => handleInputChange('work_order_number', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.work_order_number ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter work order number" />
                  <p className="mt-1 text-sm text-gray-500">{form.work_order_number.length}/50 characters</p>
                  {errors.work_order_number && <p className="mt-1 text-sm text-red-600">{errors.work_order_number}</p>}
                </div>
                <div>
                  <label htmlFor="add_work_order_date" className="block text-sm font-medium text-gray-700 mb-2">Work Order Date <span className="text-red-500">*</span></label>
                  <input id="add_work_order_date" type="date" value={form.work_order_date} onChange={(e) => handleInputChange('work_order_date', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${errors.work_order_date ? 'border-red-500' : 'border-gray-300'}`} />
                  {errors.work_order_date && <p className="mt-1 text-sm text-red-600">{errors.work_order_date}</p>}
                </div>
                <div>
                  <label htmlFor="add_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Maintenance Type <span className="text-red-500">*</span></label>
                  <select id="add_work_order_type" value={form.work_order_type} onChange={(e) => handleInputChange('work_order_type', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900">
                    <option value="Preventive">Preventive</option>
                    <option value="Corrective">Corrective</option>
                    <option value="Emergency">Emergency</option>
                    <option value="Inspection">Inspection</option>
                    <option value="Other">Other</option>
                  </select>
                  {form.work_order_type === 'Other' && (
                    <div className="mt-3">
                      <label htmlFor="add_custom_work_order_type" className="block text-sm font-medium text-gray-700 mb-2">Custom Maintenance Type <span className="text-red-500">*</span></label>
                      <input id="add_custom_work_order_type" type="text" value={form.custom_work_order_type} onChange={(e) => handleInputChange('custom_work_order_type', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.custom_work_order_type ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter custom maintenance type" />
                      <p className="mt-1 text-sm text-gray-500">{form.custom_work_order_type.length}/50 characters</p>
                      {errors.custom_work_order_type && <p className="mt-1 text-sm text-red-600">{errors.custom_work_order_type}</p>}
                    </div>
                  )}
                </div>
                <div>
                  <label htmlFor="add_reported_by" className="block text-sm font-medium text-gray-700 mb-2">Reported By <span className="text-red-500">*</span></label>
                  <input id="add_reported_by" type="text" value={form.reported_by} onChange={(e) => handleInputChange('reported_by', e.target.value)} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.reported_by ? 'border-red-500' : 'border-gray-300'}`} placeholder="Enter technician name" />
                  <p className="mt-1 text-sm text-gray-500">{form.reported_by.length}/50 characters</p>
                  {errors.reported_by && <p className="mt-1 text-sm text-red-600">{errors.reported_by}</p>}
                </div>
                <div className="md:col-span-2">
                  <label htmlFor="add_attachment" className="block text-sm font-medium text-gray-700 mb-2">Attachment</label>
                  <div className="space-y-3">
                    <input id="add_attachment" type="file" onChange={handleFileChange} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100" accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt" disabled={isUploading} />
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
                        <button type="button" onClick={() => { setSelectedFile(null); setForm((prev) => ({ ...prev, attachment: '' })); const fi = document.getElementById('add_attachment') as HTMLInputElement; if (fi) fi.value = ''; }} className="text-red-500 hover:text-red-700 transition-colors" title="Remove file">
                          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
              <div className="mt-6">
                <label htmlFor="add_action_taken" className="block text-sm font-medium text-gray-700 mb-2">Action Taken / Description <span className="text-red-500">*</span></label>
                <textarea id="add_action_taken" value={form.action_taken} onChange={(e) => handleInputChange('action_taken', e.target.value)} rows={4} className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${errors.action_taken ? 'border-red-500' : 'border-gray-300'}`} placeholder="Describe the maintenance action performed, parts replaced, issues found, etc." />
                <p className="mt-1 text-sm text-gray-500">{form.action_taken.length}/500 characters</p>
                {errors.action_taken && <p className="mt-1 text-sm text-red-600">{errors.action_taken}</p>}
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
                <button type="button" onClick={onClose} className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">Cancel</button>
                <button type="submit" disabled={isCreating} className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-blue-400 disabled:cursor-not-allowed flex items-center justify-center gap-2">
                  {isCreating ? (<><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Adding...</>) : 'Add Record'}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </>
  );
}
