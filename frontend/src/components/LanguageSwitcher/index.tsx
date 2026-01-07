import { useTranslation } from 'react-i18next';
import styles from './LanguageSwitcher.module.less';

export const LanguageSwitcher = () => {
    const { i18n } = useTranslation();

    const changeLanguage = (lng: string) => {
        i18n.changeLanguage(lng);
        if (typeof localStorage !== 'undefined') {
            // eslint-disable-next-line no-undef
            localStorage.setItem('language', lng);
        }
    };

    return (
        <div className={styles.languageSwitcher}>
            <button
                onClick={() => changeLanguage('cs')}
                className={i18n.language === 'cs' ? styles.active : ''}
                title="Čeština"
            >
                CZ
            </button>
            <button
                onClick={() => changeLanguage('en')}
                className={i18n.language === 'en' ? styles.active : ''}
                title="English"
            >
                EN
            </button>
        </div>
    );
};
