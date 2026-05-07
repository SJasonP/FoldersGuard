import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { App as AntApp, ConfigProvider } from 'antd';
import type { ColumnsType } from 'antd/es/table';
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
import { LoadShareModal } from './components/share-actions/LoadShareModal';
import { ShareSessionLayer } from './components/share-actions/ShareSessionLayer';
import { AppShell } from './components/app/AppShell';
import { ProjectSessionLayer } from './components/project-actions/ProjectSessionLayer';

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
        <AppShell
          navigation={navigation}
          onNavigationChange={setNavigation}
          onCreateProject={() => setCreateDialogOpen(true)}
          onImportProject={() => setImportDialogOpen(true)}
          onLoadShare={() => setLoadShareDialogOpen(true)}
          language={language}
          onToggleLanguage={() => setLanguage(language === 'en-US' ? 'zh-CN' : 'en-US')}
          t={t}
        >
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
        </AppShell>
        <ProjectSessionLayer
          createDialogOpen={createDialogOpen}
          createLoading={createLoading}
          settings={settings}
          defaultSourceCleanup={defaultSourceCleanup}
          onCloseCreate={() => setCreateDialogOpen(false)}
          onCreateProject={(values) => void handleCreateProject(values)}
          importDialogOpen={importDialogOpen}
          importLoading={importLoading}
          onCloseImport={() => setImportDialogOpen(false)}
          onImportProject={(values) => void handleImportProject(values)}
          projectActionsOpen={projectActionsOpen}
          selectedProject={selectedProject}
          onCloseProjectActions={() => setProjectActionsOpen(false)}
          onOpenInspect={() => setInspectDialogOpen(true)}
          onOpenVerify={() => setVerifyDialogOpen(true)}
          onOpenExport={() => setExportDialogOpen(true)}
          onOpenDelete={() => setDeleteDialogOpen(true)}
          inspectDialogOpen={inspectDialogOpen}
          inspectLoading={inspectLoading}
          onCloseInspect={() => setInspectDialogOpen(false)}
          onInspectProject={(password) => void handleInspectProject(password)}
          inspectResultOpen={inspectResultOpen}
          inspectResult={inspectResult}
          onCloseInspectResult={() => setInspectResultOpen(false)}
          verifyDialogOpen={verifyDialogOpen}
          verifyLoading={verifyLoading}
          onCloseVerify={() => setVerifyDialogOpen(false)}
          onVerifyProject={(values) => void handleVerifyProject(values)}
          verifyResultOpen={verifyResultOpen}
          verifyResult={verifyResult}
          onCloseVerifyResult={() => setVerifyResultOpen(false)}
          exportDialogOpen={exportDialogOpen}
          exportLoading={exportLoading}
          onCloseExport={() => setExportDialogOpen(false)}
          onExportProject={(values) => void handleExportProject(values)}
          deleteDialogOpen={deleteDialogOpen}
          deleteLoading={deleteLoading}
          onCloseDelete={() => setDeleteDialogOpen(false)}
          onDeleteProject={(password) => void handleDeleteProject(password)}
          t={t}
        />
        <LoadShareModal
          open={loadShareDialogOpen}
          loading={shareLoading}
          onCancel={() => setLoadShareDialogOpen(false)}
          onSubmit={(values) => void handleLoadShare(values)}
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
