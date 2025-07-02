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
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Default avatar for users without one
const getDefaultAvatar = (name: string): string => {
  // Generate initials from name
  const initials = name
    .split(' ')
    .map((word) => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2);

  // Use a placeholder service that generates avatars with initials
  return `https://ui-avatars.com/api/?name=${encodeURIComponent(initials)}&background=random&color=fff&size=150`;
};

// Function to fetch users from the file
const fetchUsersFromFile = async (): Promise<
  Array<{
    id: string;
    name: string;
    email: string;
    password: string;
    role: 'admin' | 'user';
    avatar?: string;
  }>
> => {
  try {
    const response = await fetch('/api/users');
    if (response.ok) {
      const data = await response.json();
      return data.users || [];
    }
  } catch (error) {
    console.error('Error fetching users:', error);
  }
  return [];
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [allUsers, setAllUsers] = useState<
    Array<{
      id: string;
      name: string;
      email: string;
      password: string;
      role: 'admin' | 'user';
      avatar?: string;
    }>
  >([]);

  // Load users from file on mount
  useEffect(() => {
    const loadUsers = async () => {
      const fileUsers = await fetchUsersFromFile();

      // Add default avatars to users who don't have one
      const usersWithAvatars = fileUsers.map((user) => ({
        ...user,
        avatar: user.avatar || getDefaultAvatar(user.name),
      }));

      setAllUsers(usersWithAvatars);
    };

    loadUsers();
  }, []);

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
    // Simulate API delay
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Use allUsers (from file) for authentication
    const foundUser = allUsers.find((u) => u.email === email && u.password === password);

    if (foundUser) {
      const userData: User = {
        id: foundUser.id,
        name: foundUser.name,
        email: foundUser.email,
        role: foundUser.role,
        avatar: foundUser.avatar,
      };

      setUser(userData);
      localStorage.setItem('ralts_user', JSON.stringify(userData));
      return true;
    }

    return false;
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('ralts_user');
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
