'use client';

interface FullPageLoaderProps {
  isVisible: boolean;
  message?: string;
}

export default function FullPageLoader({ isVisible, message = 'Loading...' }: FullPageLoaderProps) {
  if (!isVisible) return null;

  return (
    <div className="fixed inset-0 z-50 bg-white bg-opacity-90 backdrop-blur-sm flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-lg font-medium text-gray-700">{message}</p>
        <p className="text-sm text-gray-500 mt-2">Please wait while we load the page...</p>
      </div>
    </div>
  );
} 