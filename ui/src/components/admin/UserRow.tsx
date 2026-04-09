import Image from 'next/image';
import { User } from '@/services/userService';
import { AdminUserService, USER_STATUS, USER_ROLE, UserRole } from '@/services/adminUserService';
import StatusBadge from './StatusBadge';

interface UserRowProps {
  user: User;
  isSelected: boolean;
  onSelectionChange: (userId: number, selected: boolean) => void;
  isUpdating: boolean;
}

export default function UserRow({ 
  user, 
  isSelected, 
  onSelectionChange,
  isUpdating 
}: UserRowProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <tr className={`hover:bg-gray-50 ${isUpdating ? 'opacity-50' : ''}`}>
      {/* Selection Checkbox */}
      <td className="px-6 py-4 whitespace-nowrap">
        <input
          type="checkbox"
          checked={isSelected}
          onChange={(e) => onSelectionChange(user.id, e.target.checked)}
          disabled={isUpdating}
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
    </tr>
  );
}
