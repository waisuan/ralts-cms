import { Maintenance } from '../types/maintenance';

// Helper function to generate random dates within a range
const randomDate = (start: Date, end: Date): string => {
  return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()))
    .toISOString()
    .split('T')[0];
};

// Helper function to generate random datetime
const randomDateTime = (start: Date, end: Date): string => {
  return new Date(
    start.getTime() + Math.random() * (end.getTime() - start.getTime())
  ).toISOString();
};

// Generate extensive mock data for pagination testing
const generateMockRecords = (): Maintenance[] => {
  const records: Maintenance[] = [];
  const machineSerialNumbers = [
    'SN-001',
    'SN-002',
    'SN-003',
    'SN-004',
    'SN-005',
    'SN-006',
    'SN-007',
    'SN-008',
    'SN-009',
    'SN-010',
  ];
  const technicians = [
    'John Smith',
    'Sarah Johnson',
    'Mike Davis',
    'Lisa Chen',
    'Tom Wilson',
    'Emma Brown',
    'Alex Rodriguez',
    'Maria Garcia',
    'David Lee',
    'Jennifer White',
    'Robert Taylor',
    'Amanda Clark',
    'James Anderson',
    'Michelle Lewis',
    'Christopher Hall',
  ];
  const maintenanceTypes = ['Preventive', 'Corrective', 'Emergency', 'Inspection'] as const;
  const attachments = [
    'maintenance_report.pdf',
    'repair_log.pdf',
    'inspection_checklist.pdf',
    'parts_replacement.pdf',
    'safety_audit.pdf',
    'calibration_report.pdf',
    'troubleshooting_guide.pdf',
    'preventive_schedule.pdf',
    'emergency_protocol.pdf',
    'quality_check.pdf',
    'performance_test.pdf',
    'compliance_report.pdf',
    'warranty_claim.pdf',
    'service_history.pdf',
    'maintenance_manual.pdf',
  ];

  let workOrderCounter = 1;

  // Generate records for each machine
  machineSerialNumbers.forEach((serialNumber) => {
    // Generate 15-25 records per machine
    const numRecords = Math.floor(Math.random() * 11) + 15; // 15-25 records

    for (let i = 0; i < numRecords; i++) {
      const workOrderDate = randomDate(new Date('2023-01-01'), new Date('2024-12-31'));
      const createdDate = randomDateTime(new Date(workOrderDate), new Date());
      const updatedDate = randomDateTime(new Date(createdDate), new Date());

      const maintenanceType = maintenanceTypes[Math.floor(Math.random() * maintenanceTypes.length)];
      const technician = technicians[Math.floor(Math.random() * technicians.length)];
      const hasAttachment = Math.random() > 0.3; // 70% chance of having attachment
      const attachment = hasAttachment
        ? attachments[Math.floor(Math.random() * attachments.length)]
        : '';

      // Generate realistic action descriptions based on maintenance type
      let actionTaken = '';
      switch (maintenanceType) {
        case 'Preventive':
          actionTaken = `Completed scheduled preventive maintenance. ${
            [
              'Replaced air filters and checked fluid levels.',
              'Performed system diagnostics and calibration.',
              'Lubricated moving parts and tightened connections.',
              'Updated software and firmware versions.',
              'Conducted safety system verification.',
              'Replaced worn components as per maintenance schedule.',
              'Performed vibration analysis and alignment checks.',
              'Updated maintenance logs and documentation.',
            ][Math.floor(Math.random() * 8)]
          } All systems operating within normal parameters.`;
          break;
        case 'Corrective':
          actionTaken = `Corrective maintenance performed to address ${
            [
              'belt tension and alignment issues.',
              'hydraulic system pressure problems.',
              'electrical connection faults.',
              'mechanical wear and tear.',
              'sensor calibration drift.',
              'control system malfunctions.',
              'cooling system inefficiencies.',
              'drive mechanism problems.',
            ][Math.floor(Math.random() * 8)]
          } Repairs completed and system tested for proper operation.`;
          break;
        case 'Emergency':
          actionTaken = `Emergency repair required due to ${
            [
              'sudden hydraulic pump failure.',
              'electrical system shutdown.',
              'mechanical component breakdown.',
              'safety system activation.',
              'control panel malfunction.',
              'drive motor overheating.',
              'sensor failure causing shutdown.',
              'emergency stop activation.',
            ][Math.floor(Math.random() * 8)]
          } Immediate action taken to restore operation. Root cause analysis completed.`;
          break;
        case 'Inspection':
          actionTaken = `Comprehensive inspection conducted including ${
            [
              'visual examination of all components.',
              'performance testing and measurement.',
              'safety system verification.',
              'compliance audit and documentation review.',
              'quality control assessment.',
              'environmental impact evaluation.',
              'regulatory compliance check.',
              'operational efficiency analysis.',
            ][Math.floor(Math.random() * 8)]
          } Inspection report generated with recommendations.`;
          break;
      }

      records.push({
        machine_serial_number: serialNumber,
        work_order_number: `WO-2024-${workOrderCounter.toString().padStart(3, '0')}`,
        work_order_date: workOrderDate,
        action_taken: actionTaken,
        reported_by: technician,
        worker_order_type: maintenanceType,
        attachment: attachment,
        created_at: createdDate,
        updated_at: updatedDate,
      });

      workOrderCounter++;
    }
  });

  return records;
};

export const mockMaintenanceRecords: Maintenance[] = generateMockRecords();
