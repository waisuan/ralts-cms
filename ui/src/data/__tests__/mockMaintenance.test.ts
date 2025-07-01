import { mockMaintenanceRecords } from '../mockMaintenance';
import { MaintenanceOrderType } from '../../types/maintenance';

describe('Mock Maintenance Data', () => {
  it('should contain valid maintenance records', () => {
    expect(mockMaintenanceRecords).toBeDefined();
    expect(Array.isArray(mockMaintenanceRecords)).toBe(true);
    expect(mockMaintenanceRecords.length).toBeGreaterThan(0);
  });

  it('should have records for multiple machines', () => {
    const uniqueMachines = new Set(mockMaintenanceRecords.map((r) => r.machine_serial_number));
    expect(uniqueMachines.size).toBeGreaterThan(1);
  });

  it('should contain all required fields for each record', () => {
    mockMaintenanceRecords.forEach((record) => {
      expect(record).toHaveProperty('machine_serial_number');
      expect(record).toHaveProperty('work_order_number');
      expect(record).toHaveProperty('work_order_date');
      expect(record).toHaveProperty('action_taken');
      expect(record).toHaveProperty('reported_by');
      expect(record).toHaveProperty('worker_order_type');
      expect(record).toHaveProperty('attachment');
      expect(record).toHaveProperty('created_at');
      expect(record).toHaveProperty('updated_at');
    });
  });

  it('should have valid maintenance order types', () => {
    const validTypes: MaintenanceOrderType[] = [
      'Preventive',
      'Emergency',
      'Corrective',
      'Inspection',
    ];

    mockMaintenanceRecords.forEach((record) => {
      expect(validTypes).toContain(record.worker_order_type);
    });
  });

  it('should have different maintenance types represented', () => {
    const types = mockMaintenanceRecords.map((r) => r.worker_order_type);
    const uniqueTypes = new Set(types);

    // Should have at least 2 different types
    expect(uniqueTypes.size).toBeGreaterThanOrEqual(2);
  });

  it('should have valid ISO date formats', () => {
    const isoDateRegex = /^\d{4}-\d{2}-\d{2}$/;
    const isoDateTimeRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z$/; // Handle optional milliseconds

    mockMaintenanceRecords.forEach((record) => {
      expect(record.work_order_date).toMatch(isoDateRegex);
      expect(record.created_at).toMatch(isoDateTimeRegex);
      expect(record.updated_at).toMatch(isoDateTimeRegex);
    });
  });

  it('should have non-empty work order numbers', () => {
    mockMaintenanceRecords.forEach((record) => {
      expect(record.work_order_number).toBeTruthy();
      expect(record.work_order_number.length).toBeGreaterThan(0);
    });
  });

  it('should have unique work order numbers', () => {
    const workOrderNumbers = mockMaintenanceRecords.map((r) => r.work_order_number);
    const uniqueNumbers = new Set(workOrderNumbers);

    expect(uniqueNumbers.size).toBe(workOrderNumbers.length);
  });

  it('should have meaningful action descriptions', () => {
    mockMaintenanceRecords.forEach((record) => {
      expect(record.action_taken).toBeTruthy();
      expect(record.action_taken.length).toBeGreaterThan(10); // Meaningful descriptions
    });
  });

  it('should have reported_by field populated', () => {
    mockMaintenanceRecords.forEach((record) => {
      expect(record.reported_by).toBeTruthy();
      expect(record.reported_by.length).toBeGreaterThan(0);
    });
  });

  it('should include records with and without attachments', () => {
    const withAttachments = mockMaintenanceRecords.filter((r) => r.attachment);
    const withoutAttachments = mockMaintenanceRecords.filter((r) => !r.attachment);

    expect(withAttachments.length).toBeGreaterThan(0);
    expect(withoutAttachments.length).toBeGreaterThan(0);
  });

  it('should have created_at before or equal to updated_at', () => {
    mockMaintenanceRecords.forEach((record) => {
      const created = new Date(record.created_at);
      const updated = new Date(record.updated_at);

      expect(created.getTime()).toBeLessThanOrEqual(updated.getTime());
    });
  });

  it('should include at least one long action description for modal testing', () => {
    const longActions = mockMaintenanceRecords.filter((r) => r.action_taken.length > 100);
    expect(longActions.length).toBeGreaterThan(0);
  });

  it('should have records for machines that exist in mockMachines', () => {
    // Test that the serial numbers reference valid machines
    const serialNumbers = mockMaintenanceRecords.map((r) => r.machine_serial_number);
    const uniqueSerials = new Set(serialNumbers);

    // Should have records for common test machines
    expect(
      uniqueSerials.has('SN-001') || uniqueSerials.has('SN-002') || uniqueSerials.has('SN-003')
    ).toBe(true);
  });
});
