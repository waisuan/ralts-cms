'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { UserService } from '../services/userService';
import { isAuthError, SESSION_EXPIRED_EVENT } from '../utils/auth';
import { setSessionTokens, clearSessionAuthKeys } from '../utils/tokens';

interface AuthUser {
  id: number;
  username: string;
  email: string;
  role: string;
  approved: boolean;
  avatar?: string;
}

interface AuthContextType {
  user: AuthUser | null;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);





export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    const savedUser = localStorage.getItem('ralts_user');
    if (savedUser) {
      try {
        setUser(JSON.parse(savedUser));
      } catch (error) {
        console.error('Error parsing saved user:', error);
        localStorage.removeItem('ralts_user');
      }
    }
    setIsLoading(false);
  }, []);

  // The API client ends the session when the server stops accepting our
  // credentials. Dropping the user here swaps in the sign-in screen without a
  // page load, and unmounts the pages whose fetches would keep 401ing.
  useEffect(() => {
    const handleExpiry = () => setUser(null);
    window.addEventListener(SESSION_EXPIRED_EVENT, handleExpiry);
    return () => window.removeEventListener(SESSION_EXPIRED_EVENT, handleExpiry);
  }, []);

  const login = async (username: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await UserService.loginUser({ username, password });
      
      if (response.data) {
        const userData: AuthUser = {
          id: response.data.user.id,
          username: response.data.user.username,
          email: response.data.user.email,
          role: response.data.user.role,
          approved: response.data.user.approved,
          avatar: response.data.user.avatar,
        };

        setUser(userData);
        localStorage.setItem('ralts_user', JSON.stringify(userData));
        setSessionTokens(response.data.token, response.data.refresh_token);
        return { success: true };
      }
      
      return { success: false, error: 'Invalid username or password' };
    } catch (error) {
      console.error('Login error:', error);
      
      // Extract the error message from the API error
      let errorMessage = 'An error occurred during login';
      
      if (error && typeof error === 'object' && 'message' in error) {
        errorMessage = (error as { message: string }).message;
      } else if (error && typeof error === 'object' && 'details' in error) {
        // Handle case where error details contain the message
        const details = (error as { details: unknown }).details;
        if (typeof details === 'string') {
          errorMessage = details;
        } else if (details && typeof details === 'object' && details !== null && 'message' in details) {
          errorMessage = (details as { message: string }).message;
        }
      }
      
      // Don't redirect on login errors since we're already on the login page
      if (isAuthError(error)) {
        console.log('🔐 Login failed due to authentication error');
      }
      
      return { success: false, error: errorMessage };
    }
  };

  const logout = () => {
    setUser(null);
    const refreshToken =
      typeof window !== 'undefined' ? localStorage.getItem('ralts_refresh') : null;
    localStorage.removeItem('ralts_user');
    clearSessionAuthKeys();
    if (typeof window !== 'undefined' && refreshToken) {
      void fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    }
  };

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    logout,
    isAuthenticated: !!user,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

/**
 * Like useAuth but returns null instead of throwing when there is no provider
 * above. For components and hooks that merely tailor a supplementary detail to
 * the signed-in user and can safely fall back to showing nothing.
 */
export function useOptionalAuth(): AuthContextType | null {
  return useContext(AuthContext) ?? null;
}
