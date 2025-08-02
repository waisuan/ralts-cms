'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { UserService } from '../services/userService';
import { isAuthError } from '../utils/auth';

interface AuthUser {
  id: number;
  name: string;
  email: string;
  role: string;
  avatar?: string;
}

interface AuthContextType {
  user: AuthUser | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<boolean>;
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

  const login = async (email: string, password: string): Promise<boolean> => {
    try {
      const response = await UserService.loginUser({ email, password });
      
      if (response.data) {
        const userData: AuthUser = {
          id: response.data.user.id,
          name: response.data.user.name,
          email: response.data.user.email,
          role: response.data.user.role,
          avatar: response.data.user.avatar,
        };

        setUser(userData);
        localStorage.setItem('ralts_user', JSON.stringify(userData));
        localStorage.setItem('ralts_token', response.data.token);
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('Login error:', error);
      // Don't redirect on login errors since we're already on the login page
      if (isAuthError(error)) {
        console.log('🔐 Login failed due to authentication error');
      }
      return false;
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
