interface OverdueAlertProps {
  overdueCount: number;
  dueCount: number;
  onShowOverdue: () => void;
  onShowDue: () => void;
  onDismissOverdue?: () => void;
  onDismissDue?: () => void;
  isOverdueDismissed?: boolean;
  isDueDismissed?: boolean;
}

export default function OverdueAlert({
  overdueCount,
  dueCount,
  onShowOverdue,
  onShowDue,
  onDismissOverdue,
  onDismissDue,
  isOverdueDismissed = false,
  isDueDismissed = false,
}: OverdueAlertProps) {
  // If both are dismissed or there are no alerts to show, don't render anything
  if (
    (overdueCount === 0 || isOverdueDismissed) &&
    (dueCount === 0 || isDueDismissed)
  ) {
    return null;
  }

  return (
    <div className="mb-6">
      {/* Critical Alert for Overdue Machines */}
      {overdueCount > 0 && !isOverdueDismissed && (
        <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-3 rounded-r-md">
          <div className="flex items-center justify-between">
            <div className="flex items-center flex-1">
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
              <div className="ml-3 flex-1">
                <p className="text-sm font-semibold text-red-800">
                  {overdueCount === 1
                    ? '1 machine is overdue'
                    : `${overdueCount} machines are overdue`}{' '}
                  for PPM maintenance
                </p>
                <p className="text-xs text-red-700 mt-1">
                  Immediate attention required to avoid compliance issues
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={onShowOverdue}
                className="bg-red-100 hover:bg-red-200 text-red-800 px-3 py-1 rounded-md text-sm font-medium transition-colors"
              >
                View Overdue
              </button>
              {onDismissOverdue && (
                <button
                  onClick={onDismissOverdue}
                  className="text-red-400 hover:text-red-600 transition-colors p-1"
                  title="Dismiss alert"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Warning Alert for Due Today */}
      {dueCount > 0 && !isDueDismissed && (
        <div className="bg-orange-50 border-l-4 border-orange-400 p-4 rounded-r-md">
          <div className="flex items-center justify-between">
            <div className="flex items-center flex-1">
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
              <div className="ml-3 flex-1">
                <p className="text-sm font-semibold text-orange-800">
                  {dueCount === 1 ? '1 machine is due' : `${dueCount} machines are due`}{' '}
                  for PPM maintenance today
                </p>
                <p className="text-xs text-orange-700 mt-1">
                  Schedule maintenance to avoid becoming overdue
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={onShowDue}
                className="bg-orange-100 hover:bg-orange-200 text-orange-800 px-3 py-1 rounded-md text-sm font-medium transition-colors"
              >
                View Due
              </button>
              {onDismissDue && (
                <button
                  onClick={onDismissDue}
                  className="text-orange-400 hover:text-orange-600 transition-colors p-1"
                  title="Dismiss alert"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
