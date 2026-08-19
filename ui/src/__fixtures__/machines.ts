import { Machine } from '../types/machine';

/**
 * Realistic machine fixtures derived from production data.
 * Covers edge cases: long customer/assignee names, compound serial numbers
 * with [R] suffixes, multi-word statuses, empty fields, and all PPM states.
 */

const base: Machine = {
  serial_number: '',
  customer: '',
  state: '',
  account_type: '',
  model: '',
  status: '',
  brand: '',
  district: '',
  person_in_charge: '',
  reported_by: '',
  additional_notes: '',
  attachment: '',
  ppm_status: '',
  tnc_date: '',
  ppm_date: '',
  created_at: '',
  updated_at: '',
  updated_by: '',
};

/** Builds a machine with only the fields a test cares about set. */
export function machineFixture(overrides: Partial<Machine> = {}): Machine {
  return { ...base, ...overrides };
}

// Shorthand for the fixture list below.
const m = machineFixture;

export const FIXTURE_MACHINES: Machine[] = [
  // Short serial, long customer, empty state/district, no assignee
  m({
    serial_number: 'UA 70804141',
    customer: 'National Kidney Foundation',
    state: 'Selangor',
    status: '',
    ppm_date: '0001-12-31T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '2007-12-04T00:00:00Z',
    updated_at: '2018-02-24T00:00:00Z',
  }),
  // Serial with dash+spaces, very long status
  m({
    serial_number: 'Z17 - 1191',
    customer: 'Calibration Equipment',
    status: 'Send For Calibration',
    ppm_date: '0001-12-31T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '0001-12-31T00:00:00Z',
    person_in_charge: 'Muhammad Syukri',
    updated_at: '2016-02-25T00:00:00Z',
  }),
  // Pure numeric serial, long status
  m({
    serial_number: '1110423921',
    customer: 'Calibration Equipment',
    status: 'Done Calibration',
    ppm_date: '2017-02-18T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '2016-02-18T00:00:00Z',
    person_in_charge: 'Muhammad Syukri',
    updated_at: '2016-02-25T00:00:00Z',
  }),
  // Very long customer name spanning two lines in old layout
  m({
    serial_number: 'UB 01002742',
    customer: 'MTSB from KK Presint 9 Putrajaya',
    state: 'Wilayah Persekutuan',
    status: 'Asset',
    ppm_date: '2017-01-27T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '2010-09-30T00:00:00Z',
    updated_at: '2017-06-23T00:00:00Z',
  }),
  // Long serial, long customer, assignee present
  m({
    serial_number: 'UB-30703043',
    customer: 'Klinik Kesihatan Anika, Klang',
    state: 'Selangor',
    district: 'Klang',
    status: 'Placement',
    ppm_date: '2016-03-08T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '2015-09-09T00:00:00Z',
    person_in_charge: 'Norzilam Bt Othman',
    updated_at: '2018-04-05T00:00:00Z',
  }),
  // Long status "Asset ( In Use )", very long assignee
  m({
    serial_number: '05 - 4201',
    customer: 'Klinik Kesihatan Benta',
    state: 'Pahang',
    district: 'Kuala Lipis',
    status: 'Asset ( In Use )',
    ppm_date: '2017-06-13T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '0001-12-31T00:00:00Z',
    person_in_charge: 'Wan Nursyuhada Wan Hanafi',
    updated_at: '2018-04-06T00:00:00Z',
  }),
  // Another very long assignee
  m({
    serial_number: '306961',
    customer: 'Klinik Kesihatan Padang Tengku',
    state: 'Pahang',
    district: 'Kuala Lipis',
    status: 'Asset ( In Use )',
    ppm_date: '2017-06-14T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '0001-12-31T00:00:00Z',
    person_in_charge: 'Puan Nurhasyimah Bt Mohd Noor',
    updated_at: '2018-04-06T00:00:00Z',
  }),
  // ALL-CAPS customer, very long state+district
  m({
    serial_number: 'UG-51000344',
    customer: 'INSTITUT PERUBATAN RESPIRATORI',
    state: 'Kuala Lumpur',
    district: 'Wilayah Persekutuan Kuala Lumpur',
    status: 'Placement',
    ppm_date: '2026-06-24T00:00:00Z',
    ppm_status: '',
    tnc_date: '2022-04-08T00:00:00Z',
    person_in_charge: 'Nurul Azwa Md Dali',
    updated_at: '2025-12-30T00:00:00Z',
  }),
  // Status "Reagent Rental"
  m({
    serial_number: 'UD 20302171',
    customer: 'Hospital Mersing',
    state: 'Johor',
    district: 'Mersing',
    status: 'Reagent Rental',
    ppm_date: '2026-11-09T00:00:00Z',
    ppm_status: '',
    tnc_date: '2012-07-11T00:00:00Z',
    person_in_charge: 'Pn.Shakira',
    updated_at: '2025-12-30T00:00:00Z',
  }),
  // Long assignee with parenthetical
  m({
    serial_number: 'UG-90702783',
    customer: 'Klinik Kesihatan Kerteh',
    state: 'Terengganu',
    district: 'Kemaman',
    status: 'Placement',
    ppm_date: '2026-06-16T00:00:00Z',
    ppm_status: '',
    tnc_date: '2020-02-13T00:00:00Z',
    person_in_charge: 'Nor Sufiati Awang U32(KUP)',
    updated_at: '2025-12-30T00:00:00Z',
  }),
  // Long compound assignee
  m({
    serial_number: 'UG-90702623',
    customer: 'Klinik Kesihatan Batu 2 1/2',
    state: 'Terengganu',
    district: 'Kemaman',
    status: 'Placement',
    ppm_date: '2026-06-17T00:00:00Z',
    ppm_status: '',
    tnc_date: '2019-08-19T00:00:00Z',
    person_in_charge: 'Pn Wan Robina bt Che Wan Abas',
    updated_at: '2025-12-30T00:00:00Z',
  }),
  // Serial with [R] suffix, PPM overdue
  m({
    serial_number: 'UG-70901635[R]',
    customer: 'Klinik Kesihatan Penambang',
    state: 'Kelantan',
    district: 'Kota Bharu',
    status: 'Placement',
    ppm_date: '2026-07-01T00:00:00Z',
    ppm_status: '',
    tnc_date: '2026-01-01T00:00:00Z',
    updated_at: '2026-03-06T00:00:00Z',
  }),
  // Serial with [R], PPM "Upcoming" (almost_due)
  m({
    serial_number: 'UG-01203209',
    customer: 'Klinik Kesihatan Buloh Kasap',
    state: 'Johor',
    district: 'Segamat',
    status: 'Placement',
    ppm_date: '2026-04-13T00:00:00Z',
    ppm_status: 'almost_due',
    tnc_date: '2021-09-28T00:00:00Z',
    person_in_charge: 'Jamei bin Hasan U29',
    updated_at: '2026-03-09T00:00:00Z',
  }),
  // Another [R] serial, ALL-CAPS customer
  m({
    serial_number: 'UD-80313373[R]',
    customer: 'KLINIK KESIHATAN GUA MUSANG',
    state: 'Kelantan',
    district: 'Gua Musang',
    status: 'Placement',
    ppm_date: '2027-02-03T00:00:00Z',
    ppm_status: '',
    tnc_date: '2026-02-03T00:00:00Z',
    updated_at: '2026-02-13T00:00:00Z',
  }),
  // Multi-word [R] serial
  m({
    serial_number: 'UG-10203399 [R]',
    customer: 'KLINIK KESIHATAN BERIS KUBOR BESAR',
    state: 'Kelantan',
    district: 'Bachok',
    status: 'Placement',
    ppm_date: '2026-08-26T00:00:00Z',
    ppm_status: '',
    tnc_date: '2026-02-05T00:00:00Z',
    updated_at: '2026-02-13T00:00:00Z',
  }),
  // Long assignee on overdue row
  m({
    serial_number: 'UD-21216328',
    customer: 'Klinik Kesihatan Lundang Paku',
    state: 'Kelantan',
    district: 'Kota Bharu',
    status: 'Placement',
    ppm_date: '2025-04-18T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '2023-06-01T00:00:00Z',
    person_in_charge: 'Pn Hamisah AB Hamid',
    updated_at: '2026-04-02T00:00:00Z',
  }),
  // Machine with attachment and maintenance count
  m({
    serial_number: '770234',
    customer: 'Klinik Kesihatan Sg Koyan',
    state: 'Pahang',
    district: 'Kuala Lipis',
    status: 'Asset ( In Use )',
    ppm_date: '2017-06-14T00:00:00Z',
    ppm_status: 'overdue',
    tnc_date: '0001-12-31T00:00:00Z',
    person_in_charge: 'En. Ujang Bin Haimim',
    attachment: 'service_report_770234.pdf',
    maintenance_count: 5,
    updated_at: '2018-04-06T00:00:00Z',
    updated_by: 'admin',
  }),
  // PPM "due" status (real calendar date so UI can show server pill; sentinel dates hide pills)
  m({
    serial_number: 'UA 60101281',
    customer: 'UITM Pulau Pinang',
    state: 'Pulau Pinang',
    status: '',
    ppm_date: '2026-06-15T00:00:00Z',
    ppm_status: 'due',
    tnc_date: '2006-08-21T00:00:00Z',
    updated_at: '2016-03-03T00:00:00Z',
  }),
  // ZAWAWI BIN ABDUL RAZAK — three-line assignee in old layout
  m({
    serial_number: 'UD 51110799',
    customer: 'Hospital Segamat',
    state: 'Johor',
    district: 'Segamat',
    status: 'Placement',
    ppm_date: '2026-10-05T00:00:00Z',
    ppm_status: '',
    tnc_date: '0001-12-31T00:00:00Z',
    person_in_charge: 'ZAWAWI BIN ABDUL RAZAK',
    updated_at: '2025-12-30T00:00:00Z',
  }),
  // Pn Siti Nursalihah Mohd Anuar U29 — another very long assignee
  m({
    serial_number: 'UG-61000959',
    customer: 'Klinik Kesihatan Ajil',
    state: 'Terengganu',
    district: 'Hulu Terengganu',
    status: 'Placement',
    ppm_date: '2025-06-15T00:00:00Z',
    ppm_status: '',
    tnc_date: '2023-07-26T00:00:00Z',
    person_in_charge: 'Pn Siti Nursalihah Mohd Anuar U29',
    updated_at: '2025-12-30T00:00:00Z',
  }),
];
