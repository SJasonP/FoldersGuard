import { Alert, App as AntApp, Descriptions, Drawer, List, Typography } from 'antd';
import type { ApplyProjectChangesResultModel } from '../../types';
import { formatNumber } from '../../formatters';

type ApplyChangesResultDrawerProps = {
  open: boolean;
  result: ApplyProjectChangesResultModel | null;
  onClose: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ApplyChangesResultDrawer({ open, result, onClose, t }: ApplyChangesResultDrawerProps) {
  const { modal } = AntApp.useApp();
  const manualOperationGuide = Boolean(result?.operationGuidePath);
  const manualOperations = manualOperationGuide ? (result?.contentOperations ?? []) : [];
  const closeOrConfirm = () => {
    if (!manualOperationGuide) {
      onClose();
      return;
    }
    modal.confirm({
      title: t('operationGuideCloseTitle'),
      content: t('operationGuideCloseConfirm', { path: result?.operationGuidePath ?? '' }),
      okText: t('close'),
      cancelText: t('stay'),
      onOk: onClose,
    });
  };

  return (
    <Drawer title={t('applyChangesResult')} open={open} onClose={closeOrConfirm} width={760} maskClosable={!manualOperationGuide}>
      {result ? (
        <>
          {manualOperationGuide ? (
            <Alert
              type="warning"
              showIcon
              message={t('operationGuideCreated')}
              description={t('operationGuideCreatedDescription', { path: result.operationGuidePath })}
              style={{ marginBottom: 12 }}
            />
          ) : result.appliedContentChanges?.length ? (
            <Alert
              type="success"
              showIcon
              message={t('contentOperationsApplied')}
              description={t('contentOperationsAppliedDescription')}
              style={{ marginBottom: 12 }}
            />
          ) : null}
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
            <Descriptions.Item label={t('appliedRenames')}>{formatNumber(result.appliedRenames)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedMoves')}>{formatNumber(result.appliedMoves)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedRemoves')}>{formatNumber(result.appliedRemoves)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedAdds')}>{formatNumber(result.appliedAdds)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedCreatedFolders')}>{formatNumber(result.appliedCreatedFolders)}</Descriptions.Item>
            {result.operationGuidePath ? <Descriptions.Item label={t('operationGuidePath')}>{result.operationGuidePath}</Descriptions.Item> : null}
            {result.stagedContentPath ? <Descriptions.Item label={t('stagedContentPath')}>{result.stagedContentPath}</Descriptions.Item> : null}
            {manualOperationGuide ? (
              <Descriptions.Item label={t('manualContentOperations')}>{formatNumber(manualOperations.length)}</Descriptions.Item>
            ) : null}
          </Descriptions>
          {manualOperationGuide ? (
            <>
              <Typography.Title level={5}>{t('manualContentOperations')}</Typography.Title>
              <List
                size="small"
                bordered
                dataSource={manualOperations}
                locale={{ emptyText: t('noContentOperations') }}
                renderItem={(operation) => (
                  <List.Item>
                    <Typography.Text>
                      {t(`contentOperation_${operation.type}`)}: {operation.sourcePath ? `${operation.sourcePath} -> ` : ''}
                      {operation.targetPath}
                    </Typography.Text>
                  </List.Item>
                )}
              />
            </>
          ) : null}
        </>
      ) : null}
    </Drawer>
  );
}
