import antdEnUS from 'antd/locale/en_US';
import antdZhCN from 'antd/locale/zh_CN';
import antdAr from 'antd/locale/ar_EG';
import antdFr from 'antd/locale/fr_FR';
import antdRu from 'antd/locale/ru_RU';
import antdEs from 'antd/locale/es_ES';
import type {SupportedLanguage} from './i18n';

export const antLocales: Record<SupportedLanguage, typeof antdEnUS> = {
    'en-US': antdEnUS,
    'zh-CN': antdZhCN,
    ar: antdAr,
    fr: antdFr,
    ru: antdRu,
    es: antdEs,
};
