// Authentication utilities
import { clearSessionAuthKeys } from './tokens';

/**
 * Redirects the user to the login page and clears authentication data
 */
export function redirectToLogin(): void {
  // Clear authentication data
  if (typeof window !== 'undefined') {
    localStorage.removeItem('ralts_user');
    clearSessionAuthKeys();
  }
  
  // Redirect to login page
  if (typeof window !== 'undefined') {
    window.location.href = '/';
  }
}

/**
 * Checks if the current error is a 401 authentication error
 */
export function isAuthError(error: unknown): boolean {
  if (error && typeof error === 'object' && 'status' in error) {
    return (error as { status: number }).status === 401;
  }
  return false;
} 