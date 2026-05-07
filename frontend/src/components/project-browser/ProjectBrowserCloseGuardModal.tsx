import { Button, Modal, Space, Typography } from 'antd';

type ProjectBrowserCloseGuardModalProps = {
  open: boolean;
  applyBlocked: boolean;
  applyLoading: boolean;
  onApply: () => void;
  onCancel: () => void;
  onDiscard: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ProjectBrowserCloseGuardModal({
  open,
  applyBlocked,
  applyLoading,
  onApply,
  onCancel,
  onDiscard,
  t,
}: ProjectBrowserCloseGuardModalProps) {
  return (
    <Modal
      title={t('unappliedProjectChanges')}
      open={open}
      onCancel={onCancel}
      footer={
        <Space>
          <Button onClick={onCancel}>{t('stay')}</Button>
          <Button danger onClick={onDiscard}>
            {t('discardAndClose')}
          </Button>
          <Button type="primary" loading={applyLoading} disabled={applyBlocked} onClick={onApply}>
            {t('applyChanges')}
          </Button>
        </Space>
      }
      destroyOnHidden
    >
      <Typography.Paragraph>{t('unappliedProjectChangesConfirm')}</Typography.Paragraph>
    </Modal>
  );
}
