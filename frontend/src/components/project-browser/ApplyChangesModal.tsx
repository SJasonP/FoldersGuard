import { Descriptions, Modal, Typography } from 'antd';

type ApplyChangesModalProps = {
  open: boolean;
  loading: boolean;
  renameCount: number;
  moveCount: number;
  removeCount: number;
  contentConnected: boolean;
  onCancel: () => void;
  onConfirm: () => void;
  t: (key: string) => string;
};

export function ApplyChangesModal({
  open,
  loading,
  renameCount,
  moveCount,
  removeCount,
  contentConnected,
  onCancel,
  onConfirm,
  t,
}: ApplyChangesModalProps) {
  return (
    <Modal
      title={t('applyChanges')}
      open={open}
      confirmLoading={loading}
      onCancel={onCancel}
      onOk={onConfirm}
      okText={t('applyChanges')}
    >
      <Typography.Paragraph>{t('applyChangesConfirm')}</Typography.Paragraph>
      <Descriptions column={1} bordered size="small">
        <Descriptions.Item label={t('pendingRename')}>{renameCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingMove')}>{moveCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingRemove')}>{removeCount}</Descriptions.Item>
        <Descriptions.Item label={t('contentConnected')}>
          {contentConnected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  );
}
