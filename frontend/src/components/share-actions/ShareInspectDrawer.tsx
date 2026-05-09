import { Descriptions, Drawer } from 'antd';
import type { ShareSummaryModel } from '../../types';
import { formatNumber } from '../../formatters';

type ShareInspectDrawerProps = {
  open: boolean;
  share: ShareSummaryModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function ShareInspectDrawer({ open, share, onClose, t }: ShareInspectDrawerProps) {
  return (
    <Drawer title={t('shareDetails')} open={open} onClose={onClose} width={540}>
      {share ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('shareId')}>{share.shareId}</Descriptions.Item>
          <Descriptions.Item label={t('databaseType')}>{share.databaseType}</Descriptions.Item>
          <Descriptions.Item label={t('formatVersion')}>{share.formatVersion}</Descriptions.Item>
          <Descriptions.Item label={t('shareSummaryTopLevelItems')}>{formatNumber(share.topLevelItems)}</Descriptions.Item>
          <Descriptions.Item label={t('fileCount')}>{formatNumber(share.files)}</Descriptions.Item>
          <Descriptions.Item label={t('folderCount')}>{formatNumber(share.folders)}</Descriptions.Item>
          <Descriptions.Item label={t('partCount')}>{formatNumber(share.parts)}</Descriptions.Item>
          <Descriptions.Item label={t('storageObjects')}>{formatNumber(share.storageObjects)}</Descriptions.Item>
          <Descriptions.Item label={t('passwordProtected')}>
            {share.passwordProtected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
          </Descriptions.Item>
        </Descriptions>
      ) : null}
    </Drawer>
  );
}
