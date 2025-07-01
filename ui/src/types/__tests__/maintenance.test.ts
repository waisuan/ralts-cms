import { Maintenance, MaintenanceOrderType } from '../maintenance';

describe('Maintenance Types', () => {
  it('should have correct MaintenanceOrderType values', () => {
    const validTypes: MaintenanceOrderType[] = [
      'Preventive',
      'Emergency',
      'Corrective',
      'Inspection',
    ];

    // Test that all valid types are accepted
    validTypes.forEach((type) => {
      const maintenance: Partial<Maintenance> = {
        worker_order_type: type,
      };
      expect(maintenance.worker_order_type).toBe(type);
    });
  });

  it('should create a valid Maintenance object', () => {
    const maintenance: Maintenance = {
      machine_serial_number: 'SN-001',
      work_order_number: 'WO-001',
      work_order_date: '2024-06-15',
      action_taken: 'Performed routine maintenance',
      reported_by: 'John Doe',
      worker_order_type: 'Preventive',
      attachment: 'report.pdf',
      created_at: '2024-06-15T09:00:00Z',
      updated_at: '2024-06-15T10:30:00Z',
    };

    expect(maintenance.machine_serial_number).toBe('SN-001');
    expect(maintenance.work_order_number).toBe('WO-001');
    expect(maintenance.worker_order_type).toBe('Preventive');
    expect(maintenance.action_taken).toBe('Performed routine maintenance');
    expect(maintenance.reported_by).toBe('John Doe');
  });

  it('should allow empty attachment field', () => {
    const maintenance: Maintenance = {
      machine_serial_number: 'SN-002',
      work_order_number: 'WO-002',
      work_order_date: '2024-06-10',
      action_taken: 'Emergency repair',
      reported_by: 'Jane Smith',
      worker_order_type: 'Emergency',
      attachment: '',
      created_at: '2024-06-10T14:00:00Z',
      updated_at: '2024-06-10T16:00:00Z',
    };

    expect(maintenance.attachment).toBe('');
  });

  it('should handle all maintenance order types', () => {
    const types: MaintenanceOrderType[] = ['Preventive', 'Emergency', 'Corrective', 'Inspection'];

    types.forEach((type) => {
      const maintenance: Maintenance = {
        machine_serial_number: 'SN-003',
        work_order_number: `WO-${type}`,
        work_order_date: '2024-06-01',
        action_taken: `${type} maintenance action`,
        reported_by: 'Test User',
        worker_order_type: type,
        attachment: '',
        created_at: '2024-06-01T08:00:00Z',
        updated_at: '2024-06-01T09:00:00Z',
      };

      expect(maintenance.worker_order_type).toBe(type);
    });
  });
});
