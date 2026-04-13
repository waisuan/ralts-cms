import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="text-center">
        <div className="text-gray-400 text-6xl mb-4">🔧</div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Machine Not Found</h1>
        <p className="text-gray-600 mb-6">
          The machine with this serial number could not be found.
        </p>
        <Link
          href="/"
          className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
        >
          Go to Machines
        </Link>
      </div>
    </div>
  );
}
