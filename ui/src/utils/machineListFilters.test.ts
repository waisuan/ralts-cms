import { buildMachineListFilters } from './machineListFilters';

describe('buildMachineListFilters', () => {
  const base = {
    searchOptions: { query: '', property: 'any' },
    ppmDateRange: { from: undefined as string | undefined, to: undefined as string | undefined },
    tncDateRange: { from: undefined as string | undefined, to: undefined as string | undefined },
  };

  it('sets ppm_status_filter from banner filterType', () => {
    expect(
      buildMachineListFilters({ ...base, filterType: 'overdue', sortBy: 'newest' })
    ).toMatchObject({ ppm_status_filter: 'overdue', sort: 'updated_at_desc' });
    expect(
      buildMachineListFilters({ ...base, filterType: 'due', sortBy: 'newest' })
    ).toMatchObject({ ppm_status_filter: 'due' });
  });

  it('does not apply search q when banner lock is active', () => {
    expect(
      buildMachineListFilters({
        ...base,
        filterType: 'overdue',
        sortBy: 'newest',
        searchOptions: { query: 'acme', property: 'any' },
      })
    ).not.toHaveProperty('q');
  });

  it('applies search when filterType is all', () => {
    expect(
      buildMachineListFilters({
        ...base,
        filterType: 'all',
        sortBy: 'newest',
        searchOptions: { query: '  SN-1  ', property: 'any' },
      })
    ).toMatchObject({ q: 'SN-1' });
    expect(
      buildMachineListFilters({
        ...base,
        filterType: 'all',
        sortBy: 'newest',
        searchOptions: { query: 'overdue', property: 'ppm_status' },
      })
    ).toMatchObject({ ppm_status_filter: 'overdue' });
  });

  it('maps sort newest/oldest and passes through ppm/tnc sorts', () => {
    expect(
      buildMachineListFilters({ ...base, filterType: 'all', sortBy: 'newest' })
    ).toMatchObject({ sort: 'updated_at_desc' });
    expect(
      buildMachineListFilters({ ...base, filterType: 'all', sortBy: 'oldest' })
    ).toMatchObject({ sort: 'updated_at_asc' });
    expect(
      buildMachineListFilters({ ...base, filterType: 'all', sortBy: 'ppm_date_asc' })
    ).toMatchObject({ sort: 'ppm_date_asc' });
  });

  it('includes date range fields when set', () => {
    expect(
      buildMachineListFilters({
        ...base,
        filterType: 'all',
        sortBy: 'newest',
        ppmDateRange: { from: '2024-01-01', to: '2024-01-31' },
        tncDateRange: { from: '2024-02-01', to: undefined },
      })
    ).toMatchObject({
      ppm_date_from: '2024-01-01',
      ppm_date_to: '2024-01-31',
      tnc_date_from: '2024-02-01',
    });
  });
});
