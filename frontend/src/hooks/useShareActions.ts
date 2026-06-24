import {useState} from 'react';
import type {MessageInstance} from 'antd/es/message/interface';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';
import {DecryptShare, LoadShare, VerifyShare} from '../../wailsjs/go/main/App';
import type {DecryptShareResultModel, ShareSummaryModel, VerifyProjectResultModel} from '../types';
import {showOperationError} from '../components/common/operationError';

type UseShareActionsArgs = {
    messageApi: MessageInstance;
    modalApi: ModalHookAPI;
    t: (key: string) => string;
};

export function useShareActions({messageApi, modalApi, t}: UseShareActionsArgs) {
    const [loadShareDialogOpen, setLoadShareDialogOpen] = useState(false);
    const [shareLoading, setShareLoading] = useState(false);
    const [loadedShare, setLoadedShare] = useState<ShareSummaryModel | null>(null);
    const [loadedShareDatabasePath, setLoadedShareDatabasePath] = useState('');
    const [shareActionsOpen, setShareActionsOpen] = useState(false);
    const [inspectShareOpen, setInspectShareOpen] = useState(false);
    const [decryptShareDialogOpen, setDecryptShareDialogOpen] = useState(false);
    const [decryptShareLoading, setDecryptShareLoading] = useState(false);
    const [decryptShareResult, setDecryptShareResult] = useState<DecryptShareResultModel | null>(null);
    const [decryptShareResultOpen, setDecryptShareResultOpen] = useState(false);
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
        } catch (error) {
            showOperationError(modalApi, t('loadShareFailed'), error, t);
        } finally {
            setShareLoading(false);
        }
    };

    const handleVerifyShare = async (values: { password: string; encryptedPath: string }) => {
        setVerifyShareDialogOpen(false);
        setVerifyShareLoading(true);
        try {
            const result = await VerifyShare({
                databasePath: loadedShareDatabasePath,
                password: values.password,
                encryptedPath: values.encryptedPath,
            });
            setVerifyShareResult(result);
            setVerifyShareResultOpen(true);
            messageApi.success(t('verifyShareSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('verifyShareFailed'), error, t);
        } finally {
            setVerifyShareLoading(false);
        }
    };

    const handleDecryptShare = async (values: {
        password: string;
        encryptedPath: string;
        outputPath: string;
        force: boolean;
    }) => {
        setDecryptShareDialogOpen(false);
        setDecryptShareLoading(true);
        try {
            const result = await DecryptShare({
                databasePath: loadedShareDatabasePath,
                password: values.password,
                encryptedPath: values.encryptedPath,
                outputPath: values.outputPath,
                force: values.force,
                sourceCleanup: '',
            });
            setDecryptShareResult(result);
            setDecryptShareResultOpen(true);
            messageApi.success(t('decryptShareSucceeded'));
        } catch (error) {
            showOperationError(modalApi, t('decryptShareFailed'), error, t);
        } finally {
            setDecryptShareLoading(false);
        }
    };

    const closeShareSession = () => {
        setShareActionsOpen(false);
        setInspectShareOpen(false);
        setDecryptShareDialogOpen(false);
        setDecryptShareResultOpen(false);
        setDecryptShareResult(null);
        setVerifyShareDialogOpen(false);
        setVerifyShareResultOpen(false);
        setVerifyShareResult(null);
        setLoadedShare(null);
        setLoadedShareDatabasePath('');
    };

    return {
        closeShareSession,
        decryptShareDialogOpen,
        decryptShareLoading,
        decryptShareResult,
        decryptShareResultOpen,
        handleDecryptShare,
        handleLoadShare,
        handleVerifyShare,
        loadShareDialogOpen,
        loadedShare,
        loadedShareDatabasePath,
        inspectShareOpen,
        setLoadShareDialogOpen,
        setInspectShareOpen,
        setDecryptShareDialogOpen,
        setDecryptShareResultOpen,
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
