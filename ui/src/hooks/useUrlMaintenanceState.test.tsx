import { act, renderHook } from '@testing-library/react';
import { withNuqsTestingAdapter, type UrlUpdateEvent } from 'nuqs/adapters/testing';
import { useUrlMaintenanceState } from './useUrlMaintenanceState';

function setup(searchParams = '') {
  const onUrlUpdate = jest.fn<void, [UrlUpdateEvent]>();
  const view = renderHook(() => useUrlMaintenanceState(), {
    wrapper: withNuqsTestingAdapter({ searchParams, onUrlUpdate }),
  });
  return { ...view, onUrlUpdate };
}

function latestQuery(onUrlUpdate: jest.Mock) {
  const last = onUrlUpdate.mock.calls.at(-1)?.[0] as UrlUpdateEvent | undefined;
  return last?.queryString ?? '';
}

describe('useUrlMaintenanceState', () => {
  it('returns defaults when URL is empty', () => {
    const { result } = setup();
    expect(result.current).toEqual(
      expect.objectContaining({
        q: '',
        sort: 'updated_at_desc',
        page: 1,
        limit: 50,
      }),
    );
  });

  it('hydrates state from URL params', () => {
    const { result } = setup('?q=hello&sort=work_order_date_asc&page=3&limit=100');
    expect(result.current).toEqual(
      expect.objectContaining({
        q: 'hello',
        sort: 'work_order_date_asc',
        page: 3,
        limit: 100,
      }),
    );
  });

  it('falls back to defaults for invalid sort and limit', () => {
    const { result } = setup('?sort=bogus&limit=999');
    expect(result.current.sort).toBe('updated_at_desc');
    expect(result.current.limit).toBe(50);
  });

  it('setSearchQuery writes q and resets page', async () => {
    const { result, onUrlUpdate } = setup('?page=4');
    await act(async () => {
      await result.current.setSearchQuery('abc');
    });
    const q = latestQuery(onUrlUpdate);
    expect(q).toContain('q=abc');
    // page=1 is the default, so it is elided from the URL — the absence of
    // page=4 confirms the reset happened.
    expect(q).not.toContain('page=4');
    expect(q).not.toMatch(/page=\d+/);
  });

  it('setSort writes sort and resets page', async () => {
    const { result, onUrlUpdate } = setup('?page=5');
    await act(async () => {
      await result.current.setSort('work_order_date_desc');
    });
    const q = latestQuery(onUrlUpdate);
    expect(q).toContain('sort=work_order_date_desc');
    expect(q).not.toContain('page=5');
  });

  it('setPage writes the page and leaves other params intact', async () => {
    const { result, onUrlUpdate } = setup('?q=hi&sort=work_order_date_asc');
    await act(async () => {
      await result.current.setPage(2);
    });
    const q = latestQuery(onUrlUpdate);
    expect(q).toContain('page=2');
    expect(q).toContain('q=hi');
    expect(q).toContain('sort=work_order_date_asc');
  });

  it('setPage clamps values below 1 to 1', async () => {
    const { result, onUrlUpdate } = setup('?page=3');
    await act(async () => {
      await result.current.setPage(-5);
    });
    const q = latestQuery(onUrlUpdate);
    // The clamped value (1) is the default, so it disappears from the URL.
    expect(q).not.toMatch(/page=\d+/);
  });

  it('setLimit coerces invalid values and resets page', async () => {
    const { result, onUrlUpdate } = setup('?page=3&limit=100');
    await act(async () => {
      await result.current.setLimit(999);
    });
    const q = latestQuery(onUrlUpdate);
    // 999 is not allowed → falls back to 50 (default, elided). page also resets.
    expect(q).not.toContain('limit=999');
    expect(q).not.toContain('limit=50');
    expect(q).not.toContain('page=3');
  });

  it('setMultiple batches updates into a single URL write', async () => {
    const { result, onUrlUpdate } = setup();
    await act(async () => {
      await result.current.setMultiple({
        q: 'batched',
        sort: 'work_order_date_asc',
        limit: 100,
        page: 2,
      });
    });
    expect(onUrlUpdate).toHaveBeenCalledTimes(1);
    const q = latestQuery(onUrlUpdate);
    expect(q).toContain('q=batched');
    expect(q).toContain('sort=work_order_date_asc');
    expect(q).toContain('limit=100');
    expect(q).toContain('page=2');
  });

  it('setMultiple resets page to 1 when page is omitted', async () => {
    const { result, onUrlUpdate } = setup('?page=9');
    await act(async () => {
      await result.current.setMultiple({ q: 'x' });
    });
    const q = latestQuery(onUrlUpdate);
    expect(q).toContain('q=x');
    expect(q).not.toContain('page=9');
  });
});
