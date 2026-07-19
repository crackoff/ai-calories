import { create } from 'zustand';
import i18n from '../i18n';

interface SettingsState {
  language: string;
  timezone: string;
  setLanguage: (lang: string) => void;
  setTimezone: (tz: string) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  language: 'en',
  timezone: 'UTC',

  setLanguage: (lang) => {
    i18n.changeLanguage(lang);
    set({ language: lang });
  },

  setTimezone: (tz) => set({ timezone: tz }),
}));
