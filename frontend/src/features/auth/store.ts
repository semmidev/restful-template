import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { loginRequest, registerRequest, googleLoginRequest, logoutRequest } from './api';
import { LoginInput, RegisterInput } from '../../lib/schemas';

interface AuthState {
  userEmail: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (input: LoginInput) => Promise<{ success: boolean; error?: string }>;
  register: (input: RegisterInput) => Promise<{ success: boolean; error?: string }>;
  loginWithGoogle: (code: string, codeVerifier: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      userEmail: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (input) => {
        set({ isLoading: true, error: null });
        try {
          const res = await loginRequest(input);
          const { user } = res.data;

          set({
            userEmail: user.email,
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
          const { user } = res.data;

          set({
            userEmail: user.email,
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

      loginWithGoogle: async (code, codeVerifier) => {
        set({ isLoading: true, error: null });
        try {
          const res = await googleLoginRequest(code, codeVerifier);
          const { user } = res.data;

          set({
            userEmail: user.email,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          });
          return { success: true };
        } catch (err: any) {
          const errMsg = err.response?.data?.detail || 'Google login failed';
          set({ error: errMsg, isLoading: false });
          return { success: false, error: errMsg };
        }
      },

      logout: async () => {
        try {
          await logoutRequest();
        } catch (err) {
          // ignore or log
        }
        set({
          userEmail: null,
          isAuthenticated: false,
          error: null,
        });
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'todoapp-auth',
      partialize: (state) => ({
        userEmail: state.userEmail,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

export default useAuthStore;
