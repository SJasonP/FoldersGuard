import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { OpenProjectBrowser } from '../../wailsjs/go/main/App';
import type { ProjectBrowserStateModel } from '../types';

type UseProjectBrowserArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  selectedProjectId: string | null;
};

export function useProjectBrowser({ messageApi, t, selectedProjectId }: UseProjectBrowserArgs) {
  const [openProjectDialogOpen, setOpenProjectDialogOpen] = useState(false);
  const [browserLoading, setBrowserLoading] = useState(false);
  const [browserState, setBrowserState] = useState<ProjectBrowserStateModel | null>(null);
  const [browserOpen, setBrowserOpen] = useState(false);

  const handleOpenProjectBrowser = async (values: { password: string; encryptedPath: string }) => {
    if (!selectedProjectId) {
      return;
    }
    setBrowserLoading(true);
    try {
      const state = await OpenProjectBrowser({
        projectId: selectedProjectId,
        password: values.password,
        encryptedPath: values.encryptedPath,
      });
      setOpenProjectDialogOpen(false);
      setBrowserState(state);
      setBrowserOpen(true);
      messageApi.success(t('openProjectSucceeded'));
    } catch {
      messageApi.error(t('openProjectFailed'));
    } finally {
      setBrowserLoading(false);
    }
  };

  const closeBrowser = () => {
    setBrowserOpen(false);
    setBrowserState(null);
  };

  return {
    openProjectDialogOpen,
    browserLoading,
    browserState,
    browserOpen,
    setOpenProjectDialogOpen,
    setBrowserOpen,
    handleOpenProjectBrowser,
    closeBrowser,
  };
}
