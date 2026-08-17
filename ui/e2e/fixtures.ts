/* eslint-disable react-hooks/rules-of-hooks -- Playwright fixture API uses a parameter named `use` */
import { test as base, expect, type Page } from '@playwright/test';

export interface SeedUser {
  id: number;
  username: string;
  email: string;
  role: string;
  approved: boolean;
}

/** Roles as the API returns them; the UI compares against these exact strings. */
export const ADMIN_USER: SeedUser = {
  id: 10,
  username: 'e2e-admin',
  email: 'admin@example.com',
  role: 'ADMIN',
  approved: true,
};

export const NON_ADMIN_USER: SeedUser = {
  id: 20,
  username: 'e2e-tech',
  email: 'tech@example.com',
  role: 'NON_ADMIN',
  approved: true,
};

/**
 * Seed the signed-in user for the next navigation. Call before `page.goto`;
 * calling it again overwrites the user, because init scripts run in order on
 * every navigation.
 */
export async function signInAs(page: Page, user: SeedUser): Promise<void> {
  await page.addInitScript((seeded) => {
    localStorage.setItem('ralts_user', JSON.stringify(seeded));
  }, user);
}

/** App shell requires `ralts_user` in localStorage or `/` renders LoginPage only. */
export const test = base.extend({
  page: async ({ page }, use) => {
    await page.addInitScript(() => {
      localStorage.setItem(
        'ralts_user',
        JSON.stringify({
          id: 1,
          username: 'e2e',
          email: 'e2e@example.com',
          role: 'user',
          approved: true,
        })
      );
    });
    await use(page);
  },
});

export { expect };
