import {
  DEFAULT_SEARCH_PROPERTY,
  SEARCH_PROPERTIES,
  DATE_PROPERTIES,
  PPM_STATUSES,
  PPM_STATUS_COLORS,
} from '../constants';

describe('Constants', () => {
  describe('DEFAULT_SEARCH_PROPERTY', () => {
    it('should be set to serial_number', () => {
      expect(DEFAULT_SEARCH_PROPERTY).toBe('serial_number');
    });

    it('should exist in SEARCH_PROPERTIES', () => {
      const propertyValues = SEARCH_PROPERTIES.map((prop) => prop.value);
      expect(propertyValues).toContain(DEFAULT_SEARCH_PROPERTY);
    });
  });

  describe('SEARCH_PROPERTIES', () => {
    it('should contain all expected search properties', () => {
      const expectedProperties = [
        'serial_number',
        'customer',
        'state',
        'account_type',
        'model',
        'brand',
        'district',
        'person_in_charge',
        'reported_by',
        'status',
        'ppm_status',
        'tnc_date',
        'ppm_date',
      ];

      const actualProperties = SEARCH_PROPERTIES.map((prop) => prop.value);
      expect(actualProperties).toEqual(expectedProperties);
    });

    it('should have proper labels for all properties', () => {
      const expectedLabels = [
        'Serial Number',
        'Customer',
        'State',
        'Account Type',
        'Model',
        'Brand',
        'District',
        'Person in Charge',
        'Reported By',
        'Status',
        'PPM Status',
        'TNC Date',
        'PPM Date',
      ];

      const actualLabels = SEARCH_PROPERTIES.map((prop) => prop.label);
      expect(actualLabels).toEqual(expectedLabels);
    });

    it('should have unique values', () => {
      const values = SEARCH_PROPERTIES.map((prop) => prop.value);
      const uniqueValues = [...new Set(values)];
      expect(values).toHaveLength(uniqueValues.length);
    });

    it('should have unique labels', () => {
      const labels = SEARCH_PROPERTIES.map((prop) => prop.label);
      const uniqueLabels = [...new Set(labels)];
      expect(labels).toHaveLength(uniqueLabels.length);
    });
  });

  describe('DATE_PROPERTIES', () => {
    it('should contain the expected date properties', () => {
      const expectedDateProperties = ['tnc_date', 'ppm_date'];
      expect(DATE_PROPERTIES).toEqual(expectedDateProperties);
    });

    it('should only contain properties that exist in SEARCH_PROPERTIES', () => {
      const searchPropertyValues = SEARCH_PROPERTIES.map((prop) => prop.value);
      DATE_PROPERTIES.forEach((dateProperty) => {
        expect(searchPropertyValues).toContain(dateProperty);
      });
    });
  });

  describe('PPM_STATUSES', () => {
    it('should contain all expected PPM status values', () => {
      expect(PPM_STATUSES.OVERDUE).toBe('Overdue');
      expect(PPM_STATUSES.DUE).toBe('Due');
      expect(PPM_STATUSES.DUE_SOON).toBe('Due Soon');
      expect(PPM_STATUSES.UPCOMING).toBe('Upcoming');
    });

    it('should have all statuses defined', () => {
      const expectedStatuses = ['OVERDUE', 'DUE', 'DUE_SOON', 'UPCOMING'];
      const actualStatuses = Object.keys(PPM_STATUSES);
      expect(actualStatuses).toEqual(expectedStatuses);
    });
  });

  describe('PPM_STATUS_COLORS', () => {
    it('should have colors defined for all PPM statuses', () => {
      expect(PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE]).toBe('bg-red-100 text-red-800');
      expect(PPM_STATUS_COLORS[PPM_STATUSES.DUE]).toBe('bg-orange-100 text-orange-800');
      expect(PPM_STATUS_COLORS[PPM_STATUSES.DUE_SOON]).toBe('bg-yellow-100 text-yellow-800');
      expect(PPM_STATUS_COLORS[PPM_STATUSES.UPCOMING]).toBe('bg-green-100 text-green-800');
    });

    it('should have colors for all PPM status values', () => {
      Object.values(PPM_STATUSES).forEach((status) => {
        expect(PPM_STATUS_COLORS[status]).toBeDefined();
        expect(typeof PPM_STATUS_COLORS[status]).toBe('string');
      });
    });
  });
});
