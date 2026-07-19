import { create } from 'zustand';
import { storage } from '../services/storage';
import { authApi, AuthResponse } from '../services/api';

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  // Actions
  initialize: () => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setTokens: (resp: AuthResponse) => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  isLoading: true,

  initialize: async () => {
    const token = await storage.getAccessToken();
    set({ isAuthenticated: !!token, isLoading: false });
  },

  setTokens: async (resp: AuthResponse) => {
    await storage.setAccessToken(resp.access_token);
    await storage.setRefreshToken(resp.refresh_token);
    set({ isAuthenticated: true });
  },

  login: async (email, password) => {
    const resp = await authApi.login(email, password);
    await storage.setAccessToken(resp.access_token);
    await storage.setRefreshToken(resp.refresh_token);
    set({ isAuthenticated: true });
  },

  register: async (email, password) => {
    const resp = await authApi.register(email, password);
    await storage.setAccessToken(resp.access_token);
    await storage.setRefreshToken(resp.refresh_token);
    set({ isAuthenticated: true });
  },

  logout: async () => {
    await storage.clearTokens();
    set({ isAuthenticated: false });
  },
}));
