'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';

export default function NotFoundActions() {
  const router = useRouter();

  return (
    <div className="mt-8 flex flex-col items-stretch justify-center gap-3 sm:flex-row sm:items-center">
      <Link
        href="/"
        className="inline-flex min-w-[10rem] items-center justify-center rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
      >
        Back to machines
      </Link>
      <button
        type="button"
        onClick={() => router.back()}
        className="inline-flex min-w-[10rem] items-center justify-center rounded-lg border border-gray-300 bg-white px-5 py-2.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2"
        aria-label="Go back to the previous page"
      >
        Go back
      </button>
    </div>
  );
}
