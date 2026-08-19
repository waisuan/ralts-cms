import { renderHook } from '@testing-library/react';
import { useCanResolveFlags } from './useCanResolveFlags';
import { useOptionalAuth } from '../contexts/AuthContext';
import { USER_ROLE, type UserRole } from '../services/adminUserService';
import { machineFixture } from '../__fixtures__/machines';

jest.mock('../contexts/AuthContext', () => ({
  useOptionalAuth: jest.fn(),
}));

const signedInAs = (id: number, role: UserRole) => {
  (useOptionalAuth as jest.Mock).mockReturnValue({
    user: { id, username: 'u', email: 'e', role, approved: true },
  });
};

describe('useCanResolveFlags', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('lets an admin resolve flags on any machine', () => {
    signedInAs(1, USER_ROLE.ADMIN);

    const { result } = renderHook(() => useCanResolveFlags());

    expect(result.current(machineFixture({ assigned_user_id: 42 }))).toBe(true);
    expect(result.current(machineFixture({ assigned_user_id: null }))).toBe(true);
  });

  it('lets other users resolve flags on machines assigned to them', () => {
    signedInAs(7, USER_ROLE.NON_ADMIN);

    const { result } = renderHook(() => useCanResolveFlags());

    expect(result.current(machineFixture({ assigned_user_id: 7 }))).toBe(true);
    expect(result.current(machineFixture({ assigned_user_id: 8 }))).toBe(false);
  });

  it('does not let other users resolve flags on a free-text assignee machine', () => {
    // Such a machine has a name but no user id, so nobody but an admin owns it.
    signedInAs(7, USER_ROLE.NON_ADMIN);

    const { result } = renderHook(() => useCanResolveFlags());

    expect(result.current(machineFixture({ person_in_charge: 'Bob the contractor' }))).toBe(false);
  });

  it('refuses everything when nobody is signed in', () => {
    (useOptionalAuth as jest.Mock).mockReturnValue(null);

    const { result } = renderHook(() => useCanResolveFlags());

    expect(result.current(machineFixture({ assigned_user_id: 7 }))).toBe(false);
  });
});
