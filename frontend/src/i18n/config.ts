import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import cs from './locales/cs.json';
import en from './locales/en.json';

const getSavedLanguage = (): string => {
    if (typeof window === 'undefined' || !window.localStorage) {
        return 'cs';
    }
    return window.localStorage.getItem('language') || 'cs';
};

const savedLanguage = getSavedLanguage();

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
