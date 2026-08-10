'use client';

import React, { createContext, useContext, useState } from 'react';
import { useRouter } from 'next/navigation';
import { User } from '../types';
import { loginUser, logoutUser, changePassword } from '@/app/api/auth';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  changePassword: (form: { currentPassword: string; newPassword: string; confirmPassword: string }) => Promise<void>;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode; initialUser?: User | null }> = ({ children, initialUser = null }) => {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(initialUser);
  const [loading, setLoading] = useState(false);

  const login = async (username: string, password: string) => {
    setLoading(true);
    try {
      const res = await loginUser({ form: { username, password } });
      if (!res.success) {
        throw new Error(res.error);
      }
      setUser(res.user);
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    setLoading(true);
    try {
      await logoutUser();
      setUser(null);
      router.refresh();
      router.push('/login');
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const changePasswordAction = async (form: { currentPassword: string; newPassword: string; confirmPassword: string }) => {
    setLoading(true);
    try {
      const res = await changePassword(form);
      if (!res.success) {
        throw new Error(res.error);
      }
      setUser(res.user || null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        changePassword: changePasswordAction,
        isAuthenticated: !!user,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth deve ser utilizado dentro de um AuthProvider');
  }
  return context;
};
