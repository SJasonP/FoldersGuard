import type {TFunction} from 'i18next';
import {ChangePasswordModal} from '../common/ChangePasswordModal';
import {ShareActionsDrawer} from './ShareActionsDrawer';
import {ShareInspectDrawer} from './ShareInspectDrawer';
import {DecryptShareDrawer} from './DecryptShareDrawer';
import {DecryptShareModal} from './DecryptShareModal';
import {VerifyShareModal} from './VerifyShareModal';
import {VerifyProjectDrawer} from '../project-actions/VerifyProjectDrawer';
import type {DecryptShareResultModel, SettingsModel, ShareSummaryModel, VerifyProjectResultModel} from '../../types';

type ShareSessionLayerProps = {
    decryptShareDialogOpen: boolean;
    decryptShareLoading: boolean;
    decryptShareResult: DecryptShareResultModel | null;
    decryptShareResultOpen: boolean;
    settings: SettingsModel | null;
    shareActionsOpen: boolean;
    verifyShareDialogOpen: boolean;
    verifyShareLoading: boolean;
    verifyShareResult: VerifyProjectResultModel | null;
    verifyShareResultOpen: boolean;
    loadedShare: ShareSummaryModel | null;
    inspectShareOpen: boolean;
    onCloseShareSession: () => void;
    onOpenDecryptShare: () => void;
    onOpenVerifyShare: () => void;
    onOpenInspectShare: () => void;
    onCloseInspectShare: () => void;
    onCloseDecryptShare: () => void;
    onDecryptShare: (values: {
        password: string;
        encryptedPath: string;
        outputPath: string;
        force: boolean;
        resume: boolean;
        continueOnError: boolean;
    }) => void;
    onCloseDecryptShareResult: () => void;
    onCloseVerifyShare: () => void;
    onVerifyShare: (values: { password: string; encryptedPath: string }) => void;
    onCloseVerifyShareResult: () => void;
    changeSharePasswordOpen: boolean;
    changeSharePasswordLoading: boolean;
    onOpenChangeSharePassword: () => void;
    onCloseChangeSharePassword: () => void;
    onChangeSharePassword: (values: { oldPassword: string; newPassword: string }) => void;
    t: TFunction;
};

export function ShareSessionLayer({
                                      decryptShareDialogOpen,
                                      decryptShareLoading,
                                      decryptShareResult,
                                      decryptShareResultOpen,
                                      settings,
                                      shareActionsOpen,
                                      verifyShareDialogOpen,
                                      verifyShareLoading,
                                      verifyShareResult,
                                      verifyShareResultOpen,
                                      loadedShare,
                                      inspectShareOpen,
                                      onCloseShareSession,
                                      onOpenDecryptShare,
                                      onOpenVerifyShare,
                                      onOpenInspectShare,
                                      onCloseInspectShare,
                                      onCloseDecryptShare,
                                      onDecryptShare,
                                      onCloseDecryptShareResult,
                                      onCloseVerifyShare,
                                      onVerifyShare,
                                      onCloseVerifyShareResult,
                                      changeSharePasswordOpen,
                                      changeSharePasswordLoading,
                                      onOpenChangeSharePassword,
                                      onCloseChangeSharePassword,
                                      onChangeSharePassword,
                                      t,
                                  }: ShareSessionLayerProps) {
    return (
        <>
            <ShareActionsDrawer
                open={shareActionsOpen}
                share={loadedShare}
                onClose={onCloseShareSession}
                onInspect={onOpenInspectShare}
                onDecrypt={onOpenDecryptShare}
                onVerify={onOpenVerifyShare}
                onChangePassword={onOpenChangeSharePassword}
                t={t}
            />
            <ShareInspectDrawer open={inspectShareOpen} share={loadedShare} onClose={onCloseInspectShare} t={t}/>
            <DecryptShareModal
                open={decryptShareDialogOpen}
                loading={decryptShareLoading}
                sourceCleanupMode={settings?.sourceCleanupMode ?? 'delete'}
                defaultFailureHandling={settings?.failureHandling ?? 'abort'}
                onCancel={onCloseDecryptShare}
                onSubmit={(values) => onDecryptShare(values)}
                t={t}
            />
            <DecryptShareDrawer
                open={decryptShareResultOpen}
                result={decryptShareResult}
                onClose={onCloseDecryptShareResult}
                t={t}
            />
            <VerifyShareModal
                open={verifyShareDialogOpen}
                loading={verifyShareLoading}
                onCancel={onCloseVerifyShare}
                onSubmit={(values) => onVerifyShare(values)}
                t={t}
            />
            <VerifyProjectDrawer
                open={verifyShareResultOpen}
                result={verifyShareResult}
                onClose={onCloseVerifyShareResult}
                title={t('verifyShare')}
                identityLabel={t('shareId')}
                t={t}
            />
            <ChangePasswordModal
                open={changeSharePasswordOpen}
                loading={changeSharePasswordLoading}
                title={t('changeSharePassword')}
                onCancel={onCloseChangeSharePassword}
                onSubmit={onChangeSharePassword}
                t={t}
            />
        </>
    );
}
