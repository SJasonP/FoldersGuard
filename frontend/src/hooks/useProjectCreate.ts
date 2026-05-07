import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { CreateProject } from '../../wailsjs/go/main/App';
import type { CreateProjectResultModel, SettingsModel } from '../types';
import { formatNumber } from '../formatters';

type CreateProjectValues = {
  sourcePath: string;
  contentOutput: string;
  password: string;
  passwordConfirm: string;
  maxPartSize?: number;
  useDefaultMaxPartSize: boolean;
  force: boolean;
  sourceCleanup: string;
  databaseExport?: string;
};

type UseProjectCreateArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  settings: SettingsModel | null;
  reloadProjects: () => Promise<void>;
};

export function useProjectCreate({ messageApi, t, settings, reloadProjects }: UseProjectCreateArgs) {
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
        databaseExport: values.databaseExport?.trim() ?? '',
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
    } catch {
      messageApi.error(t('createProjectFailed'));
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
