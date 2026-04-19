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
