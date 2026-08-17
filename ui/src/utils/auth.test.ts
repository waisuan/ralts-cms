import { endSession, isAuthError, SESSION_EXPIRED_EVENT } from './auth';

describe('isAuthError', () => {
  it('recognises a 401 and nothing else', () => {
    expect(isAuthError({ status: 401 })).toBe(true);
    expect(isAuthError({ status: 403 })).toBe(false);
    expect(isAuthError(new Error('boom'))).toBe(false);
    expect(isAuthError(null)).toBe(false);
  });
});

describe('endSession', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('forgets the user and tokens, then announces the expiry', () => {
    localStorage.setItem('ralts_user', JSON.stringify({ id: 1 }));
    localStorage.setItem('ralts_token', 'access');
    localStorage.setItem('ralts_refresh', 'refresh');
    const listener = jest.fn();
    window.addEventListener(SESSION_EXPIRED_EVENT, listener);

    endSession();

    expect(localStorage.getItem('ralts_user')).toBeNull();
    expect(localStorage.getItem('ralts_token')).toBeNull();
    expect(localStorage.getItem('ralts_refresh')).toBeNull();
    expect(listener).toHaveBeenCalledTimes(1);

    window.removeEventListener(SESSION_EXPIRED_EVENT, listener);
  });

  it('leaves the location alone, so a page still loading is not aborted', () => {
    const before = window.location.href;

    endSession();

    expect(window.location.href).toBe(before);
  });
});
