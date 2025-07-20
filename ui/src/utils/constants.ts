// Search-related constants
export const DEFAULT_SEARCH_PROPERTY = 'serial_number' as const;

// Other application constants can be added here as needed
export const SEARCH_PROPERTIES = [
  { value: 'serial_number', label: 'Serial Number' },
  { value: 'customer', label: 'Customer' },
  { value: 'state', label: 'State' },
  { value: 'account_type', label: 'Account Type' },
  { value: 'model', label: 'Model' },
  { value: 'brand', label: 'Brand' },
  { value: 'district', label: 'District' },
  { value: 'person_in_charge', label: 'Person in Charge' },
  { value: 'reported_by', label: 'Reported By' },
  { value: 'status', label: 'Status' },
  { value: 'ppm_status', label: 'PPM Status' },
  { value: 'tnc_date', label: 'TNC Date' },
  { value: 'ppm_date', label: 'PPM Date' },
] as const;

// Properties that should use date picker instead of text input
export const DATE_PROPERTIES = ['tnc_date', 'ppm_date'] as const;

// Properties that should use dropdown selection instead of text input
export const PPM_STATUS_PROPERTY = 'ppm_status' as const;

// PPM Status constants
export const PPM_STATUSES = {
  OVERDUE: 'overdue',
  DUE: 'due',
  ALMOST_DUE: 'almost_due',
} as const;

export const PPM_STATUS_COLORS = {
  [PPM_STATUSES.OVERDUE]: 'bg-red-100 text-red-800',
  [PPM_STATUSES.DUE]: 'bg-orange-100 text-orange-800',
  [PPM_STATUSES.ALMOST_DUE]: 'bg-yellow-100 text-yellow-800',
} as const;

// Available PPM statuses for search (only ones that actually appear in the UI)
export const SEARCHABLE_PPM_STATUSES = [
  { value: PPM_STATUSES.OVERDUE, label: 'Overdue' },
  { value: PPM_STATUSES.DUE, label: 'Due' },
  { value: PPM_STATUSES.ALMOST_DUE, label: 'Almost Due' },
] as const;

// Type for search property values
export type SearchPropertyValue = (typeof SEARCH_PROPERTIES)[number]['value'];
export type DateProperty = (typeof DATE_PROPERTIES)[number];
export type PPMStatus = (typeof PPM_STATUSES)[keyof typeof PPM_STATUSES];

// Type guard to check if a property is a date property
export function isDateProperty(property: string): property is DateProperty {
  return (DATE_PROPERTIES as readonly string[]).includes(property);
}

// Type guard to check if a property is the PPM status property
export function isPPMStatusProperty(property: string): property is typeof PPM_STATUS_PROPERTY {
  return property === PPM_STATUS_PROPERTY;
}

// Malaysian states
export const MALAYSIAN_STATES = [
  'Johor',
  'Kedah',
  'Kelantan',
  'Kuala Lumpur',
  'Labuan',
  'Malacca',
  'Negeri Sembilan',
  'Pahang',
  'Penang',
  'Perak',
  'Perlis',
  'Putrajaya',
  'Sabah',
  'Sarawak',
  'Selangor',
  'Terengganu',
] as const;

// Malaysian districts organized by state
export const MALAYSIAN_DISTRICTS_BY_STATE = {
  Johor: [
    'Batu Pahat',
    'Johor Bahru',
    'Kluang',
    'Kota Tinggi',
    'Kulai',
    'Mersing',
    'Muar',
    'Pontian',
    'Segamat',
    'Tangkak',
  ],
  Kedah: [
    'Baling',
    'Bandar Baharu',
    'Kota Setar',
    'Kuala Muda',
    'Kubang Pasu',
    'Kulim',
    'Langkawi',
    'Padang Terap',
    'Pendang',
    'Pokok Sena',
    'Sik',
    'Yan',
  ],
  Kelantan: [
    'Bachok',
    'Gua Musang',
    'Jeli',
    'Kota Bharu',
    'Kuala Krai',
    'Machang',
    'Pasir Mas',
    'Pasir Puteh',
    'Tanah Merah',
    'Tumpat',
  ],
  'Kuala Lumpur': ['Kuala Lumpur'],
  Labuan: ['Labuan'],
  Malacca: ['Alor Gajah', 'Jasin', 'Melaka Tengah'],
  'Negeri Sembilan': [
    'Jelebu',
    'Jempol',
    'Kuala Pilah',
    'Port Dickson',
    'Rembau',
    'Seremban',
    'Tampin',
  ],
  Pahang: [
    'Bentong',
    'Bera',
    'Cameron Highlands',
    'Jerantut',
    'Kuantan',
    'Lipis',
    'Maran',
    'Pekan',
    'Raub',
    'Rompin',
    'Temerloh',
  ],
  Penang: [
    'Barat Daya',
    'Seberang Perai Selatan',
    'Seberang Perai Tengah',
    'Seberang Perai Utara',
    'Timur Laut',
  ],
  Perak: [
    'Batang Padang',
    'Hilir Perak',
    'Hulu Perak',
    'Kampar',
    'Kerian',
    'Kinta',
    'Kuala Kangsar',
    'Larut, Matang dan Selama',
    'Manjung',
    'Muallim',
    'Perak Tengah',
  ],
  Perlis: ['Perlis'],
  Putrajaya: ['Putrajaya'],
  Sabah: [
    'Beaufort',
    'Beluran',
    'Keningau',
    'Kinabalu',
    'Kinabatangan',
    'Kota Belud',
    'Kota Kinabalu',
    'Kota Marudu',
    'Kuala Penyu',
    'Kudat',
    'Kunak',
    'Lahad Datu',
    'Nabawan',
    'Papar',
    'Penampang',
    'Pitas',
    'Putatan',
    'Ranau',
    'Sandakan',
    'Semporna',
    'Sipitang',
    'Tambunan',
    'Tawau',
    'Telupid',
    'Tenom',
    'Tongod',
    'Tuaran',
  ],
  Sarawak: [
    'Asajaya',
    'Baram',
    'Belaga',
    'Beluru',
    'Betong',
    'Bintulu',
    'Bukit Mabong',
    'Dalat',
    'Daro',
    'Julau',
    'Kabong',
    'Kanowit',
    'Kapit',
    'Kuching',
    'Limbang',
    'Lubok Antu',
    'Lundu',
    'Maradong',
    'Marudi',
    'Meradong',
    'Miri',
    'Mukah',
    'Pakan',
    'Pusa',
    'Samarahan',
    'Saratok',
    'Sarikei',
    'Sebauh',
    'Selangau',
    'Serian',
    'Sibu',
    'Simunjan',
    'Song',
    'Sri Aman',
    'Subis',
    'Tanjung Manis',
    'Tatau',
    'Tebedu',
    'Telang Usan',
  ],
  Selangor: [
    'Gombak',
    'Hulu Langat',
    'Hulu Selangor',
    'Klang',
    'Kuala Langat',
    'Kuala Selangor',
    'Petaling',
    'Sabak Bernam',
    'Sepang',
  ],
  Terengganu: [
    'Besut',
    'Dungun',
    'Hulu Terengganu',
    'Kemaman',
    'Kuala Nerus',
    'Kuala Terengganu',
    'Marang',
    'Setiu',
  ],
} as const;

// Type for Malaysian state
export type MalaysianState = (typeof MALAYSIAN_STATES)[number];

// Helper function to get districts for a state
export function getDistrictsForState(state: MalaysianState): readonly string[] {
  return MALAYSIAN_DISTRICTS_BY_STATE[state] || [];
}
