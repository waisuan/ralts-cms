import Link from 'next/link';
import { CHANGELOG_PATH } from '@/utils/constants';

const SUPPORT_EMAIL = 'e-sia@outlook.com';

export default function AppFooter() {
  return (
    <footer className="mt-auto border-t border-gray-200 bg-white">
      <div className="container mx-auto flex flex-col items-center justify-center gap-2 px-4 py-4 text-sm text-gray-600 sm:flex-row sm:gap-6">
        <Link
          href={CHANGELOG_PATH}
          className="font-medium text-gray-700 underline-offset-2 hover:text-blue-700 hover:underline"
        >
          Changelog
        </Link>
        <span className="hidden text-gray-300 sm:inline" aria-hidden>
          |
        </span>
        <p className="text-center sm:text-left">
          <span className="block sm:inline">Need help? </span>
          <a
            href={`mailto:${SUPPORT_EMAIL}?subject=${encodeURIComponent('Ralts CMS support')}`}
            className="font-medium text-blue-600 underline-offset-2 hover:text-blue-800 hover:underline"
          >
            {SUPPORT_EMAIL}
          </a>
        </p>
      </div>
    </footer>
  );
}
