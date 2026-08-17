import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, flag, machine, machinesListBody } from './factories';

const isMachinesList = (url: URL) => url.pathname === '/api/v1/machines';
const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isOpenFlagsBatch = (url: URL) => url.pathname === '/api/v1/machines/flags/open-by-machine';

const rows = [machine({ serial_number: 'SN-1' }), machine({ serial_number: 'SN-2' })];

test.describe('flag badges on the machines list', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(isMachinesList, (route) => fulfilJSON(route, machinesListBody(rows)));
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 0 }));
  });

  test('an admin sees a badge on the flagged machine only', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isOpenFlagsBatch, (route) =>
      fulfilJSON(route, {
        flags: {
          'SN-1': [flag({ reason: 'missing_values', note: 'District is blank' })],
        },
      })
    );

    await page.goto('/');
    await expect(page.getByRole('link', { name: 'SN-1' })).toBeVisible();

    // The badge carries the reason, the raise time and the note in its tooltip so
    // a machine can be triaged straight from the table.
    const badge = page.getByTitle(/missing_values · flagged [\s\S]+District is blank/);
    await expect(badge).toBeVisible();
    await expect(badge).toHaveText('1');
    await expect(page.getByTitle(/missing_values/)).toHaveCount(1);
  });

  test('a non-admin sees the same badges, on any machine', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isOpenFlagsBatch, (route) =>
      fulfilJSON(route, { flags: { 'SN-1': [flag({ note: 'District is blank' })] } })
    );

    await page.goto('/');
    await expect(page.getByRole('link', { name: 'SN-2' })).toBeVisible();

    // Seeing that a machine is flagged is not gated on assignment or on being
    // able to raise the flag in the first place.
    const badge = page.getByTitle(/missing_values · flagged [\s\S]+District is blank/);
    await expect(badge).toBeVisible();
    await expect(badge).toHaveText('1');
  });

  test('a signed-out visitor sees no badges and the batch request is never made', async ({
    page,
  }) => {
    // Runs after the fixture's seed script, leaving the page unauthenticated.
    await page.addInitScript(() => localStorage.removeItem('ralts_user'));
    let batchRequests = 0;
    await page.route(isOpenFlagsBatch, (route) => {
      batchRequests += 1;
      return fulfilJSON(route, { flags: { 'SN-1': [flag()] } });
    });

    await page.goto('/');

    await expect(page.getByTitle(/missing_values/)).toHaveCount(0);
    expect(batchRequests).toBe(0);
  });

  test('an admin viewing unflagged machines sees no badges', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isOpenFlagsBatch, (route) => fulfilJSON(route, { flags: {} }));

    await page.goto('/');
    await expect(page.getByRole('link', { name: 'SN-1' })).toBeVisible();
    await expect(page.getByTitle(/missing_values/)).toHaveCount(0);
  });
});
