import {useState} from 'react';
import type {MessageInstance} from 'antd/es/message/interface';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';
import {
    ChangeProjectPassword,
    DecryptProject,
    DeleteProject,
    ExportProject,
    InspectProject,
    ListProjectBackups,
    RestoreProjectBackup,
    SaveLocalProjectName,
    VerifyProject
} from '../../wailsjs/go/main/App';
import type {
    DecryptProjectResultModel,
    DeleteProjectResultModel,
    ExportProjectResultModel,
    InspectProjectResultModel,
    LocalProjectSummary,
    ProjectBackupInfoModel,
    VerifyProjectResultModel,
} from '../types';
import {showOperationError} from '../components/common/operationError';

type UseProjectActionsArgs = {
    messageApi: MessageInstance;
    modalApi: ModalHookAPI;
    t: (key: string) => string;
    selectedProjectId: string | null;
    selectedProject: LocalProjectSummary | null;
    reloadProjects: () => Promise<void>;
    clearSelectedProject: () => void;
};

export function useProjectActions({
                                      messageApi,
                                      modalApi,
                                      t,
                                      selectedProjectId,
                                      selectedProject,
                                      reloadProjects,
                                      clearSelectedProject,
                                  }: UseProjectActionsArgs) {
    const [projectActionsOpen, setProjectActionsOpen] = useState(false);
    const [inspectDialogOpen, setInspectDialogOpen] = useState(false);
    const [inspectLoading, setInspectLoading] = useState(false);
    const [inspectResult, setInspectResult] = useState<InspectProjectResultModel | null>(null);
    const [inspectResultOpen, setInspectResultOpen] = useState(false);
    const [verifyDialogOpen, setVerifyDialogOpen] = useState(false);
    const [verifyLoading, setVerifyLoading] = useState(false);
    const [verifyResult, setVerifyResult] = useState<VerifyProjectResultModel | null>(null);
    const [verifyResultOpen, setVerifyResultOpen] = useState(false);
    const [decryptDialogOpen, setDecryptDialogOpen] = useState(false);
    const [decryptLoading, setDecryptLoading] = useState(false);
    const [decryptResult, setDecryptResult] = useState<DecryptProjectResultModel | null>(null);
    const [decryptResultOpen, setDecryptResultOpen] = useState(false);
    const [exportDialogOpen, setExportDialogOpen] = useState(false);
    const [exportLoading, setExportLoading] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deleteLoading, setDeleteLoading] = useState(false);
    const [restoreBackupOpen, setRestoreBackupOpen] = useState(false);
    const [backups, setBackups] = useState<ProjectBackupInfoModel[]>([]);
    const [backupsLoading, setBackupsLoading] = useState(false);
    const [restoreBackupLoading, setRestoreBackupLoading] = useState(false);
    const [changePasswordOpen, setChangePasswordOpen] = useState(false);
    const [changePasswordLoading, setChangePasswordLoading] = useState(false);
    const [projectNameSaving, setProjectNameSaving] = useState(false);

    const openProjectActions = () => {
        if (!selectedProjectId) {
            return;
        }
        setProjectActionsOpen(true);
    };

    const handleInspectProject = async (password: string) => {
        if (!selectedProjectId) {
            return;
        }
        setInspectLoading(true);
        try {
            const result = await InspectProject({
                projectId: selectedProjectId,
                password,
            });
            setInspectDialogOpen(false);
            setProjectActionsOpen(false);
            setInspectResult(result);
            setInspectResultOpen(true);
        } catch (error) {
            showOperationError(modalApi, t('inspectProjectFailed'), error, t);
        } finally {
            setInspectLoading(false);
        }
    };

    const handleExportProject = async (values: { password: string; outputPath: string; force: boolean }) => {
        if (!selectedProjectId) {
            return;
        }
        setExportDialogOpen(false);
        setProjectActionsOpen(false);
        setExportLoading(true);
        try {
            const result: ExportProjectResultModel = await ExportProject({
                projectId: selectedProjectId,
                password: values.password,
                outputPath: values.outputPath,
                force: values.force,
            });
            messageApi.success(`${t('exportProjectSucceeded')}: ${result.outputPath}`);
        } catch (error) {
            showOperationError(modalApi, t('exportProjectFailed'), error, t);
        } finally {
            setExportLoading(false);
        }
    };

    const handleVerifyProject = async (values: { password: string; encryptedPath: string }) => {
        if (!selectedProjectId) {
            return;
        }
        setVerifyDialogOpen(false);
        setProjectActionsOpen(false);
        setVerifyLoading(true);
        try {
            const result: VerifyProjectResultModel = await VerifyProject({
                projectId: selectedProjectId,
                password: values.password,
                encryptedPath: values.encryptedPath,
            });
            setVerifyResult(result);
            setVerifyResultOpen(true);
            messageApi.success(t('verifyProjectSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('verifyProjectFailed'), error, t);
        } finally {
            setVerifyLoading(false);
        }
    };

    const handleDecryptProject = async (values: {
        password: string;
        encryptedPath: string;
        outputPath: string;
        force: boolean;
        resume: boolean;
        continueOnError: boolean;
    }) => {
        if (!selectedProjectId) {
            return;
        }
        setDecryptDialogOpen(false);
        setProjectActionsOpen(false);
        setDecryptLoading(true);
        try {
            const result: DecryptProjectResultModel = await DecryptProject({
                projectId: selectedProjectId,
                password: values.password,
                encryptedPath: values.encryptedPath,
                outputPath: values.outputPath,
                force: values.force,
                sourceCleanup: '',
                resume: values.resume,
                failureHandling: values.continueOnError ? 'continue' : 'abort',
            });
            setDecryptResult(result);
            setDecryptResultOpen(true);
            messageApi.success(t('decryptProjectSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('decryptProjectFailed'), error, t);
        } finally {
            setDecryptLoading(false);
        }
    };

    const handleDeleteProject = async (password: string) => {
        if (!selectedProjectId) {
            return;
        }
        setDeleteDialogOpen(false);
        setProjectActionsOpen(false);
        setDeleteLoading(true);
        try {
            const result: DeleteProjectResultModel = await DeleteProject({
                projectId: selectedProjectId,
                password,
            });
            setInspectResultOpen(false);
            setInspectResult(null);
            clearSelectedProject();
            await reloadProjects();
            messageApi.success(`${t('deleteProjectSucceeded')}: ${result.projectId}`);
        } catch (error) {
            showOperationError(modalApi, t('deleteProjectFailed'), error, t);
        } finally {
            setDeleteLoading(false);
        }
    };

    const handleOpenRestoreBackup = async () => {
        if (!selectedProjectId) {
            return;
        }
        setBackups([]);
        setRestoreBackupOpen(true);
        setBackupsLoading(true);
        try {
            const result = await ListProjectBackups(selectedProjectId);
            setBackups(result ?? []);
        } catch (error) {
            showOperationError(modalApi, t('restoreBackupFailed'), error, t);
        } finally {
            setBackupsLoading(false);
        }
    };

    const handleRestoreBackup = async (backupId: string) => {
        if (!selectedProjectId) {
            return;
        }
        setRestoreBackupLoading(true);
        try {
            await RestoreProjectBackup({
                projectId: selectedProjectId,
                backupId,
                force: true,
            });
            setRestoreBackupOpen(false);
            setProjectActionsOpen(false);
            await reloadProjects();
            messageApi.success(t('restoreBackupSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('restoreBackupFailed'), error, t);
        } finally {
            setRestoreBackupLoading(false);
        }
    };

    const handleChangePassword = async (values: { oldPassword: string; newPassword: string }) => {
        if (!selectedProjectId) {
            return;
        }
        setChangePasswordOpen(false);
        setProjectActionsOpen(false);
        setChangePasswordLoading(true);
        try {
            await ChangeProjectPassword({
                projectId: selectedProjectId,
                oldPassword: values.oldPassword,
                newPassword: values.newPassword,
            });
            messageApi.success(t('changePasswordSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('changePasswordFailed'), error, t);
        } finally {
            setChangePasswordLoading(false);
        }
    };

    const handleSaveProjectName = async (projectName: string) => {
        if (!selectedProjectId) {
            return;
        }
        setProjectNameSaving(true);
        try {
            await SaveLocalProjectName({
                projectId: selectedProjectId,
                projectName,
            });
            await reloadProjects();
            messageApi.success(t('projectNameSaved'));
        } catch (error) {
            showOperationError(modalApi, t('projectNameSaveFailed'), error, t);
        } finally {
            setProjectNameSaving(false);
        }
    };

    return {
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
        restoreBackupOpen,
        backups,
        backupsLoading,
        restoreBackupLoading,
        changePasswordOpen,
        changePasswordLoading,
        selectedProject,
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
        setRestoreBackupOpen,
        setChangePasswordOpen,
        setVerifyDialogOpen,
        setVerifyResultOpen,
        openProjectActions,
        handleDecryptProject,
        handleDeleteProject,
        handleExportProject,
        handleInspectProject,
        handleOpenRestoreBackup,
        handleRestoreBackup,
        handleChangePassword,
        handleSaveProjectName,
        handleVerifyProject,
    };
}
