import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { storage } from './storage';

// ---------------------------------------------------------------------------
// Base URL — override with EXPO_PUBLIC_API_URL env var
// ---------------------------------------------------------------------------
const BASE_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

// ---------------------------------------------------------------------------
// Axios instance
// ---------------------------------------------------------------------------
const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  timeout: 15_000,
});

// Attach Bearer token to every request
api.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  const token = await storage.getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// On 401 try to refresh once, then give up
let refreshing = false;
api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry && !refreshing) {
      original._retry = true;
      refreshing = true;
      try {
        const refreshToken = await storage.getRefreshToken();
        if (!refreshToken) throw new Error('no refresh token');
        const { data } = await axios.post<AuthResponse>(`${BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });
        await storage.setAccessToken(data.access_token);
        await storage.setRefreshToken(data.refresh_token);
        original.headers.Authorization = `Bearer ${data.access_token}`;
        return api(original);
      } catch {
        await storage.clearTokens();
        return Promise.reject(error);
      } finally {
        refreshing = false;
      }
    }
    return Promise.reject(error);
  }
);

// ---------------------------------------------------------------------------
// Types matching Go backend responses
// ---------------------------------------------------------------------------
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
}

export interface FoodEntryResponse {
  id: number;
  food_item: string;
  weight: number;
  calories: number;
  protein: number;
  fat: number;
  carbohydrates: number;
  from_cache: boolean;
  timestamp: string;
}

export interface FoodCacheSearchResult {
  id: number;
  food_name: string;
  calories_100g: number;
  protein_100g: number;
  fat_100g: number;
  carbs_100g: number;
  image_url: string | null;
}

export interface MealGroup {
  period: 'Morning' | 'Afternoon' | 'Evening';
  entries: FoodEntryResponse[];
}

export interface MacrosBreakdown {
  protein_pct: number;
  fat_pct: number;
  carbs_pct: number;
}

export interface NutritionSummary {
  date: string;
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbohydrates: number;
  meals: MealGroup[];
  macros_breakdown: MacrosBreakdown;
}

export interface HistoryDataPoint {
  date: string;
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
}

export interface FoodHistoryResponse {
  period: string;
  data: HistoryDataPoint[];
}

export interface UserProfileResponse {
  id: number;
  email: string | null;
  auth_provider: string | null;
  language: string;
  timezone: string;
}

export interface CurrentPaymentResponse {
  sku: string;
  payment_date: string;
  expiration_date: string;
  amount: number;
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------
export const authApi = {
  register: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/register', { email, password }).then((r) => r.data),

  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }).then((r) => r.data),

  googleLogin: (idToken: string) =>
    api.post<AuthResponse>('/auth/google', { id_token: idToken }).then((r) => r.data),

  appleLogin: (idToken: string) =>
    api.post<AuthResponse>('/auth/apple', { id_token: idToken }).then((r) => r.data),

  refresh: (refreshToken: string) =>
    api.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken }).then((r) => r.data),
};

// ---------------------------------------------------------------------------
// Food
// ---------------------------------------------------------------------------
export interface LogFoodRequest {
  food_cache_id?: number;
  free_text?: string;
  input_mode: 'grams' | 'kcal';
  value: number;
}

export const foodApi = {
  log: (req: LogFoodRequest) =>
    api.post<FoodEntryResponse>('/food', req).then((r) => r.data),

  getTodaySummary: () =>
    api.get<NutritionSummary>('/food/summary/today').then((r) => r.data),

  getDateSummary: (date: string) =>
    api.get<NutritionSummary>(`/food/summary/${date}`).then((r) => r.data),

  getHistory: (period: 'week' | 'month' | 'year') =>
    api.get<FoodHistoryResponse>('/food/history', { params: { period } }).then((r) => r.data),

  deleteLast: () => api.delete('/food/last'),

  deleteById: (id: number) => api.delete(`/food/${id}`),
};

// ---------------------------------------------------------------------------
// Food cache (autocomplete)
// ---------------------------------------------------------------------------
export const foodCacheApi = {
  search: (q: string) =>
    api.get<FoodCacheSearchResult[]>('/food-cache/search', { params: { q } }).then((r) => r.data),

  getById: (id: number) =>
    api.get<FoodCacheSearchResult>(`/food-cache/${id}`).then((r) => r.data),
};

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------
export const userApi = {
  getProfile: () =>
    api.get<UserProfileResponse>('/user/profile').then((r) => r.data),

  updateTimezone: (timezone: string) =>
    api.put('/user/timezone', { timezone }),

  updateLanguage: (language: string) =>
    api.put('/user/language', { language }),
};

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------
export const paymentApi = {
  getCurrent: () =>
    api.get<CurrentPaymentResponse | null>('/payments/current').then((r) => r.data),

  getHistory: () =>
    api.get<CurrentPaymentResponse[]>('/payments/history').then((r) => r.data),
};
