import { ITEMS_PER_PAGE_TABLE } from '@/components/recordsList/recordsListConstants';

export const LIMIT_VALUES = [ITEMS_PER_PAGE_TABLE, 100] as const;
export type LimitValue = (typeof LIMIT_VALUES)[number];

export function coerceLimit(value: number): LimitValue {
  if ((LIMIT_VALUES as readonly number[]).includes(value)) {
    return value as LimitValue;
  }
  if (process.env.NODE_ENV !== 'production') {
    console.warn(
      `coerceLimit: unexpected limit ${value}, falling back to ${ITEMS_PER_PAGE_TABLE}`,
    );
  }
  return ITEMS_PER_PAGE_TABLE;
}

export function clampPage(page: number): number {
  if (!Number.isFinite(page)) return 1;
  return Math.max(1, Math.floor(page));
}
