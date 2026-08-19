import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import {
  fulfilJSON,
  emptyMaintenanceBody,
  flag,
  machine,
  machineAssignedTo,
  machinesListBody,
} from './factories';

const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isMachinesList = (url: URL) => url.pathname === '/api/v1/machines';
const isMachineDetail = (url: URL) => url.pathname === '/api/v1/machines/SN-1';
const isMaintenanceList = (url: URL) => url.pathname === '/api/v1/machines/SN-1/maintenance';
const isMachineFlags = (url: URL) => url.pathname === '/api/v1/machines/SN-1/flags';
const isOpenFlagsBatch = (url: URL) => url.pathname === '/api/v1/machines/flags/open-by-machine';

const assignedMachine = machineAssignedTo(NON_ADMIN_USER);

test.describe('flagging a machine from the record actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 0 }));
    await page.route(isMaintenanceList, (route) => fulfilJSON(route, emptyMaintenanceBody));
    await page.route(isMachineDetail, (route) => fulfilJSON(route, assignedMachine));
  });

  test('an admin flags a machine from the detail page actions', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    // GET is the modal's lookup for an existing open flag; POST raises the flag.
    await page.route(isMachineFlags, (route) => {
      if (route.request().method() === 'GET') return fulfilJSON(route, { flags: [] });
      return fulfilJSON(route, flag({ reason: 'other', note: 'Awaiting customer sign-off' }), 201);
    });
    const createRequest = page.waitForRequest(
      (req) => isMachineFlags(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/machines/SN-1');

    // The flag action sits alongside edit and delete on the record.
    await expect(page.getByRole('button', { name: 'Edit Machine' })).toBeVisible();
    await page.getByRole('button', { name: 'Flag Machine' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Flag Machine' })).toBeVisible();
    await expect(dialog.getByText('SN-1')).toBeVisible();

    await dialog.getByLabel(/Reason/).selectOption('other');
    await dialog.getByLabel(/^Note/).fill('Awaiting customer sign-off');
    await dialog.getByRole('button', { name: 'Flag Machine' }).click();

    const request = await createRequest;
    expect(request.postDataJSON()).toEqual({
      reason: 'other',
      note: 'Awaiting customer sign-off',
    });
    await expect(page.getByRole('dialog')).toHaveCount(0);
  });

  test('reason "other" requires a note before anything is sent', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    let postRequests = 0;
    await page.route(isMachineFlags, (route) => {
      if (route.request().method() === 'GET') return fulfilJSON(route, { flags: [] });
      postRequests += 1;
      return fulfilJSON(route, flag(), 201);
    });

    await page.goto('/machines/SN-1');
    await page.getByRole('button', { name: 'Flag Machine' }).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/Reason/).selectOption('other');
    await dialog.getByRole('button', { name: 'Flag Machine' }).click();

    await expect(dialog.getByText('A note is required when the reason is "Other".')).toBeVisible();
    expect(postRequests).toBe(0);
    await expect(dialog).toBeVisible();
  });

  test('flagging an already-flagged machine updates its open flag', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    const existing = flag({
      reason: 'other',
      note: 'Awaiting parts',
      created_by_username: 'e2e-admin',
    });
    await page.route(isMachineFlags, (route) => {
      if (route.request().method() === 'GET') return fulfilJSON(route, { flags: [existing] });
      // A machine holds one open flag, so the API rewrites it and answers 200.
      return fulfilJSON(route, { ...existing, note: 'Parts arrived, still broken' }, 200);
    });
    const postRequest = page.waitForRequest(
      (req) => isMachineFlags(new URL(req.url())) && req.method() === 'POST'
    );

    await page.goto('/machines/SN-1');
    await page.getByRole('button', { name: 'Flag Machine' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Update Flag' })).toBeVisible();
    // The banner dates the flag being replaced as well as naming who raised it.
    await expect(dialog.getByText(/already flagged .+ by e2e-admin/i)).toBeVisible();
    await expect(dialog.getByLabel(/Reason/)).toHaveValue('other');
    await expect(dialog.getByLabel(/^Note/)).toHaveValue('Awaiting parts');

    await dialog.getByLabel(/^Note/).fill('Parts arrived, still broken');
    await dialog.getByRole('button', { name: 'Update Flag' }).click();

    expect((await postRequest).postDataJSON()).toEqual({
      reason: 'other',
      note: 'Parts arrived, still broken',
    });
    await expect(page.getByRole('dialog')).toHaveCount(0);
  });

  test('a non-admin gets no flag action at all', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);

    await page.goto('/machines/SN-1');
    await expect(page.getByRole('button', { name: 'Edit Machine' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Flag Machine' })).toHaveCount(0);
  });

  test('an admin flags a machine from the list row actions and the badge appears', async ({
    page,
  }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isMachinesList, (route) =>
      fulfilJSON(route, machinesListBody([machine({ serial_number: 'SN-1' })]))
    );

    let created = false;
    await page.route(isMachineFlags, (route) => {
      if (route.request().method() === 'GET') return fulfilJSON(route, { flags: [] });
      created = true;
      return fulfilJSON(route, flag({ note: 'District is blank' }), 201);
    });
    // The badge lookup reflects the new flag only after it has been raised, so
    // seeing a badge proves the list refreshed itself.
    await page.route(isOpenFlagsBatch, (route) =>
      fulfilJSON(route, {
        flags: created ? { 'SN-1': [flag({ note: 'District is blank' })] } : {},
      })
    );

    await page.goto('/');
    await expect(page.getByRole('link', { name: 'SN-1' })).toBeVisible();
    await expect(page.getByTitle(/missing_values/)).toHaveCount(0);

    await page.getByTitle('Actions').click();
    await page.getByRole('button', { name: 'Flag' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: 'Flag Machine' })).toBeVisible();
    await dialog.getByLabel(/^Note/).fill('District is blank');
    await dialog.getByRole('button', { name: 'Flag Machine' }).click();

    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(
      page.getByTitle(/missing_values · flagged [\s\S]+District is blank/)
    ).toBeVisible();
  });

  test('a non-admin sees no flag action in the list row actions', async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isMachinesList, (route) =>
      fulfilJSON(route, machinesListBody([machine({ serial_number: 'SN-1' })]))
    );

    await page.goto('/');
    await page.getByTitle('Actions').click();

    await expect(page.getByRole('button', { name: 'Edit' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Flag' })).toHaveCount(0);
  });
});
