import type { Page } from '@playwright/test';
import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, directoryBody, emptyMaintenanceBody, machine } from './factories';

const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isMachineDetail = (url: URL) => url.pathname === '/api/v1/machines/SN-1';
const isMaintenanceList = (url: URL) => url.pathname === '/api/v1/machines/SN-1/maintenance';
const isDirectory = (url: URL) => url.pathname === '/api/v1/users/directory';

const FREE_TEXT_OPTION = '__free_text__';

const assignedMachine = machine({
  serial_number: 'SN-1',
  assigned_user_id: NON_ADMIN_USER.id,
  assigned_user: {
    id: NON_ADMIN_USER.id,
    username: NON_ADMIN_USER.username,
    email: NON_ADMIN_USER.email,
  },
  person_in_charge: NON_ADMIN_USER.username,
});

/** Open the edit modal for SN-1, which is pre-filled from the machine record. */
async function openEditModal(page: Page) {
  await page.goto('/machines/SN-1');
  await page.getByRole('button', { name: 'Edit Machine' }).click();
  await expect(page.getByRole('heading', { name: 'Edit Machine' })).toBeVisible();
}

const assigneeSelect = (page: Page) => page.getByRole('combobox', { name: /^Assignee/ });
const assigneeFreeText = (page: Page) => page.getByLabel('Assignee name');

test.describe('assignee picker', () => {
  test.beforeEach(async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 0 }));
    await page.route(isMaintenanceList, (route) => fulfilJSON(route, emptyMaintenanceBody));
    await page.route(isDirectory, (route) => fulfilJSON(route, directoryBody));
    // Serves both the initial GET and the PUT the modal issues on save.
    await page.route(isMachineDetail, (route) => fulfilJSON(route, assignedMachine));
  });

  test('is a dropdown pre-selected to the linked user, with no free text in sight', async ({
    page,
  }) => {
    await openEditModal(page);

    const select = assigneeSelect(page);
    await expect(select).toHaveValue(String(NON_ADMIN_USER.id));
    await expect(select.locator('option')).toHaveText([
      'Select a user',
      'e2e-tech (tech@example.com)',
      'other-tech (other@example.com)',
      'Someone else (enter a name)',
    ]);
    await expect(assigneeFreeText(page)).toHaveCount(0);
    await expect(
      page.getByText('Notifications will go to e2e-tech (tech@example.com).')
    ).toBeVisible();
  });

  test('picking another registered user links the machine to that account', async ({ page }) => {
    await openEditModal(page);

    await assigneeSelect(page).selectOption('30');
    await expect(
      page.getByText('Notifications will go to other-tech (other@example.com).')
    ).toBeVisible();

    const save = page.waitForRequest(
      (req) => isMachineDetail(new URL(req.url())) && req.method() === 'PUT'
    );
    await page.getByRole('button', { name: 'Update Machine' }).click();

    const body = (await save).postDataJSON();
    expect(body.assigned_user_id).toBe(30);
    expect(body.person_in_charge).toBe('other-tech');
  });

  test('free text is only used when explicitly chosen', async ({ page }) => {
    await openEditModal(page);

    await assigneeSelect(page).selectOption(FREE_TEXT_OPTION);
    const freeText = assigneeFreeText(page);
    await expect(freeText).toBeVisible();
    await expect(freeText).toHaveValue('');
    await expect(
      page.getByText('This assignee has no account, so they will not receive notifications.')
    ).toBeVisible();

    await freeText.fill('Jane Contractor');

    const save = page.waitForRequest(
      (req) => isMachineDetail(new URL(req.url())) && req.method() === 'PUT'
    );
    await page.getByRole('button', { name: 'Update Machine' }).click();

    const body = (await save).postDataJSON();
    expect(body.assigned_user_id).toBeNull();
    expect(body.person_in_charge).toBe('Jane Contractor');
  });

  test('an empty free-text name blocks the save', async ({ page }) => {
    await openEditModal(page);

    let putRequests = 0;
    page.on('request', (req) => {
      if (isMachineDetail(new URL(req.url())) && req.method() === 'PUT') putRequests += 1;
    });

    await assigneeSelect(page).selectOption(FREE_TEXT_OPTION);
    await page.getByRole('button', { name: 'Update Machine' }).click();

    await expect(page.getByText('Enter the assignee name')).toBeVisible();
    expect(putRequests).toBe(0);
  });

  test('no assignee selected blocks the save', async ({ page }) => {
    await openEditModal(page);

    let putRequests = 0;
    page.on('request', (req) => {
      if (isMachineDetail(new URL(req.url())) && req.method() === 'PUT') putRequests += 1;
    });

    await assigneeSelect(page).selectOption('');
    await page.getByRole('button', { name: 'Update Machine' }).click();

    await expect(page.getByText('Assignee is required')).toBeVisible();
    expect(putRequests).toBe(0);
  });

  test('a free-text assignee reopens in free-text mode', async ({ page }) => {
    await page.unroute(isMachineDetail);
    await page.route(isMachineDetail, (route) =>
      fulfilJSON(
        route,
        machine({
          serial_number: 'SN-1',
          assigned_user_id: null,
          assigned_user: null,
          person_in_charge: 'Jane Contractor',
        })
      )
    );

    await openEditModal(page);

    await expect(assigneeSelect(page)).toHaveValue(FREE_TEXT_OPTION);
    await expect(assigneeFreeText(page)).toHaveValue('Jane Contractor');
  });
});
