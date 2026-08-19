import type { Route } from '@playwright/test';

/** Fulfil a route with a JSON body. */
export async function fulfilJSON(route: Route, body: unknown, status = 200): Promise<void> {
  await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
}

/** Fulfil a route with an empty 204, as the write endpoints do. */
export async function fulfilNoContent(route: Route): Promise<void> {
  await route.fulfill({ status: 204, body: '' });
}

const ISO = '2026-01-15T10:00:00Z';

export function machine(overrides: Record<string, unknown> = {}) {
  return {
    serial_number: 'SN-1',
    customer: 'Acme',
    state: 'Selangor',
    account_type: 'Standard',
    model: 'M1',
    status: 'Operational',
    brand: 'BrandX',
    district: 'Petaling',
    person_in_charge: 'e2e-tech',
    assigned_user_id: null,
    assigned_user: null,
    reported_by: 'reporter',
    additional_notes: '',
    attachment: '',
    ppm_status: 'ok',
    tnc_date: ISO,
    ppm_date: ISO,
    created_at: ISO,
    updated_at: ISO,
    updated_by: 'e2e-admin',
    maintenance_count: 0,
    ...overrides,
  };
}

/**
 * A machine assigned to a registered user, which is what makes them its
 * notification recipient and lets them resolve its flags.
 */
export function machineAssignedTo(
  user: { id: number; username: string; email: string },
  overrides: Record<string, unknown> = {}
) {
  return machine({
    assigned_user_id: user.id,
    assigned_user: { id: user.id, username: user.username, email: user.email },
    person_in_charge: user.username,
    ...overrides,
  });
}

export function machinesListBody(machines: ReturnType<typeof machine>[]) {
  return {
    machines,
    overdue_count: 0,
    due_count: 0,
    almost_due_count: 0,
    count: machines.length,
    offset: 0,
    limit: 50,
    sort: 'updated_at_desc',
  };
}

export const emptyMaintenanceBody = {
  maintenance: [],
  preventative_count: 0,
  corrective_count: 0,
  emergency_count: 0,
  inspection_count: 0,
  other_count: 0,
  count: 0,
  limit: 10,
  offset: 0,
  sort: 'work_order_date_desc',
};

export function flag(overrides: Record<string, unknown> = {}) {
  return {
    id: 'flag-1',
    machine_serial_number: 'SN-1',
    reason: 'missing_values',
    note: '',
    status: 'open',
    created_by: 10,
    created_by_username: 'e2e-admin',
    created_at: ISO,
    resolved_by: null,
    resolved_by_username: null,
    resolved_at: null,
    ...overrides,
  };
}

export function notification(overrides: Record<string, unknown> = {}) {
  return {
    id: 'n-1',
    user_id: 20,
    type: 'assigned',
    machine_serial_number: 'SN-1',
    flag_id: null,
    flag_status: null,
    title: 'You were assigned to machine SN-1',
    body: 'Customer: Acme',
    actor_user_id: 10,
    actor_username: 'e2e-admin',
    read_at: null,
    created_at: ISO,
    ...overrides,
  };
}

export function notificationsBody(
  notifications: ReturnType<typeof notification>[],
  unreadCount = notifications.filter((n) => !n.read_at).length
) {
  return {
    notifications,
    count: notifications.length,
    unread_count: unreadCount,
    limit: 25,
    offset: 0,
  };
}

export const directoryBody = {
  users: [
    { id: 20, username: 'e2e-tech', email: 'tech@example.com' },
    { id: 30, username: 'other-tech', email: 'other@example.com' },
  ],
};
