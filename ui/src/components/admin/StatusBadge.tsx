import { AdminUserService, UserStatus, USER_STATUS } from '@/services/adminUserService';

interface StatusBadgeProps {
  status: UserStatus | string;
  className?: string;
}

export default function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  // Ensure we have a valid status
  const validStatus = AdminUserService.isValidStatus(status) ? status : USER_STATUS.PENDING_APPROVAL;
  
  const statusLabel = AdminUserService.getStatusLabel(validStatus);
  const baseClassName = AdminUserService.getStatusClassName(validStatus);

  // Define status-specific styles
  const statusStyles = {
    'status-pending': 'bg-yellow-100 text-yellow-800 border-yellow-200',
    'status-approved': 'bg-green-100 text-green-800 border-green-200',
    'status-suspended': 'bg-red-100 text-red-800 border-red-200',
    'status-inactive': 'bg-gray-100 text-gray-800 border-gray-200',
    'status-unknown': 'bg-gray-100 text-gray-800 border-gray-200',
  };

  const statusStyle = statusStyles[baseClassName as keyof typeof statusStyles] || statusStyles['status-unknown'];

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${statusStyle} ${className}`}
      title={`Status: ${statusLabel}`}
    >
      {statusLabel}
    </span>
  );
}
