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
import { useProjectCreate } from './hooks/useProjectCreate';
import { useProjectImport } from './hooks/useProjectImport';
import { useProjectActions } from './hooks/useProjectActions';
import { useShareActions } from './hooks/useShareActions';
import { HomeView } from './views/HomeView';
import { SettingsView } from './views/SettingsView';
import { AboutView } from './views/AboutView';
import { CreateProjectModal } from './components/project-actions/CreateProjectModal';
import { ImportProjectModal } from './components/project-actions/ImportProjectModal';
import { ProjectActionsDrawer } from './components/project-actions/ProjectActionsDrawer';
import { InspectProjectModal } from './components/project-actions/InspectProjectModal';
import { InspectProjectDrawer } from './components/project-actions/InspectProjectDrawer';
import { VerifyProjectModal } from './components/project-actions/VerifyProjectModal';
import { VerifyProjectDrawer } from './components/project-actions/VerifyProjectDrawer';
import { ExportProjectModal } from './components/project-actions/ExportProjectModal';
import { DeleteProjectModal } from './components/project-actions/DeleteProjectModal';
import { LoadShareModal } from './components/share-actions/LoadShareModal';
import { ShareSessionLayer } from './components/share-actions/ShareSessionLayer';

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
    createDialogOpen,
    createLoading,
    defaultSourceCleanup,
    setCreateDialogOpen,
    handleCreateProject,
  } = useProjectCreate({
    messageApi: antApp.message,
    t,
    settings,
    reloadProjects: loadProjects,
  });

  const {
    importDialogOpen,
    importLoading,
    setImportDialogOpen,
    handleImportProject,
  } = useProjectImport({
    messageApi: antApp.message,
    t,
    reloadProjects: loadProjects,
  });

  const {
    closeShareSession,
    decryptShareDialogOpen,
    decryptShareLoading,
    decryptShareResult,
    decryptShareResultOpen,
    handleDecryptShare,
    handleLoadShare,
    handleVerifyShare,
    inspectShareOpen,
    loadShareDialogOpen,
    loadedShare,
    setLoadShareDialogOpen,
    setDecryptShareDialogOpen,
    setDecryptShareResultOpen,
    setInspectShareOpen,
    setVerifyShareDialogOpen,
    setVerifyShareResultOpen,
    shareActionsOpen,
    shareLoading,
    verifyShareDialogOpen,
    verifyShareLoading,
    verifyShareResult,
    verifyShareResultOpen,
  } = useShareActions({
    messageApi: antApp.message,
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
    verifyDialogOpen,
    verifyLoading,
    verifyResult,
    verifyResultOpen,
    setDeleteDialogOpen,
    setExportDialogOpen,
    setInspectDialogOpen,
    setInspectResultOpen,
    setProjectActionsOpen,
    setVerifyDialogOpen,
    setVerifyResultOpen,
    openProjectActions,
    handleDeleteProject,
    handleExportProject,
    handleInspectProject,
    handleVerifyProject,
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
                <Button icon={<FolderAddOutlined />} type="primary" onClick={() => setCreateDialogOpen(true)}>
                  {t('createProject')}
                </Button>
                <Button icon={<ImportOutlined />} onClick={() => setImportDialogOpen(true)}>
                  {t('importProject')}
                </Button>
                <Button icon={<ShareAltOutlined />} onClick={() => setLoadShareDialogOpen(true)}>
                  {t('loadShare')}
                </Button>
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
          onVerify={() => setVerifyDialogOpen(true)}
          onExport={() => setExportDialogOpen(true)}
          onDelete={() => setDeleteDialogOpen(true)}
          t={t}
        />
        <CreateProjectModal
          open={createDialogOpen}
          loading={createLoading}
          settings={settings}
          defaultSourceCleanup={defaultSourceCleanup}
          onCancel={() => setCreateDialogOpen(false)}
          onSubmit={(values) => void handleCreateProject(values)}
          t={t}
        />
        <ImportProjectModal
          open={importDialogOpen}
          loading={importLoading}
          onCancel={() => setImportDialogOpen(false)}
          onSubmit={(values) => void handleImportProject(values)}
          t={t}
        />
        <LoadShareModal
          open={loadShareDialogOpen}
          loading={shareLoading}
          onCancel={() => setLoadShareDialogOpen(false)}
          onSubmit={(values) => void handleLoadShare(values)}
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
        <VerifyProjectModal
          open={verifyDialogOpen}
          loading={verifyLoading}
          onCancel={() => setVerifyDialogOpen(false)}
          onSubmit={(values) => void handleVerifyProject(values)}
          t={t}
        />
        <VerifyProjectDrawer
          open={verifyResultOpen}
          result={verifyResult}
          onClose={() => setVerifyResultOpen(false)}
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
        <ShareSessionLayer
          decryptShareDialogOpen={decryptShareDialogOpen}
          decryptShareLoading={decryptShareLoading}
          decryptShareResult={decryptShareResult}
          decryptShareResultOpen={decryptShareResultOpen}
          defaultSourceCleanup={defaultSourceCleanup}
          shareActionsOpen={shareActionsOpen}
          verifyShareDialogOpen={verifyShareDialogOpen}
          verifyShareLoading={verifyShareLoading}
          verifyShareResult={verifyShareResult}
          verifyShareResultOpen={verifyShareResultOpen}
          loadedShare={loadedShare}
          inspectShareOpen={inspectShareOpen}
          onCloseShareSession={closeShareSession}
          onOpenDecryptShare={() => setDecryptShareDialogOpen(true)}
          onOpenInspectShare={() => setInspectShareOpen(true)}
          onCloseInspectShare={() => setInspectShareOpen(false)}
          onCloseDecryptShare={() => setDecryptShareDialogOpen(false)}
          onDecryptShare={(values) => void handleDecryptShare(values)}
          onCloseDecryptShareResult={() => setDecryptShareResultOpen(false)}
          onOpenVerifyShare={() => setVerifyShareDialogOpen(true)}
          onCloseVerifyShare={() => setVerifyShareDialogOpen(false)}
          onVerifyShare={(values) => void handleVerifyShare(values)}
          onCloseVerifyShareResult={() => setVerifyShareResultOpen(false)}
          t={t}
        />
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
