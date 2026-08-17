# Notifications and flagging

Ralts CMS lets you **assign a machine to someone** and **flag machines that need attention**. When the assignee is a registered user, they are told about both events through an **in-app inbox**: a bell in the header with an unread count, plus a full `/inbox` page. Flags are durable records with an `open` → `resolved` lifecycle, so the machine carries a badge until the flag is cleared.

There is no email or websocket delivery, and no background polling. Inbox state is read when a page loads.

## Concepts

| Piece | What it is | Where it lives |
|-------|------------|----------------|
| **Assignee** | Who is responsible for a machine. Either a registered user (a real foreign key) or a free-text name for someone with no account yet. | `machines."assignedUserId"` → `users(id)` and `machines."personInCharge"` (migration `000023`). |
| **Flag** | A durable "this machine needs attention" record with a reason and an `open`/`resolved` status. | Table `machine_flags` (migrations `000024`, `000026`, `000027`). |
| **Notification** | One inbox row for one recipient, created when they are assigned a machine, when a machine assigned to them is flagged, or when a flag they raised is resolved by an assignee. | Table `notifications` (migration `000025`). |

### Registered users and free-text assignees

A machine can be assigned two ways, and the API decides between them based on what the client sends:

| Sent | Result | Notifications |
|------|--------|---------------|
| `assigned_user_id` | Linked to that user, who must be an account the directory offers — approved and active. The API derives `personInCharge` from their username and **ignores any `person_in_charge` in the request**. | Yes |
| `person_in_charge` only | Saved as free text, capped at 200 characters to match the column. `assignedUserId` stays `NULL`. | No |

The free-text option exists so you can record who is responsible before their account exists. Such an assignee is sometimes called a *ghost* assignee: they show up everywhere a name is displayed, but there is no user to deliver anything to. Assign the machine to their account once they have one, and notifications start working with no other change.

`personInCharge` is populated in both cases, which is what lets CSV export and the trigram `search_text` column (migration `000020`) read a single field for the assignee name regardless of how the machine was assigned.

## How it's meant to be used

**Assigning a machine.** Open a machine in Add or Edit mode. The *Assignee* field is a dropdown of registered users from the user directory (`GET /api/v1/users/directory`), so the normal path is picking the account that will be notified. Its last entry, *Someone else (enter a name)*, reveals a text box for a free-text assignee — free text is never used unless you ask for it explicitly. A hint under the field says who will be notified, or that a typed name has no account and so will not be notified.

Saving with a registered assignee sends them an `assigned` notification, unless they assigned the machine to themselves, in which case it is skipped.

**Receiving notifications.** The bell in the header shows an unread count read once per page load and again on each navigation. Opening it lists the 10 most recent notifications; clicking one marks it read and jumps to the relevant machine. *Mark all read* clears the badge. The `/inbox` page shows everything with 25 per page, an *Unread only* filter, and per-row and bulk mark-as-read. *Mark all read* is only rendered while something is unread.

The inbox is **read-only** as far as the work itself goes: it tells you what happened and, for a notification about a flag that is still open, links to the flagged records page where the flag can be cleared. Rows whose flag has since been resolved say so.

**Flagging a machine (admin only).** Admins get a *Flag* action wherever a machine can be edited or deleted: the row action menu in the records table, the buttons on a record card, and *Flag Machine* on the machine detail page. It opens a modal in the same style as adding or editing a record: choose a reason, add a note, submit. The machine's current assignee is notified if they are a registered user. Non-admins have no flag action at all.

**A machine carries one open flag at a time.** Flagging a machine that is already flagged does not stack a second flag: it rewrites the open one, keeping its identity while taking the new reason, note, flagger and time. The form reflects that — it opens on the existing reason and note, is titled *Update Flag*, and says when the machine was already flagged and by whom. The assignee is notified again, because a new note is a new instruction. The invariant is enforced by a partial unique index (migration `000027`), so it holds even for concurrent writers, and the API implements the replacement as an upsert against it. Resolving a flag frees the machine, so the next flag is a fresh record and the resolved one stays as history.

**Reviewing and resolving flags.** *Flagged Records* in the user menu (`/flags`) is where flags are worked off, and it is open to everyone. What it lists depends on who is looking, and the API decides:

| Signed in as | Page title | Lists |
|--------------|------------|-------|
| Admin | *Flagged records* | Every flag on every machine. |
| Anyone else | *My flagged records* | Open flags on machines assigned to them — the same set they are allowed to resolve — plus resolved flags **they closed themselves**. |

The second half of that rule matters when a machine changes hands: assignment decides what you can work on, so a machine's open flags follow it, but resolved flags closed by somebody else stay with them rather than appearing as history you never took part in. The page says as much on the *Resolved* and *All* tabs. Admins are unfiltered and still see every resolution. Note this is a tidiness rule, not a confidentiality boundary: `GET /api/v1/machines/{serial}/flags?include_resolved=true` deliberately serves any machine's full flag history to any authenticated user, because it backs the badge and flag-form lookups.

Either way the rows are newest first and show the machine, reason and note. Tabs switch between *Open* (the default), *Resolved* and *All*, and each open row has *Mark resolved*.

**Each event carries its own timestamp.** A row reads *Flagged {when} by {who}* and, once closed, *Resolved {when} by {who}* on a second line, so a resolved flag is never mistaken for one closed the moment it was raised. Both times come from the database clock — `created_at` on insert, `resolved_at` on resolution — so they are directly comparable, and resolving a flag leaves its raise time untouched. Re-flagging a machine is the one case where a time moves: rewriting the open flag re-dates it, because the row now represents the newer request. The flag form says so when it opens on an existing flag, and the badge tooltip carries the raise time too.

*Mark resolved* opens a confirmation modal that repeats why the machine was flagged and offers an **optional resolution note** — a line about what was actually done. Cancelling leaves the flag untouched; confirming with the note blank resolves it with no note. Resolved flags stay listed for the audit trail with `resolved_by`, `resolved_at` and the resolution note, which is shown on the row under the original note.

**Reporting back.** When a non-admin resolves a flag, the admin who raised it gets a `flag_resolved` notification carrying the resolution note as its body, closing the loop on the request they made. Resolutions by admins notify nobody: admins work off the same page across every machine, and reporting each clearance to whichever colleague raised it would be noise rather than news.

Because the scope matches the resolve permission exactly, everything on your page is something you can act on, and there is nothing to resolve anywhere else: machine pages deliberately carry no flags list, and the inbox only points here.

**Spotting flags at a glance.** Machines with open flags show an orange badge with the open-flag count on the records table and card view. Every signed-in user sees it, on any machine and not just their own, because knowing a machine is flagged is useful to whoever is looking at it — raising a flag stays admin-only, reading one does not. Signed-out visitors see no badges and the batch request that backs them is not issued. Hovering shows each reason, when it was flagged, and the note.

## Flag reasons

| Reason | Who can raise it | Notes |
|--------|------------------|-------|
| `missing_values` | Admin | Note is optional; use it to say which fields are missing. |
| `other` | Admin | Note is **required**; a request without one returns `400`. |
| `requires_attention` | Nobody. Reserved for the system. | Creating one through the API returns `400`. Flags of this reason carry a `ppm_status` snapshot; no other reason uses that field. |

Statuses are `open` and `resolved`. Each has its own timestamp: `created_at` for the raise and `resolved_at` for the closure, both stamped by the database rather than by the API process, with `resolved_by`/`resolution_note` alongside the latter.

Notes are capped at **500 characters** (counted as characters, not bytes, so accented and non-Latin text is not penalised). The flag form counts down against that limit and stops accepting input at it; the API enforces the same bound and returns `400` for anything longer, whatever the reason. Notes are trimmed of surrounding whitespace before being stored, so whitespace alone will not satisfy the required note for `other`. The limit exists because the note becomes the body of the notification sent to the assignee.

Resolution notes (`resolution_note`, migration `000026`) follow the same rules with one difference: they are always optional. They are trimmed, bounded at 500 characters, and a whitespace-only note is stored as no note at all.

Every flag is raised deliberately by an admin — nothing in the system opens one on its own. In particular, a machine's PPM status is computed from `ppmDate` and surfaced on records, badges, and filters entirely separately, and never produces a flag.

## Notification types

| Type | Trigger | Recipient |
|------|---------|-----------|
| `assigned` | A machine's `assigned_user_id` is set or changed on create or update. | The new assignee. |
| `flagged` | A flag is created. | The machine's assignee at the moment the flag is created. |
| `flag_resolved` | A **non-admin** resolves a flag. Admin resolutions are silent. | The admin who raised the flag. Nobody, if that account has since been deleted. |

Three rules apply throughout: a notification is **never** sent to the user who caused it (actor equals recipient is dropped), a machine with **no assignee** produces no `assigned`/`flagged` notifications, and a machine with a **free-text assignee** produces none either, because there is no user to address.

The first rule covers the admin-flags-their-own-machine case: the flag itself is created, badged and listed as normal, and can be resolved from the flags page — only the inbox notification is skipped, since it would tell you something you just did.

## API

All routes below require a Bearer access token. Flag routes do their role checks **inside** the handlers rather than via the `/admin/` subrouter: the per-machine routes keep the natural REST nesting under `/machines/{serial}`, and the cross-machine listing is open to every user but scoped by role, which middleware cannot express.

### Notifications

Every endpoint is scoped to the caller's own user ID from the JWT. There is no way to read another user's inbox.

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/v1/notifications` | List. Query: `limit` (1–200, default 50), `offset`, `unread_only=true`. Returns `{ notifications, count, unread_count, limit, offset }`. Rows referencing a flag also carry `flag_status`, which is what lets the inbox distinguish a flag still needing work from one already resolved. |
| `GET` | `/api/v1/notifications/unread-count` | Returns `{ unread_count }`. The cheap endpoint the bell reads on load. |
| `POST` | `/api/v1/notifications/{id}/read` | Marks one notification read. `204` on success, and it is idempotent. |
| `POST` | `/api/v1/notifications/read-all` | Marks all the caller's unread notifications read. Returns `{ updated }`. |

### Flags

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/v1/machines/{serial}/flags` | Any user | Open flags for one machine. Add `?include_resolved=true` for the full history. Returns `{ flags }`. |
| `GET` | `/api/v1/machines/flags/open-by-machine` | Any user | Batch lookup. `?serials=A,B,C` (max 200). Returns `{ flags: { serial: [...] } }`. Backs the table badges with one request per page instead of one per row. |
| `GET` | `/api/v1/flags` | Any user, **scoped by role** | Flags across machines, newest first: every machine for an admin; otherwise machines assigned to the caller, with resolved rows narrowed to the ones the caller resolved. Query: `status=open` (default) \| `resolved` \| `all`, `limit` (1–200, default 50), `offset`. Returns `{ flags, count, limit, offset, scope }`, where `count` is the whole filtered set rather than the page and `scope` is `all` or `assigned`. Backs the *Flagged Records* page. |
| `POST` | `/api/v1/machines/{serial}/flags` | **Admin** | Body `{ "reason": "missing_values" \| "other", "note": "..." }`. `201` when the machine had no open flag, `200` when its open flag was rewritten (same flag `id`, new reason/note); `403` non-admin; `404` unknown machine; `400` invalid reason, missing required note, or a note over 500 characters. |
| `POST` | `/api/v1/machines/{serial}/flags/{id}/resolve` | **Admin or the machine's assignee** | Optional body `{ "note": "..." }` recording what was done; an empty or absent body is valid. A non-admin caller also sends a `flag_resolved` notification to the flag's raiser. `204` on success; `400` malformed body or a note over 500 characters; `403` for a non-admin who is not the current assignee; `404` if the flag does not exist or is already resolved. |

### Assignee

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/v1/users/directory` | Approved, active users as `{ users: [{ id, username, email }] }`, ordered by username. Deliberately slim — no roles, no password material. Available to any authenticated user, unlike the admin-only `GET /api/v1/admin/users`. |

Machine create and update (`POST /api/v1/machines`, `PUT /api/v1/machines/{serial}`) accept either `assigned_user_id` or a free-text `person_in_charge`, as described above. Send both as null/empty to leave the machine unassigned. An `assigned_user_id` that does not resolve to a user returns `400`, as does a `person_in_charge` longer than 200 characters.

Two rules protect an existing assignment on update:

- **Naming a new assignee requires an assignable account** — approved and active, the same condition `GET /users/directory` applies — and anything else returns `400`. Re-sending the assignee a machine already has is always accepted, so a machine whose assignee was later suspended can still be edited without either failing or losing them.
- **A body that mentions neither `assigned_user_id` nor `person_in_charge` leaves the assignee alone.** Otherwise a partial update would read as a free-text assignee, unlink the user, and take their notifications and their right to resolve the machine's flags with it. Sending either field still sets the assignee as normal, including `assigned_user_id: null` to hand the machine to a free-text name, which is what the UI does. Every other field follows normal `PUT` replacement semantics.

## Environment variables

None. This feature has no configuration of its own.

## Database

- Migrations: `db/migrations/000023_add_machine_assignee_fk.*`, `000024_add_machine_flags_table.*`, `000025_add_notifications_table.*`, `000026_add_flag_resolution_note.*`, `000027_one_open_flag_per_machine.*`
- Repositories: `internal/flags/repository.go`, `internal/notifications/repository.go`, `ListDirectory` in `internal/users/repository.go`

`machine_flags` carries a partial unique index on `machine_serial_number WHERE status = 'open'`, so a machine can never hold two open flags. `Repository.Create` inserts against it with `ON CONFLICT ... DO UPDATE`, which makes flagging idempotent by construction: it returns the stored flag's id and a bool saying whether a row was inserted, which is what the handler turns into `201` or `200`. Migration `000027` closed any pre-existing duplicates (keeping the newest per machine) before adding the index.

Deletion behaviour worth knowing: deleting a **machine** cascades to its flags and notifications. Deleting a **user** cascades to their notifications (their inbox goes away) but only nulls them out as a machine assignee or as a flag's `created_by`/`resolved_by`, so flag history survives. A machine whose assignee is deleted becomes unassigned, keeping whatever name is in `personInCharge`.

## Frontend behavior

- `ui/src/components/NotificationBell.tsx` — bell, unread badge, dropdown. Reads the unread count on mount and again on each pathname change, so the badge is current as of the page you are on. A failed read is logged at debug level and retried on the next page load.
- `ui/src/app/inbox/page.tsx` — full list, 25 per page, unread filter, mark-read controls, and a link to `/flags` on rows whose flag is still open. No flag is ever mutated from here.
- `ui/src/app/flags/page.tsx` — the *Flagged Records* page, for any signed-in user. It words its own heading from the `scope` the API reports rather than deciding who sees what.
- `ui/src/components/FlagMachineModal.tsx` — the flag dialog, opened from the record actions in `RecordsList`, `RecordsTable`, `RecordCard` and `MachineInfoCard`. Those components render a flag action only when handed an `onFlag` callback, which the parents supply for admins only.
- `ui/src/components/FlagBadge.tsx` + `ui/src/hooks/useOpenFlagsForMachines.ts` — badges, batch-fetched per page of machines. The hook skips the request when nobody is signed in, so that rule lives in one place rather than in each consumer. Its `reloadToken` argument is bumped after a flag is raised so the new badge appears immediately.
- `ui/src/components/MachineModal.tsx` — assignee dropdown of directory users, loaded when the modal opens so newly approved users appear without a reload, plus the explicit free-text option. A machine linked to a user who has dropped out of the directory keeps that user as a selectable option, so editing another field cannot silently unlink them.
- Services: `ui/src/services/notificationService.ts`, `flagService.ts`, `userDirectoryService.ts`.

## Operations

- **Apply migrations before deploying** (`make migrate-up` or your pipeline). The API writes `assignedUserId` on every machine save and will fail without it.
- **Auditing silent assignees.** Machines with a free-text assignee receive no notifications, which is easy to miss because the name still displays everywhere. To list them and decide whether the person now has an account to link:
  ```sql
  SELECT "serialNumber", "personInCharge"
  FROM machines
  WHERE "assignedUserId" IS NULL AND COALESCE("personInCharge", '') <> '';
  ```
- **Notification rows are never pruned.** Unlike audit logs there is no retention job, so the table grows without bound. If it becomes a problem, add a cleanup worker modelled on `internal/audit/service.go`.

## Permissions summary

| Action | Non-admin | Admin |
|--------|-----------|-------|
| See flag badges on the records table and cards | Yes, on any machine | Yes |
| Open the *Flagged Records* page | Yes, listing open flags on machines assigned to them plus resolutions of their own | Yes, listing every flag |
| Assign a machine | Yes | Yes |
| Raise a manual flag | No (no action shown; `403` from the API) | Yes |
| Resolve a flag | Only on machines assigned to them, from the *Flagged Records* page | Any flag |
| Read own inbox | Yes | Yes |
| Read someone else's inbox | No | No |

## Known limitations

- One assignee per machine. There is no concept of a team or watcher list.
- A free-text assignee is never linked to an account after the fact. Creating an account whose username equals an existing free-text name does **not** retroactively link machines; edit each machine and pick the account from the dropdown.
- The assignee dropdown loads the whole directory with no search. Fine at the current user count; it would need a typeahead at a few hundred users.
- The *Flagged Records* page paginates in blocks of 50 with no search or filtering beyond status, on the assumption that open flags stay few. An admin cannot narrow it to a single assignee or machine.
- A non-admin's page follows the **current** assignee, not who was assigned when the flag was raised. Reassigning a machine moves its open flags to the new assignee's page, even though the notification stays in the original assignee's inbox.
- A resolution is reported to the admin who raised the flag and to nobody else. Other admins find out from the *Flagged Records* page, and a flag raised by a since-deleted account has no one to report to.
- Inbox state is only as fresh as the last page load. A notification arriving while you sit on a page will not appear until you navigate or reload.

## Related code

- Flags: `internal/flags/` (model, repository, service), handlers in `internal/handlers/flags.go`
- Notifications: `internal/notifications/`, handlers in `internal/handlers/notifications.go`
- Assignee wiring: `internal/machines/machine.go`, `internal/machines/repository.go`, `resolveAssignee` in `internal/handlers/machines.go`
- Service wiring: `internal/deps/deps.go`
- Router: `internal/router/router.go`
- Unit tests: `internal/flags/{repository,service}_test.go`, `internal/notifications/{repository,service}_test.go`, `internal/handlers/{flags,notifications,machines}_test.go`, `ui/src/app/inbox/page.test.tsx`, `ui/src/app/flags/page.test.tsx`, `ui/src/components/{NotificationBell,MachineModal,FlagMachineModal,ResolveFlagModal}.test.tsx`, `ui/src/hooks/useOpenFlagsForMachines.test.tsx`
- Integration tests: `internal/integration/notifications_flags_test.go` — the assignee, flag permission, and notification delivery flows end to end against Postgres. Docker required:

```bash
go test -tags=integration -count=1 -timeout=15m ./internal/integration/...
```

- Browser tests: `ui/e2e/{notification-bell,inbox-flags,flag-badges,flag-machine-modal,flags-page,machine-assignee}.spec.ts` — the bell, the read-only inbox, badges for both roles, the flag modal and where it can be opened from, the flagged-records page for both roles, and the assignee dropdown, all against mocked API responses:

```bash
cd ui && npm run build && npm run test:e2e
```

## See also

- [Backend deployment](./backend-deployment.md) — running migrations and the full env var table
- [Authentication: access and refresh tokens](./auth-refresh-tokens.md) — how the Bearer token these endpoints require is issued
