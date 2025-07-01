import { Maintenance } from '../types/maintenance';

export const mockMaintenanceRecords: Maintenance[] = [
  {
    machine_serial_number: 'SN-001',
    work_order_number: 'WO-2024-001',
    work_order_date: '2024-06-15',
    action_taken:
      'Completed routine preventive maintenance. Replaced air filters, checked fluid levels, and performed system diagnostics. All components functioning within normal parameters.',
    reported_by: 'John Smith',
    worker_order_type: 'Preventive',
    attachment: 'maintenance_report_001.pdf',
    created_at: '2024-06-15T09:00:00.000Z',
    updated_at: '2024-06-15T14:30:00.000Z',
  },
  {
    machine_serial_number: 'SN-001',
    work_order_number: 'WO-2024-002',
    work_order_date: '2024-05-20',
    action_taken:
      'Emergency repair due to hydraulic pump failure. Replaced faulty pump assembly and tested system pressure. Machine returned to full operational status.',
    reported_by: 'Sarah Johnson',
    worker_order_type: 'Emergency',
    attachment: 'emergency_repair_002.pdf',
    created_at: '2024-05-20T08:15:00.000Z',
    updated_at: '2024-05-20T16:45:00.000Z',
  },
  {
    machine_serial_number: 'SN-001',
    work_order_number: 'WO-2024-003',
    work_order_date: '2024-04-10',
    action_taken:
      'Corrective maintenance to address belt tension issues. Adjusted tension on drive belts and replaced worn components. System calibration completed.',
    reported_by: 'Mike Davis',
    worker_order_type: 'Corrective',
    attachment: '',
    created_at: '2024-04-10T10:30:00.000Z',
    updated_at: '2024-04-10T15:20:00.000Z',
  },
  {
    machine_serial_number: 'SN-002',
    work_order_number: 'WO-2024-004',
    work_order_date: '2024-06-10',
    action_taken:
      'Monthly inspection and lubrication service. Checked all moving parts, applied lubricants as per schedule, and verified safety systems.',
    reported_by: 'Lisa Chen',
    worker_order_type: 'Inspection',
    attachment: 'inspection_004.pdf',
    created_at: '2024-06-10T11:00:00.000Z',
    updated_at: '2024-06-10T13:45:00.000Z',
  },
  {
    machine_serial_number: 'SN-002',
    work_order_number: 'WO-2024-005',
    work_order_date: '2024-05-05',
    action_taken:
      'Replaced worn-out conveyor belt and realigned transport mechanism. Performed load testing to ensure proper operation.',
    reported_by: 'Tom Wilson',
    worker_order_type: 'Corrective',
    attachment: 'belt_replacement_005.pdf',
    created_at: '2024-05-05T09:45:00.000Z',
    updated_at: '2024-05-05T17:30:00.000Z',
  },
  {
    machine_serial_number: 'SN-003',
    work_order_number: 'WO-2024-006',
    work_order_date: '2024-06-25',
    action_taken:
      'Quarterly preventive maintenance cycle. Complete system overhaul including filter replacement, fluid changes, and comprehensive testing.',
    reported_by: 'Emma Brown',
    worker_order_type: 'Preventive',
    attachment: 'quarterly_maintenance_006.pdf',
    created_at: '2024-06-25T08:00:00.000Z',
    updated_at: '2024-06-25T16:00:00.000Z',
  },
  {
    machine_serial_number: 'SN-003',
    work_order_number: 'WO-2024-007',
    work_order_date: '2024-03-15',
    action_taken:
      'Emergency shutdown investigation. Found and repaired electrical fault in control panel. Updated safety protocols and tested all emergency systems.',
    reported_by: 'Alex Rodriguez',
    worker_order_type: 'Emergency',
    attachment: 'emergency_electrical_007.pdf',
    created_at: '2024-03-15T14:20:00.000Z',
    updated_at: '2024-03-15T20:15:00.000Z',
  },
];
