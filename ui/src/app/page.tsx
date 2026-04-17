import { Suspense } from 'react';
import HomeContent from './HomeContent';
import FullPageLoader from '@/components/FullPageLoader';

export default function Home() {
  return (
    <Suspense fallback={<FullPageLoader isVisible message="Loading..." />}>
      <HomeContent />
    </Suspense>
  );
}
