import { act, render, screen } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { SESSION_EXPIRED_EVENT } from '../utils/auth';

function SessionState() {
  const { isAuthenticated, user } = useAuth();
  return <div>{isAuthenticated ? `signed in as ${user?.username}` : 'signed out'}</div>;
}

describe('AuthProvider', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('restores a saved session on mount', async () => {
    localStorage.setItem(
      'ralts_user',
      JSON.stringify({ id: 1, username: 'ada', email: 'a@b.c', role: 'ADMIN', approved: true })
    );

    render(
      <AuthProvider>
        <SessionState />
      </AuthProvider>
    );

    expect(await screen.findByText('signed in as ada')).toBeInTheDocument();
  });

  it('signs the user out when the API reports the session has expired', async () => {
    localStorage.setItem(
      'ralts_user',
      JSON.stringify({ id: 1, username: 'ada', email: 'a@b.c', role: 'ADMIN', approved: true })
    );

    render(
      <AuthProvider>
        <SessionState />
      </AuthProvider>
    );
    expect(await screen.findByText('signed in as ada')).toBeInTheDocument();

    act(() => {
      window.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
    });

    expect(screen.getByText('signed out')).toBeInTheDocument();
  });
});
