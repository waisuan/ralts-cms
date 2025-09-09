import { useState, useEffect, useRef } from 'react';
import Image from 'next/image';
import { User } from '@/services/userService';
import { AdminUserService, UserStatus, USER_STATUS, USER_ROLE, UserRole } from '@/services/adminUserService';
import StatusBadge from './StatusBadge';

interface UserRowProps {
  user: User;
  isSelected: boolean;
  onSelectionChange: (userId: number, selected: boolean) => void;
  onStatusUpdate: (userId: number, newStatus: UserStatus) => Promise<void>;
  isUpdating: boolean;
}

export default function UserRow({ 
  user, 
  isSelected, 
  onSelectionChange, 
  onStatusUpdate,
  isUpdating 
}: UserRowProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isLocalUpdating, setIsLocalUpdating] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Handle click outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    if (isDropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => {
        document.removeEventListener('mousedown', handleClickOutside);
      };
    }
  }, [isDropdownOpen]);

  const handleStatusChange = async (newStatus: UserStatus) => {
    setIsLocalUpdating(true);
    setIsDropdownOpen(false);
    
    try {
      await onStatusUpdate(user.id, newStatus);
    } catch (error) {
      console.error('Failed to update user status:', error);
    } finally {
      setIsLocalUpdating(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const isUpdatingRow = isUpdating || isLocalUpdating;

  return (
    <tr className={`hover:bg-gray-50 ${isUpdatingRow ? 'opacity-50' : ''}`}>
      {/* Selection Checkbox */}
      <td className="px-6 py-4 whitespace-nowrap">
        <input
          type="checkbox"
          checked={isSelected}
          onChange={(e) => onSelectionChange(user.id, e.target.checked)}
          disabled={isUpdatingRow}
          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded disabled:opacity-50"
        />
      </td>

      {/* User Info */}
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center">
          <div className="flex-shrink-0 h-10 w-10">
            {user.avatar ? (
              <Image
                className="h-10 w-10 rounded-full object-cover"
                src={user.avatar}
                alt={user.username}
                width={40}
                height={40}
              />
            ) : (
              <div className="h-10 w-10 rounded-full bg-blue-500 flex items-center justify-center">
                <span className="text-white text-sm font-medium">
                  {user.username.charAt(0).toUpperCase()}
                </span>
              </div>
            )}
          </div>
          <div className="ml-4">
            <div className="text-sm font-medium text-gray-900">{user.username}</div>
            <div className="text-sm text-gray-500">{user.email}</div>
          </div>
        </div>
      </td>

      {/* Role */}
      <td className="px-6 py-4 whitespace-nowrap">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
          user.role === USER_ROLE.ADMIN 
            ? 'bg-purple-100 text-purple-800 border border-purple-200' 
            : 'bg-gray-100 text-gray-800 border border-gray-200'
        }`}>
          {AdminUserService.getRoleLabel(user.role as UserRole)}
        </span>
      </td>

      {/* Status */}
      <td className="px-6 py-4 whitespace-nowrap">
        <StatusBadge status={user.status || USER_STATUS.PENDING_APPROVAL} />
      </td>

      {/* Created Date */}
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
        {formatDate(user.created_at)}
      </td>

      {/* Actions */}
      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
        <div className="relative" ref={dropdownRef}>
          <button
            type="button"
            onClick={() => setIsDropdownOpen(!isDropdownOpen)}
            disabled={isUpdatingRow}
            className="inline-flex items-center px-3 py-1 border border-gray-300 shadow-sm text-xs font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {isUpdatingRow ? (
              <>
                <svg className="animate-spin -ml-1 mr-2 h-3 w-3 text-gray-500" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Updating...
              </>
            ) : (
              <>
                Change Status
                <svg className="ml-1 h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </>
            )}
          </button>

          {isDropdownOpen && !isUpdatingRow && (
            <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg z-50 border border-gray-200">
              <div className="py-1">
                {Object.values(USER_STATUS).map((status) => (
                  <button
                    key={status}
                    onClick={() => handleStatusChange(status)}
                    className={`block w-full text-left px-4 py-2 text-sm hover:bg-gray-100 focus:outline-none focus:bg-gray-100 ${
                      user.status === status ? 'bg-blue-50 text-blue-700' : 'text-gray-700'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <StatusBadge status={status} />
                      {user.status === status && (
                        <svg className="h-4 w-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                        </svg>
                      )}
                    </div>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </td>
    </tr>
  );
}
