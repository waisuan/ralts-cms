import { Machine } from '../types/machine';

// Helper function to get dates for status examples
const getDateForStatus = (status: 'overdue' | 'due' | 'due_soon' | 'future') => {
  const today = new Date();
  const todayStr = today.toISOString().split('T')[0]; // YYYY-MM-DD format

  switch (status) {
    case 'overdue':
      // 5 days ago
      const overdue = new Date(today);
      overdue.setDate(today.getDate() - 5);
      return overdue.toISOString().split('T')[0];
    case 'due':
      // Today
      return todayStr;
    case 'due_soon':
      // 3 days from now
      const dueSoon = new Date(today);
      dueSoon.setDate(today.getDate() + 3);
      return dueSoon.toISOString().split('T')[0];
    case 'future':
      // 30 days from now
      const future = new Date(today);
      future.setDate(today.getDate() + 30);
      return future.toISOString().split('T')[0];
  }
};

// Mock Machine data with different status examples
export const mockMachines: Machine[] = [
  {
    serial_number: 'SN-001',
    customer: 'Acme Corp',
    state: 'CA',
    account_type: 'Premium',
    model: 'X100',
    status: 'Active', // Regular status
    brand: 'BrandA',
    district: 'North',
    person_in_charge: 'Alice Johnson',
    reported_by: 'Bob Smith',
    additional_notes: 'Needs inspection after last maintenance check. Unit has been running well but requires quarterly review.',
    attachment: 'inspection_report_SN001.pdf',
    ppm_status: '',
    tnc_date: '2024-07-01',
    ppm_date: getDateForStatus('overdue'), // Will show "Overdue" (Red)
    created_at: '2024-01-01',
    updated_at: '2024-06-01',
  },
  {
    serial_number: 'SN-002',
    customer: 'Beta LLC',
    state: 'NY',
    account_type: 'Standard',
    model: 'Y200',
    status: '', // Empty status
    brand: 'BrandB',
    district: 'East',
    person_in_charge: 'Charlie Brown',
    reported_by: '', // Empty reported_by
    additional_notes: 'Regular maintenance scheduled',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-10',
    ppm_date: getDateForStatus('due'), // Will show "Due" (Orange)
    created_at: '2024-02-01',
    updated_at: '2024-06-10',
  },
  {
    serial_number: 'SN-003',
    customer: 'Gamma Inc',
    state: 'TX',
    account_type: 'Basic',
    model: 'Z300',
    status: 'Maintenance',
    brand: 'BrandC',
    district: 'South',
    person_in_charge: 'Eve Wilson',
    reported_by: 'Frank Miller',
    additional_notes: '', // Empty notes
    attachment: 'maintenance_schedule.xlsx',
    ppm_status: '',
    tnc_date: '2024-06-01',
    ppm_date: getDateForStatus('due_soon'), // Will show "Due Soon" (Yellow)
    created_at: '2024-03-01',
    updated_at: '2024-06-15',
  },
  {
    serial_number: 'SN-004',
    customer: 'Delta Co',
    state: 'FL',
    account_type: 'Premium',
    model: 'A400',
    status: 'Inactive',
    brand: 'BrandD',
    district: 'West',
    person_in_charge: 'Grace Lee',
    reported_by: 'Heidi Klum',
    additional_notes: 'Recently serviced, performing optimally. Next service due in 6 months.',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-20',
    ppm_date: getDateForStatus('future'), // Will show no status badge
    created_at: '2024-04-01',
    updated_at: '2024-06-20',
  },
  {
    serial_number: 'SN-005',
    customer: 'Echo Systems',
    state: 'WA',
    account_type: 'Enterprise',
    model: 'B500',
    status: 'Active',
    brand: 'BrandE',
    district: 'Northwest',
    person_in_charge: 'Ian Cooper',
    reported_by: '', // Empty reported_by
    additional_notes: 'High priority equipment - critical for production line',
    attachment: 'priority_specs.pdf',
    ppm_status: '',
    tnc_date: '2024-08-01',
    ppm_date: getDateForStatus('overdue'), // Another "Overdue" example
    created_at: '2024-05-01',
    updated_at: '2024-06-25',
  },
  {
    serial_number: 'SN-006',
    customer: 'Foxtrot Industries',
    state: 'CO',
    account_type: 'Standard',
    model: 'C600',
    status: '', // Empty status
    brand: 'BrandF',
    district: 'Mountain',
    person_in_charge: 'Kate Martinez',
    reported_by: 'Liam O\'Connor',
    additional_notes: '', // Empty notes
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-07-15',
    ppm_date: getDateForStatus('due_soon'), // Another "Due Soon" example
    created_at: '2024-06-01',
    updated_at: '2024-06-28',
  },
  {
    serial_number: 'SN-007',
    customer: 'Golf Manufacturing',
    state: 'OH',
    account_type: 'Standard',
    model: 'D700',
    status: 'Active',
    brand: 'BrandG',
    district: 'Midwest',
    person_in_charge: 'Mike Chen',
    reported_by: 'Nancy Davis',
    additional_notes: 'Production line equipment - requires special handling procedures',
    attachment: 'handling_procedures.docx',
    ppm_status: '',
    tnc_date: '2024-07-25',
    ppm_date: getDateForStatus('future'),
    created_at: '2024-06-05',
    updated_at: '2024-06-30',
  },
  {
    serial_number: 'SN-008',
    customer: 'Hotel Services',
    state: 'NV',
    account_type: 'Premium',
    model: 'E800',
    status: 'Active',
    brand: 'BrandH',
    district: 'Southwest',
    person_in_charge: 'Oscar Rodriguez',
    reported_by: 'Patricia Wilson',
    additional_notes: '24/7 operation - any downtime must be scheduled during maintenance windows',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-08-05',
    ppm_date: getDateForStatus('due_soon'),
    created_at: '2024-06-10',
    updated_at: '2024-07-01',
  },
  {
    serial_number: 'SN-009',
    customer: 'India Tech',
    state: 'GA',
    account_type: 'Enterprise',
    model: 'F900',
    status: 'Maintenance',
    brand: 'BrandI',
    district: 'Southeast',
    person_in_charge: 'Quinn Taylor',
    reported_by: '', // Empty reported_by
    additional_notes: '', // Empty notes
    attachment: 'calibration_cert.pdf',
    ppm_status: '',
    tnc_date: '2024-07-30',
    ppm_date: getDateForStatus('overdue'),
    created_at: '2024-06-15',
    updated_at: '2024-07-02',
  },
  {
    serial_number: 'SN-010',
    customer: 'Juliet Logistics',
    state: 'TN',
    account_type: 'Basic',
    model: 'G1000',
    status: 'Inactive',
    brand: 'BrandJ',
    district: 'Central',
    person_in_charge: 'Sam Thompson',
    reported_by: 'Tina Brown',
    additional_notes: 'Warehouse automation system - integrated with main inventory system',
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-08-10',
    ppm_date: getDateForStatus('due'),
    created_at: '2024-06-20',
    updated_at: '2024-07-03',
  },
  {
    serial_number: 'SN-011',
    customer: 'Kilo Solutions',
    state: 'NC',
    account_type: 'Standard',
    model: 'H1100',
    status: '', // Empty status
    brand: 'BrandK',
    district: 'Atlantic',
    person_in_charge: 'Uma Patel',
    reported_by: 'Victor Martinez',
    additional_notes: 'Research facility equipment - handle with extreme care',
    attachment: 'safety_protocols.pdf',
    ppm_status: '',
    tnc_date: '2024-08-15',
    ppm_date: getDateForStatus('future'),
    created_at: '2024-06-25',
    updated_at: '2024-07-04',
  },
  {
    serial_number: 'SN-012',
    customer: 'Lima Industries',
    state: 'SC',
    account_type: 'Premium',
    model: 'I1200',
    status: 'Active',
    brand: 'BrandL',
    district: 'Carolina',
    person_in_charge: 'Wendy Johnson',
    reported_by: 'Xavier Lee',
    additional_notes: '', // Empty notes
    attachment: '',
    ppm_status: '',
    tnc_date: '2024-08-20',
    ppm_date: getDateForStatus('due_soon'),
    created_at: '2024-06-30',
    updated_at: '2024-07-05',
  },
];
