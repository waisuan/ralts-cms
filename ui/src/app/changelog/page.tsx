import { format, parseISO } from 'date-fns';
import { getChangelogEntries } from '@/data/changelog';

export const metadata = {
  title: 'Changelog | Ralts CMS',
  description: 'Release notes and changes in Ralts CMS',
};

function formatReleaseDate(iso: string) {
  return format(parseISO(iso), 'MMMM d, yyyy');
}

function releaseAnchorId(date: string) {
  return `release-${date}`;
}

export default function ChangelogPage() {
  const entries = getChangelogEntries();

  return (
    <div className="container mx-auto max-w-6xl px-4 py-8">
      <h1 className="text-2xl font-semibold text-gray-900">Changelog</h1>
      <p className="mt-2 text-gray-600">
        Release notes for Ralts CMS — newest updates first. Older entries stay below for reference.
      </p>

      <div className="mt-10 flex flex-col gap-10 lg:flex-row lg:items-start lg:gap-12">
        <nav
          aria-label="Jump to release by date"
          className="shrink-0 lg:sticky lg:top-24 lg:w-52 lg:self-start"
        >
          <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">On this page</p>
          <ul className="mt-4 space-y-0">
            {entries.map((entry, i) => (
              <li key={entry.date} className="flex items-stretch gap-3">
                <div className="flex w-4 shrink-0 flex-col items-center self-stretch">
                  <span
                    className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-blue-500 ring-4 ring-gray-50"
                    aria-hidden
                  />
                  {i < entries.length - 1 ? (
                    <span className="mt-1 w-px flex-1 min-h-[2rem] bg-gray-200" aria-hidden />
                  ) : null}
                </div>
                <a
                  href={`#${releaseAnchorId(entry.date)}`}
                  className={`block min-w-0 flex-1 text-sm text-gray-700 transition-colors hover:text-blue-700 ${i < entries.length - 1 ? 'pb-8' : ''}`}
                >
                  <span className="font-medium leading-snug">{formatReleaseDate(entry.date)}</span>
                  <span className="mt-0.5 block text-xs font-normal text-gray-500 line-clamp-2">
                    {entry.title}
                  </span>
                </a>
              </li>
            ))}
          </ul>
        </nav>

        <div className="min-w-0 flex-1 space-y-14 border-t border-gray-200 pt-10 lg:border-t-0 lg:pt-0">
          {entries.map((entry) => (
            <article
              key={entry.date}
              id={releaseAnchorId(entry.date)}
              className="scroll-mt-24 border-b border-gray-200 pb-14 last:border-0 last:pb-0"
            >
              <header className="border-l-4 border-blue-500 pl-4">
                <time
                  dateTime={entry.date}
                  className="text-sm font-medium uppercase tracking-wide text-blue-700"
                >
                  {formatReleaseDate(entry.date)}
                </time>
                <h2 className="mt-1 text-xl font-semibold text-gray-900">{entry.title}</h2>
                {entry.summary ? (
                  <p className="mt-3 leading-relaxed text-gray-700">{entry.summary}</p>
                ) : null}
              </header>

              <div className="mt-8 space-y-8">
                {entry.sections.map((section) => (
                  <section key={section.heading}>
                    <h3 className="text-base font-semibold text-gray-900">{section.heading}</h3>
                    <ul className="mt-3 list-disc space-y-2 pl-5 leading-relaxed text-gray-700">
                      {section.items.map((item, i) => (
                        <li key={`${section.heading}-${i}`}>{item}</li>
                      ))}
                    </ul>
                  </section>
                ))}
              </div>
            </article>
          ))}
        </div>
      </div>
    </div>
  );
}
