import { OverdueStats } from '../hooks/useOverdueStats';

interface OverdueAlertProps {
  stats: OverdueStats;
  onShowOverdue: () => void;
  onShowDue: () => void;
}

export default function OverdueAlert({ stats, onShowOverdue, onShowDue }: OverdueAlertProps) {
  if (stats.overdueCount === 0 && stats.dueCount === 0) {
    return null;
  }

  return (
    <div className="mb-6">
      {/* Critical Alert for Overdue Machines */}
      {stats.overdueCount > 0 && (
        <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-3 rounded-r-md">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-semibold text-red-800">
                  {stats.overdueCount === 1
                    ? '1 machine is overdue'
                    : `${stats.overdueCount} machines are overdue`}{' '}
                  for PPM maintenance
                </p>
                <p className="text-xs text-red-700 mt-1">
                  Immediate attention required to avoid compliance issues
                </p>
              </div>
            </div>
            <button
              onClick={onShowOverdue}
              className="bg-red-100 hover:bg-red-200 text-red-800 px-3 py-1 rounded-md text-sm font-medium transition-colors"
            >
              View Overdue
            </button>
          </div>
        </div>
      )}

      {/* Warning Alert for Due Today */}
      {stats.dueCount > 0 && (
        <div className="bg-orange-50 border-l-4 border-orange-400 p-4 rounded-r-md">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-orange-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-semibold text-orange-800">
                  {stats.dueCount === 1 ? '1 machine is due' : `${stats.dueCount} machines are due`}{' '}
                  for PPM maintenance today
                </p>
                <p className="text-xs text-orange-700 mt-1">
                  Schedule maintenance to avoid becoming overdue
                </p>
              </div>
            </div>
            <button
              onClick={onShowDue}
              className="bg-orange-100 hover:bg-orange-200 text-orange-800 px-3 py-1 rounded-md text-sm font-medium transition-colors"
            >
              View Due
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
