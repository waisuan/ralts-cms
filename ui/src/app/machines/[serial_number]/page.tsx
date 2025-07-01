'use client';

import { notFound, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import { mockMachines } from '@/data/mockMachines';
import { Machine } from '@/types/machine';

interface MachinePageProps {
  params: Promise<{
    serial_number: string;
  }>;
}

function MachineContent({ machine }: { machine: Machine }) {
  const router = useRouter();

  const handleBack = () => {
    router.push('/');
  };

  return <MaintenanceHistory machine={machine} onBack={handleBack} />;
}

export default function MachinePage({ params }: MachinePageProps) {
  const [machine, setMachine] = useState<Machine | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadMachine() {
      try {
        // Await the params in Next.js 15
        const resolvedParams = await params;

        // Decode the serial number from the URL
        const serialNumber = decodeURIComponent(resolvedParams.serial_number);

        // Find the machine by serial number
        const foundMachine = mockMachines.find((m) => m.serial_number === serialNumber);

        if (!foundMachine) {
          notFound();
          return;
        }

        setMachine(foundMachine);
      } catch (error) {
        console.error('Error loading machine:', error);
        notFound();
      } finally {
        setIsLoading(false);
      }
    }

    loadMachine();
  }, [params]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading machine information...</p>
        </div>
      </div>
    );
  }

  if (!machine) {
    notFound();
    return null;
  }

  return <MachineContent machine={machine} />;
}
