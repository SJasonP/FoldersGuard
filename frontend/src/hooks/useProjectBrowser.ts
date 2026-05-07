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

export type PendingMove = {
  itemId: string;
  itemPath: string;
  targetFolderPath: string;
};

export type PendingRemove = {
  itemId: string;
  itemPath: string;
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
  const [pendingMoves, setPendingMoves] = useState<PendingMove[]>([]);
  const [pendingRemoves, setPendingRemoves] = useState<PendingRemove[]>([]);

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
      setPendingMoves([]);
      setPendingRemoves([]);
      messageApi.success(t('openProjectSucceeded'));
    } catch {
      messageApi.error(t('openProjectFailed'));
    } finally {
      setBrowserLoading(false);
    }
  };

  const addPendingRename = (rename: PendingRename) => {
    setPendingRemoves((current) => current.filter((item) => item.itemId !== rename.itemId));
    setPendingRenames((current) => [...current.filter((item) => item.itemId !== rename.itemId), rename]);
  };

  const addPendingMove = (move: PendingMove) => {
    setPendingRemoves((current) => current.filter((item) => item.itemId !== move.itemId));
    setPendingMoves((current) => [...current.filter((item) => item.itemId !== move.itemId), move]);
  };

  const addPendingRemove = (remove: PendingRemove) => {
    setPendingRenames((current) => current.filter((item) => item.itemId !== remove.itemId));
    setPendingMoves((current) => current.filter((item) => item.itemId !== remove.itemId));
    setPendingRemoves((current) => [...current.filter((item) => item.itemId !== remove.itemId), remove]);
  };

  const discardPendingRename = (itemId: string) => {
    setPendingRenames((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardPendingMove = (itemId: string) => {
    setPendingMoves((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardPendingRemove = (itemId: string) => {
    setPendingRemoves((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardAllPendingChanges = () => {
    setPendingRenames([]);
    setPendingMoves([]);
    setPendingRemoves([]);
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
        moveChanges: pendingMoves.map((move) => ({
          itemPath: move.itemPath,
          targetFolderPath: move.targetFolderPath,
        })),
        removeChanges: pendingRemoves.map((remove) => ({
          itemPath: remove.itemPath,
        })),
      }));
      setBrowserState(result.browserState);
      setPendingRenames([]);
      setPendingMoves([]);
      setPendingRemoves([]);
      messageApi.success(t('applyChangesSucceeded'));
      if (result.operationGuidePath) {
        messageApi.info(`${t('operationGuidePath')}: ${result.operationGuidePath}`);
      }
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
    setPendingMoves([]);
    setPendingRemoves([]);
  };

  return {
    openProjectDialogOpen,
    applyLoading,
    browserLoading,
    browserState,
    browserOpen,
    pendingRenames,
    pendingMoves,
    pendingRemoves,
    setOpenProjectDialogOpen,
    setBrowserOpen,
    addPendingRename,
    addPendingMove,
    addPendingRemove,
    discardPendingRename,
    discardPendingMove,
    discardPendingRemove,
    discardAllPendingChanges,
    handleApplyProjectChanges,
    handleOpenProjectBrowser,
    closeBrowser,
  };
}
