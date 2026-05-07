import { Button, Drawer, Space, Typography } from 'antd';
import { SafetyCertificateOutlined } from '@ant-design/icons';
import type { ShareSummaryModel } from '../../types';

type ShareActionsDrawerProps = {
  open: boolean;
  share: ShareSummaryModel | null;
  onClose: () => void;
  onInspect: () => void;
  onDecrypt: () => void;
  onVerify: () => void;
  t: (key: string) => string;
};

export function ShareActionsDrawer({ open, share, onClose, onInspect, onDecrypt, onVerify, t }: ShareActionsDrawerProps) {
  return (
    <Drawer title={t('shareActions')} open={open} onClose={onClose} width={360}>
      <Space direction="vertical" size="middle" className="content-stack">
        {share ? (
          <Typography.Text type="secondary">
            {share.shareId} / {share.databaseType}
          </Typography.Text>
        ) : null}
        <Button block type="primary" onClick={onInspect}>
          {t('inspectShare')}
        </Button>
        <Button block onClick={onDecrypt}>
          {t('decryptShare')}
        </Button>
        <Button block icon={<SafetyCertificateOutlined />} onClick={onVerify}>
          {t('verifyShare')}
        </Button>
      </Space>
    </Drawer>
  );
}
