import { test } from './fixtures';

const machinesListBody = {
  machines: [] as unknown[],
  overdue_count: 2,
  due_count: 1,
  almost_due_count: 0,
  count: 0,
  offset: 0,
  limit: 50,
  sort: 'updated_at_desc',
};

test.describe('home banners', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/machines**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(machinesListBody),
      });
    });
  });

  test('View Overdue triggers list fetch with ppm_status_filter=overdue', async ({ page }) => {
    await page.goto('/');
    await page.waitForResponse((r) => r.url().includes('/api/v1/machines'));

    const filtered = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/machines') &&
        req.url().includes('ppm_status_filter=overdue')
    );
    await page.getByRole('button', { name: 'View Overdue' }).click();
    await filtered;
  });

  test('switching search property to Any clears overdue filter', async ({ page }) => {
    await page.goto('/');
    await page.waitForResponse((r) => r.url().includes('/api/v1/machines'));

    const overdueRequest = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/machines') && req.url().includes('ppm_status_filter=overdue')
    );
    await page.getByRole('button', { name: 'View Overdue' }).click();
    await overdueRequest;

    const afterClear = page.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/machines') && !req.url().includes('ppm_status_filter=overdue')
    );

    await page.getByRole('button', { name: /Search by PPM Status/i }).click();
    await page.getByRole('button', { name: 'Any', exact: true }).click();
    await afterClear;
  });
});
