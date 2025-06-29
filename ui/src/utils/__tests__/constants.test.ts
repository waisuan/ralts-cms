import { DEFAULT_SEARCH_PROPERTY, SEARCH_PROPERTIES, DATE_PROPERTIES } from '../constants';

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
});
