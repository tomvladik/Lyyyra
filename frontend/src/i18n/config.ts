import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import cs from './locales/cs.json';
import en from './locales/en.json';

// eslint-disable-next-line no-undef
const savedLanguage = typeof localStorage !== 'undefined' ? localStorage.getItem('language') || 'cs' : 'cs';

i18n
    .use(initReactI18next)
    .init({
        resources: {
            cs: { translation: cs },
            en: { translation: en }
        },
        lng: savedLanguage,
        fallbackLng: 'cs',
        interpolation: {
            escapeValue: false // React already escapes values
        }
    });

export default i18n;
