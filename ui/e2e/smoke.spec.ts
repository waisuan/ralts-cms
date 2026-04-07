import { test, expect } from './fixtures';

test.describe('smoke', () => {
  test('home loads with search and machines section', async ({ page }) => {
    await page.route('**/api/v1/machines**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          machines: [],
          overdue_count: 0,
          due_count: 0,
          almost_due_count: 0,
          count: 0,
          offset: 0,
          limit: 50,
          sort: 'updated_at_desc',
        }),
      });
    });

    await page.goto('/');

    await expect(page.getByPlaceholder('Search by any...')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Machines' }).first()).toBeVisible();
  });
});
