import { Descriptions, Drawer, List, Typography } from 'antd';
import type { ApplyProjectChangesResultModel } from '../../types';
import { formatNumber } from '../../formatters';

type ApplyChangesResultDrawerProps = {
  open: boolean;
  result: ApplyProjectChangesResultModel | null;
  onClose: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ApplyChangesResultDrawer({ open, result, onClose, t }: ApplyChangesResultDrawerProps) {
  const operations = result?.appliedContentChanges?.length ? result.appliedContentChanges : (result?.contentOperations ?? []);

  return (
    <Drawer title={t('applyChangesResult')} open={open} onClose={onClose} width={720}>
      {result ? (
        <>
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
            <Descriptions.Item label={t('appliedRenames')}>{formatNumber(result.appliedRenames)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedMoves')}>{formatNumber(result.appliedMoves)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedRemoves')}>{formatNumber(result.appliedRemoves)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedAdds')}>{formatNumber(result.appliedAdds)}</Descriptions.Item>
            <Descriptions.Item label={t('appliedCreatedFolders')}>{formatNumber(result.appliedCreatedFolders)}</Descriptions.Item>
            <Descriptions.Item label={t('operationGuidePath')}>{result.operationGuidePath}</Descriptions.Item>
            <Descriptions.Item label={t('stagedContentPath')}>{result.stagedContentPath}</Descriptions.Item>
            <Descriptions.Item label={t('contentOperations')}>{formatNumber(operations.length)}</Descriptions.Item>
          </Descriptions>
          <Typography.Title level={5}>{t('contentOperations')}</Typography.Title>
          <List
            size="small"
            bordered
            dataSource={operations}
            locale={{ emptyText: t('noContentOperations') }}
            renderItem={(operation) => (
              <List.Item>
                <Typography.Text>
                  {operation.type}: {operation.sourcePath ? `${operation.sourcePath} -> ` : ''}
                  {operation.targetPath}
                </Typography.Text>
              </List.Item>
            )}
          />
        </>
      ) : null}
    </Drawer>
  );
}
