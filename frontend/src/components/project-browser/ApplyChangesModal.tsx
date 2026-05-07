import { Alert, Descriptions, List, Modal, Typography } from 'antd';

type ApplyChangesModalProps = {
  open: boolean;
  loading: boolean;
  renameCount: number;
  moveCount: number;
  removeCount: number;
  addCount: number;
  createFolderCount: number;
  contentConnected: boolean;
  willWriteOperationGuide: boolean;
  blockingConflicts: string[];
  warnings: string[];
  onCancel: () => void;
  onConfirm: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ApplyChangesModal({
  open,
  loading,
  renameCount,
  moveCount,
  removeCount,
  addCount,
  createFolderCount,
  contentConnected,
  willWriteOperationGuide,
  blockingConflicts,
  warnings,
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
      {blockingConflicts.length > 0 ? (
        <Alert
          type="error"
          showIcon
          message={t('blockingConflicts')}
          description={
            <List
              size="small"
              dataSource={blockingConflicts}
              renderItem={(conflict) => <List.Item>{conflict}</List.Item>}
            />
          }
          style={{ marginBottom: 12 }}
        />
      ) : null}
      {warnings.length > 0 ? (
        <Alert
          type="warning"
          showIcon
          message={t('applyWarnings')}
          description={
            <List
              size="small"
              dataSource={warnings}
              renderItem={(warning) => <List.Item>{warning}</List.Item>}
            />
          }
          style={{ marginBottom: 12 }}
        />
      ) : null}
      <Descriptions column={1} bordered size="small">
        <Descriptions.Item label={t('pendingRename')}>{renameCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingMove')}>{moveCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingRemove')}>{removeCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingAdd')}>{addCount}</Descriptions.Item>
        <Descriptions.Item label={t('pendingCreateFolder')}>{createFolderCount}</Descriptions.Item>
        <Descriptions.Item label={t('contentConnected')}>
          {contentConnected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
        </Descriptions.Item>
        <Descriptions.Item label={t('operationGuidePath')}>
          {willWriteOperationGuide ? t('operationGuideWillBeWritten') : t('passwordProtectedNo')}
        </Descriptions.Item>
      </Descriptions>
    </Modal>
  );
}
