'use client';

import React, { useState, useEffect, useMemo, useRef } from 'react';
import { Machine } from '../types/machine';
import { MALAYSIAN_STATES } from '../utils/constants';
import { backendDateToHtmlDate, htmlDateToBackendDate, isMachineDateUnset } from '../utils/dateUtils';
import { AttachmentService } from '../services/attachmentService';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

type MachineModalMode = 'add' | 'edit';

interface MachineModalProps {
  isOpen: boolean;
  mode: MachineModalMode;
  machine?: Machine | null; // Required for edit mode, optional for add mode
  onClose: () => void;
  onSubmit: (machine: Machine | Omit<Machine, 'created_at' | 'updated_at'>) => Promise<void>;
}

/** Edit mode: API sent a sentinel date (shown as empty in the date input) — border highlight until user picks a date. */
function editSentinelDateNeedsPick(
  mode: MachineModalMode,
  machine: Machine | null | undefined,
  backendIso: string,
  htmlDate: string
): boolean {
  return mode === 'edit' && !!machine && isMachineDateUnset(backendIso) && !htmlDate;
}

export default function MachineModal({
  isOpen,
  mode,
  machine,
  onClose,
  onSubmit,
}: MachineModalProps) {
  useLockBodyScroll(isOpen);

  const [formData, setFormData] = useState({
    serial_number: '',
    customer: '',
    state: '',
    account_type: '',
    model: '',
    status: '',
    brand: '',
    district: '',
    person_in_charge: '',
    reported_by: '',
    additional_notes: '',
    attachment: '',
    ppm_status: '',
    tnc_date: '',
    ppm_date: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const [showCancelConfirm, setShowCancelConfirm] = useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [attachmentChanged, setAttachmentChanged] = useState<boolean>(false);
  
  // Use ref to track progress intervals for cleanup
  const progressIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup interval on unmount
  useEffect(() => {
    return () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
    };
  }, []);

  // Pre-populate form when in edit mode and machine data is available
  useEffect(() => {
    if (mode === 'edit' && machine && isOpen) {
      setFormData({
        serial_number: machine.serial_number,
        customer: machine.customer,
        state: machine.state,
        account_type: machine.account_type,
        model: machine.model,
        status: machine.status,
        brand: machine.brand,
        district: machine.district,
        person_in_charge: machine.person_in_charge,
        reported_by: machine.reported_by,
        additional_notes: machine.additional_notes,
        attachment: machine.attachment,
        ppm_status: machine.ppm_status,
        tnc_date: backendDateToHtmlDate(machine.tnc_date),
        ppm_date: backendDateToHtmlDate(machine.ppm_date),
      });
      // Reset submission state when opening modal
      setIsSubmitting(false);
      setSubmitError(null);
      setAttachmentChanged(false);
    } else if (mode === 'add' && isOpen) {
      // Reset form for add mode
      setFormData({
        serial_number: '',
        customer: '',
        state: '',
        account_type: '',
        model: '',
        status: '',
        brand: '',
        district: '',
        person_in_charge: '',
        reported_by: '',
        additional_notes: '',
        attachment: '',
        ppm_status: '',
        tnc_date: '',
        ppm_date: '',
      });
      // Reset submission state when opening modal
      setIsSubmitting(false);
      setSubmitError(null);
      setAttachmentChanged(false);
    }
  }, [mode, machine, isOpen]);



  // Modal configuration based on mode
  const modalConfig = useMemo(() => {
    if (mode === 'edit') {
      return {
        title: 'Edit Machine',
        submitButtonText: 'Update Machine',
        submitButtonClass:
          'px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors',
      };
    } else {
      return {
        title: 'Add New Machine',
        submitButtonText: 'Add Machine',
        submitButtonClass:
          'px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors',
      };
    }
  }, [mode]);

  const tncNeedsAttention = editSentinelDateNeedsPick(mode, machine, machine?.tnc_date ?? '', formData.tnc_date);
  const ppmNeedsAttention = editSentinelDateNeedsPick(mode, machine, machine?.ppm_date ?? '', formData.ppm_date);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    // Required fields validation with character limits
    if (!formData.serial_number.trim()) {
      newErrors.serial_number = 'Serial number is required';
    } else if (formData.serial_number.length > 50) {
      newErrors.serial_number = 'Serial number must not exceed 50 characters';
    }
    
    if (!formData.customer.trim()) {
      newErrors.customer = 'Customer is required';
    } else if (formData.customer.length > 50) {
      newErrors.customer = 'Customer must not exceed 50 characters';
    }
    
    if (!formData.state) {
      newErrors.state = 'State is required';
    }
    
    if (!formData.model.trim()) {
      newErrors.model = 'Model is required';
    } else if (formData.model.length > 50) {
      newErrors.model = 'Model must not exceed 50 characters';
    }
    
    if (!formData.brand.trim()) {
      newErrors.brand = 'Brand is required';
    } else if (formData.brand.length > 50) {
      newErrors.brand = 'Brand must not exceed 50 characters';
    }
    
    if (!formData.district) {
      newErrors.district = 'District is required';
    } else if (formData.district.length > 50) {
      newErrors.district = 'District must not exceed 50 characters';
    }
    
    if (!formData.person_in_charge.trim()) {
      newErrors.person_in_charge = 'Person in charge is required';
    } else if (formData.person_in_charge.length > 50) {
      newErrors.person_in_charge = 'Person in charge must not exceed 50 characters';
    }
    
    if (!formData.reported_by.trim()) {
      newErrors.reported_by = 'Reported by is required';
    } else if (formData.reported_by.length > 50) {
      newErrors.reported_by = 'Reported by must not exceed 50 characters';
    }
    
    if (!formData.tnc_date) {
      newErrors.tnc_date = 'TNC date is required';
    }
    
    if (!formData.ppm_date) {
      newErrors.ppm_date = 'PPM date is required';
    }

    // Optional fields with character limits
    if (formData.account_type.length > 50) {
      newErrors.account_type = 'Account type must not exceed 50 characters';
    }
    
    if (formData.status.length > 50) {
      newErrors.status = 'Status must not exceed 50 characters';
    }
    
    if (formData.additional_notes.length > 500) {
      newErrors.additional_notes = 'Additional notes must not exceed 500 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    try {
      // Convert HTML date format back to backend format
      const submissionData = {
        ...formData,
        tnc_date: htmlDateToBackendDate(formData.tnc_date),
        ppm_date: htmlDateToBackendDate(formData.ppm_date),
        // For now, just set attachment to filename if file is selected
        attachment: selectedFile ? selectedFile.name : formData.attachment,
      };

      let savedMachine: Machine;

      if (mode === 'edit' && machine) {
        // Create updated machine with existing timestamps
        const updatedMachine: Machine = {
          ...submissionData,
          created_at: machine.created_at,
          updated_at: new Date().toISOString(),
        };
        await onSubmit(updatedMachine);
        savedMachine = updatedMachine;
      } else {
        // Create new machine (timestamps will be added by parent)
        await onSubmit(submissionData);
        // For new machines, we need to use the submissionData with a serial number
        savedMachine = {
          ...submissionData,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        } as Machine;
      }

      // Upload attachment if user actually changed it
      if (attachmentChanged && savedMachine.serial_number) {
        try {

          // Determine if this is a replacement or new upload
          const originalAttachment = mode === 'edit' ? machine?.attachment || '' : '';
          const hasOriginalAttachment = originalAttachment.trim() !== '';
          
          if (selectedFile) {
            // User selected a new file
            if (hasOriginalAttachment) {
              // Use PUT to replace existing attachment
              console.log('🔄 Replacing existing attachment:', originalAttachment, '→', selectedFile.name);
              await AttachmentService.replaceMachineAttachment(
                savedMachine.serial_number,
                originalAttachment, // old attachment name
                selectedFile
              );
            } else {
              // Use POST to create new attachment (no existing attachment)
              console.log('📎 Creating new attachment:', selectedFile.name);
              await AttachmentService.uploadMachineAttachment(
                savedMachine.serial_number,
                selectedFile
              );
            }
          } else if (hasOriginalAttachment) {
            // User removed the attachment
            console.log('🗑️ Removing existing attachment:', originalAttachment);
            await AttachmentService.deleteMachineAttachment(
              savedMachine.serial_number,
              originalAttachment
            );
          }
          
          console.log('✅ Attachment changes processed successfully');
        } catch (uploadError) {
          console.error('❌ Attachment upload failed:', uploadError);
          // Show a warning but don't prevent machine creation
          setSubmitError(
            'Machine saved successfully, but attachment upload failed. You can try uploading the attachment again later.'
          );
          setIsSubmitting(false);
          return; // Don't close modal so user can see the error
        }
      }

      // Only close if everything successful
      handleClose();
    } catch (machineError) {
      console.error('❌ Machine save failed:', machineError);
      setSubmitError('An error occurred while saving the machine. Please try again.');
      setIsSubmitting(false);
    }
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      // Validate file type (PDF only as per Go backend)
      if (!file.name.toLowerCase().endsWith('.pdf')) {
        setSubmitError('Only PDF files are allowed for attachments.');
        // Clear the file input
        e.target.value = '';
        return;
      }

      // Validate file size (5MB max as per Go backend)
      const maxSize = 5 * 1024 * 1024; // 5MB
      if (file.size > maxSize) {
        setSubmitError('File size must be less than 5MB.');
        // Clear the file input
        e.target.value = '';
        return;
      }

      setSelectedFile(file);
      setIsUploading(true);
      setUploadProgress(0);
      setSubmitError(null); // Clear any previous errors
      setAttachmentChanged(true); // Mark that user changed the attachment

      // Clear any existing interval
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
      
      // Simulate upload progress for better UX
      progressIntervalRef.current = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 100) {
            if (progressIntervalRef.current) {
              clearInterval(progressIntervalRef.current);
              progressIntervalRef.current = null;
            }
            setIsUploading(false);
            return 100;
          }
          const increment = Math.random() * 15 + 5; // Random increment between 5-20%
          const newProgress = prev + increment;
          return Math.min(newProgress, 100); // Ensure progress never exceeds 100%
        });
      }, 150);

      // Update form data to show the filename
      setFormData((prev) => ({
        ...prev,
        attachment: file.name,
      }));
    } else {
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setFormData((prev) => ({
        ...prev,
        attachment: '',
      }));
    }
  };

  const handleRemoveFile = () => {
    // Clear any progress intervals
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }
    
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
    setAttachmentChanged(true); // Mark that user removed the attachment
    setFormData((prev) => ({
      ...prev,
      attachment: '',
    }));

    // Clear the file input
    const fileInput = document.getElementById('attachment') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
  };

  const handleCancelClick = () => {
    // Check if form data has been changed
    const hasChanges =
      mode === 'edit'
        ? machine &&
          Object.keys(formData).some((key) => {
            const formValue = formData[key as keyof typeof formData];
            const machineValue = machine[key as keyof Machine];
            
            // Special handling for date fields - convert machine dates to HTML format for comparison
            if (key === 'tnc_date' || key === 'ppm_date') {
              const convertedMachineValue = backendDateToHtmlDate(machineValue as string);
              return formValue !== convertedMachineValue;
            }
            
            return formValue !== machineValue;
          })
        : Object.values(formData).some((value) => value !== '') || selectedFile;

    if (hasChanges) {
      setShowCancelConfirm(true);
    } else {
      handleClose();
    }
  };

  const handleConfirmCancel = () => {
    setShowCancelConfirm(false);
    handleClose();
  };

  const handleClose = () => {
    setFormData({
      serial_number: '',
      customer: '',
      state: '',
      account_type: '',
      model: '',
      status: '',
      brand: '',
      district: '',
      person_in_charge: '',
      reported_by: '',
      additional_notes: '',
      attachment: '',
      ppm_status: '',
      tnc_date: '',
      ppm_date: '',
    });
    setErrors({});
    // Clear any progress intervals
    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }
    
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
    setShowCancelConfirm(false);
    setSubmitError(null);
    setAttachmentChanged(false);
    onClose();
  };

  if (!isOpen || (mode === 'edit' && !machine)) return null;

  return (
    <div
      className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
      onClick={handleCancelClick}
    >
      <div className="flex min-h-full items-center justify-center p-0 sm:p-4">
        <div
          className="relative w-full max-w-4xl bg-white sm:rounded-lg shadow-xl min-h-screen sm:min-h-0"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-semibold text-gray-900">{modalConfig.title}</h2>
            <button
              onClick={handleCancelClick}
              aria-label="Close"
              className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
            >
              <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="px-6 py-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Serial Number */}
              <div>
                <label
                  htmlFor="serial_number"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Serial Number <span className="text-red-500">*</span>
                </label>
                <input
                  id="serial_number"
                  type="text"
                  value={formData.serial_number}
                  onChange={(e) => handleInputChange('serial_number', e.target.value)}
                  disabled={mode === 'edit'}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.serial_number ? 'border-red-500' : 'border-gray-300'
                  } ${mode === 'edit' ? 'bg-gray-100 cursor-not-allowed' : ''}`}
                  placeholder="Enter serial number"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.serial_number.length}/50 characters
                </p>
                {errors.serial_number && (
                  <p className="mt-1 text-sm text-red-600">{errors.serial_number}</p>
                )}
              </div>

              {/* Customer */}
              <div>
                <label htmlFor="customer" className="block text-sm font-medium text-gray-700 mb-2">
                  Customer <span className="text-red-500">*</span>
                </label>
                <input
                  id="customer"
                  type="text"
                  value={formData.customer}
                  onChange={(e) => handleInputChange('customer', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.customer ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter customer name"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.customer.length}/50 characters
                </p>
                {errors.customer && <p className="mt-1 text-sm text-red-600">{errors.customer}</p>}
              </div>

              {/* State */}
              <div>
                <label htmlFor="state" className="block text-sm font-medium text-gray-700 mb-2">
                  State <span className="text-red-500">*</span>
                </label>
                <select
                  id="state"
                  value={formData.state}
                  onChange={(e) => handleInputChange('state', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                    errors.state ? 'border-red-500' : 'border-gray-300'
                  }`}
                >
                  <option value="">Select state</option>
                  {MALAYSIAN_STATES.map((state) => (
                    <option key={state} value={state}>
                      {state}
                    </option>
                  ))}
                </select>
                {errors.state && <p className="mt-1 text-sm text-red-600">{errors.state}</p>}
              </div>

              {/* Account Type */}
              <div>
                <label
                  htmlFor="account_type"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Account Type
                </label>
                <input
                  id="account_type"
                  type="text"
                  value={formData.account_type}
                  onChange={(e) => handleInputChange('account_type', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.account_type ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter account type"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.account_type.length}/50 characters
                </p>
                {errors.account_type && (
                  <p className="mt-1 text-sm text-red-600">{errors.account_type}</p>
                )}
              </div>

              {/* Model */}
              <div>
                <label htmlFor="model" className="block text-sm font-medium text-gray-700 mb-2">
                  Model <span className="text-red-500">*</span>
                </label>
                <input
                  id="model"
                  type="text"
                  value={formData.model}
                  onChange={(e) => handleInputChange('model', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.model ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter model number"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.model.length}/50 characters
                </p>
                {errors.model && <p className="mt-1 text-sm text-red-600">{errors.model}</p>}
              </div>

              {/* Status */}
              <div>
                <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-2">
                  Status
                </label>
                <input
                  id="status"
                  type="text"
                  value={formData.status}
                  onChange={(e) => handleInputChange('status', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.status ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter status"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.status.length}/50 characters
                </p>
                {errors.status && (
                  <p className="mt-1 text-sm text-red-600">{errors.status}</p>
                )}
              </div>

              {/* Brand */}
              <div>
                <label htmlFor="brand" className="block text-sm font-medium text-gray-700 mb-2">
                  Brand <span className="text-red-500">*</span>
                </label>
                <input
                  id="brand"
                  type="text"
                  value={formData.brand}
                  onChange={(e) => handleInputChange('brand', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.brand ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter brand name"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.brand.length}/50 characters
                </p>
                {errors.brand && <p className="mt-1 text-sm text-red-600">{errors.brand}</p>}
              </div>

              {/* District */}
              <div>
                <label htmlFor="district" className="block text-sm font-medium text-gray-700 mb-2">
                  District <span className="text-red-500">*</span>
                </label>
                <input
                  id="district"
                  type="text"
                  value={formData.district}
                  onChange={(e) => handleInputChange('district', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.district ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter district name"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.district.length}/50 characters
                </p>
                {errors.district && <p className="mt-1 text-sm text-red-600">{errors.district}</p>}
              </div>

              {/* Person in Charge */}
              <div>
                <label
                  htmlFor="person_in_charge"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Person in Charge <span className="text-red-500">*</span>
                </label>
                <input
                  id="person_in_charge"
                  type="text"
                  value={formData.person_in_charge}
                  onChange={(e) => handleInputChange('person_in_charge', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.person_in_charge ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter person in charge"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.person_in_charge.length}/50 characters
                </p>
                {errors.person_in_charge && (
                  <p className="mt-1 text-sm text-red-600">{errors.person_in_charge}</p>
                )}
              </div>

              {/* Reported By */}
              <div>
                <label
                  htmlFor="reported_by"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Reported By <span className="text-red-500">*</span>
                </label>
                <input
                  id="reported_by"
                  type="text"
                  value={formData.reported_by}
                  onChange={(e) => handleInputChange('reported_by', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.reported_by ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter reporter name"
                />
                <p className="mt-1 text-sm text-gray-500">
                  {formData.reported_by.length}/50 characters
                </p>
                {errors.reported_by && (
                  <p className="mt-1 text-sm text-red-600">{errors.reported_by}</p>
                )}
              </div>

              {/* TNC Date */}
              <div>
                <label htmlFor="tnc_date" className="block text-sm font-medium text-gray-700 mb-2">
                  TNC Date <span className="text-red-500">*</span>
                </label>
                <input
                  id="tnc_date"
                  type="date"
                  value={formData.tnc_date}
                  onChange={(e) => handleInputChange('tnc_date', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                    errors.tnc_date || tncNeedsAttention ? 'border-red-500' : 'border-gray-300'
                  }`}
                />
                {errors.tnc_date && <p className="mt-1 text-sm text-red-600">{errors.tnc_date}</p>}
              </div>

              {/* PPM Date */}
              <div>
                <label htmlFor="ppm_date" className="block text-sm font-medium text-gray-700 mb-2">
                  PPM Date <span className="text-red-500">*</span>
                </label>
                <input
                  id="ppm_date"
                  type="date"
                  value={formData.ppm_date}
                  onChange={(e) => handleInputChange('ppm_date', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                    errors.ppm_date || ppmNeedsAttention ? 'border-red-500' : 'border-gray-300'
                  }`}
                />
                {errors.ppm_date && <p className="mt-1 text-sm text-red-600">{errors.ppm_date}</p>}
              </div>

              {/* Attachment */}
              <div>
                <label
                  htmlFor="attachment"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Attachment
                </label>
                <div className="space-y-3">
                  <input
                    id="attachment"
                    type="file"
                    onChange={handleFileChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                    accept=".pdf"
                    disabled={isUploading || isSubmitting}
                  />

                  {/* Upload Progress */}
                  {isUploading && (
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm text-gray-600">
                        <span>Processing file...</span>
                        <span>{Math.round(uploadProgress)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-blue-600 h-2 rounded-full transition-all duration-300 ease-out"
                          style={{ width: `${uploadProgress}%` }}
                        />
                      </div>
                    </div>
                  )}

                  {/* File Info - Selected but not uploaded yet */}
                  {selectedFile && !isUploading && (
                    <div className="flex items-center justify-between bg-blue-50 border border-blue-200 rounded-lg p-3">
                      <div className="flex items-center space-x-2">
                        <svg
                          className="h-5 w-5 text-blue-600"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                        <div>
                          <p className="text-sm font-medium text-blue-800">{selectedFile.name}</p>
                          <p className="text-xs text-blue-600">
                            {(selectedFile.size / 1024).toFixed(1)} KB - Ready to upload
                          </p>
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={handleRemoveFile}
                        className="text-red-500 hover:text-red-700 focus:outline-none focus:text-red-700 transition-colors"
                        title="Remove file"
                        disabled={isSubmitting}
                      >
                        <svg
                          className="h-5 w-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  )}

                  {/* Show existing attachment in edit mode if no new file */}
                  {mode === 'edit' && !selectedFile && !isUploading && formData.attachment && (
                    <div className="flex items-center justify-between bg-blue-50 border border-blue-200 rounded-lg p-3">
                      <div className="flex items-center space-x-2">
                        <svg
                          className="h-5 w-5 text-blue-600"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                        <div>
                          <p className="text-sm font-medium text-blue-800">{formData.attachment}</p>
                          <p className="text-xs text-blue-600">Current attachment</p>
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => {
                          handleInputChange('attachment', '');
                          setAttachmentChanged(true);
                        }}
                        className="text-red-500 hover:text-red-700 focus:outline-none focus:text-red-700 transition-colors"
                        title="Remove attachment"
                      >
                        <svg
                          className="h-5 w-5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Additional Notes */}
            <div className="mt-6">
              <label
                htmlFor="additional_notes"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Additional Notes
              </label>
              <textarea
                id="additional_notes"
                value={formData.additional_notes}
                onChange={(e) => handleInputChange('additional_notes', e.target.value)}
                rows={4}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                  errors.additional_notes ? 'border-red-500' : 'border-gray-300'
                }`}
                placeholder="Enter any additional notes or comments"
              />
              <p className="mt-1 text-sm text-gray-500">
                {formData.additional_notes.length}/500 characters
              </p>
              {errors.additional_notes && (
                <p className="mt-1 text-sm text-red-600">{errors.additional_notes}</p>
              )}
            </div>

            {/* Error Display - moved closer to buttons */}
            {submitError && (
              <div className="mt-3 mb-3 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <p className="text-sm font-medium">{submitError}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex justify-end gap-3 mt-6">
              <button
                type="button"
                onClick={handleCancelClick}
                className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
              >
                Cancel
              </button>
              <button 
                type="submit" 
                disabled={isSubmitting}
                className={`${modalConfig.submitButtonClass} disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2`}
              >
                {isSubmitting ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                    {mode === 'edit' ? 'Updating...' : 'Adding...'}
                  </>
                ) : (
                  modalConfig.submitButtonText
                )}
              </button>
            </div>
          </form>
        </div>
      </div>

      {/* Cancel Confirmation Dialog */}
      {showCancelConfirm && (
        <div
          className="fixed inset-0 z-[60] overflow-y-auto bg-black bg-opacity-50"
          onClick={(e) => {
            e.stopPropagation();
            setShowCancelConfirm(false);
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
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                      />
                    </svg>
                  </div>
                </div>
                <div className="text-center">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Discard changes?</h3>
                  <p className="text-sm text-gray-500 mb-6">
                    You have unsaved changes. Are you sure you want to discard them and close the
                    form?
                  </p>
                </div>
                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={() => setShowCancelConfirm(false)}
                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Keep editing
                  </button>
                  <button
                    type="button"
                    onClick={handleConfirmCancel}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
                  >
                    Discard changes
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
