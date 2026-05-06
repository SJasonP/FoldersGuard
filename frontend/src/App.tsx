import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { App as AntApp, Button, ConfigProvider, Layout, Menu, Space, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  FolderAddOutlined,
  HomeOutlined,
  ImportOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  SettingOutlined,
  ShareAltOutlined,
} from '@ant-design/icons';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import { AppInfo } from '../wailsjs/go/main/App';
import type { SupportedLanguage } from './i18n';
import i18n from './i18n';
import { resolveTheme, themeAlgorithm, type ThemeMode } from './theme';
import type { AppInfoModel, LocalProjectRow, NavigationKey } from './types';
import { useAppSettings } from './hooks/useAppSettings';
import { useLocalProjects } from './hooks/useLocalProjects';
import { useProjectActions } from './hooks/useProjectActions';
import { HomeView } from './views/HomeView';
import { SettingsView } from './views/SettingsView';
import { AboutView } from './views/AboutView';
import { ProjectActionsDrawer } from './components/project-actions/ProjectActionsDrawer';
import { InspectProjectModal } from './components/project-actions/InspectProjectModal';
import { InspectProjectDrawer } from './components/project-actions/InspectProjectDrawer';
import { ExportProjectModal } from './components/project-actions/ExportProjectModal';
import { DeleteProjectModal } from './components/project-actions/DeleteProjectModal';

const antLocales: Record<SupportedLanguage, typeof enUS> = {
  'en-US': enUS,
  'zh-CN': zhCN,
};

function App() {
  const { t } = useTranslation();
  const antApp = AntApp.useApp();
  const [navigation, setNavigation] = useState<NavigationKey>('home');
  const [language, setLanguage] = useState<'en-US' | 'zh-CN'>('en-US');
  const [themeMode, setThemeMode] = useState<ThemeMode>('system');
  const [resolvedTheme, setResolvedTheme] = useState(resolveTheme(themeMode));
  const [info, setInfo] = useState<AppInfoModel | null>(null);

  useEffect(() => {
    AppInfo().then(setInfo).catch(() => setInfo(null));
  }, []);

  useEffect(() => {
    void i18n.changeLanguage(language);
  }, [language]);

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const update = () => setResolvedTheme(resolveTheme(themeMode));
    update();
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, [themeMode]);

  const {
    settings,
    settingsLoading,
    settingsSaving,
    handleSaveSettings,
    handleClearRecentPaths,
  } = useAppSettings({
    messageApi: antApp.message,
    t,
    setLanguage,
    setThemeMode,
  });

  const {
    projectSearch,
    setProjectSearch,
    projectsLoading,
    projectsError,
    selectedProject,
    selectedProjectId,
    setSelectedProjectId,
    visibleProjects,
    loadProjects,
  } = useLocalProjects({
    language,
    t,
  });

  const {
    deleteDialogOpen,
    deleteLoading,
    exportDialogOpen,
    exportLoading,
    inspectDialogOpen,
    inspectLoading,
    inspectResult,
    inspectResultOpen,
    projectActionsOpen,
    setDeleteDialogOpen,
    setExportDialogOpen,
    setInspectDialogOpen,
    setInspectResultOpen,
    setProjectActionsOpen,
    openProjectActions,
    handleDeleteProject,
    handleExportProject,
    handleInspectProject,
  } = useProjectActions({
    messageApi: antApp.message,
    t,
    selectedProjectId,
    selectedProject,
    reloadProjects: loadProjects,
    clearSelectedProject: () => setSelectedProjectId(null),
  });

  const columns = useMemo<ColumnsType<LocalProjectRow>>(
    () => [
      { title: t('projectId'), dataIndex: 'projectId', key: 'projectId' },
      { title: t('projectName'), dataIndex: 'fileName', key: 'fileName' },
      { title: t('modifiedTime'), dataIndex: 'modifiedTime', key: 'modifiedTime' },
      { title: t('availabilityStatus'), dataIndex: 'availabilityStatus', key: 'availabilityStatus' },
    ],
    [t],
  );

  return (
    <ConfigProvider
      locale={antLocales[language]}
      theme={{
        algorithm: themeAlgorithm(resolvedTheme),
        token: {
          borderRadius: 6,
          colorPrimary: '#1677ff',
        },
      }}
    >
      <AntApp>
        <Layout className="app-shell">
          <Layout.Sider width={236} className="app-sidebar">
            <div className="app-brand">
              <Typography.Title level={4}>{t('foldersGuard')}</Typography.Title>
              <Typography.Text type="secondary">{t('startSubtitle')}</Typography.Text>
            </div>
            <Menu
              mode="inline"
              selectedKeys={[navigation]}
              onClick={({ key }) => setNavigation(key as NavigationKey)}
              items={[
                { key: 'home', icon: <HomeOutlined />, label: t('home') },
                { key: 'settings', icon: <SettingOutlined />, label: t('settings') },
                { key: 'about', icon: <InfoCircleOutlined />, label: t('about') },
              ]}
            />
          </Layout.Sider>
          <Layout>
            <Layout.Header className="app-header">
              <Space>
                <Button icon={<FolderAddOutlined />} type="primary">
                  {t('createProject')}
                </Button>
                <Button icon={<ImportOutlined />}>{t('importProject')}</Button>
                <Button icon={<ShareAltOutlined />}>{t('loadShare')}</Button>
              </Space>
              <Space>
                <Button onClick={() => setLanguage(language === 'en-US' ? 'zh-CN' : 'en-US')}>
                  {language}
                </Button>
              </Space>
            </Layout.Header>
            <Layout.Content className="app-content">
              {navigation === 'home' && (
                <HomeView
                  columns={columns}
                  loading={projectsLoading}
                  projects={visibleProjects}
                  projectSearch={projectSearch}
                  projectsError={projectsError}
                  selectedProjectId={selectedProjectId}
                  onProjectSearchChange={setProjectSearch}
                  onRefresh={() => void loadProjects()}
                  onSelectProject={setSelectedProjectId}
                  onOpenProjectActions={openProjectActions}
                  t={t}
                />
              )}
              {navigation === 'settings' && (
                <SettingsView
                  settings={settings}
                  loading={settingsLoading}
                  saving={settingsSaving}
                  onSave={(values) => void handleSaveSettings(values)}
                  onClearRecentPaths={() => void handleClearRecentPaths()}
                  t={t}
                />
              )}
              {navigation === 'about' && <AboutView info={info} t={t} />}
            </Layout.Content>
          </Layout>
        </Layout>
        <ProjectActionsDrawer
          open={projectActionsOpen}
          project={selectedProject}
          onClose={() => setProjectActionsOpen(false)}
          onInspect={() => setInspectDialogOpen(true)}
          onExport={() => setExportDialogOpen(true)}
          onDelete={() => setDeleteDialogOpen(true)}
          t={t}
        />
        <InspectProjectModal
          open={inspectDialogOpen}
          loading={inspectLoading}
          onCancel={() => setInspectDialogOpen(false)}
          onSubmit={(password) => void handleInspectProject(password)}
          t={t}
        />
        <InspectProjectDrawer
          open={inspectResultOpen}
          result={inspectResult}
          onClose={() => setInspectResultOpen(false)}
          t={t}
        />
        <ExportProjectModal
          open={exportDialogOpen}
          loading={exportLoading}
          onCancel={() => setExportDialogOpen(false)}
          onSubmit={(values) => void handleExportProject(values)}
          t={t}
        />
        <DeleteProjectModal
          open={deleteDialogOpen}
          loading={deleteLoading}
          onCancel={() => setDeleteDialogOpen(false)}
          onSubmit={(password) => void handleDeleteProject(password)}
          t={t}
        />
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
