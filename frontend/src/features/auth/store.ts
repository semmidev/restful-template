import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { loginRequest, registerRequest } from './api';
import { LoginInput, RegisterInput } from '../../lib/schemas';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresIn: number | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (input: LoginInput) => Promise<{ success: boolean; error?: string }>;
  register: (input: RegisterInput) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      expiresIn: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (input) => {
        set({ isLoading: true, error: null });
        try {
          const res = await loginRequest(input);
          const { access_token, refresh_token, expires_in } = res.data.data;

          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);
          localStorage.setItem('expires_in', expires_in.toString());

          set({
            accessToken: access_token,
            refreshToken: refresh_token,
            expiresIn: expires_in,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
          return { success: true };
        } catch (err: any) {
          const errMsg = err.response?.data?.detail || 'Invalid email or password';
          set({ error: errMsg, isLoading: false });
          return { success: false, error: errMsg };
        }
      },

      register: async (input) => {
        set({ isLoading: true, error: null });
        try {
          const res = await registerRequest({
            email: input.email,
            password: input.password,
          });
          const { access_token, refresh_token, expires_in } = res.data.data;

          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);
          localStorage.setItem('expires_in', expires_in.toString());

          set({
            accessToken: access_token,
            refreshToken: refresh_token,
            expiresIn: expires_in,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
          return { success: true };
        } catch (err: any) {
          const errMsg = err.response?.data?.detail || 'Registration failed. Password must be 8-72 chars.';
          set({ error: errMsg, isLoading: false });
          return { success: false, error: errMsg };
        }
      },

      logout: () => {
        localStorage.clear();
        set({
          accessToken: null,
          refreshToken: null,
          expiresIn: null,
          isAuthenticated: false,
          error: null,
        });
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'todoapp-auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        expiresIn: state.expiresIn,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
export default useAuthStore;
