import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { CreateProject } from '../../wailsjs/go/main/App';
import type { CreateProjectResultModel, SettingsModel } from '../types';
import { formatNumber } from '../formatters';
import { showOperationError } from '../components/common/operationError';

type CreateProjectValues = {
  sourcePath: string;
  contentOutput: string;
  password: string;
  passwordConfirm: string;
  maxPartSize?: number;
  useDefaultMaxPartSize: boolean;
  force: boolean;
  sourceCleanup: string;
};

type UseProjectCreateArgs = {
  messageApi: MessageInstance;
  modalApi: ModalHookAPI;
  t: (key: string) => string;
  settings: SettingsModel | null;
  reloadProjects: () => Promise<void>;
};

export function useProjectCreate({ messageApi, modalApi, t, settings, reloadProjects }: UseProjectCreateArgs) {
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);

  const handleCreateProject = async (values: CreateProjectValues) => {
    setCreateLoading(true);
    try {
      const result: CreateProjectResultModel = await CreateProject({
        sourcePath: values.sourcePath,
        contentOutput: values.contentOutput,
        password: values.password,
        maxPartSize: values.useDefaultMaxPartSize ? 0 : Math.trunc(values.maxPartSize ?? 0),
        force: values.force,
        sourceCleanup: values.sourceCleanup,
        databaseExport: '',
      });
      setCreateDialogOpen(false);
      await reloadProjects();
      messageApi.success(
        [
          t('createProjectSucceeded'),
          `${t('createSummaryProjectId')}: ${result.projectId}`,
          `${t('createSummaryProjectName')}: ${result.projectName}`,
          `${t('createSummaryEncryptedFiles')}: ${formatNumber(result.encryptedFiles)}`,
          `${t('createSummaryEncryptedFolders')}: ${formatNumber(result.encryptedFolders)}`,
          `${t('createSummaryEncryptedParts')}: ${formatNumber(result.encryptedParts)}`,
          `${t('createSummaryDeletedCleartextFiles')}: ${formatNumber(result.deletedCleartextFiles)}`,
        ].join(' | '),
      );
    } catch (error) {
      showOperationError(modalApi, t('createProjectFailed'), error, t);
    } finally {
      setCreateLoading(false);
    }
  };

  const defaultSourceCleanup =
    settings?.sourceCleanupMode && settings.sourceCleanupMode !== 'ask' ? settings.sourceCleanupMode : 'keep';

  return {
    createDialogOpen,
    createLoading,
    defaultSourceCleanup,
    setCreateDialogOpen,
    handleCreateProject,
  };
}
