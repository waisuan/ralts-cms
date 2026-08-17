import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, fulfilNoContent, flag } from './factories';

const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isFlagsList = (url: URL) => url.pathname === '/api/v1/flags';
const isResolve = (url: URL) =>
  /^\/api\/v1\/machines\/SN-1\/flags\/flag-1\/resolve$/.test(url.pathname);

function flagsBody(flags: ReturnType<typeof flag>[], scope: 'all' | 'assigned' = 'all') {
  return { flags, count: flags.length, limit: 50, offset: 0, scope };
}

const emptyMachines = {
  machines: [],
  overdue_count: 0,
  due_count: 0,
  almost_due_count: 0,
  count: 0,
  offset: 0,
  limit: 50,
  sort: 'updated_at_desc',
};

test.describe('flagged records page', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 0 }));
  });

  test('an admin sees flags across machines and can reach the machine', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isFlagsList, (route) =>
      fulfilJSON(route, flagsBody([flag({ note: 'District is blank' })]))
    );

    await page.goto('/flags');

    await expect(page.getByRole('heading', { name: 'Flagged records' })).toBeVisible();
    await expect(page.getByText('Every machine flagged for follow-up.')).toBeVisible();
    await expect(page.getByText('Missing values')).toBeVisible();
    await expect(page.getByText('District is blank')).toBeVisible();
    await expect(page.locator('p', { hasText: /^Flagged .+ by e2e-admin$/ })).toBeVisible();
    await expect(page.getByRole('link', { name: 'SN-1' })).toHaveAttribute(
      'href',
      '/machines/SN-1'
    );
  });

  test('a non-admin sees only their own flags and resolves one', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    let resolved = false;
    await page.route(isFlagsList, (route) =>
      fulfilJSON(
        route,
        flagsBody(resolved ? [] : [flag({ note: 'District is blank' })], 'assigned')
      )
    );
    await page.route(isResolve, (route) => {
      resolved = true;
      return fulfilNoContent(route);
    });
    const resolveRequest = page.waitForRequest(
      (req) => isResolve(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/flags');

    await expect(page.getByRole('heading', { name: 'My flagged records' })).toBeVisible();
    await expect(page.getByText('Flags raised on machines assigned to you.')).toBeVisible();

    await page.getByRole('button', { name: 'Mark resolved' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText('SN-1 · Missing values')).toBeVisible();
    await dialog.getByLabel(/resolution note/i).fill('Filled in the district');
    await dialog.getByRole('button', { name: 'Resolve Flag' }).click();

    expect((await resolveRequest).postDataJSON()).toEqual({ note: 'Filled in the district' });
    await expect(page.getByText('No open flags. Nothing needs attention.')).toBeVisible();
  });

  test('the page defaults to open flags and can switch to resolved', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    const requestedStatuses: (string | null)[] = [];
    await page.route(isFlagsList, (route) => {
      const status = new URL(route.request().url()).searchParams.get('status');
      requestedStatuses.push(status);
      return fulfilJSON(
        route,
        flagsBody(
          status === 'resolved'
            ? [
                flag({
                  status: 'resolved',
                  resolved_by_username: 'e2e-admin',
                  resolved_at: '2024-06-03T08:30:00Z',
                  resolution_note: 'Filled in the district',
                }),
              ]
            : [flag({ note: 'District is blank' })]
        )
      );
    });

    await page.goto('/flags');
    await expect(page.getByText('District is blank')).toBeVisible();

    await page.getByRole('button', { name: 'Resolved', exact: true }).click();

    // Both events are dated, so a resolved row is not read as having been
    // resolved the moment it was raised.
    await expect(page.locator('p', { hasText: /^Flagged .+ by e2e-admin$/ })).toBeVisible();
    await expect(page.locator('p', { hasText: /^Resolved .+ by e2e-admin$/ })).toBeVisible();
    await expect(page.getByText('Filled in the district')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Mark resolved' })).toHaveCount(0);
    expect(requestedStatuses).toEqual(['open', 'resolved']);
  });

  test('an admin resolves a flag and the list stops showing it', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    let resolved = false;
    await page.route(isFlagsList, (route) =>
      fulfilJSON(route, flagsBody(resolved ? [] : [flag({ note: 'District is blank' })]))
    );
    await page.route(isResolve, (route) => {
      resolved = true;
      return fulfilNoContent(route);
    });
    const resolveRequest = page.waitForRequest(
      (req) => isResolve(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/flags');
    await page.getByRole('button', { name: 'Mark resolved' }).click();
    // The note is optional, so confirming straight away is a valid resolution.
    await page.getByRole('dialog').getByRole('button', { name: 'Resolve Flag' }).click();

    expect((await resolveRequest).postDataJSON()).toEqual({ note: '' });
    await expect(page.getByText('No open flags. Nothing needs attention.')).toBeVisible();
  });

  test('a signed-out visitor gets the sign-in screen and no flags are fetched', async ({
    page,
  }) => {
    // Runs after the fixture's seed script, leaving the page unauthenticated.
    await page.addInitScript(() => localStorage.removeItem('ralts_user'));
    let flagRequests = 0;
    await page.route(isFlagsList, (route) => {
      flagRequests += 1;
      return fulfilJSON(route, flagsBody([flag()]));
    });

    await page.goto('/flags');

    await expect(page.getByRole('heading', { name: 'Sign in to your account' })).toBeVisible();
    await expect(page.getByRole('heading', { name: /flagged records/i })).toHaveCount(0);
    expect(flagRequests).toBe(0);
  });

  test('the user menu links to the page for non-admins too', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(
      (url) => url.pathname === '/api/v1/machines',
      (route) => fulfilJSON(route, emptyMachines)
    );
    await page.route(isFlagsList, (route) => fulfilJSON(route, flagsBody([], 'assigned')));

    await page.goto('/');
    await page.getByRole('button', { name: /e2e-tech/ }).click();
    await page.getByRole('link', { name: 'Flagged Records' }).click();

    await expect(page.getByRole('heading', { name: 'My flagged records' })).toBeVisible();
  });
});
