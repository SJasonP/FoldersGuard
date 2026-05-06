import { useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { LoadShare, VerifyShare } from '../../wailsjs/go/main/App';
import type { ShareSummaryModel, VerifyProjectResultModel } from '../types';

type UseShareActionsArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
};

export function useShareActions({ messageApi, t }: UseShareActionsArgs) {
  const [loadShareDialogOpen, setLoadShareDialogOpen] = useState(false);
  const [shareLoading, setShareLoading] = useState(false);
  const [loadedShare, setLoadedShare] = useState<ShareSummaryModel | null>(null);
  const [loadedShareDatabasePath, setLoadedShareDatabasePath] = useState('');
  const [shareActionsOpen, setShareActionsOpen] = useState(false);
  const [inspectShareOpen, setInspectShareOpen] = useState(false);
  const [verifyShareDialogOpen, setVerifyShareDialogOpen] = useState(false);
  const [verifyShareLoading, setVerifyShareLoading] = useState(false);
  const [verifyShareResult, setVerifyShareResult] = useState<VerifyProjectResultModel | null>(null);
  const [verifyShareResultOpen, setVerifyShareResultOpen] = useState(false);

  const handleLoadShare = async (values: { databasePath: string; password: string }) => {
    setShareLoading(true);
    try {
      const result = await LoadShare({
        databasePath: values.databasePath,
        password: values.password,
      });
      setLoadedShare(result);
      setLoadedShareDatabasePath(values.databasePath);
      setLoadShareDialogOpen(false);
      setShareActionsOpen(true);
      messageApi.success(t('loadShareSucceeded'));
    } catch {
      messageApi.error(t('loadShareFailed'));
    } finally {
      setShareLoading(false);
    }
  };

  const handleVerifyShare = async (values: { password: string; encryptedPath: string }) => {
    setVerifyShareLoading(true);
    try {
      const result = await VerifyShare({
        databasePath: loadedShareDatabasePath,
        password: values.password,
        encryptedPath: values.encryptedPath,
      });
      setVerifyShareDialogOpen(false);
      setVerifyShareResult(result);
      setVerifyShareResultOpen(true);
      messageApi.success(t('verifyShareSucceeded'));
    } catch {
      messageApi.error(t('verifyShareFailed'));
    } finally {
      setVerifyShareLoading(false);
    }
  };

  const closeShareSession = () => {
    setShareActionsOpen(false);
    setInspectShareOpen(false);
    setVerifyShareDialogOpen(false);
    setVerifyShareResultOpen(false);
    setVerifyShareResult(null);
    setLoadedShare(null);
    setLoadedShareDatabasePath('');
  };

  return {
    closeShareSession,
    handleLoadShare,
    handleVerifyShare,
    loadShareDialogOpen,
    loadedShare,
    loadedShareDatabasePath,
    inspectShareOpen,
    setLoadShareDialogOpen,
    setInspectShareOpen,
    setVerifyShareDialogOpen,
    setVerifyShareResultOpen,
    shareActionsOpen,
    shareLoading,
    verifyShareDialogOpen,
    verifyShareLoading,
    verifyShareResult,
    verifyShareResultOpen,
  };
}
