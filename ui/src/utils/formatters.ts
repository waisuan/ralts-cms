export function formatDate(dateString: string): string {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function formatDateTime(dateString: string): string {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function getTypeColor(type: string): string {
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
}
