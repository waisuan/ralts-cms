import type { Metadata } from 'next';
import NotFoundActions from './not-found-actions';

export const metadata: Metadata = {
  title: 'Page not found | Ralts CMS',
  description: 'This URL does not match any page in Ralts CMS.',
};

export default function NotFound() {
  return (
    <div className="mx-auto w-full max-w-lg px-4 py-12 sm:py-16">
      <div className="rounded-2xl border border-gray-200 bg-white px-6 py-10 text-center shadow-sm sm:px-10 sm:py-12">
        <p className="text-sm font-semibold uppercase tracking-wide text-blue-600">Error 404</p>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl">
          This page doesn&apos;t exist
        </h1>
        <p className="mt-4 text-base leading-relaxed text-gray-600">
          Oops — we couldn&apos;t find anything at this address. The link may be wrong, or the page may
          have been moved.
        </p>
        <NotFoundActions />
      </div>
    </div>
  );
}
