/* eslint-disable react-hooks/rules-of-hooks -- Playwright fixture API uses a parameter named `use` */
import { test as base, expect } from '@playwright/test';

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
