import type { AuthResult } from '@/services/auth';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthUser {
  userId: string;
  email: string;
  role: string;
}

interface AuthState {
  accessToken: string | null;
  isAuthenticated: boolean;
  user: AuthUser | null;
  clearSession: () => void;
  setSession: (result: AuthResult) => void;
}

const toAuthUser = (result: AuthResult): AuthUser => ({
  userId: result.user_id,
  email: result.email,
  role: result.role,
});

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      isAuthenticated: false,
      user: null,

      clearSession: () => {
        set({
          accessToken: null,
          isAuthenticated: false,
          user: null,
        });
      },

      setSession: (result) => {
        set({
          accessToken: result.access_token,
          isAuthenticated: true,
          user: toAuthUser(result),
        });
      },
    }),
    {
      name: 'linkflow.auth',
    },
  ),
);
