import { nextFilterTypeAfterSearchChange } from './bannerSearchLock';

describe('nextFilterTypeAfterSearchChange', () => {
  it('leaves filter when not in banner lock', () => {
    expect(
      nextFilterTypeAfterSearchChange('all', { property: 'ppm_status', query: 'due' })
    ).toBe('all');
  });

  it('stays on overdue when ppm_status query stays overdue', () => {
    expect(
      nextFilterTypeAfterSearchChange('overdue', { property: 'ppm_status', query: 'overdue' })
    ).toBe('overdue');
  });

  it('clears to all when switching property away from ppm_status', () => {
    expect(
      nextFilterTypeAfterSearchChange('overdue', { property: 'any', query: '' })
    ).toBe('all');
    expect(
      nextFilterTypeAfterSearchChange('due', { property: 'any', query: 'x' })
    ).toBe('all');
  });

  it('clears to all when ppm_status query no longer matches banner lock', () => {
    expect(
      nextFilterTypeAfterSearchChange('overdue', { property: 'ppm_status', query: 'due' })
    ).toBe('all');
    expect(
      nextFilterTypeAfterSearchChange('due', { property: 'ppm_status', query: 'overdue' })
    ).toBe('all');
  });

  it('stays on due when ppm_status query stays due', () => {
    expect(
      nextFilterTypeAfterSearchChange('due', { property: 'ppm_status', query: 'due' })
    ).toBe('due');
  });
});
