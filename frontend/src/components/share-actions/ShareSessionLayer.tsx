import type { TFunction } from 'i18next';
import { ShareActionsDrawer } from './ShareActionsDrawer';
import { ShareInspectDrawer } from './ShareInspectDrawer';
import { VerifyShareModal } from './VerifyShareModal';
import { VerifyProjectDrawer } from '../project-actions/VerifyProjectDrawer';
import type { ShareSummaryModel, VerifyProjectResultModel } from '../../types';

type ShareSessionLayerProps = {
  shareActionsOpen: boolean;
  verifyShareDialogOpen: boolean;
  verifyShareLoading: boolean;
  verifyShareResult: VerifyProjectResultModel | null;
  verifyShareResultOpen: boolean;
  loadedShare: ShareSummaryModel | null;
  inspectShareOpen: boolean;
  onCloseShareSession: () => void;
  onOpenVerifyShare: () => void;
  onOpenInspectShare: () => void;
  onCloseInspectShare: () => void;
  onCloseVerifyShare: () => void;
  onVerifyShare: (values: { password: string; encryptedPath: string }) => void;
  onCloseVerifyShareResult: () => void;
  t: TFunction;
};

export function ShareSessionLayer({
  shareActionsOpen,
  verifyShareDialogOpen,
  verifyShareLoading,
  verifyShareResult,
  verifyShareResultOpen,
  loadedShare,
  inspectShareOpen,
  onCloseShareSession,
  onOpenVerifyShare,
  onOpenInspectShare,
  onCloseInspectShare,
  onCloseVerifyShare,
  onVerifyShare,
  onCloseVerifyShareResult,
  t,
}: ShareSessionLayerProps) {
  return (
    <>
      <ShareActionsDrawer
        open={shareActionsOpen}
        share={loadedShare}
        onClose={onCloseShareSession}
        onInspect={onOpenInspectShare}
        onVerify={onOpenVerifyShare}
        t={t}
      />
      <ShareInspectDrawer open={inspectShareOpen} share={loadedShare} onClose={onCloseInspectShare} t={t} />
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
        t={t}
      />
    </>
  );
}
