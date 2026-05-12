import {useEffect, useMemo, useRef, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {App as AntApp, ConfigProvider} from 'antd';
import type {ColumnsType} from 'antd/es/table';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import {AppInfo, SetLongRunningOperationActive, SetManualContentGuideCloseGuardActive} from '../wailsjs/go/main/App';
import i18n, {resolveSystemLanguage, type SupportedLanguage} from './i18n';
import {resolveTheme, themeAlgorithm, type ThemeMode} from './theme';
import type {AppInfoModel, LocalProjectRow, NavigationKey} from './types';
import {useAppSettings} from './hooks/useAppSettings';
import {useLocalProjects} from './hooks/useLocalProjects';
import {useProjectCreate} from './hooks/useProjectCreate';
import {useProjectImport} from './hooks/useProjectImport';
import {useProjectActions} from './hooks/useProjectActions';
import {useProjectShare} from './hooks/useProjectShare';
import {useProjectBrowser} from './hooks/useProjectBrowser';
import {useShareActions} from './hooks/useShareActions';
import {HomeView} from './views/HomeView';
import {SettingsView} from './views/SettingsView';
import {AboutView} from './views/AboutView';
import {LoadShareModal} from './components/share-actions/LoadShareModal';
import {ShareSessionLayer} from './components/share-actions/ShareSessionLayer';
import {AppShell} from './components/app/AppShell';
import {ProjectSessionLayer} from './components/project-actions/ProjectSessionLayer';
import {ProjectBrowserLayer} from './components/project-browser/ProjectBrowserLayer';
import {showStartupError} from './components/common/operationError';

const antLocales: Record<SupportedLanguage, typeof enUS> = {
    'en-US': enUS,
    'zh-CN': zhCN,
};

function App() {
    const {t} = useTranslation();
    const antApp = AntApp.useApp();
    const [navigation, setNavigation] = useState<NavigationKey>('home');
    const [systemLanguage, setSystemLanguage] = useState<SupportedLanguage>(resolveSystemLanguage);
    const [language, setLanguage] = useState<SupportedLanguage>(systemLanguage);
    const [themeMode, setThemeMode] = useState<ThemeMode>('system');
    const [resolvedTheme, setResolvedTheme] = useState(resolveTheme(themeMode));
    const [info, setInfo] = useState<AppInfoModel | null>(null);
    const [infoLoaded, setInfoLoaded] = useState(false);
    const shownStartupError = useRef<string | null>(null);

    useEffect(() => {
        AppInfo()
            .then(setInfo)
            .catch(() => setInfo(null))
            .finally(() => setInfoLoaded(true));
    }, []);

    useEffect(() => {
        void i18n.changeLanguage(language);
    }, [language]);

    useEffect(() => {
        const update = () => setSystemLanguage(resolveSystemLanguage());
        window.addEventListener('languagechange', update);
        return () => window.removeEventListener('languagechange', update);
    }, []);

    useEffect(() => {
        const media = window.matchMedia('(prefers-color-scheme: dark)');
        const update = () => setResolvedTheme(resolveTheme(themeMode));
        update();
        media.addEventListener('change', update);
        return () => media.removeEventListener('change', update);
    }, [themeMode]);

    const startupError = info?.startupError ?? '';
    const startupBlocked = startupError !== '';
    const dataServicesEnabled = infoLoaded && !startupBlocked;

    useEffect(() => {
        if (!startupBlocked || shownStartupError.current === startupError) {
            return;
        }
        shownStartupError.current = startupError;
        showStartupError(antApp.modal, t('dataDirectoryUnavailable'), info?.dataDir ?? '', startupError, t);
    }, [antApp.modal, info?.dataDir, startupBlocked, startupError, t]);

    const {
        settings,
        settingsLoading,
        settingsSaving,
        handleSaveSettings,
    } = useAppSettings({
        enabled: dataServicesEnabled,
        messageApi: antApp.message,
        modalApi: antApp.modal,
        t,
        systemLanguage,
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
        enabled: dataServicesEnabled,
        language,
        modalApi: antApp.modal,
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
        modalApi: antApp.modal,
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
        modalApi: antApp.modal,
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
        modalApi: antApp.modal,
        t,
    });

    const {
        decryptDialogOpen,
        decryptLoading,
        decryptResult,
        decryptResultOpen,
        deleteDialogOpen,
        deleteLoading,
        exportDialogOpen,
        exportLoading,
        inspectDialogOpen,
        inspectLoading,
        inspectResult,
        inspectResultOpen,
        projectActionsOpen,
        projectNameSaving,
        verifyDialogOpen,
        verifyLoading,
        verifyResult,
        verifyResultOpen,
        setDeleteDialogOpen,
        setDecryptDialogOpen,
        setDecryptResultOpen,
        setExportDialogOpen,
        setInspectDialogOpen,
        setInspectResultOpen,
        setProjectActionsOpen,
        setVerifyDialogOpen,
        setVerifyResultOpen,
        openProjectActions,
        handleDecryptProject,
        handleDeleteProject,
        handleExportProject,
        handleInspectProject,
        handleSaveProjectName,
        handleVerifyProject,
    } = useProjectActions({
        messageApi: antApp.message,
        modalApi: antApp.modal,
        t,
        selectedProjectId,
        selectedProject,
        reloadProjects: loadProjects,
        clearSelectedProject: () => setSelectedProjectId(null),
    });

    const {
        sharePasswordDialogOpen,
        shareBrowserOpen,
        shareSelectionOpen,
        shareLoading: createShareLoading,
        shareBrowserState,
        selectedShareItemCount,
        createShareResult,
        createShareResultOpen,
        setSharePasswordDialogOpen,
        setShareBrowserOpen,
        setShareSelectionOpen,
        setCreateShareResultOpen,
        handleOpenShareSelection,
        handleConfirmShareSelection,
        handleCreateShare,
    } = useProjectShare({
        messageApi: antApp.message,
        modalApi: antApp.modal,
        t,
        selectedProjectId,
    });

    const handleOpenProjectActionsFromHome = (projectId?: string) => {
        if (projectId) {
            setSelectedProjectId(projectId);
            setProjectActionsOpen(true);
            return;
        }
        openProjectActions();
    };

    const {
        openProjectDialogOpen,
        applyLoading,
        browserLoading,
        browserState,
        browserOpen,
        applyResult,
        applyResultOpen,
        pendingRenames,
        pendingMoves,
        pendingRemoves,
        pendingAdds,
        pendingCreateFolders,
        setOpenProjectDialogOpen,
        setApplyResultOpen,
        handleOpenProjectBrowser,
        addPendingAdd,
        addPendingCreateFolder,
        addPendingRename,
        addPendingMove,
        addPendingRemove,
        discardPendingRename,
        discardPendingMove,
        discardPendingRemove,
        discardPendingAdd,
        discardPendingCreateFolder,
        discardAllPendingChanges,
        handleApplyProjectChanges,
        closeBrowser,
    } = useProjectBrowser({
        messageApi: antApp.message,
        modalApi: antApp.modal,
        t,
        selectedProjectId,
    });

    const columns = useMemo<ColumnsType<LocalProjectRow>>(
        () => [
            {
                title: t('projectId'),
                dataIndex: 'projectId',
                key: 'projectId',
                sorter: (left, right) => left.projectId.localeCompare(right.projectId),
            },
            {
                title: t('projectName'),
                dataIndex: 'projectName',
                key: 'projectName',
                sorter: (left, right) => left.projectName.localeCompare(right.projectName),
            },
            {
                title: t('modifiedTime'),
                dataIndex: 'modifiedTime',
                key: 'modifiedTime',
                defaultSortOrder: 'descend',
                sorter: (left, right) => left.modifiedAtMs - right.modifiedAtMs,
            },
            {
                title: t('availabilityStatus'),
                dataIndex: 'availabilityStatus',
                key: 'availabilityStatus',
                sorter: (left, right) => left.availabilityStatus.localeCompare(right.availabilityStatus),
            },
        ],
        [t],
    );

    const activeOperationLabel = useMemo(() => {
        const operations = [
            {active: createLoading, label: t('createProject')},
            {active: importLoading, label: t('importProject')},
            {active: shareLoading, label: t('loadShare')},
            {active: createShareLoading, label: t('createShare')},
            {active: decryptShareLoading, label: t('decryptShare')},
            {active: verifyShareLoading, label: t('verifyShare')},
            {active: browserLoading, label: t('openProject')},
            {active: applyLoading, label: t('applyChanges')},
            {active: decryptLoading, label: t('decryptProject')},
            {active: verifyLoading, label: t('verifyProject')},
            {active: exportLoading, label: t('exportProject')},
            {active: deleteLoading, label: t('deleteProject')},
        ];
        return operations.find((operation) => operation.active)?.label ?? null;
    }, [
        applyLoading,
        browserLoading,
        createLoading,
        createShareLoading,
        decryptLoading,
        decryptShareLoading,
        deleteLoading,
        exportLoading,
        importLoading,
        shareLoading,
        t,
        verifyLoading,
        verifyShareLoading,
    ]);

    useEffect(() => {
        void SetLongRunningOperationActive(activeOperationLabel !== null);
    }, [activeOperationLabel]);

    useEffect(() => {
        void SetManualContentGuideCloseGuardActive(Boolean(applyResultOpen && applyResult?.manualContentGuide), language);
    }, [applyResult?.manualContentGuide, applyResultOpen, language]);

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
                    activeOperationLabel={activeOperationLabel}
                    resolvedTheme={resolvedTheme}
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
                            disabled={!dataServicesEnabled}
                            operationActive={activeOperationLabel !== null}
                            onCreateProject={() => setCreateDialogOpen(true)}
                            onImportProject={() => setImportDialogOpen(true)}
                            onLoadShare={() => setLoadShareDialogOpen(true)}
                            onProjectSearchChange={setProjectSearch}
                            onRefresh={() => void loadProjects()}
                            onSelectProject={setSelectedProjectId}
                            onOpenProjectActions={handleOpenProjectActionsFromHome}
                            t={t}
                        />
                    )}
                    {navigation === 'settings' && (
                        <SettingsView
                            settings={settings}
                            loading={settingsLoading}
                            saving={settingsSaving}
                            disabled={!dataServicesEnabled}
                            onSave={(values) => void handleSaveSettings(values)}
                            t={t}
                        />
                    )}
                    {navigation === 'about' && <AboutView info={info} t={t}/>}
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
                    dataDirectory={info?.dataDir ?? ''}
                    selectedProject={selectedProject}
                    onCloseProjectActions={() => setProjectActionsOpen(false)}
                    projectNameSaving={projectNameSaving}
                    onSaveProjectName={handleSaveProjectName}
                    onOpenInspect={() => setInspectDialogOpen(true)}
                    onOpenModify={() => setOpenProjectDialogOpen(true)}
                    onOpenVerify={() => setVerifyDialogOpen(true)}
                    onOpenDecrypt={() => setDecryptDialogOpen(true)}
                    onOpenCreateShare={() => setSharePasswordDialogOpen(true)}
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
                    decryptDialogOpen={decryptDialogOpen}
                    decryptLoading={decryptLoading}
                    onCloseDecrypt={() => setDecryptDialogOpen(false)}
                    onDecryptProject={(values) => void handleDecryptProject(values)}
                    decryptResultOpen={decryptResultOpen}
                    decryptResult={decryptResult}
                    onCloseDecryptResult={() => setDecryptResultOpen(false)}
                    createSharePasswordDialogOpen={sharePasswordDialogOpen}
                    createShareBrowserOpen={shareBrowserOpen}
                    createShareDialogOpen={shareSelectionOpen}
                    createShareLoading={createShareLoading}
                    createShareBrowserState={shareBrowserState}
                    selectedShareItemCount={selectedShareItemCount}
                    createShareResultOpen={createShareResultOpen}
                    createShareResult={createShareResult}
                    onCloseCreateSharePassword={() => setSharePasswordDialogOpen(false)}
                    onLoadShareableItems={(password) => void handleOpenShareSelection(password)}
                    onCloseShareSelectionBrowser={() => setShareBrowserOpen(false)}
                    onConfirmShareSelection={handleConfirmShareSelection}
                    onCloseCreateShare={() => setShareSelectionOpen(false)}
                    onCreateShare={(values) => void handleCreateShare(values)}
                    onCloseCreateShareResult={() => setCreateShareResultOpen(false)}
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
                <ProjectBrowserLayer
                    openProjectDialogOpen={openProjectDialogOpen}
                    browserLoading={browserLoading}
                    applyLoading={applyLoading}
                    browserOpen={browserOpen}
                    browserState={browserState}
                    applyResult={applyResult}
                    applyResultOpen={applyResultOpen}
                    pendingRenames={pendingRenames}
                    pendingMoves={pendingMoves}
                    pendingRemoves={pendingRemoves}
                    pendingAdds={pendingAdds}
                    pendingCreateFolders={pendingCreateFolders}
                    onCloseOpenProject={() => setOpenProjectDialogOpen(false)}
                    onOpenProject={(values) => void handleOpenProjectBrowser(values)}
                    onCloseBrowser={closeBrowser}
                    onCloseApplyResult={() => setApplyResultOpen(false)}
                    onAdd={addPendingAdd}
                    onCreateFolder={addPendingCreateFolder}
                    onRename={addPendingRename}
                    onMove={addPendingMove}
                    onRemove={addPendingRemove}
                    onDiscardRename={discardPendingRename}
                    onDiscardMove={discardPendingMove}
                    onDiscardRemove={discardPendingRemove}
                    onDiscardAdd={discardPendingAdd}
                    onDiscardCreateFolder={discardPendingCreateFolder}
                    onDiscardAll={discardAllPendingChanges}
                    onApply={() => void handleApplyProjectChanges()}
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
