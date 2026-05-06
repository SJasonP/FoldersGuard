import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { DeleteProject, ExportProject, InspectProject } from '../../wailsjs/go/main/App';
import type {
  DeleteProjectResultModel,
  ExportProjectResultModel,
  InspectProjectResultModel,
  LocalProjectSummary,
} from '../types';

type UseProjectActionsArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  selectedProjectId: string | null;
  selectedProject: LocalProjectSummary | null;
  reloadProjects: () => Promise<void>;
  clearSelectedProject: () => void;
};

export function useProjectActions({
  messageApi,
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
  const [exportDialogOpen, setExportDialogOpen] = useState(false);
  const [exportLoading, setExportLoading] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);

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
    } catch {
      messageApi.error(t('inspectProjectFailed'));
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
    } catch {
      messageApi.error(t('exportProjectFailed'));
    } finally {
      setExportLoading(false);
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
    } catch {
      messageApi.error(t('deleteProjectFailed'));
    } finally {
      setDeleteLoading(false);
    }
  };

  return {
    deleteDialogOpen,
    deleteLoading,
    exportDialogOpen,
    exportLoading,
    inspectDialogOpen,
    inspectLoading,
    inspectResult,
    inspectResultOpen,
    projectActionsOpen,
    selectedProject,
    setDeleteDialogOpen,
    setExportDialogOpen,
    setInspectDialogOpen,
    setInspectResultOpen,
    setProjectActionsOpen,
    openProjectActions,
    handleDeleteProject,
    handleExportProject,
    handleInspectProject,
  };
}
