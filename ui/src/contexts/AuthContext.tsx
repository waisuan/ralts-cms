'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { UserService } from '../services/userService';
import { isAuthError } from '../utils/auth';

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
        localStorage.setItem('ralts_token', response.data.token);
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
    localStorage.removeItem('ralts_user');
    localStorage.removeItem('ralts_token');
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
