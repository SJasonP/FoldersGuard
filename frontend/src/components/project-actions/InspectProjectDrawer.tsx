import { Descriptions, Drawer } from 'antd';
import type { InspectProjectResultModel } from '../../types';

type InspectProjectDrawerProps = {
  open: boolean;
  result: InspectProjectResultModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function InspectProjectDrawer({ open, result, onClose, t }: InspectProjectDrawerProps) {
  return (
    <Drawer title={t('projectDetails')} open={open} onClose={onClose} width={540}>
      {result ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
          <Descriptions.Item label={t('projectName')}>{result.rootName}</Descriptions.Item>
          <Descriptions.Item label={t('rootFolderId')}>{result.rootFolderId}</Descriptions.Item>
          <Descriptions.Item label={t('rootName')}>{result.rootName}</Descriptions.Item>
          <Descriptions.Item label={t('formatVersion')}>{result.formatVersion}</Descriptions.Item>
          <Descriptions.Item label={t('schemaVersion')}>{result.schemaVersion}</Descriptions.Item>
          <Descriptions.Item label={t('databaseType')}>{result.databaseType}</Descriptions.Item>
          <Descriptions.Item label={t('itemCount')}>{result.items}</Descriptions.Item>
          <Descriptions.Item label={t('folderCount')}>{result.folders}</Descriptions.Item>
          <Descriptions.Item label={t('fileCount')}>{result.files}</Descriptions.Item>
          <Descriptions.Item label={t('partCount')}>{result.parts}</Descriptions.Item>
          <Descriptions.Item label={t('storageObjects')}>{result.storageObjects}</Descriptions.Item>
        </Descriptions>
      ) : null}
    </Drawer>
  );
}
