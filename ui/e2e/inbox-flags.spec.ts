import { test, expect, signInAs, ADMIN_USER, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, fulfilNoContent, notification, notificationsBody } from './factories';

const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isNotificationsList = (url: URL) => url.pathname === '/api/v1/notifications';
const isResolve = (url: URL) => /\/flags\/[^/]+\/resolve$/.test(url.pathname);
const isMarkRead = (url: URL) => url.pathname === '/api/v1/notifications/n-1/read';

const openFlagNotification = notification({
  id: 'n-1',
  type: 'flagged',
  flag_id: 'flag-1',
  flag_status: 'open',
  title: 'Machine SN-1 flagged: missing values',
  body: 'District is blank',
});

test.describe('inbox is read-only for flags', () => {
  test.beforeEach(async ({ page }) => {
    // A non-admin assignee: they read the request here and act on the flags page.
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 1 }));
  });

  test('an open flag links to the flagged records page instead of resolving inline', async ({
    page,
  }) => {
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([openFlagNotification]))
    );
    let resolveRequests = 0;
    await page.route(isResolve, (route) => {
      resolveRequests += 1;
      return fulfilNoContent(route);
    });
    await page.route(
      (url) => url.pathname === '/api/v1/flags',
      (route) => fulfilJSON(route, { flags: [], count: 0, limit: 50, offset: 0, scope: 'assigned' })
    );

    await page.goto('/inbox');
    await expect(page.getByText('Machine SN-1 flagged: missing values')).toBeVisible();

    const link = page.getByRole('link', { name: 'Resolve in flagged records' });
    await expect(link).toHaveAttribute('href', '/flags');
    await expect(page.getByRole('button', { name: /resolve/i })).toHaveCount(0);

    await link.click();
    await expect(page).toHaveURL('/flags');
    expect(resolveRequests).toBe(0);
  });

  test('an already resolved flag is marked resolved with no action', async ({ page }) => {
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(
        route,
        notificationsBody([notification({ ...openFlagNotification, flag_status: 'resolved' })])
      )
    );

    await page.goto('/inbox');
    await expect(page.getByText('Flag resolved')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Resolve in flagged records' })).toHaveCount(0);
  });

  test('an assignment notification offers no flag action', async ({ page }) => {
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([notification({ id: 'n-1' })]))
    );

    await page.goto('/inbox');
    await expect(page.getByText('You were assigned to machine SN-1')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Resolve in flagged records' })).toHaveCount(0);
  });

  test('marking a notification read updates the row and the counter', async ({ page }) => {
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([notification({ id: 'n-1' })]))
    );
    const markRead = page.waitForRequest(
      (req) => isMarkRead(new URL(req.url())) && req.method() === 'POST'
    );
    await page.route(isMarkRead, fulfilNoContent);

    await page.goto('/inbox');
    await expect(page.getByText('1 unread')).toBeVisible();

    await page.getByRole('button', { name: 'Mark as read' }).click();
    await markRead;

    await expect(page.getByText('All caught up')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Mark as read' })).toHaveCount(0);
  });

  test('mark-all-read is offered only while something is unread', async ({ page }) => {
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([notification({ id: 'n-1' })]))
    );
    const markAllRead = page.waitForRequest(
      (req) =>
        new URL(req.url()).pathname === '/api/v1/notifications/read-all' && req.method() === 'POST'
    );
    await page.route((url) => url.pathname === '/api/v1/notifications/read-all', fulfilNoContent);

    await page.goto('/inbox');
    const button = page.getByRole('button', { name: 'Mark all read' });
    await expect(button).toBeVisible();

    await button.click();
    await markAllRead;

    // Hidden once read, so no unreadable greyed-out button is left behind.
    await expect(button).toHaveCount(0);
  });

  test('an empty inbox shows the all-caught-up state once', async ({ page }) => {
    await page.route(isNotificationsList, (route) => fulfilJSON(route, notificationsBody([], 0)));

    await page.goto('/inbox');
    await expect(page.getByText('All caught up 🎉')).toBeVisible();
    // Nothing to mark, so the action is hidden rather than shown disabled.
    await expect(page.getByRole('button', { name: 'Mark all read' })).toHaveCount(0);
  });
});

test.describe('the raising admin hears about a resolution', () => {
  test('the notification names the resolver and carries their note', async ({ page }) => {
    await signInAs(page, ADMIN_USER);
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 1 }));
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(
        route,
        notificationsBody([
          notification({
            id: 'n-1',
            user_id: ADMIN_USER.id,
            type: 'flag_resolved',
            flag_id: 'flag-1',
            flag_status: 'resolved',
            title: 'Machine SN-1: flag resolved',
            body: 'Filled in the district',
            actor_user_id: NON_ADMIN_USER.id,
            actor_username: NON_ADMIN_USER.username,
          }),
        ])
      )
    );

    await page.goto('/inbox');
    await expect(page.getByText('Machine SN-1: flag resolved')).toBeVisible();
    await expect(page.getByText('Filled in the district')).toBeVisible();
    await expect(page.getByText(`By ${NON_ADMIN_USER.username}`)).toBeVisible();
    // The type chip carries the status, so the separate badge would be a repeat.
    await expect(page.getByText('Flag resolved', { exact: true })).toHaveCount(1);
    await expect(page.getByRole('link', { name: 'Resolve in flagged records' })).toHaveCount(0);
  });
});
