import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

export const resources = {
  'en-US': {
    translation: {
      about: 'About',
      appId: 'App ID',
      availabilityStatus: 'Availability Status',
      cliAlias: 'CLI Alias',
      cliExecutable: 'CLI Executable',
      createProject: 'Create Project',
      dataDirectory: 'Data Directory',
      foldersGuard: 'FoldersGuard',
      formatVersion: 'Format Version',
      home: 'Home',
      importProject: 'Import Project',
      loadShare: 'Load Share',
      localProjects: 'Local Projects',
      modifiedTime: 'Modified Time',
      noProjects: 'No local projects found.',
      projectId: 'Project ID',
      projectName: 'Project Name',
      refresh: 'Refresh',
      schemaVersion: 'Schema Version',
      searchProjects: 'Search projects',
      settings: 'Settings',
      startSubtitle: 'Protect folder contents with encrypted metadata and content objects.',
      unavailable: 'Unavailable',
    },
  },
  'zh-CN': {
    translation: {
      about: '关于',
      appId: 'App ID',
      availabilityStatus: '可用状态',
      cliAlias: 'CLI 短名称',
      cliExecutable: 'CLI 可执行名称',
      createProject: '创建项目',
      dataDirectory: '数据目录',
      foldersGuard: 'FoldersGuard',
      formatVersion: '格式版本',
      home: '首页',
      importProject: '导入项目',
      loadShare: '加载分享',
      localProjects: '本地项目',
      modifiedTime: '修改时间',
      noProjects: '未找到本地项目。',
      projectId: '项目 ID',
      projectName: '项目名称',
      refresh: '刷新',
      schemaVersion: 'Schema 版本',
      searchProjects: '搜索项目',
      settings: '设置',
      startSubtitle: '使用加密元数据和加密内容对象保护文件夹内容。',
      unavailable: '不可用',
    },
  },
} as const;

export type SupportedLanguage = keyof typeof resources;

void i18n.use(initReactI18next).init({
  resources,
  lng: 'en-US',
  fallbackLng: 'en-US',
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
