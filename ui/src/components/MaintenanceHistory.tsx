'use client';

import React, { useState, useMemo } from 'react';
import { Maintenance, MaintenanceOrderType } from '../types/maintenance';
import { mockMaintenanceRecords } from '../data/mockMaintenance';
import { Machine } from '../types/machine';

interface MaintenanceHistoryProps {
  machine: Machine;
  onBack: () => void;
}

type SortField = 'work_order_date' | 'work_order_number' | 'worker_order_type' | 'reported_by';
type SortDirection = 'asc' | 'desc';

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
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      );
    case 'Emergency':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      );
    case 'Corrective':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
        </svg>
      );
    case 'Inspection':
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      );
    default:
      return (
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      );
  }
};

export default function MaintenanceHistory({ machine, onBack }: MaintenanceHistoryProps) {
  const [sortField, setSortField] = useState<SortField>('work_order_date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [selectedAction, setSelectedAction] = useState<{ workOrder: string; action: string } | null>(null);

  // Filter maintenance records for this specific machine
  const maintenanceRecords = useMemo(() => {
    return mockMaintenanceRecords.filter(
      (record) => record.machine_serial_number === machine.serial_number
    );
  }, [machine.serial_number]);

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

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
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

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) {
      return (
        <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
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
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
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
                    <div className="font-medium text-gray-900">{machine.status || 'Not specified'}</div>
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
              </div>

              <div>
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
            </div>

            {/* Created and Updated Row */}
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
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-900">{maintenanceRecords.length}</div>
              <div className="text-sm text-gray-600">Total Records</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-green-600">
                {maintenanceRecords.filter(r => r.worker_order_type === 'Preventive').length}
              </div>
              <div className="text-sm text-gray-600">Preventive</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-red-600">
                {maintenanceRecords.filter(r => r.worker_order_type === 'Emergency').length}
              </div>
              <div className="text-sm text-gray-600">Emergency</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-blue-600">
                {maintenanceRecords.filter(r => r.worker_order_type === 'Corrective').length}
              </div>
              <div className="text-sm text-gray-600">Corrective</div>
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
                  {sortedRecords.map((record) => {
                    const actionSummary = record.action_taken.length > 100 
                      ? record.action_taken.substring(0, 100) + '...' 
                      : record.action_taken;

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
                            <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${getMaintenanceTypeColor(record.worker_order_type)}`}>
                              {getMaintenanceTypeIcon(record.worker_order_type)}
                              {record.worker_order_type}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-900 max-w-md">
                            <div className="line-clamp-2">
                              {actionSummary}
                            </div>
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
                                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                <span className="text-xs">Download</span>
                              </button>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-center">
                            {record.action_taken.length > 100 && (
                              <button
                                onClick={() => openActionModal(record.work_order_number, record.action_taken)}
                                className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                              >
                                View Details
                              </button>
                            )}
                          </td>
                        </tr>
                      </React.Fragment>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Action Details Modal */}
      {selectedAction && (
        <div className="fixed inset-0 backdrop-blur-md flex items-center justify-center z-50" onClick={closeActionModal}>
          <div className="bg-white rounded-lg border-2 border-gray-800 max-w-2xl w-full mx-4 max-h-96 overflow-hidden" onClick={(e) => e.stopPropagation()}>
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
    </div>
  );
} 