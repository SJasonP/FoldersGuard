import { useMemo, useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { CreateShare, OpenProjectBrowser } from '../../wailsjs/go/main/App';
import type { CreateShareResultModel, ProjectBrowserStateModel } from '../types';
import { showOperationError } from '../components/common/operationError';

type UseProjectShareArgs = {
  messageApi: MessageInstance;
  modalApi: ModalHookAPI;
  t: (key: string) => string;
  selectedProjectId: string | null;
};

export function useProjectShare({ messageApi, modalApi, t, selectedProjectId }: UseProjectShareArgs) {
  const [sharePasswordDialogOpen, setSharePasswordDialogOpen] = useState(false);
  const [shareBrowserOpen, setShareBrowserOpen] = useState(false);
  const [shareSelectionOpen, setShareSelectionOpen] = useState(false);
  const [shareLoading, setShareLoading] = useState(false);
  const [shareBrowserState, setShareBrowserState] = useState<ProjectBrowserStateModel | null>(null);
  const [selectedShareItemPaths, setSelectedShareItemPaths] = useState<string[]>([]);
  const [projectPassword, setProjectPassword] = useState('');
  const [createShareResult, setCreateShareResult] = useState<CreateShareResultModel | null>(null);
  const [createShareResultOpen, setCreateShareResultOpen] = useState(false);

  const selectedShareItemCount = useMemo(() => selectedShareItemPaths.length, [selectedShareItemPaths]);

  const handleOpenShareSelection = async (password: string) => {
    if (!selectedProjectId) {
      return;
    }
    setShareLoading(true);
    try {
      const state = await OpenProjectBrowser({
        projectId: selectedProjectId,
        password,
        encryptedPath: '',
      });
      setProjectPassword(password);
      setShareBrowserState(state);
      setSelectedShareItemPaths([]);
      setSharePasswordDialogOpen(false);
      setShareBrowserOpen(true);
    } catch (error) {
      showOperationError(modalApi, t('loadShareableItemsFailed'), error, t);
    } finally {
      setShareLoading(false);
    }
  };

  const handleCreateShare = async (values: {
    outputPath: string;
    force: boolean;
    passwordProtected: boolean;
    sharePassword?: string;
    sharePasswordConfirm?: string;
  }) => {
    if (!selectedProjectId) {
      return;
    }
    setShareLoading(true);
    try {
      const result = await CreateShare({
        projectId: selectedProjectId,
        projectPassword,
        itemPaths: selectedShareItemPaths,
        outputPath: values.outputPath,
        force: values.force,
        passwordProtected: values.passwordProtected,
        sharePassword: values.passwordProtected ? values.sharePassword ?? '' : '',
      });
      setShareSelectionOpen(false);
      setCreateShareResult(result);
      setCreateShareResultOpen(true);
      messageApi.success(t('createShareSucceeded'));
    } catch (error) {
      showOperationError(modalApi, t('createShareFailed'), error, t);
    } finally {
      setShareLoading(false);
    }
  };

  const handleConfirmShareSelection = (itemPaths: string[]) => {
    setSelectedShareItemPaths(itemPaths);
    setShareBrowserOpen(false);
    setShareSelectionOpen(true);
  };

  const closeShareFlow = () => {
    setSharePasswordDialogOpen(false);
    setShareBrowserOpen(false);
    setShareSelectionOpen(false);
    setShareBrowserState(null);
    setSelectedShareItemPaths([]);
    setProjectPassword('');
  };

  return {
    sharePasswordDialogOpen,
    shareBrowserOpen,
    shareSelectionOpen,
    shareLoading,
    shareBrowserState,
    selectedShareItemCount,
    createShareResult,
    createShareResultOpen,
    setSharePasswordDialogOpen,
    setShareBrowserOpen,
    setShareSelectionOpen,
    setCreateShareResultOpen,
    closeShareFlow,
    handleOpenShareSelection,
    handleConfirmShareSelection,
    handleCreateShare,
  };
}
