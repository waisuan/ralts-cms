'use client';

import { useState, useMemo } from 'react';
import { Machine } from '../types/machine';
import { MALAYSIAN_STATES, getDistrictsForState, MalaysianState } from '../utils/constants';

interface AddMachineModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAdd: (machine: Omit<Machine, 'created_at' | 'updated_at'>) => void;
}

export default function AddMachineModal({ isOpen, onClose, onAdd }: AddMachineModalProps) {
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
    ppm_status: '',
    tnc_date: '',
    ppm_date: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Get available districts based on selected state
  const availableDistricts = useMemo(() => {
    if (!formData.state) return [];
    return getDistrictsForState(formData.state as MalaysianState);
  }, [formData.state]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    // Required fields validation
    if (!formData.serial_number.trim()) {
      newErrors.serial_number = 'Serial number is required';
    }
    if (!formData.customer.trim()) {
      newErrors.customer = 'Customer is required';
    }
    if (!formData.state) {
      newErrors.state = 'State is required';
    }
    if (!formData.model.trim()) {
      newErrors.model = 'Model is required';
    }
    if (!formData.brand.trim()) {
      newErrors.brand = 'Brand is required';
    }
    if (!formData.district) {
      newErrors.district = 'District is required';
    }
    if (!formData.person_in_charge.trim()) {
      newErrors.person_in_charge = 'Person in charge is required';
    }
    if (!formData.reported_by.trim()) {
      newErrors.reported_by = 'Reported by is required';
    }
    if (!formData.ppm_date) {
      newErrors.ppm_date = 'PPM date is required';
    }

    // Date validation
    if (formData.tnc_date && formData.ppm_date) {
      const tncDate = new Date(formData.tnc_date);
      const ppmDate = new Date(formData.ppm_date);
      if (ppmDate < tncDate) {
        newErrors.ppm_date = 'PPM date cannot be earlier than TNC date';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    const machineData = {
      ...formData,
      attachment: '', // Add empty attachment since it's required by Machine type
    };
    onAdd(machineData);

    // Reset form
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
      ppm_status: '',
      tnc_date: '',
      ppm_date: '',
    });
    setErrors({});
    onClose();
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => {
      const newData = { ...prev, [field]: value };
      // Clear district when state changes since available districts will change
      if (field === 'state') {
        newData.district = '';
      }
      return newData;
    });
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
    // Clear district error when state changes
    if (field === 'state' && errors.district) {
      setErrors((prev) => ({ ...prev, district: '' }));
    }
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
      ppm_status: '',
      tnc_date: '',
      ppm_date: '',
    });
    setErrors({});
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-4xl bg-white rounded-lg shadow-xl">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h2 className="text-xl font-semibold text-gray-900">Add New Machine</h2>
            <button
              onClick={handleClose}
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
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900 ${
                    errors.serial_number ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="Enter serial number"
                />
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
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900"
                  placeholder="Enter account type"
                />
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
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900"
                  placeholder="Enter status"
                />
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
                {errors.brand && <p className="mt-1 text-sm text-red-600">{errors.brand}</p>}
              </div>

              {/* District */}
              <div>
                <label htmlFor="district" className="block text-sm font-medium text-gray-700 mb-2">
                  District <span className="text-red-500">*</span>
                </label>
                <select
                  id="district"
                  value={formData.district}
                  onChange={(e) => handleInputChange('district', e.target.value)}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 ${
                    errors.district ? 'border-red-500' : 'border-gray-300'
                  }`}
                >
                  <option value="">Select district</option>
                  {availableDistricts.map((district) => (
                    <option key={district} value={district}>
                      {district}
                    </option>
                  ))}
                </select>
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
                {errors.reported_by && (
                  <p className="mt-1 text-sm text-red-600">{errors.reported_by}</p>
                )}
              </div>

              {/* TNC Date */}
              <div>
                <label htmlFor="tnc_date" className="block text-sm font-medium text-gray-700 mb-2">
                  TNC Date
                </label>
                <input
                  id="tnc_date"
                  type="date"
                  value={formData.tnc_date}
                  onChange={(e) => handleInputChange('tnc_date', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
                />
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
                    errors.ppm_date ? 'border-red-500' : 'border-gray-300'
                  }`}
                />
                {errors.ppm_date && <p className="mt-1 text-sm text-red-600">{errors.ppm_date}</p>}
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
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-600 text-gray-900"
                placeholder="Enter any additional notes or comments"
              />
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-3 mt-8">
              <button
                type="button"
                onClick={handleClose}
                className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
              >
                Add Machine
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
