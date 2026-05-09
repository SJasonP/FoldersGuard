import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { DecryptProject, DeleteProject, ExportProject, InspectProject, SaveLocalProjectName, VerifyProject } from '../../wailsjs/go/main/App';
import type {
  DecryptProjectResultModel,
  DeleteProjectResultModel,
  ExportProjectResultModel,
  InspectProjectResultModel,
  LocalProjectSummary,
  VerifyProjectResultModel,
} from '../types';
import { showOperationError } from '../components/common/operationError';

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
    setExportLoading(true);
    try {
      const result: ExportProjectResultModel = await ExportProject({
        projectId: selectedProjectId,
        password: values.password,
        outputPath: values.outputPath,
        force: values.force,
      });
      setExportDialogOpen(false);
      setProjectActionsOpen(false);
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
    setVerifyLoading(true);
    try {
      const result: VerifyProjectResultModel = await VerifyProject({
        projectId: selectedProjectId,
        password: values.password,
        encryptedPath: values.encryptedPath,
      });
      setVerifyDialogOpen(false);
      setProjectActionsOpen(false);
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
    sourceCleanup: string;
  }) => {
    if (!selectedProjectId) {
      return;
    }
    setDecryptLoading(true);
    try {
      const result: DecryptProjectResultModel = await DecryptProject({
        projectId: selectedProjectId,
        password: values.password,
        encryptedPath: values.encryptedPath,
        outputPath: values.outputPath,
        force: values.force,
        sourceCleanup: values.sourceCleanup,
      });
      setDecryptDialogOpen(false);
      setProjectActionsOpen(false);
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
    setDeleteLoading(true);
    try {
      const result: DeleteProjectResultModel = await DeleteProject({
        projectId: selectedProjectId,
        password,
      });
      setDeleteDialogOpen(false);
      setProjectActionsOpen(false);
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
    setVerifyDialogOpen,
    setVerifyResultOpen,
    openProjectActions,
    handleDecryptProject,
    handleDeleteProject,
    handleExportProject,
    handleInspectProject,
    handleSaveProjectName,
    handleVerifyProject,
  };
}
