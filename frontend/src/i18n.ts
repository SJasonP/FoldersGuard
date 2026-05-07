import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import enUS from './locales/en-US';
import zhCN from './locales/zh-CN';

export const resources = {
  'en-US': {
    translation: enUS,
  },
  'zh-CN': {
    translation: zhCN,
  },
} as const;

export type SupportedLanguage = keyof typeof resources;
export type LanguageSetting = SupportedLanguage | 'system';

export function resolveSupportedLanguage(language: string | undefined): SupportedLanguage {
  return language?.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US';
}

export function resolveSystemLanguage(): SupportedLanguage {
  return resolveSupportedLanguage(typeof navigator !== 'undefined' ? navigator.language : undefined);
}

export function resolveLanguageSetting(setting: string | undefined, systemLanguage: SupportedLanguage): SupportedLanguage {
  if (setting === 'zh-CN' || setting === 'en-US') {
    return setting;
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
