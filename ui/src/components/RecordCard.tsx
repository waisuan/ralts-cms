import { Machine } from '../types/machine';

interface RecordCardProps {
  machine: Machine;
  onView: (serial_number: string) => void;
  onEdit: (serial_number: string) => void;
  onDelete: (serial_number: string) => void;
}

function getStatusFromPPMDate(ppm_date: string): { label: string; color: string } | null {
  if (!ppm_date) return null;
  const today = new Date();
  const ppm = new Date(ppm_date);
  // Remove time for accurate day comparison
  today.setHours(0, 0, 0, 0);
  ppm.setHours(0, 0, 0, 0);
  const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
  if (diffDays < 0) {
    return { label: 'Overdue', color: 'bg-red-100 text-red-800' };
  } else if (diffDays === 0) {
    return { label: 'Due', color: 'bg-orange-100 text-orange-800' };
  } else if (diffDays > 0 && diffDays <= 7) {
    return { label: 'Due Soon', color: 'bg-yellow-100 text-yellow-800' };
  }
  return null;
}

export default function RecordCard({ machine, onView, onEdit, onDelete }: RecordCardProps) {
  const status = getStatusFromPPMDate(machine.ppm_date);

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 overflow-hidden">
      <div className="p-6">
        <div className="flex justify-between items-start mb-4">
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900 truncate">
              {machine.model}{' '}
              <span className="text-xs text-gray-500">({machine.serial_number})</span>
            </h3>
            <div className="text-xs text-gray-500 mt-1">
              {machine.brand} &middot; {machine.state}
            </div>
          </div>
          {status && (
            <span className={`px-2 py-1 text-xs font-medium rounded-full ${status.color}`}>
              {status.label}
            </span>
          )}
        </div>
        <div className="mb-4">
          <div className="text-sm text-gray-700 font-medium">
            Customer: <span className="font-normal">{machine.customer}</span>
          </div>
          <div className="text-sm text-gray-700 font-medium">
            Person In Charge: <span className="font-normal">{machine.person_in_charge}</span>
          </div>
          <div className="text-sm text-gray-700 font-medium">
            District: <span className="font-normal">{machine.district}</span>
          </div>
        </div>
        <div className="text-xs text-gray-500 mb-4 space-y-1">
          <div>TNC Date: {formatDate(machine.tnc_date)}</div>
          <div>PPM Date: {formatDate(machine.ppm_date)}</div>
          <div>Created: {formatDate(machine.created_at)}</div>
          <div>Updated: {formatDate(machine.updated_at)}</div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => onView(machine.serial_number)}
            className="flex-1 bg-blue-50 hover:bg-blue-100 text-blue-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            View
          </button>
          <button
            onClick={() => onEdit(machine.serial_number)}
            className="flex-1 bg-yellow-50 hover:bg-yellow-100 text-yellow-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            Edit
          </button>
          <button
            onClick={() => onDelete(machine.serial_number)}
            className="flex-1 bg-red-50 hover:bg-red-100 text-red-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}
