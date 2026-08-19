/**
 * Append-only release log for /changelog (newest entry first).
 *
 * To add a release: copy the shape below, add a new object at the **top** of the array,
 * set `date` to ISO `YYYY-MM-DD`, and keep bullets concise.
 */
export type ChangelogSection = {
  heading: string;
  items: string[];
};

export type ChangelogEntry = {
  /** ISO date YYYY-MM-DD (displayed in the user's locale) */
  date: string;
  title: string;
  summary?: string;
  sections: ChangelogSection[];
};

const changelogEntries: ChangelogEntry[] = [
  {
    date: '2026-08-19',
    title: 'Assign machines to people, flag the ones needing attention, and get an inbox',
    summary:
      'A machine can now be assigned to someone with an account, so the app can tell them when it needs doing something. Admins can flag a machine that needs attention, and whoever it is assigned to sees it in a new inbox and can clear the flag once it is sorted.',
    sections: [
      {
        heading: 'Assigning machines',
        items: [
          'Pick an assignee from the list of users when adding or editing a machine, and they are notified straight away — unless you assigned it to yourself.',
          'Someone without an account yet? Choose "Someone else (enter a name)" and type it in as before. Those names still show on the record, but there is nobody to notify.',
        ],
      },
      {
        heading: 'Flagging a machine (admins)',
        items: [
          'A Flag action sits alongside View, Edit and Delete in the records table, on record cards, and on the machine detail page. Choose a reason — missing values, or something else you describe — and add a note of up to 500 characters.',
          'The machine\u2019s assignee is notified if they have an account. A machine carries one flag at a time, so flagging an already-flagged machine updates it rather than adding a second one.',
        ],
      },
      {
        heading: 'Spotting and clearing flags',
        items: [
          'Flagged machines carry an orange badge in the records list and on their card, visible to everyone; hover it for the reason, when it was raised, and the note.',
          'The new Flagged Records page, under your account menu, lists what is outstanding. Admins see every flag; everyone else sees the ones on machines assigned to them.',
          'Resolve a flag from that page or straight from the machine\u2019s row or card, adding a note about what you did if it is worth saying. When you clear a flag an admin raised, they are told it is done.',
        ],
      },
      {
        heading: 'Your inbox',
        items: [
          'A bell in the header shows how many notifications you have not read yet, with the ten most recent a click away. The Inbox page has the full history, 25 at a time, with an Unread only filter.',
          'The inbox is for reading: a notification about a flag that is still open links to where you can clear it.',
        ],
      },
    ],
  },
  {
    date: '2026-08-02',
    title: 'See who last updated a machine or maintenance record',
    summary:
      'Machine and maintenance records now show who last saved them, alongside the existing last-updated timestamp.',
    sections: [
      {
        heading: 'Machines and maintenance',
        items: [
          'An "Updated By" field shows the username of whoever last saved the record — set automatically based on who is signed in, so it cannot be edited by hand.',
          'Visible on machine cards, the machine detail page, and maintenance records once you expand their details.',
          'Included as a column in CSV exports for machines and maintenance.',
        ],
      },
    ],
  },
  {
    date: '2026-04-26',
    title: 'Smoother sessions and sign-in security',
    summary:
      'The app now keeps you signed in more naturally while you use it, and uses a shorter-lived access token behind the scenes. If you stay idle for a long time, you may be asked to sign in again.',
    sections: [
      {
        heading: 'Sessions',
        items: [
          'You get a short-lived access token plus a separate refresh token — you will not notice the difference day to day; the app refreshes your session when needed.',
          'Changing your password signs out other refresh-based sessions tied to your account (as stored on the server).',
        ],
      },
      {
        heading: 'What you might see',
        items: [
          'If you open the app after a very long break, you may need to sign in again — that is expected when the refresh period has ended.',
        ],
      },
    ],
  },
  {
    date: '2026-04-19',
    title: 'Ralts CMS replaces clown-cms',
    summary:
      'Ralts CMS replaces the previous clown-cms web app. Your machines, maintenance history, and attachments were brought over; you can sign in with your existing username and password (your password is upgraded securely on first login).',
    sections: [
      {
        heading: 'Layout and devices',
        items: [
          'Full mobile support — layouts adapt to phones and tablets, with larger touch targets where it matters.',
          'Table and card views for the machine list — switch between a dense table (similar to the old grid) and cards. On small screens, cards are used automatically so you avoid horizontal scrolling.',
          'Column visibility on the table — show or hide columns; your choices are remembered in this browser.',
        ],
      },
      {
        heading: 'Finding and exporting data',
        items: [
          'Improved search — tuned for partial matches and real-world names, not only exact phrases.',
          'CSV export — download machine lists and maintenance history as spreadsheets from the app (respecting filters where supported).',
        ],
      },
      {
        heading: 'Machines and maintenance',
        items: [
          'Maintenance attachments — attach files to individual maintenance records, not only at the machine level.',
          'PPM and due tracking — clearer overdue and due-soon indicators.',
        ],
      },
      {
        heading: 'Account and administration',
        items: [
          'Change your password from the app when signed in.',
          'Admins — updated user management; audit history is available where enabled.',
        ],
      },
      {
        heading: 'What stayed familiar',
        items: [
          'Core workflows — machines, maintenance, attachments, search, and roles — should feel familiar if you used the old app.',
        ],
      },
    ],
  },
];

/** Newest date first (safe if entries are appended out of order). */
export function getChangelogEntries(): ChangelogEntry[] {
  return [...changelogEntries].sort((a, b) => b.date.localeCompare(a.date));
}
