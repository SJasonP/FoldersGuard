import { useMemo, useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { CreateShare, ListShareableItems } from '../../wailsjs/go/main/App';
import type { CreateShareResultModel, ShareableItemModel } from '../types';
import { showOperationError } from '../components/common/operationError';

type UseProjectShareArgs = {
  messageApi: MessageInstance;
  modalApi: ModalHookAPI;
  t: (key: string) => string;
  selectedProjectId: string | null;
};

export function useProjectShare({ messageApi, modalApi, t, selectedProjectId }: UseProjectShareArgs) {
  const [sharePasswordDialogOpen, setSharePasswordDialogOpen] = useState(false);
  const [shareSelectionOpen, setShareSelectionOpen] = useState(false);
  const [shareLoading, setShareLoading] = useState(false);
  const [shareItems, setShareItems] = useState<ShareableItemModel[]>([]);
  const [projectPassword, setProjectPassword] = useState('');
  const [createShareResult, setCreateShareResult] = useState<CreateShareResultModel | null>(null);
  const [createShareResultOpen, setCreateShareResultOpen] = useState(false);

  const selectableItems = useMemo(
    () =>
      shareItems.map((item) => ({
        value: item.path,
        label: `${item.path} (${item.type})`,
      })),
    [shareItems],
  );

  const handleOpenShareSelection = async (password: string) => {
    if (!selectedProjectId) {
      return;
    }
    setShareLoading(true);
    try {
      const items = await ListShareableItems({
        projectId: selectedProjectId,
        password,
      });
      setProjectPassword(password);
      setShareItems(items);
      setSharePasswordDialogOpen(false);
      setShareSelectionOpen(true);
    } catch (error) {
      showOperationError(modalApi, t('loadShareableItemsFailed'), error, t);
    } finally {
      setShareLoading(false);
    }
  };

  const handleCreateShare = async (values: {
    itemPaths: string[];
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
        itemPaths: values.itemPaths,
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

  const closeShareFlow = () => {
    setSharePasswordDialogOpen(false);
    setShareSelectionOpen(false);
    setShareItems([]);
    setProjectPassword('');
  };

  return {
    sharePasswordDialogOpen,
    shareSelectionOpen,
    shareLoading,
    shareItems,
    selectableItems,
    createShareResult,
    createShareResultOpen,
    setSharePasswordDialogOpen,
    setShareSelectionOpen,
    setCreateShareResultOpen,
    closeShareFlow,
    handleOpenShareSelection,
    handleCreateShare,
  };
}
