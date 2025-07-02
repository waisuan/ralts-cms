'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'user';
  avatar?: string;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  isAuthenticated: boolean;
  error: string;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Mock user data
const mockUsers = [
  {
    id: '1',
    name: 'John Doe',
    email: 'john@example.com',
    password: 'password123',
    role: 'admin' as const,
    avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=150&h=150&fit=crop&crop=face',
  },
  {
    id: '2',
    name: 'Jane Smith',
    email: 'jane@example.com',
    password: 'password123',
    role: 'user' as const,
    avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=150&h=150&fit=crop&crop=face',
  },
];

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

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
    setIsLoading(true);
    setError(''); // Clear any previous errors
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const mockUser = mockUsers.find(u => u.email === email && u.password === password);
    
    if (mockUser) {
      const userData: User = {
        id: mockUser.id,
        name: mockUser.name,
        email: mockUser.email,
        role: mockUser.role,
        avatar: mockUser.avatar,
      };
      
      setUser(userData);
      localStorage.setItem('ralts_user', JSON.stringify(userData));
      setIsLoading(false);
      return true;
    }
    
    setError('Invalid email or password');
    setIsLoading(false);
    return false;
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('ralts_user');
  };

  const clearError = () => {
    setError('');
  };

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    logout,
    isAuthenticated: !!user,
    error,
    clearError,
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