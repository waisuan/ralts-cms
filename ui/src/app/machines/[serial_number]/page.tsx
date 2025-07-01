'use client';

import { notFound, useRouter } from 'next/navigation';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import { mockMachines } from '@/data/mockMachines';

interface MachinePageProps {
  params: {
    serial_number: string;
  };
}

export default function MachinePage({ params }: MachinePageProps) {
  const router = useRouter();

  // Decode the serial number from the URL
  const serialNumber = decodeURIComponent(params.serial_number);

  // Find the machine by serial number
  const machine = mockMachines.find((m) => m.serial_number === serialNumber);

  // If machine not found, show 404
  if (!machine) {
    notFound();
  }

  return (
    <MaintenanceHistory
      machine={machine}
      onBack={() => {
        router.push('/');
      }}
    />
  );
}
