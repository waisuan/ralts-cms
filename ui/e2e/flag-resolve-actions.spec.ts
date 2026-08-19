import type { Page } from '@playwright/test';
import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, fulfilNoContent, flag, machineAssignedTo, machinesListBody } from './factories';

const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isMachinesList = (url: URL) => url.pathname === '/api/v1/machines';
const isOpenFlagsBatch = (url: URL) => url.pathname === '/api/v1/machines/flags/open-by-machine';
const isResolve = (url: URL) =>
  url.pathname === '/api/v1/machines/SN-1/flags/flag-1/resolve';

const openFlag = flag({ note: 'District is blank' });

/** SN-1 assigned to the non-admin, so they may clear its flag themselves. */
const myMachine = machineAssignedTo(NON_ADMIN_USER);

/** The same machine assigned to a colleague instead. */
const someoneElsesMachine = machineAssignedTo({
  id: 30,
  username: 'other-tech',
  email: 'other@example.com',
});

test.describe('resolving a flag from the record actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 0 }));
  });

  /**
   * Serves the open flag until it is resolved, then nothing, so a badge that
   * disappears proves the list refreshed itself off the back of the resolution.
   */
  async function routeFlagUntilResolved(page: Page) {
    let resolved = false;
    await page.route(isOpenFlagsBatch, (route) =>
      fulfilJSON(route, { flags: resolved ? {} : { 'SN-1': [openFlag] } })
    );
    await page.route(isResolve, (route) => {
      resolved = true;
      return fulfilNoContent(route);
    });
  }

  test('an assignee resolves their machine from the list row actions', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isMachinesList, (route) => fulfilJSON(route, machinesListBody([myMachine])));
    await routeFlagUntilResolved(page);
    const resolveRequest = page.waitForRequest(
      (req) => isResolve(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/');
    await expect(page.getByTitle(/missing_values · flagged/)).toBeVisible();

    await page.getByTitle('Actions').click();
    await page.getByRole('button', { name: 'Resolve flag' }).click();

    // The same modal the flagged records page uses, so the note travels with it.
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Resolve Flag' })).toBeVisible();
    await expect(dialog.getByText('District is blank')).toBeVisible();
    await dialog.getByLabel(/Resolution note/).fill('Filled in the district');
    await dialog.getByRole('button', { name: 'Resolve Flag' }).click();

    expect((await resolveRequest).postDataJSON()).toEqual({ note: 'Filled in the district' });
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(page.getByTitle(/missing_values/)).toHaveCount(0);
  });

  test('an assignee resolves their machine from the card actions', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.addInitScript(() => localStorage.setItem('ralts-view-mode', 'cards'));
    await page.route(isMachinesList, (route) => fulfilJSON(route, machinesListBody([myMachine])));
    await routeFlagUntilResolved(page);
    const resolveRequest = page.waitForRequest(
      (req) => isResolve(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/');
    await expect(page.getByTitle(/missing_values · flagged/)).toBeVisible();

    await page.getByRole('button', { name: 'Resolve' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Resolve Flag' })).toBeVisible();
    await dialog.getByRole('button', { name: 'Resolve Flag' }).click();

    // Leaving the note blank is allowed; the flag closes with none.
    expect((await resolveRequest).postDataJSON()).toEqual({ note: '' });
    await expect(page.getByTitle(/missing_values/)).toHaveCount(0);
  });

  test('an admin resolves a flag on a machine assigned to somebody else', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isMachinesList, (route) =>
      fulfilJSON(route, machinesListBody([someoneElsesMachine]))
    );
    await routeFlagUntilResolved(page);

    await page.goto('/');
    await expect(page.getByTitle(/missing_values · flagged/)).toBeVisible();

    await page.getByTitle('Actions').click();
    // Admins get both: raising a flag rewrites the open one, resolving closes it.
    await expect(page.getByRole('button', { name: 'Flag', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Resolve flag' })).toBeVisible();
  });

  test('a non-admin gets no resolve action on a colleague\u2019s machine', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isMachinesList, (route) =>
      fulfilJSON(route, machinesListBody([someoneElsesMachine]))
    );
    await page.route(isOpenFlagsBatch, (route) =>
      fulfilJSON(route, { flags: { 'SN-1': [openFlag] } })
    );

    await page.goto('/');
    // The badge still shows: reading a flag is open to everyone, clearing it is not.
    await expect(page.getByTitle(/missing_values · flagged/)).toBeVisible();

    await page.getByTitle('Actions').click();
    await expect(page.getByRole('button', { name: 'Edit' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Resolve flag' })).toHaveCount(0);
  });

  test('an unflagged machine offers nothing to resolve', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isMachinesList, (route) => fulfilJSON(route, machinesListBody([myMachine])));
    await page.route(isOpenFlagsBatch, (route) => fulfilJSON(route, { flags: {} }));

    await page.goto('/');
    await expect(page.getByRole('link', { name: 'SN-1' })).toBeVisible();

    await page.getByTitle('Actions').click();
    await expect(page.getByRole('button', { name: 'Edit' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Resolve flag' })).toHaveCount(0);
  });
});
