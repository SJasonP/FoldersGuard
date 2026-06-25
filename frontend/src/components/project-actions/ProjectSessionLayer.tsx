import {CreateProjectModal} from './CreateProjectModal';
import {CreateShareModal} from './CreateShareModal';
import {CreateSharePasswordModal} from './CreateSharePasswordModal';
import {CreateShareResultDrawer} from './CreateShareResultDrawer';
import {DecryptProjectDrawer} from './DecryptProjectDrawer';
import {DecryptProjectModal} from './DecryptProjectModal';
import {DeleteProjectModal} from './DeleteProjectModal';
import {ExportProjectModal} from './ExportProjectModal';
import {ImportProjectModal} from './ImportProjectModal';
import {InspectProjectDrawer} from './InspectProjectDrawer';
import {InspectProjectModal} from './InspectProjectModal';
import {ChangePasswordModal} from '../common/ChangePasswordModal';
import {ProjectActionsDrawer} from './ProjectActionsDrawer';
import {RestoreBackupModal} from './RestoreBackupModal';
import {ShareSelectionDrawer} from './ShareSelectionDrawer';
import {VerifyProjectDrawer} from './VerifyProjectDrawer';
import {VerifyProjectModal} from './VerifyProjectModal';
import type {
    CreateShareResultModel,
    DecryptProjectResultModel,
    InspectProjectResultModel,
    LocalProjectSummary,
    ProjectBackupInfoModel,
    ProjectBrowserStateModel,
    SettingsModel,
    VerifyProjectResultModel,
} from '../../types';

type ProjectSessionLayerProps = {
    createDialogOpen: boolean;
    createLoading: boolean;
    settings: SettingsModel | null;
    onCloseCreate: () => void;
    onCreateProject: (values: {
        sourcePath: string;
        contentOutput: string;
        password: string;
        passwordConfirm: string;
        force: boolean;
    }) => void;
    importDialogOpen: boolean;
    importLoading: boolean;
    onCloseImport: () => void;
    onImportProject: (values: { inputPath: string; password: string; force: boolean }) => void;
    projectActionsOpen: boolean;
    dataDirectory: string;
    selectedProject: LocalProjectSummary | null;
    onCloseProjectActions: () => void;
    projectNameSaving: boolean;
    onSaveProjectName: (projectName: string) => void;
    onOpenInspect: () => void;
    onOpenModify: () => void;
    onOpenVerify: () => void;
    onOpenDecrypt: () => void;
    onOpenCreateShare: () => void;
    onOpenExport: () => void;
    onOpenChangePassword: () => void;
    onOpenRestoreBackup: () => void;
    onOpenDelete: () => void;
    inspectDialogOpen: boolean;
    inspectLoading: boolean;
    onCloseInspect: () => void;
    onInspectProject: (password: string) => void;
    inspectResultOpen: boolean;
    inspectResult: InspectProjectResultModel | null;
    onCloseInspectResult: () => void;
    verifyDialogOpen: boolean;
    verifyLoading: boolean;
    onCloseVerify: () => void;
    onVerifyProject: (values: { password: string; encryptedPath: string }) => void;
    verifyResultOpen: boolean;
    verifyResult: VerifyProjectResultModel | null;
    onCloseVerifyResult: () => void;
    decryptDialogOpen: boolean;
    decryptLoading: boolean;
    onCloseDecrypt: () => void;
    onDecryptProject: (values: {
        password: string;
        encryptedPath: string;
        outputPath: string;
        force: boolean;
    }) => void;
    decryptResultOpen: boolean;
    decryptResult: DecryptProjectResultModel | null;
    onCloseDecryptResult: () => void;
    createSharePasswordDialogOpen: boolean;
    createShareBrowserOpen: boolean;
    createShareDialogOpen: boolean;
    createShareLoading: boolean;
    createShareBrowserState: ProjectBrowserStateModel | null;
    selectedShareItemCount: number;
    createShareResultOpen: boolean;
    createShareResult: CreateShareResultModel | null;
    onCloseCreateSharePassword: () => void;
    onLoadShareableItems: (password: string) => void;
    onCloseShareSelectionBrowser: () => void;
    onConfirmShareSelection: (itemPaths: string[]) => void;
    onCloseCreateShare: () => void;
    onCreateShare: (values: {
        outputPath: string;
        force: boolean;
        passwordProtected: boolean;
        sharePassword?: string;
        sharePasswordConfirm?: string;
    }) => void;
    onCloseCreateShareResult: () => void;
    exportDialogOpen: boolean;
    exportLoading: boolean;
    onCloseExport: () => void;
    onExportProject: (values: { password: string; outputPath: string; force: boolean }) => void;
    deleteDialogOpen: boolean;
    deleteLoading: boolean;
    onCloseDelete: () => void;
    onDeleteProject: (password: string) => void;
    restoreBackupOpen: boolean;
    backups: ProjectBackupInfoModel[];
    backupsLoading: boolean;
    restoreBackupLoading: boolean;
    onCloseRestoreBackup: () => void;
    onRestoreBackup: (backupId: string) => void;
    changePasswordOpen: boolean;
    changePasswordLoading: boolean;
    onCloseChangePassword: () => void;
    onChangePassword: (values: { oldPassword: string; newPassword: string }) => void;
    t: (key: string, options?: Record<string, unknown>) => string;
};

export function ProjectSessionLayer({
                                        createDialogOpen,
                                        createLoading,
                                        settings,
                                        onCloseCreate,
                                        onCreateProject,
                                        importDialogOpen,
                                        importLoading,
                                        onCloseImport,
                                        onImportProject,
                                        projectActionsOpen,
                                        dataDirectory,
                                        selectedProject,
                                        onCloseProjectActions,
                                        projectNameSaving,
                                        onSaveProjectName,
                                        onOpenInspect,
                                        onOpenModify,
                                        onOpenVerify,
                                        onOpenDecrypt,
                                        onOpenCreateShare,
                                        onOpenExport,
                                        onOpenChangePassword,
                                        onOpenRestoreBackup,
                                        onOpenDelete,
                                        inspectDialogOpen,
                                        inspectLoading,
                                        onCloseInspect,
                                        onInspectProject,
                                        inspectResultOpen,
                                        inspectResult,
                                        onCloseInspectResult,
                                        verifyDialogOpen,
                                        verifyLoading,
                                        onCloseVerify,
                                        onVerifyProject,
                                        verifyResultOpen,
                                        verifyResult,
                                        onCloseVerifyResult,
                                        decryptDialogOpen,
                                        decryptLoading,
                                        onCloseDecrypt,
                                        onDecryptProject,
                                        decryptResultOpen,
                                        decryptResult,
                                        onCloseDecryptResult,
                                        createSharePasswordDialogOpen,
                                        createShareBrowserOpen,
                                        createShareDialogOpen,
                                        createShareLoading,
                                        createShareBrowserState,
                                        selectedShareItemCount,
                                        createShareResultOpen,
                                        createShareResult,
                                        onCloseCreateSharePassword,
                                        onLoadShareableItems,
                                        onCloseShareSelectionBrowser,
                                        onConfirmShareSelection,
                                        onCloseCreateShare,
                                        onCreateShare,
                                        onCloseCreateShareResult,
                                        exportDialogOpen,
                                        exportLoading,
                                        onCloseExport,
                                        onExportProject,
                                        deleteDialogOpen,
                                        deleteLoading,
                                        onCloseDelete,
                                        onDeleteProject,
                                        restoreBackupOpen,
                                        backups,
                                        backupsLoading,
                                        restoreBackupLoading,
                                        onCloseRestoreBackup,
                                        onRestoreBackup,
                                        changePasswordOpen,
                                        changePasswordLoading,
                                        onCloseChangePassword,
                                        onChangePassword,
                                        t,
                                    }: ProjectSessionLayerProps) {
    return (
        <>
            <ProjectActionsDrawer
                open={projectActionsOpen}
                project={selectedProject}
                projectNameSaving={projectNameSaving}
                onClose={onCloseProjectActions}
                onSaveProjectName={onSaveProjectName}
                onInspect={onOpenInspect}
                onModify={onOpenModify}
                onVerify={onOpenVerify}
                onDecrypt={onOpenDecrypt}
                onCreateShare={onOpenCreateShare}
                onExport={onOpenExport}
                onChangePassword={onOpenChangePassword}
                onRestoreBackup={onOpenRestoreBackup}
                onDelete={onOpenDelete}
                t={t}
            />
            <CreateProjectModal
                open={createDialogOpen}
                loading={createLoading}
                settings={settings}
                onCancel={onCloseCreate}
                onSubmit={onCreateProject}
                t={t}
            />
            <ImportProjectModal open={importDialogOpen} loading={importLoading} onCancel={onCloseImport}
                                onSubmit={onImportProject} t={t}/>
            <InspectProjectModal open={inspectDialogOpen} loading={inspectLoading} onCancel={onCloseInspect}
                                 onSubmit={onInspectProject} t={t}/>
            <InspectProjectDrawer open={inspectResultOpen} result={inspectResult} onClose={onCloseInspectResult} t={t}/>
            <VerifyProjectModal open={verifyDialogOpen} loading={verifyLoading} onCancel={onCloseVerify}
                                onSubmit={onVerifyProject} t={t}/>
            <VerifyProjectDrawer open={verifyResultOpen} result={verifyResult} onClose={onCloseVerifyResult} t={t}/>
            <DecryptProjectModal
                open={decryptDialogOpen}
                loading={decryptLoading}
                sourceCleanupMode={settings?.sourceCleanupMode ?? 'delete'}
                onCancel={onCloseDecrypt}
                onSubmit={onDecryptProject}
                t={t}
            />
            <DecryptProjectDrawer open={decryptResultOpen} result={decryptResult} onClose={onCloseDecryptResult} t={t}/>
            <CreateSharePasswordModal
                open={createSharePasswordDialogOpen}
                loading={createShareLoading}
                onCancel={onCloseCreateSharePassword}
                onSubmit={onLoadShareableItems}
                t={t}
            />
            <ShareSelectionDrawer
                open={createShareBrowserOpen}
                loading={createShareLoading}
                state={createShareBrowserState}
                onCancel={onCloseShareSelectionBrowser}
                onContinue={onConfirmShareSelection}
                t={t}
            />
            <CreateShareModal
                open={createShareDialogOpen}
                loading={createShareLoading}
                selectedItemCount={selectedShareItemCount}
                onCancel={onCloseCreateShare}
                onSubmit={onCreateShare}
                t={t}
            />
            <CreateShareResultDrawer open={createShareResultOpen} result={createShareResult}
                                     onClose={onCloseCreateShareResult} t={t}/>
            <ExportProjectModal open={exportDialogOpen} loading={exportLoading} onCancel={onCloseExport}
                                onSubmit={onExportProject} t={t}/>
            <DeleteProjectModal
                open={deleteDialogOpen}
                loading={deleteLoading}
                dataDirectory={dataDirectory}
                project={selectedProject}
                onCancel={onCloseDelete}
                onSubmit={onDeleteProject}
                t={t}
            />
            <RestoreBackupModal
                open={restoreBackupOpen}
                loading={backupsLoading}
                restoreLoading={restoreBackupLoading}
                backups={backups}
                onRestore={onRestoreBackup}
                onCancel={onCloseRestoreBackup}
                t={t}
            />
            <ChangePasswordModal
                open={changePasswordOpen}
                loading={changePasswordLoading}
                title={t('changePassword')}
                onCancel={onCloseChangePassword}
                onSubmit={onChangePassword}
                t={t}
            />
        </>
    );
}
