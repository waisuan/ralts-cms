import { MaintenanceService } from './maintenanceService';
import { apiClient } from '../utils/api';

jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

const getMock = apiClient.get as jest.Mock;

describe('MaintenanceService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    getMock.mockResolvedValue({ data: { maintenance: [], count: 0 } });
  });

  it('getMaintenanceList encodes machine serial and passes pagination + filters', async () => {
    await MaintenanceService.getMaintenanceList('SN#1', 2, 20, {
      sort: 'work_order_date_desc',
      q: 'oil',
    });

    expect(getMock).toHaveBeenCalledWith('/api/v1/machines/SN%231/maintenance', {
      limit: '20',
      offset: '20',
      sort: 'work_order_date_desc',
      q: 'oil',
    });
  });
});
