import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

// Web doesn't support SecureStore — fall back to memory
const memStore: Record<string, string> = {};

async function set(key: string, value: string): Promise<void> {
  if (Platform.OS === 'web') {
    memStore[key] = value;
    return;
  }
  await SecureStore.setItemAsync(key, value);
}

async function get(key: string): Promise<string | null> {
  if (Platform.OS === 'web') {
    return memStore[key] ?? null;
  }
  return SecureStore.getItemAsync(key);
}

async function remove(key: string): Promise<void> {
  if (Platform.OS === 'web') {
    delete memStore[key];
    return;
  }
  await SecureStore.deleteItemAsync(key);
}

export const storage = {
  getAccessToken: () => get(ACCESS_TOKEN_KEY),
  setAccessToken: (token: string) => set(ACCESS_TOKEN_KEY, token),
  getRefreshToken: () => get(REFRESH_TOKEN_KEY),
  setRefreshToken: (token: string) => set(REFRESH_TOKEN_KEY, token),
  clearTokens: async () => {
    await remove(ACCESS_TOKEN_KEY);
    await remove(REFRESH_TOKEN_KEY);
  },
};
