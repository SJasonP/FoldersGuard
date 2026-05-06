import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { ImportProject } from '../../wailsjs/go/main/App';
import type { ImportProjectResultModel } from '../types';

type UseProjectImportArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  reloadProjects: () => Promise<void>;
};

export function useProjectImport({ messageApi, t, reloadProjects }: UseProjectImportArgs) {
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [importLoading, setImportLoading] = useState(false);

  const handleImportProject = async (values: { inputPath: string; password: string; force: boolean }) => {
    setImportLoading(true);
    try {
      const result: ImportProjectResultModel = await ImportProject({
        inputPath: values.inputPath,
        password: values.password,
        force: values.force,
      });
      setImportDialogOpen(false);
      await reloadProjects();
      messageApi.success(`${t('importProjectSucceeded')}: ${result.projectId}`);
    } catch {
      messageApi.error(t('importProjectFailed'));
    } finally {
      setImportLoading(false);
    }
  };

  return {
    importDialogOpen,
    importLoading,
    setImportDialogOpen,
    handleImportProject,
  };
}
