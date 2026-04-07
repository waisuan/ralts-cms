import { renderHook, act, waitFor } from '@testing-library/react';
import { useMachine } from './useMachine';
import { MachineService } from '../services/machineService';
import { ApiError } from '../utils/api';

jest.mock('../utils/auth', () => ({
  isAuthError: jest.fn(() => false),
}));

jest.mock('../services/machineService', () => ({
  MachineService: {
    getMachine: jest.fn(),
    createMachine: jest.fn(),
    updateMachine: jest.fn(),
    deleteMachine: jest.fn(),
  },
}));

const machine = {
  serial_number: 'SN-1',
  customer: 'Acme',
  state: 'S',
  account_type: 'T',
  model: 'M',
  status: '',
  brand: 'B',
  district: 'D',
  person_in_charge: 'P',
  reported_by: 'R',
  additional_notes: '',
  attachment: '',
  ppm_status: '',
  tnc_date: '2024-01-01',
  ppm_date: '2024-01-01',
  created_at: '',
  updated_at: '',
};

describe('useMachine', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('fetchMachine loads machine and clears loading', async () => {
    (MachineService.getMachine as jest.Mock).mockResolvedValue({ data: machine });

    const { result } = renderHook(() => useMachine());

    await act(async () => {
      await result.current.fetchMachine('SN-1');
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.machine).toEqual(machine);
    expect(result.current.error).toBeNull();
  });

  it('fetchMachine sets error and clears machine when response has no data', async () => {
    (MachineService.getMachine as jest.Mock).mockResolvedValue({});

    const { result } = renderHook(() => useMachine());

    await act(async () => {
      await result.current.fetchMachine('SN-1');
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.machine).toBeNull();
    expect(result.current.error?.message).toBe('The server returned an empty machine response.');
  });

  it('fetchMachine sets error when request fails', async () => {
    (MachineService.getMachine as jest.Mock).mockRejectedValue(new ApiError('gone', 404));

    const { result } = renderHook(() => useMachine());

    await act(async () => {
      await result.current.fetchMachine('missing');
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.machine).toBeNull();
    expect(result.current.error?.message).toBe('gone');
  });

  it('createMachine returns created machine and updates state', async () => {
    (MachineService.createMachine as jest.Mock).mockResolvedValue({ data: machine });

    const { result } = renderHook(() => useMachine());
    let created: typeof machine | null = null;

    await act(async () => {
      created = await result.current.createMachine({
        serial_number: 'SN-1',
        customer: 'Acme',
        state: 'S',
        account_type: 'T',
        model: 'M',
        status: '',
        brand: 'B',
        district: 'D',
        person_in_charge: 'P',
        reported_by: 'R',
        ppm_status: '',
        tnc_date: '2024-01-01',
        ppm_date: '2024-01-01',
      });
    });

    expect(created).toEqual(machine);
    expect(result.current.machine).toEqual(machine);
  });

  it('createMachine returns null and sets error when response has no data', async () => {
    (MachineService.createMachine as jest.Mock).mockResolvedValue({});

    const { result } = renderHook(() => useMachine());
    let created: typeof machine | null = null;

    await act(async () => {
      created = await result.current.createMachine({
        serial_number: 'SN-1',
        customer: 'Acme',
        state: 'S',
        account_type: 'T',
        model: 'M',
        status: '',
        brand: 'B',
        district: 'D',
        person_in_charge: 'P',
        reported_by: 'R',
        ppm_status: '',
        tnc_date: '2024-01-01',
        ppm_date: '2024-01-01',
      });
    });

    expect(created).toBeNull();
    expect(result.current.machine).toBeNull();
    expect(result.current.error?.message).toBe('The server returned an empty machine response.');
  });

  it('updateMachine leaves machine unchanged when response has no data', async () => {
    (MachineService.getMachine as jest.Mock).mockResolvedValue({ data: machine });
    (MachineService.updateMachine as jest.Mock).mockResolvedValue({});

    const { result } = renderHook(() => useMachine());

    await act(async () => {
      await result.current.fetchMachine('SN-1');
    });

    let updated: typeof machine | null = null;
    await act(async () => {
      updated = await result.current.updateMachine('SN-1', {
        serial_number: 'SN-1',
        customer: 'Other',
      });
    });

    expect(updated).toBeNull();
    expect(result.current.machine).toEqual(machine);
    expect(result.current.error?.message).toBe('The server returned an empty machine response.');
  });

  it('deleteMachine clears machine on success', async () => {
    (MachineService.getMachine as jest.Mock).mockResolvedValue({ data: machine });
    (MachineService.deleteMachine as jest.Mock).mockResolvedValue({ data: undefined });

    const { result } = renderHook(() => useMachine());

    await act(async () => {
      await result.current.fetchMachine('SN-1');
    });

    let ok = false;
    await act(async () => {
      ok = await result.current.deleteMachine('SN-1');
    });

    expect(ok).toBe(true);
    expect(result.current.machine).toBeNull();
  });
});
