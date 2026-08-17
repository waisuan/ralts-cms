import {
  isPublicAuthPath,
  shouldSuppressAuthRedirectOn401,
} from './tokens';

describe('tokens', () => {
  describe('isPublicAuthPath', () => {
    it('marks login and refresh as public', () => {
      expect(isPublicAuthPath('/api/v1/users/login')).toBe(true);
      expect(isPublicAuthPath('/api/v1/auth/refresh')).toBe(true);
    });
    it('marks registration POST path as public', () => {
      expect(isPublicAuthPath('/api/v1/users')).toBe(true);
    });
    it('does not mark protected routes as public', () => {
      expect(isPublicAuthPath('/api/v1/machines')).toBe(false);
    });
  });

  describe('shouldSuppressAuthRedirectOn401', () => {
    it('suppresses for public auth paths', () => {
      expect(shouldSuppressAuthRedirectOn401('/api/v1/users/login')).toBe(true);
    });
    it('suppresses for password change so wrong current password does not log user out', () => {
      expect(shouldSuppressAuthRedirectOn401('/api/v1/users/password')).toBe(true);
    });
    it('suppresses for the unread badge so a decorative fetch cannot log the user out', () => {
      expect(
        shouldSuppressAuthRedirectOn401('/api/v1/notifications/unread-count')
      ).toBe(true);
    });
    it('allows redirect for typical API 401', () => {
      expect(shouldSuppressAuthRedirectOn401('/api/v1/machines')).toBe(false);
      expect(shouldSuppressAuthRedirectOn401('/api/v1/notifications')).toBe(false);
    });
  });
});
