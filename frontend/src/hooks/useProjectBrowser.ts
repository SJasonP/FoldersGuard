import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { ApplyProjectChanges, OpenProjectBrowser } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import type { ProjectBrowserStateModel } from '../types';

export type PendingRename = {
  itemId: string;
  itemPath: string;
  oldName: string;
  newName: string;
};

type UseProjectBrowserArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  selectedProjectId: string | null;
};

export function useProjectBrowser({ messageApi, t, selectedProjectId }: UseProjectBrowserArgs) {
  const [openProjectDialogOpen, setOpenProjectDialogOpen] = useState(false);
  const [browserLoading, setBrowserLoading] = useState(false);
  const [applyLoading, setApplyLoading] = useState(false);
  const [browserState, setBrowserState] = useState<ProjectBrowserStateModel | null>(null);
  const [browserOpen, setBrowserOpen] = useState(false);
  const [browserPassword, setBrowserPassword] = useState('');
  const [browserEncryptedPath, setBrowserEncryptedPath] = useState('');
  const [pendingRenames, setPendingRenames] = useState<PendingRename[]>([]);

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
      setBrowserPassword(values.password);
      setBrowserEncryptedPath(values.encryptedPath);
      setPendingRenames([]);
      messageApi.success(t('openProjectSucceeded'));
    } catch {
      messageApi.error(t('openProjectFailed'));
    } finally {
      setBrowserLoading(false);
    }
  };

  const addPendingRename = (rename: PendingRename) => {
    setPendingRenames((current) => [...current.filter((item) => item.itemId !== rename.itemId), rename]);
  };

  const discardPendingRename = (itemId: string) => {
    setPendingRenames((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardAllPendingChanges = () => {
    setPendingRenames([]);
  };

  const handleApplyProjectChanges = async () => {
    if (!browserState) {
      return;
    }
    setApplyLoading(true);
    try {
      const result = await ApplyProjectChanges(new main.ApplyProjectChangesRequest({
        projectId: browserState.projectId,
        password: browserPassword,
        encryptedPath: browserEncryptedPath,
        renameChanges: pendingRenames.map((rename) => ({
          itemPath: rename.itemPath,
          newName: rename.newName,
        })),
      }));
      setBrowserState(result.browserState);
      setPendingRenames([]);
      messageApi.success(t('applyChangesSucceeded'));
    } catch {
      messageApi.error(t('applyChangesFailed'));
    } finally {
      setApplyLoading(false);
    }
  };

  const closeBrowser = () => {
    setBrowserOpen(false);
    setBrowserState(null);
    setBrowserPassword('');
    setBrowserEncryptedPath('');
    setPendingRenames([]);
  };

  return {
    openProjectDialogOpen,
    applyLoading,
    browserLoading,
    browserState,
    browserOpen,
    pendingRenames,
    setOpenProjectDialogOpen,
    setBrowserOpen,
    addPendingRename,
    discardPendingRename,
    discardAllPendingChanges,
    handleApplyProjectChanges,
    handleOpenProjectBrowser,
    closeBrowser,
  };
}
