import { test, expect, signInAs, NON_ADMIN_USER } from './fixtures';
import { fulfilJSON, machinesListBody, notification, notificationsBody } from './factories';

// Exact-path predicates rather than globs, so the inbox list, the unread count
// and the per-notification writes never shadow one another.
const isMachinesList = (url: URL) => url.pathname === '/api/v1/machines';
const isUnreadCount = (url: URL) => url.pathname === '/api/v1/notifications/unread-count';
const isNotificationsList = (url: URL) => url.pathname === '/api/v1/notifications';
const isMarkAllRead = (url: URL) => url.pathname === '/api/v1/notifications/read-all';

test.describe('notification bell', () => {
  test.beforeEach(async ({ page }) => {
    await signInAs(page, NON_ADMIN_USER);
    await page.route(isMachinesList, (route) => fulfilJSON(route, machinesListBody([])));
  });

  test('reads the unread count once per page load and never polls', async ({ page }) => {
    // Install a fake clock before navigating so we can jump far enough ahead to
    // catch a leftover interval without waiting in real time.
    await page.clock.install();

    let unreadRequests = 0;
    await page.route(isUnreadCount, (route) => {
      unreadRequests += 1;
      return fulfilJSON(route, { unread_count: 3 });
    });

    await page.goto('/');
    const bell = page.getByRole('button', { name: 'Notifications' });
    await expect(bell).toContainText('3');
    expect(unreadRequests).toBe(1);

    await page.clock.fastForward('10:00');
    await expect(bell).toContainText('3');
    expect(unreadRequests).toBe(1);
  });

  test('refreshes the count when the user navigates to the inbox', async ({ page }) => {
    let unreadRequests = 0;
    await page.route(isUnreadCount, (route) => {
      unreadRequests += 1;
      // The second read models notifications arriving between page loads.
      return fulfilJSON(route, { unread_count: unreadRequests === 1 ? 1 : 4 });
    });
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([notification()], 4))
    );

    await page.goto('/');
    const bell = page.getByRole('button', { name: 'Notifications' });
    await expect(bell).toContainText('1');

    await bell.click();
    await page.getByRole('link', { name: 'View all' }).click();

    await expect(page.getByRole('heading', { name: 'Inbox' })).toBeVisible();
    await expect(page.getByText('4 unread')).toBeVisible();
    await expect(bell).toContainText('4');
  });

  test('dropdown lists recent notifications and can mark them all read', async ({ page }) => {
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: 2 }));
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(
        route,
        notificationsBody([
          notification({ id: 'n-1' }),
          notification({
            id: 'n-2',
            type: 'flagged',
            flag_id: 'flag-1',
            flag_status: 'open',
            title: 'Machine SN-1 flagged: missing values',
            body: 'Serial plate photo missing',
          }),
        ])
      )
    );

    const markAllRead = page.waitForRequest(
      (req) => req.url().includes('/api/v1/notifications/read-all') && req.method() === 'POST'
    );
    await page.route(isMarkAllRead, (route) => fulfilJSON(route, { updated: 2 }));

    await page.goto('/');
    const bell = page.getByRole('button', { name: 'Notifications' });
    await expect(bell).toContainText('2');
    await bell.click();

    await expect(page.getByText('You were assigned to machine SN-1')).toBeVisible();
    await expect(page.getByText('Machine SN-1 flagged: missing values')).toBeVisible();
    await expect(page.getByText('Serial plate photo missing')).toBeVisible();

    await page.getByRole('button', { name: 'Mark all read' }).click();
    await markAllRead;

    // The badge disappears entirely once nothing is unread.
    await expect(bell).not.toContainText('2');
    await expect(page.getByRole('button', { name: 'Mark all read' })).toBeDisabled();
  });

  test('the badge clears when the inbox page marks everything read', async ({ page }) => {
    let unread = 1;
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: unread }));
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(route, notificationsBody([notification({ read_at: null })], unread))
    );
    await page.route(isMarkAllRead, (route) => {
      unread = 0;
      return fulfilJSON(route, { updated: 1 });
    });

    await page.goto('/inbox');
    const bell = page.getByRole('button', { name: 'Notifications' });
    await expect(bell).toContainText('1');

    // The page's own button, the only one on screen while the bell is closed.
    await page.getByRole('button', { name: 'Mark all read' }).click();

    // No navigation involved: the bell hears about it and re-reads the count.
    await expect(bell).not.toContainText('1');
    await expect(page.getByText('All caught up')).toBeVisible();
  });

  test('marking all read from the bell updates the inbox list behind it', async ({ page }) => {
    let unread = 1;
    await page.route(isUnreadCount, (route) => fulfilJSON(route, { unread_count: unread }));
    await page.route(isNotificationsList, (route) =>
      fulfilJSON(
        route,
        notificationsBody(
          [notification({ read_at: unread === 0 ? '2026-01-15T11:00:00Z' : null })],
          unread
        )
      )
    );
    await page.route(isMarkAllRead, (route) => {
      unread = 0;
      return fulfilJSON(route, { updated: 1 });
    });

    await page.goto('/inbox');
    await expect(page.getByText('1 unread')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Mark as read' })).toBeVisible();

    const bell = page.getByRole('button', { name: 'Notifications' });
    await bell.click();
    // Both the dropdown and the page have a "Mark all read", so reach for the
    // one inside the bell's own container.
    await bell.locator('xpath=..').getByRole('button', { name: 'Mark all read' }).click();
    // Close the dropdown, whose own button is only disabled rather than removed,
    // so the assertions below are about the page's controls.
    await page.getByRole('heading', { name: 'Inbox' }).click();

    // The page re-read its list, so its rows and header agree with the bell.
    await expect(page.getByText('All caught up')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Mark as read' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Mark all read' })).toHaveCount(0);
  });

  test('is hidden when nobody is signed in', async ({ page }) => {
    await page.addInitScript(() => localStorage.removeItem('ralts_user'));
    let unreadRequests = 0;
    await page.route(isUnreadCount, (route) => {
      unreadRequests += 1;
      return fulfilJSON(route, { unread_count: 7 });
    });

    await page.goto('/');
    await expect(page.getByRole('button', { name: 'Notifications' })).toHaveCount(0);
    expect(unreadRequests).toBe(0);
  });
});
