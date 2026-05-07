import { Descriptions, Empty, Typography } from 'antd';
import type { ProjectBrowserItemModel } from '../../types';
import type { PendingRename } from '../../hooks/useProjectBrowser';
import { displayNameForItem } from './projectBrowserView';
import { formatDateTime, formatFileSize, formatNumber } from '../../formatters';
import { displayItemType } from '../../itemDisplay';

type ProjectBrowserDetailsPanelProps = {
  item: ProjectBrowserItemModel | null;
  pendingByID: Map<string, PendingRename>;
  pendingStateByID: Map<string, string>;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ProjectBrowserDetailsPanel({ item, pendingByID, pendingStateByID, t }: ProjectBrowserDetailsPanelProps) {
  return (
    <div className="project-browser-details">
      <Typography.Title level={5}>{t('itemDetails')}</Typography.Title>
      {item ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('itemName')}>{displayNameForItem(item, pendingByID)}</Descriptions.Item>
          <Descriptions.Item label={t('itemType')}>{displayItemType(item.type, t)}</Descriptions.Item>
          <Descriptions.Item label={t('itemId')}>{item.id}</Descriptions.Item>
          <Descriptions.Item label={t('itemPath')}>{item.path}</Descriptions.Item>
          <Descriptions.Item label={t('parentPath')}>{item.parentPath}</Descriptions.Item>
          <Descriptions.Item label={t('fileSize')}>{formatFileSize(item.size)}</Descriptions.Item>
          <Descriptions.Item label={t('childCount')}>{formatNumber(item.childCount)}</Descriptions.Item>
          <Descriptions.Item label={t('modifiedTime')}>{formatDateTime(item.modifiedAt)}</Descriptions.Item>
          <Descriptions.Item label={t('metadataCaptured')}>
            {item.metadataCaptured ? t('passwordProtectedYes') : t('passwordProtectedNo')}
          </Descriptions.Item>
          <Descriptions.Item label={t('contentStatus')}>
            {item.contentAvailable ? t('available') : t('unavailable')}
          </Descriptions.Item>
          <Descriptions.Item label={t('pendingState')}>{pendingStateByID.get(item.id) ?? ''}</Descriptions.Item>
        </Descriptions>
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('noItemSelected')} />
      )}
    </div>
  );
}
