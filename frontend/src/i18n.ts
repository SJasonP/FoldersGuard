import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import enUS from './locales/en-US';
import zhCN from './locales/zh-CN';
import ar from './locales/ar';
import fr from './locales/fr';
import ru from './locales/ru';
import es from './locales/es';

export const resources = {
    'en-US': {
        translation: enUS,
    },
    'zh-CN': {
        translation: zhCN,
    },
    ar: {translation: ar},
    fr: {translation: fr},
    ru: {translation: ru},
    es: {translation: es},
} as const;

export type SupportedLanguage = keyof typeof resources;
export type LanguageSetting = SupportedLanguage | 'system';

export function resolveSupportedLanguage(language: string | undefined): SupportedLanguage {
    const normalized = language?.toLowerCase() ?? '';
    if (normalized.startsWith('zh')) return 'zh-CN';
    if (normalized.startsWith('ar')) return 'ar';
    if (normalized.startsWith('fr')) return 'fr';
    if (normalized.startsWith('ru')) return 'ru';
    if (normalized.startsWith('es')) return 'es';
    return 'en-US';
}

export function resolveSystemLanguage(): SupportedLanguage {
    return resolveSupportedLanguage(typeof navigator !== 'undefined' ? navigator.language : undefined);
}

export function resolveLanguageSetting(setting: string | undefined, systemLanguage: SupportedLanguage): SupportedLanguage {
    if (setting && setting in resources) {
        return setting as SupportedLanguage;
    }
    return systemLanguage;
}

void i18n.use(initReactI18next).init({
    resources,
    lng: 'en-US',
    fallbackLng: 'en-US',
    interpolation: {
        escapeValue: false,
    },
});

export default i18n;
