'use client';

import React, { useState, useMemo, useEffect } from 'react';
import { mockMaintenanceRecords } from '../data/mockMaintenance';
import { Machine } from '../types/machine';
import { Maintenance, MaintenanceOrderType } from '../types/maintenance';

interface MaintenanceHistoryProps {
  machine: Machine;
  onBack: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

type SortField = 'work_order_date' | 'work_order_number' | 'worker_order_type' | 'reported_by';
type SortDirection = 'asc' | 'desc';

// Pagination configuration
const ITEMS_PER_PAGE = 10;
const PAGE_SIZE_OPTIONS = [5, 10, 20, 50];

const getMaintenanceTypeColor = (type: string): string => {
  switch (type) {
    case 'Preventive':
      return 'bg-green-100 text-green-800';
    case 'Corrective':
      return 'bg-blue-100 text-blue-800';
    case 'Emergency':
      return 'bg-red-100 text-red-800';
    case 'Inspection':
      return 'bg-purple-100 text-purple-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getMaintenanceTypeIcon = (type: string): React.ReactElement => {
  switch (type) {
    case 'Preventive':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    case 'Emergency':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
          />
        </svg>
      );
    case 'Corrective':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
          />
        </svg>
      );
    case 'Inspection':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      );
    default:
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
          />
        </svg>
      );
  }
};

export default function MaintenanceHistory({
  machine,
  onBack,
  onEdit,
  onDelete,
}: MaintenanceHistoryProps) {
  const [sortField, setSortField] = useState<SortField>('work_order_date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [selectedAction, setSelectedAction] = useState<{
    workOrder: string;
    action: string;
  } | null>(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(ITEMS_PER_PAGE);

  // Local state for maintenance records
  const [localMaintenanceRecords, setLocalMaintenanceRecords] = useState<Maintenance[]>(() =>
    mockMaintenanceRecords.filter(
      (record) => record.machine_serial_number === machine.serial_number
    )
  );

  // New record form state
  const [isAddRecordModalOpen, setIsAddRecordModalOpen] = useState(false);
  const [newRecordForm, setNewRecordForm] = useState({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    worker_order_type: 'Preventive' as MaintenanceOrderType,
    attachment: '',
  });
  const [newRecordErrors, setNewRecordErrors] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  // Edit record state
  const [isEditRecordModalOpen, setIsEditRecordModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Maintenance | null>(null);
  const [editRecordForm, setEditRecordForm] = useState({
    work_order_number: '',
    work_order_date: '',
    action_taken: '',
    reported_by: '',
    worker_order_type: 'Preventive' as MaintenanceOrderType,
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
    worker_order_type: 'Preventive' as MaintenanceOrderType,
    attachment: '',
  });

  // Delete confirmation state
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [recordToDelete, setRecordToDelete] = useState<Maintenance | null>(null);

  // Dropdown menu state
  const [openDropdownId, setOpenDropdownId] = useState<string | null>(null);
  const [dropdownPosition, setDropdownPosition] = useState<'above' | 'below'>('below');

  // Use local maintenance records
  const maintenanceRecords = localMaintenanceRecords;

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (openDropdownId && !(event.target as Element).closest('.relative')) {
        setOpenDropdownId(null);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [openDropdownId]);

  // Sort maintenance records
  const sortedRecords = useMemo(() => {
    return [...maintenanceRecords].sort((a, b) => {
      let aValue: string | number;
      let bValue: string | number;

      switch (sortField) {
        case 'work_order_date':
          aValue = new Date(a.work_order_date).getTime();
          bValue = new Date(b.work_order_date).getTime();
          break;
        case 'work_order_number':
          aValue = a.work_order_number;
          bValue = b.work_order_number;
          break;
        case 'worker_order_type':
          aValue = a.worker_order_type;
          bValue = b.worker_order_type;
          break;
        case 'reported_by':
          aValue = a.reported_by;
          bValue = b.reported_by;
          break;
        default:
          return 0;
      }

      if (sortDirection === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });
  }, [maintenanceRecords, sortField, sortDirection]);

  // Pagination calculations
  const totalItems = sortedRecords.length;
  const totalPages = Math.ceil(totalItems / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentRecords = sortedRecords.slice(startIndex, endIndex);

  // Reset to first page when sorting changes
  useEffect(() => {
    setCurrentPage(1);
  }, [sortField, sortDirection]);

  // Reset to first page when items per page changes
  useEffect(() => {
    setCurrentPage(1);
  }, [itemsPerPage]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  // Pagination handlers
  const goToPage = (page: number) => {
    setCurrentPage(Math.max(1, Math.min(page, totalPages)));
  };

  const goToNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(currentPage + 1);
    }
  };

  const goToPreviousPage = () => {
    if (currentPage > 1) {
      setCurrentPage(currentPage - 1);
    }
  };

  const handleItemsPerPageChange = (newItemsPerPage: number) => {
    setItemsPerPage(newItemsPerPage);
  };

  // Generate page numbers for pagination
  const getPageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;

    if (totalPages <= maxVisiblePages) {
      // Show all pages if total is small
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show pages around current page
      let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
      let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

      // Adjust if we're near the end
      if (endPage - startPage + 1 < maxVisiblePages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }

      for (let i = startPage; i <= endPage; i++) {
        pages.push(i);
      }
    }

    return pages;
  };

  const openActionModal = (workOrder: string, action: string) => {
    setSelectedAction({ workOrder, action });
  };

  const closeActionModal = () => {
    setSelectedAction(null);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleDownloadAttachment = (attachment: string) => {
    if (attachment) {
      console.log('Downloading attachment:', attachment);
      alert(`Downloading attachment: ${attachment}`);
    }
  };

  // New record form handlers
  const openAddRecordModal = () => {
    setIsAddRecordModalOpen(true);
    setNewRecordForm({
      work_order_number: '',
      work_order_date: '',
      action_taken: '',
      reported_by: '',
      worker_order_type: 'Preventive',
      attachment: '',
    });
    setNewRecordErrors({});
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
  };

  const closeAddRecordModal = () => {
    setIsAddRecordModalOpen(false);
    setNewRecordForm({
      work_order_number: '',
      work_order_date: '',
      action_taken: '',
      reported_by: '',
      worker_order_type: 'Preventive',
      attachment: '',
    });
    setNewRecordErrors({});
    setSelectedFile(null);
    setUploadProgress(0);
    setIsUploading(false);
  };

  const handleNewRecordInputChange = (field: string, value: string) => {
    setNewRecordForm((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
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

      // Simulate upload progress
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 100) {
            clearInterval(progressInterval);
            setIsUploading(false);
            return 100;
          }
          const increment = Math.random() * 25 + 10;
          return Math.min(prev + increment, 100);
        });
      }, 150);

      setNewRecordForm((prev) => ({ ...prev, attachment: file.name }));
    } else {
      setSelectedFile(null);
      setUploadProgress(0);
      setIsUploading(false);
      setNewRecordForm((prev) => ({ ...prev, attachment: '' }));
    }
  };

  const validateNewRecordForm = () => {
    const errors: Record<string, string> = {};

    if (!newRecordForm.work_order_number.trim()) {
      errors.work_order_number = 'Work order number is required';
    }
    if (!newRecordForm.work_order_date) {
      errors.work_order_date = 'Work order date is required';
    }
    if (!newRecordForm.action_taken.trim()) {
      errors.action_taken = 'Action taken is required';
    }
    if (!newRecordForm.reported_by.trim()) {
      errors.reported_by = 'Reported by is required';
    }

    // Check if work order number already exists
    const existingRecord = localMaintenanceRecords.find(
      (record) => record.work_order_number === newRecordForm.work_order_number.trim()
    );
    if (existingRecord) {
      errors.work_order_number = 'Work order number already exists';
    }

    setNewRecordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleAddNewRecord = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateNewRecordForm()) {
      return;
    }

    const now = new Date().toISOString();
    const newRecord: Maintenance = {
      machine_serial_number: machine.serial_number,
      work_order_number: newRecordForm.work_order_number.trim(),
      work_order_date: newRecordForm.work_order_date,
      action_taken: newRecordForm.action_taken.trim(),
      reported_by: newRecordForm.reported_by.trim(),
      worker_order_type: newRecordForm.worker_order_type,
      attachment: newRecordForm.attachment,
      created_at: now,
      updated_at: now,
    };

    setLocalMaintenanceRecords((prev) => [newRecord, ...prev]);
    closeAddRecordModal();
  };

  // Edit record handlers
  const openEditRecordModal = (record: Maintenance) => {
    const formData = {
      work_order_number: record.work_order_number,
      work_order_date: record.work_order_date.split('T')[0], // Extract date part for input
      action_taken: record.action_taken,
      reported_by: record.reported_by,
      worker_order_type: record.worker_order_type as MaintenanceOrderType,
      attachment: record.attachment,
    };

    setEditingRecord(record);
    setEditRecordForm(formData);
    setOriginalEditFormData(formData); // Store original data for comparison
    setEditRecordErrors({});
    setEditSelectedFile(null);
    setEditUploadProgress(0);
    setEditIsUploading(false);
    setIsEditRecordModalOpen(true);
    setOpenDropdownId(null);
  };

  const closeEditRecordModal = () => {
    setIsEditRecordModalOpen(false);
    setEditingRecord(null);
    setEditRecordForm({
      work_order_number: '',
      work_order_date: '',
      action_taken: '',
      reported_by: '',
      worker_order_type: 'Preventive',
      attachment: '',
    });
    setOriginalEditFormData({
      work_order_number: '',
      work_order_date: '',
      action_taken: '',
      reported_by: '',
      worker_order_type: 'Preventive',
      attachment: '',
    });
    setEditRecordErrors({});
    setEditSelectedFile(null);
    setEditUploadProgress(0);
    setEditIsUploading(false);
    setShowEditCancelConfirm(false);
  };

  const handleEditCancelClick = () => {
    // Check if form data has been changed
    const hasChanges =
      Object.keys(editRecordForm).some((key) => {
        return (
          editRecordForm[key as keyof typeof editRecordForm] !==
          originalEditFormData[key as keyof typeof originalEditFormData]
        );
      }) || editSelectedFile;

    if (hasChanges) {
      setShowEditCancelConfirm(true);
    } else {
      closeEditRecordModal();
    }
  };

  const handleEditConfirmCancel = () => {
    setShowEditCancelConfirm(false);
    closeEditRecordModal();
  };

  const handleEditRecordInputChange = (field: string, value: string) => {
    setEditRecordForm((prev) => ({ ...prev, [field]: value }));
    if (editRecordErrors[field]) {
      setEditRecordErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  const handleEditRecordFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setEditSelectedFile(file);
      setEditIsUploading(true);
      setEditUploadProgress(0);

      const progressInterval = setInterval(() => {
        setEditUploadProgress((prev) => {
          if (prev >= 100) {
            clearInterval(progressInterval);
            setEditIsUploading(false);
            return 100;
          }
          const increment = Math.random() * 25 + 10;
          return Math.min(prev + increment, 100);
        });
      }, 150);

      setEditRecordForm((prev) => ({ ...prev, attachment: file.name }));
    } else {
      setEditSelectedFile(null);
      setEditUploadProgress(0);
      setEditIsUploading(false);
      setEditRecordForm((prev) => ({ ...prev, attachment: '' }));
    }
  };

  const validateEditRecordForm = () => {
    const errors: Record<string, string> = {};

    if (!editRecordForm.work_order_number.trim()) {
      errors.work_order_number = 'Work order number is required';
    }
    if (!editRecordForm.work_order_date) {
      errors.work_order_date = 'Work order date is required';
    }
    if (!editRecordForm.action_taken.trim()) {
      errors.action_taken = 'Action taken is required';
    }
    if (!editRecordForm.reported_by.trim()) {
      errors.reported_by = 'Reported by is required';
    }

    // Check if work order number already exists (excluding current record)
    const existingRecord = localMaintenanceRecords.find(
      (record) =>
        record.work_order_number === editRecordForm.work_order_number.trim() &&
        record.work_order_number !== editingRecord?.work_order_number
    );
    if (existingRecord) {
      errors.work_order_number = 'Work order number already exists';
    }

    setEditRecordErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleEditRecord = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateEditRecordForm() || !editingRecord) {
      return;
    }

    const updatedRecord: Maintenance = {
      ...editingRecord,
      work_order_number: editRecordForm.work_order_number.trim(),
      work_order_date: editRecordForm.work_order_date,
      action_taken: editRecordForm.action_taken.trim(),
      reported_by: editRecordForm.reported_by.trim(),
      worker_order_type: editRecordForm.worker_order_type,
      attachment: editRecordForm.attachment,
      updated_at: new Date().toISOString(),
    };

    setLocalMaintenanceRecords((prev) =>
      prev.map((record) =>
        record.work_order_number === editingRecord.work_order_number ? updatedRecord : record
      )
    );
    closeEditRecordModal();
  };

  // Delete record handlers
  const openDeleteConfirm = (record: Maintenance) => {
    setRecordToDelete(record);
    setShowDeleteConfirm(true);
    setOpenDropdownId(null);
  };

  const closeDeleteConfirm = () => {
    setRecordToDelete(null);
    setShowDeleteConfirm(false);
  };

  const handleDeleteRecord = () => {
    if (recordToDelete) {
      setLocalMaintenanceRecords((prev) =>
        prev.filter((record) => record.work_order_number !== recordToDelete.work_order_number)
      );
    }
    closeDeleteConfirm();
  };

  // Dropdown menu handlers
  const toggleDropdown = (workOrderNumber: string, event: React.MouseEvent<HTMLButtonElement>) => {
    if (openDropdownId === workOrderNumber) {
      setOpenDropdownId(null);
    } else {
      const position = getDropdownPosition(event.currentTarget);
      setDropdownPosition(position);
      setOpenDropdownId(workOrderNumber);
    }
  };

  // Calculate dropdown position to prevent it from going off-screen
  const getDropdownPosition = (buttonElement: HTMLElement) => {
    const rect = buttonElement.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const dropdownHeight = 130; // Approximate height of dropdown menu with padding

    // Check if there's enough space below
    const spaceBelow = viewportHeight - rect.bottom;

    if (spaceBelow >= dropdownHeight) {
      // Position below (default)
      return 'below';
    } else {
      // Position above
      return 'above';
    }
  };

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) {
      return (
        <svg
          className="h-4 w-4 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
          />
        </svg>
      );
    }

    return sortDirection === 'asc' ? (
      <svg className="h-4 w-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
      </svg>
    ) : (
      <svg className="h-4 w-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-4 mb-4">
            <button
              onClick={onBack}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
              Back to Machines
            </button>
          </div>

          <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-4">Machine Information</h1>

            {/* Machine Details Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <div>
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
                    <div className="font-medium text-gray-900">
                      {machine.status || 'Not specified'}
                    </div>
                  </div>
                </div>
              </div>

              <div>
                <div className="space-y-2">
                  <div>
                    <span className="text-sm text-gray-500">Customer:</span>
                    <div className="font-medium text-gray-900">{machine.customer}</div>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500">Account Type:</span>
                    <div className="font-medium text-gray-900">
                      {machine.account_type || 'Not specified'}
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500">Location:</span>
                    <div className="font-medium text-gray-900">
                      {machine.district}, {machine.state}
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500">Person in Charge:</span>
                    <div className="font-medium text-gray-900">{machine.person_in_charge}</div>
                  </div>
                </div>
              </div>

              <div>
                <div className="space-y-2">
                  <div>
                    <span className="text-sm text-gray-500">TNC Date:</span>
                    <div className="font-medium text-gray-900">
                      {machine.tnc_date ? formatDate(machine.tnc_date) : 'Not set'}
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500">PPM Date:</span>
                    <div className="font-medium text-gray-900">
                      {machine.ppm_date ? formatDate(machine.ppm_date) : 'Not set'}
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-gray-500">Reported By:</span>
                    <div className="font-medium text-gray-900">
                      {machine.reported_by || 'Not specified'}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Created and Updated Row */}
            <div className="mt-6 pt-4 border-t border-gray-200">
              <div className="flex flex-wrap items-center justify-center gap-8 text-sm">
                <div className="flex items-center gap-2">
                  <svg
                    className="h-4 w-4 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                    />
                  </svg>
                  <span className="text-gray-500">Created:</span>
                  <span className="font-medium text-gray-900">
                    {formatDateTime(machine.created_at)}
                  </span>
                </div>
                <div className="hidden sm:block text-gray-300">•</div>
                <div className="flex items-center gap-2">
                  <svg
                    className="h-4 w-4 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  <span className="text-gray-500">Last Updated:</span>
                  <span className="font-medium text-gray-900">
                    {formatDateTime(machine.updated_at)}
                  </span>
                </div>
              </div>
            </div>

            {/* Additional Notes and Attachment */}
            {(machine.additional_notes || machine.attachment) && (
              <div className="mt-6 pt-6 border-t border-gray-200">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {machine.additional_notes && (
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-2">Additional Notes</h4>
                      <div className="text-sm text-gray-700 bg-gray-100 rounded-lg p-3">
                        {machine.additional_notes}
                      </div>
                    </div>
                  )}
                  {machine.attachment && (
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-2">Attachment</h4>
                      <button
                        onClick={() => handleDownloadAttachment(machine.attachment)}
                        className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-800 transition-colors"
                        title="Download attachment"
                      >
                        <svg
                          className="h-4 w-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
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
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-900">{maintenanceRecords.length}</div>
              <div className="text-sm text-gray-600">Total Records</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-green-600">
                {maintenanceRecords.filter((r) => r.worker_order_type === 'Preventive').length}
              </div>
              <div className="text-sm text-gray-600">Preventive</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-red-600">
                {maintenanceRecords.filter((r) => r.worker_order_type === 'Emergency').length}
              </div>
              <div className="text-sm text-gray-600">Emergency</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-blue-600">
                {maintenanceRecords.filter((r) => r.worker_order_type === 'Corrective').length}
              </div>
              <div className="text-sm text-gray-600">Corrective</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-purple-600">
                {maintenanceRecords.filter((r) => r.worker_order_type === 'Inspection').length}
              </div>
              <div className="text-sm text-gray-600">Inspection</div>
            </div>
          </div>

          {/* Add New Record Button */}
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-gray-900">Maintenance History</h2>
            <div className="flex items-center gap-3">
              {onEdit && (
                <button
                  onClick={onEdit}
                  className="bg-yellow-600 hover:bg-yellow-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
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
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                  Delete Machine
                </button>
              )}
              <button
                onClick={openAddRecordModal}
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 4v16m8-8H4"
                  />
                </svg>
                Add New Record
              </button>
            </div>
          </div>
        </div>

        {/* Maintenance Records Table */}
        <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
          {sortedRecords.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-gray-400 text-6xl mb-4">🔧</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">No Maintenance Records</h3>
              <p className="text-gray-500">No maintenance history found for this machine.</p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <button
                          onClick={() => handleSort('work_order_number')}
                          className="flex items-center gap-1 hover:text-gray-700"
                        >
                          Work Order
                          {getSortIcon('work_order_number')}
                        </button>
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <button
                          onClick={() => handleSort('work_order_date')}
                          className="flex items-center gap-1 hover:text-gray-700"
                        >
                          Date
                          {getSortIcon('work_order_date')}
                        </button>
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <button
                          onClick={() => handleSort('worker_order_type')}
                          className="flex items-center gap-1 hover:text-gray-700"
                        >
                          Type
                          {getSortIcon('worker_order_type')}
                        </button>
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Action Summary
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <button
                          onClick={() => handleSort('reported_by')}
                          className="flex items-center gap-1 hover:text-gray-700"
                        >
                          Reported By
                          {getSortIcon('reported_by')}
                        </button>
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Attachment
                      </th>
                      <th className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {currentRecords.map((record) => {
                      const actionSummary =
                        record.action_taken.length > 100
                          ? record.action_taken.substring(0, 100) + '...'
                          : record.action_taken;
                      const isActionTruncated = record.action_taken.length > 100;

                      return (
                        <React.Fragment key={record.work_order_number}>
                          <tr className="hover:bg-gray-50">
                            <td className="px-6 py-4 whitespace-nowrap">
                              <div className="text-sm font-medium text-gray-900">
                                {record.work_order_number}
                              </div>
                              <div className="text-xs text-gray-500 space-y-1">
                                <div>Created: {formatDateTime(record.created_at)}</div>
                                <div>Updated: {formatDateTime(record.updated_at)}</div>
                              </div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {formatDate(record.work_order_date)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span
                                className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${getMaintenanceTypeColor(record.worker_order_type)}`}
                              >
                                {getMaintenanceTypeIcon(record.worker_order_type)}
                                {record.worker_order_type}
                              </span>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-900 max-w-md">
                              {isActionTruncated ? (
                                <button
                                  onClick={() =>
                                    openActionModal(record.work_order_number, record.action_taken)
                                  }
                                  className="text-left hover:text-blue-600 cursor-pointer transition-colors"
                                  title="Click to view full details"
                                >
                                  <div className="line-clamp-2">{actionSummary}</div>
                                </button>
                              ) : (
                                <div className="line-clamp-2">{actionSummary}</div>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {record.reported_by}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                              {record.attachment ? (
                                <button
                                  onClick={() => handleDownloadAttachment(record.attachment)}
                                  className="flex items-center gap-1 text-blue-600 hover:text-blue-800"
                                  title="Download attachment"
                                >
                                  <svg
                                    className="h-4 w-4"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                  >
                                    <path
                                      strokeLinecap="round"
                                      strokeLinejoin="round"
                                      strokeWidth={2}
                                      d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                    />
                                  </svg>
                                  <span className="text-xs">Download</span>
                                </button>
                              ) : (
                                <span className="text-gray-400">-</span>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-center">
                              <div className="relative">
                                <button
                                  onClick={(e) => toggleDropdown(record.work_order_number, e)}
                                  className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 rounded-full"
                                  aria-label="More options"
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
                                      d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z"
                                    />
                                  </svg>
                                </button>

                                {/* Dropdown Menu */}
                                {openDropdownId === record.work_order_number && (
                                  <div
                                    className={`absolute right-0 w-48 bg-white rounded-md shadow-lg border border-gray-200 z-20 ${
                                      dropdownPosition === 'above'
                                        ? 'bottom-full mb-2'
                                        : 'top-full mt-2'
                                    }`}
                                  >
                                    <div className="py-1">
                                      <button
                                        onClick={() => openEditRecordModal(record)}
                                        className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
                                      >
                                        <svg
                                          className="h-4 w-4"
                                          fill="none"
                                          stroke="currentColor"
                                          viewBox="0 0 24 24"
                                        >
                                          <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth={2}
                                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                                          />
                                        </svg>
                                        Edit Record
                                      </button>
                                      <button
                                        onClick={() => openDeleteConfirm(record)}
                                        className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                                      >
                                        <svg
                                          className="h-4 w-4"
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
                                        Delete Record
                                      </button>
                                    </div>
                                  </div>
                                )}
                              </div>
                            </td>
                          </tr>
                        </React.Fragment>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              {/* Pagination Controls */}
              {totalPages > 1 && (
                <div className="bg-white px-6 py-4 border-t border-gray-200">
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    {/* Items per page selector */}
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-900 font-medium">Show:</span>
                      <select
                        value={itemsPerPage}
                        onChange={(e) => handleItemsPerPageChange(Number(e.target.value))}
                        className="border border-gray-300 rounded-md px-3 py-1 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        {PAGE_SIZE_OPTIONS.map((option) => (
                          <option key={option} value={option}>
                            {option}
                          </option>
                        ))}
                      </select>
                      <span className="text-sm text-gray-900 font-medium">per page</span>
                    </div>

                    {/* Pagination info */}
                    <div className="text-sm text-gray-900 font-medium">
                      Showing {startIndex + 1} to {Math.min(endIndex, totalItems)} of {totalItems}{' '}
                      results
                    </div>

                    {/* Pagination navigation */}
                    <div className="flex items-center gap-2">
                      {/* Previous button */}
                      <button
                        onClick={goToPreviousPage}
                        disabled={currentPage === 1}
                        className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 text-gray-900 font-medium"
                      >
                        Previous
                      </button>

                      {/* Page numbers */}
                      <div className="flex items-center gap-1">
                        {getPageNumbers().map((page) => (
                          <button
                            key={page}
                            onClick={() => goToPage(page)}
                            className={`px-3 py-1 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 font-medium ${
                              currentPage === page
                                ? 'bg-blue-600 text-white border-blue-600'
                                : 'border-gray-300 text-gray-900 hover:bg-gray-50'
                            }`}
                          >
                            {page}
                          </button>
                        ))}
                      </div>

                      {/* Next button */}
                      <button
                        onClick={goToNextPage}
                        disabled={currentPage === totalPages}
                        className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 text-gray-900 font-medium"
                      >
                        Next
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Action Details Modal */}
      {selectedAction && (
        <div
          className="fixed inset-0 backdrop-blur-md flex items-center justify-center z-50"
          onClick={closeActionModal}
        >
          <div
            className="bg-white rounded-lg border-2 border-gray-800 max-w-2xl w-full mx-4 max-h-96 overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">
                Action Details - {selectedAction.workOrder}
              </h3>
              <button
                onClick={closeActionModal}
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
            <div className="p-6 overflow-y-auto max-h-80">
              <div className="text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
                {selectedAction.action}
              </div>
            </div>
            <div className="flex justify-end gap-3 p-6 border-t border-gray-200">
              <button
                onClick={closeActionModal}
                className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg font-medium transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add New Record Modal */}
      {isAddRecordModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
            onClick={closeAddRecordModal}
          />

          {/* Modal */}
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl">
              {/* Header */}
              <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">Add New Maintenance Record</h2>
                <button
                  onClick={closeAddRecordModal}
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
              <form onSubmit={handleAddNewRecord} className="px-6 py-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Work Order Number */}
                  <div>
                    <label
                      htmlFor="work_order_number"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Work Order Number <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="work_order_number"
                      type="text"
                      value={newRecordForm.work_order_number}
                      onChange={(e) =>
                        handleNewRecordInputChange('work_order_number', e.target.value)
                      }
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                        newRecordErrors.work_order_number ? 'border-red-500' : 'border-gray-300'
                      }`}
                      placeholder="Enter work order number"
                    />
                    {newRecordErrors.work_order_number && (
                      <p className="mt-1 text-sm text-red-600">
                        {newRecordErrors.work_order_number}
                      </p>
                    )}
                  </div>

                  {/* Work Order Date */}
                  <div>
                    <label
                      htmlFor="work_order_date"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Work Order Date <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="work_order_date"
                      type="date"
                      value={newRecordForm.work_order_date}
                      onChange={(e) =>
                        handleNewRecordInputChange('work_order_date', e.target.value)
                      }
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                        newRecordErrors.work_order_date ? 'border-red-500' : 'border-gray-300'
                      }`}
                    />
                    {newRecordErrors.work_order_date && (
                      <p className="mt-1 text-sm text-red-600">{newRecordErrors.work_order_date}</p>
                    )}
                  </div>

                  {/* Maintenance Type */}
                  <div>
                    <label
                      htmlFor="worker_order_type"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Maintenance Type <span className="text-red-500">*</span>
                    </label>
                    <select
                      id="worker_order_type"
                      value={newRecordForm.worker_order_type}
                      onChange={(e) =>
                        handleNewRecordInputChange('worker_order_type', e.target.value)
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
                    >
                      <option value="Preventive">Preventive</option>
                      <option value="Corrective">Corrective</option>
                      <option value="Emergency">Emergency</option>
                      <option value="Inspection">Inspection</option>
                    </select>
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
                      value={newRecordForm.reported_by}
                      onChange={(e) => handleNewRecordInputChange('reported_by', e.target.value)}
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                        newRecordErrors.reported_by ? 'border-red-500' : 'border-gray-300'
                      }`}
                      placeholder="Enter technician name"
                    />
                    {newRecordErrors.reported_by && (
                      <p className="mt-1 text-sm text-red-600">{newRecordErrors.reported_by}</p>
                    )}
                  </div>

                  {/* Attachment */}
                  <div className="md:col-span-2">
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
                        onChange={handleNewRecordFileChange}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                        accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt"
                        disabled={isUploading}
                      />

                      {/* Upload Progress */}
                      {isUploading && (
                        <div className="space-y-2">
                          <div className="flex justify-between text-sm text-gray-600">
                            <span>Uploading...</span>
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

                      {/* File Info */}
                      {selectedFile && !isUploading && (
                        <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2">
                            <svg
                              className="h-5 w-5 text-green-600"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                              />
                            </svg>
                            <div>
                              <p className="text-sm font-medium text-green-800">
                                {selectedFile.name}
                              </p>
                              <p className="text-xs text-green-600">
                                {(selectedFile.size / 1024).toFixed(1)} KB - Upload successful
                              </p>
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() => {
                              setSelectedFile(null);
                              setNewRecordForm((prev) => ({ ...prev, attachment: '' }));
                              const fileInput = document.getElementById(
                                'attachment'
                              ) as HTMLInputElement;
                              if (fileInput) fileInput.value = '';
                            }}
                            className="text-red-500 hover:text-red-700 focus:outline-none focus:text-red-700 transition-colors"
                            title="Remove file"
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

                {/* Action Taken */}
                <div className="mt-6">
                  <label
                    htmlFor="action_taken"
                    className="block text-sm font-medium text-gray-700 mb-2"
                  >
                    Action Taken / Description <span className="text-red-500">*</span>
                  </label>
                  <textarea
                    id="action_taken"
                    value={newRecordForm.action_taken}
                    onChange={(e) => handleNewRecordInputChange('action_taken', e.target.value)}
                    rows={4}
                    className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                      newRecordErrors.action_taken ? 'border-red-500' : 'border-gray-300'
                    }`}
                    placeholder="Describe the maintenance action performed, parts replaced, issues found, etc."
                  />
                  {newRecordErrors.action_taken && (
                    <p className="mt-1 text-sm text-red-600">{newRecordErrors.action_taken}</p>
                  )}
                </div>

                {/* Action Buttons */}
                <div className="flex justify-end gap-3 mt-8">
                  <button
                    type="button"
                    onClick={closeAddRecordModal}
                    className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                  >
                    Add Record
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
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
            onClick={handleEditCancelClick}
          />

          {/* Modal */}
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative w-full max-w-2xl bg-white rounded-lg shadow-xl">
              {/* Header */}
              <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
                <h2 className="text-xl font-semibold text-gray-900">Edit Maintenance Record</h2>
                <button
                  onClick={handleEditCancelClick}
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
              <form onSubmit={handleEditRecord} className="px-6 py-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Work Order Number */}
                  <div>
                    <label
                      htmlFor="edit_work_order_number"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Work Order Number <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="edit_work_order_number"
                      type="text"
                      value={editRecordForm.work_order_number}
                      onChange={(e) =>
                        handleEditRecordInputChange('work_order_number', e.target.value)
                      }
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                        editRecordErrors.work_order_number ? 'border-red-500' : 'border-gray-300'
                      }`}
                      placeholder="Enter work order number"
                    />
                    {editRecordErrors.work_order_number && (
                      <p className="mt-1 text-sm text-red-600">
                        {editRecordErrors.work_order_number}
                      </p>
                    )}
                  </div>

                  {/* Work Order Date */}
                  <div>
                    <label
                      htmlFor="edit_work_order_date"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Work Order Date <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="edit_work_order_date"
                      type="date"
                      value={editRecordForm.work_order_date}
                      onChange={(e) =>
                        handleEditRecordInputChange('work_order_date', e.target.value)
                      }
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                        editRecordErrors.work_order_date ? 'border-red-500' : 'border-gray-300'
                      }`}
                    />
                    {editRecordErrors.work_order_date && (
                      <p className="mt-1 text-sm text-red-600">
                        {editRecordErrors.work_order_date}
                      </p>
                    )}
                  </div>

                  {/* Maintenance Type */}
                  <div>
                    <label
                      htmlFor="edit_worker_order_type"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Maintenance Type <span className="text-red-500">*</span>
                    </label>
                    <select
                      id="edit_worker_order_type"
                      value={editRecordForm.worker_order_type}
                      onChange={(e) =>
                        handleEditRecordInputChange('worker_order_type', e.target.value)
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
                    >
                      <option value="Preventive">Preventive</option>
                      <option value="Corrective">Corrective</option>
                      <option value="Emergency">Emergency</option>
                      <option value="Inspection">Inspection</option>
                    </select>
                  </div>

                  {/* Reported By */}
                  <div>
                    <label
                      htmlFor="edit_reported_by"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Reported By <span className="text-red-500">*</span>
                    </label>
                    <input
                      id="edit_reported_by"
                      type="text"
                      value={editRecordForm.reported_by}
                      onChange={(e) => handleEditRecordInputChange('reported_by', e.target.value)}
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                        editRecordErrors.reported_by ? 'border-red-500' : 'border-gray-300'
                      }`}
                      placeholder="Enter technician name"
                    />
                    {editRecordErrors.reported_by && (
                      <p className="mt-1 text-sm text-red-600">{editRecordErrors.reported_by}</p>
                    )}
                  </div>

                  {/* Attachment */}
                  <div className="md:col-span-2">
                    <label
                      htmlFor="edit_attachment"
                      className="block text-sm font-medium text-gray-700 mb-2"
                    >
                      Attachment
                    </label>
                    <div className="space-y-3">
                      <input
                        id="edit_attachment"
                        type="file"
                        onChange={handleEditRecordFileChange}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                        accept=".pdf,.doc,.docx,.png,.jpg,.jpeg,.txt"
                        disabled={editIsUploading}
                      />

                      {/* Upload Progress */}
                      {editIsUploading && (
                        <div className="space-y-2">
                          <div className="flex justify-between text-sm text-gray-600">
                            <span>Uploading...</span>
                            <span>{Math.round(editUploadProgress)}%</span>
                          </div>
                          <div className="w-full bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-blue-600 h-2 rounded-full transition-all duration-300 ease-out"
                              style={{ width: `${editUploadProgress}%` }}
                            />
                          </div>
                        </div>
                      )}

                      {/* File Info */}
                      {editSelectedFile && !editIsUploading && (
                        <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-3">
                          <div className="flex items-center space-x-2">
                            <svg
                              className="h-5 w-5 text-green-600"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                              />
                            </svg>
                            <div>
                              <p className="text-sm font-medium text-green-800">
                                {editSelectedFile.name}
                              </p>
                              <p className="text-xs text-green-600">
                                {(editSelectedFile.size / 1024).toFixed(1)} KB - Upload successful
                              </p>
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() => {
                              setEditSelectedFile(null);
                              setEditRecordForm((prev) => ({ ...prev, attachment: '' }));
                              const fileInput = document.getElementById(
                                'edit_attachment'
                              ) as HTMLInputElement;
                              if (fileInput) fileInput.value = '';
                            }}
                            className="text-red-500 hover:text-red-700 focus:outline-none focus:text-red-700 transition-colors"
                            title="Remove file"
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
                      {!editSelectedFile && !editIsUploading && editRecordForm.attachment && (
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
                              <p className="text-sm font-medium text-blue-800">
                                {editRecordForm.attachment}
                              </p>
                              <p className="text-xs text-blue-600">Current attachment</p>
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() => handleEditRecordInputChange('attachment', '')}
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

                {/* Action Taken */}
                <div className="mt-6">
                  <label
                    htmlFor="edit_action_taken"
                    className="block text-sm font-medium text-gray-700 mb-2"
                  >
                    Action Taken / Description <span className="text-red-500">*</span>
                  </label>
                  <textarea
                    id="edit_action_taken"
                    value={editRecordForm.action_taken}
                    onChange={(e) => handleEditRecordInputChange('action_taken', e.target.value)}
                    rows={4}
                    className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                      editRecordErrors.action_taken ? 'border-red-500' : 'border-gray-300'
                    }`}
                    placeholder="Describe the maintenance action performed, parts replaced, issues found, etc."
                  />
                  {editRecordErrors.action_taken && (
                    <p className="mt-1 text-sm text-red-600">{editRecordErrors.action_taken}</p>
                  )}
                </div>

                {/* Action Buttons */}
                <div className="flex justify-end gap-3 mt-8">
                  <button
                    type="button"
                    onClick={handleEditCancelClick}
                    className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                  >
                    Update Record
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
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    Delete Maintenance Record
                  </h3>
                  <p className="text-sm text-gray-500 mb-2">
                    Are you sure you want to delete maintenance record{' '}
                    <strong>{recordToDelete.work_order_number}</strong>?
                  </p>
                  <p className="text-sm text-gray-500 mb-6">
                    This action cannot be undone. All data associated with this maintenance record
                    will be permanently removed.
                  </p>
                </div>
                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={closeDeleteConfirm}
                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleDeleteRecord}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
                  >
                    Delete Record
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Cancel Confirmation Dialog */}
      {showEditCancelConfirm && (
        <div className="fixed inset-0 z-60 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50" />
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
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                      />
                    </svg>
                  </div>
                </div>
                <div className="text-center">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Discard changes?</h3>
                  <p className="text-sm text-gray-500 mb-6">
                    You have unsaved changes to this maintenance record. Are you sure you want to
                    discard them and close the form?
                  </p>
                </div>
                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={() => setShowEditCancelConfirm(false)}
                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Keep editing
                  </button>
                  <button
                    type="button"
                    onClick={handleEditConfirmCancel}
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
