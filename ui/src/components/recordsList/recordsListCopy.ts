import type { MachineListFilterType } from '@/utils/machineListFilters';

export function getRecordsListFilterStatusSuffix(filterType: MachineListFilterType): string {
  if (filterType === 'overdue') return ' (overdue only)';
  if (filterType === 'due') return ' (due today only)';
  return '';
}

export function getRecordsListEmptyState(filterType: MachineListFilterType, searchQuery: string) {
  if (filterType === 'overdue') {
    return {
      title: 'No overdue machines found',
      subtitle: 'Great! All machines are up to date with their PPM maintenance.',
    };
  }
  if (filterType === 'due') {
    return {
      title: 'No machines due today',
      subtitle: 'No machines require PPM maintenance today.',
    };
  }
  if (searchQuery) {
    return {
      title: 'No machines found',
      subtitle: `No machines match "${searchQuery}". Try a different search term.`,
    };
  }
  return {
    title: 'No machines found',
    subtitle: 'Get started by creating your first machine.',
  };
}
