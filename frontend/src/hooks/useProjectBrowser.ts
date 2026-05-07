import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { ApplyProjectChanges, OpenProjectBrowser } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import type { ApplyProjectChangesResultModel, ProjectBrowserStateModel } from '../types';
import { showOperationError } from '../components/common/operationError';

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

export type PendingAdd = {
  itemId: string;
  sourcePath: string;
  targetFolderPath: string;
  maxPartSize: number;
};

export type PendingCreateFolder = {
  itemId: string;
  targetFolderPath: string;
  name: string;
};

type UseProjectBrowserArgs = {
  messageApi: MessageInstance;
  modalApi: ModalHookAPI;
  t: (key: string, values?: Record<string, string | number>) => string;
  selectedProjectId: string | null;
};

export function useProjectBrowser({ messageApi, modalApi, t, selectedProjectId }: UseProjectBrowserArgs) {
  const [openProjectDialogOpen, setOpenProjectDialogOpen] = useState(false);
  const [browserLoading, setBrowserLoading] = useState(false);
  const [applyLoading, setApplyLoading] = useState(false);
  const [browserState, setBrowserState] = useState<ProjectBrowserStateModel | null>(null);
  const [browserOpen, setBrowserOpen] = useState(false);
  const [applyResult, setApplyResult] = useState<ApplyProjectChangesResultModel | null>(null);
  const [applyResultOpen, setApplyResultOpen] = useState(false);
  const [browserPassword, setBrowserPassword] = useState('');
  const [browserEncryptedPath, setBrowserEncryptedPath] = useState('');
  const [pendingRenames, setPendingRenames] = useState<PendingRename[]>([]);
  const [pendingMoves, setPendingMoves] = useState<PendingMove[]>([]);
  const [pendingRemoves, setPendingRemoves] = useState<PendingRemove[]>([]);
  const [pendingAdds, setPendingAdds] = useState<PendingAdd[]>([]);
  const [pendingCreateFolders, setPendingCreateFolders] = useState<PendingCreateFolder[]>([]);

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
      setApplyResult(null);
      setApplyResultOpen(false);
      setPendingRenames([]);
      setPendingMoves([]);
      setPendingRemoves([]);
      setPendingAdds([]);
      setPendingCreateFolders([]);
      messageApi.success(t('openProjectSucceeded'));
    } catch (error) {
      showOperationError(modalApi, t('openProjectFailed'), error, t);
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

  const addPendingAdd = (add: PendingAdd) => {
    setPendingAdds((current) => [...current.filter((item) => item.itemId !== add.itemId), add]);
  };

  const addPendingCreateFolder = (createFolder: PendingCreateFolder) => {
    setPendingCreateFolders((current) => [...current.filter((item) => item.itemId !== createFolder.itemId), createFolder]);
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

  const discardPendingAdd = (itemId: string) => {
    setPendingAdds((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardPendingCreateFolder = (itemId: string) => {
    setPendingCreateFolders((current) => current.filter((item) => item.itemId !== itemId));
  };

  const discardAllPendingChanges = () => {
    setPendingRenames([]);
    setPendingMoves([]);
    setPendingRemoves([]);
    setPendingAdds([]);
    setPendingCreateFolders([]);
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
        addChanges: pendingAdds.map((add) => ({
          sourcePath: add.sourcePath,
          targetFolderPath: add.targetFolderPath,
          maxPartSize: add.maxPartSize,
        })),
        createFolderChanges: pendingCreateFolders.map((createFolder) => ({
          targetFolderPath: createFolder.targetFolderPath,
          name: createFolder.name,
        })),
      }));
      setBrowserState(result.browserState);
      setApplyResult(result);
      setApplyResultOpen(true);
      setPendingRenames([]);
      setPendingMoves([]);
      setPendingRemoves([]);
      setPendingAdds([]);
      setPendingCreateFolders([]);
      messageApi.success(t('applyChangesSucceeded'));
      if (result.operationGuidePath) {
        messageApi.info(`${t('operationGuidePath')}: ${result.operationGuidePath}`);
      }
    } catch (error) {
      showOperationError(modalApi, t('applyChangesFailed'), error, t);
    } finally {
      setApplyLoading(false);
    }
  };

  const closeBrowser = () => {
    setBrowserOpen(false);
    setBrowserState(null);
    setApplyResult(null);
    setApplyResultOpen(false);
    setBrowserPassword('');
    setBrowserEncryptedPath('');
    setPendingRenames([]);
    setPendingMoves([]);
    setPendingRemoves([]);
    setPendingAdds([]);
    setPendingCreateFolders([]);
  };

  return {
    openProjectDialogOpen,
    applyLoading,
    browserLoading,
    browserState,
    browserOpen,
    applyResult,
    applyResultOpen,
    pendingRenames,
    pendingMoves,
    pendingRemoves,
    pendingAdds,
    pendingCreateFolders,
    setOpenProjectDialogOpen,
    setBrowserOpen,
    setApplyResultOpen,
    addPendingRename,
    addPendingMove,
    addPendingRemove,
    addPendingAdd,
    addPendingCreateFolder,
    discardPendingRename,
    discardPendingMove,
    discardPendingRemove,
    discardPendingAdd,
    discardPendingCreateFolder,
    discardAllPendingChanges,
    handleApplyProjectChanges,
    handleOpenProjectBrowser,
    closeBrowser,
  };
}
