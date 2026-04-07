import { getRecordsListFilterStatusSuffix, getRecordsListEmptyState } from './recordsListCopy';

describe('recordsListCopy', () => {
  it('filter status suffix', () => {
    expect(getRecordsListFilterStatusSuffix('all')).toBe('');
    expect(getRecordsListFilterStatusSuffix('overdue')).toBe(' (overdue only)');
    expect(getRecordsListFilterStatusSuffix('due')).toBe(' (due today only)');
  });

  it('empty state messages', () => {
    expect(getRecordsListEmptyState('overdue', '')).toMatchObject({
      title: 'No overdue machines found',
    });
    expect(getRecordsListEmptyState('all', 'SN-1')).toMatchObject({
      title: 'No machines found',
      subtitle: expect.stringContaining('SN-1'),
    });
  });
});
