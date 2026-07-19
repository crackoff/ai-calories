import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from './en';
import es419 from './es-419';
import ptBR from './pt-BR';
import ru from './ru';
import de from './de';
import fr from './fr';

i18n.use(initReactI18next).init({
  resources: {
    en:       { translation: en },
    'es-419': { translation: es419 },
    'pt-BR':  { translation: ptBR },
    ru:       { translation: ru },
    de:       { translation: de },
    fr:       { translation: fr },
  },
  lng: 'en',
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
});

export default i18n;
