import { MachineService } from './machineService';
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

describe('MachineService', () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    jest.clearAllMocks();
    getMock.mockResolvedValue({ data: { machines: [], count: 0 } });
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('getMachines builds offset from page and passes filters as query params', async () => {
    await MachineService.getMachines(3, 25, {
      sort: 'ppm_date_asc',
      q: 'acme',
      ppm_status_filter: 'overdue',
      ppm_date_from: '2024-01-01',
    });

    expect(getMock).toHaveBeenCalledWith('/api/v1/machines', {
      limit: '25',
      offset: '50',
      sort: 'ppm_date_asc',
      q: 'acme',
      ppm_status_filter: 'overdue',
      ppm_date_from: '2024-01-01',
    });
  });

  it('getMachine encodes serial for path', async () => {
    await MachineService.getMachine('SN 001');
    expect(getMock).toHaveBeenCalledWith('/api/v1/machines/SN%20001');
  });

  it('updateMachine and deleteMachine encode serial', async () => {
    (apiClient.put as jest.Mock).mockResolvedValue({ data: {} });
    (apiClient.delete as jest.Mock).mockResolvedValue({ data: undefined });

    await MachineService.updateMachine('A/B', { serial_number: 'A/B', customer: 'c' } as never);
    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/machines/A%2FB', expect.any(Object));

    await MachineService.deleteMachine('x y');
    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/machines/x%20y');
  });

  it('exportMachinesCSV calls fetch with relative URL', async () => {
    const blobMock = new Blob(['csv-data'], { type: 'text/csv' });
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      blob: async () => blobMock,
      headers: { get: () => null },
    });
    window.URL.createObjectURL = jest.fn(() => 'blob:mock');
    window.URL.revokeObjectURL = jest.fn();

    await MachineService.exportMachinesCSV({ q: 'acme' });

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/machines/export/csv?q=acme',
      expect.objectContaining({
        headers: expect.any(Object),
      })
    );

    const calledUrl = (global.fetch as jest.Mock).mock.calls[0][0] as string;
    expect(calledUrl).not.toMatch(/^https?:\/\//);
  });

  it('exportMaintenanceCSV calls fetch with relative URL', async () => {
    const blobMock = new Blob(['csv-data'], { type: 'text/csv' });
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      blob: async () => blobMock,
      headers: { get: () => null },
    });
    window.URL.createObjectURL = jest.fn(() => 'blob:mock');
    window.URL.revokeObjectURL = jest.fn();

    await MachineService.exportMaintenanceCSV('SN-001');

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/v1/machines/SN-001/maintenance/export/csv',
      expect.objectContaining({
        headers: expect.any(Object),
      })
    );

    const calledUrl = (global.fetch as jest.Mock).mock.calls[0][0] as string;
    expect(calledUrl).not.toMatch(/^https?:\/\//);
  });
});
