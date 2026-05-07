import { Button, Flex, Input, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { ProjectBrowserItemModel } from '../../types';
import type { PendingRename } from '../../hooks/useProjectBrowser';
import { displayNameForItem } from './projectBrowserView';

type ProjectBrowserItemTableProps = {
  items: ProjectBrowserItemModel[];
  pendingByID: Map<string, PendingRename>;
  selectedItem: ProjectBrowserItemModel | null;
  rootFolderID: string;
  searchQuery: string;
  applyLoading: boolean;
  pendingCount: number;
  onSearchChange: (value: string) => void;
  onSelectItem: (item: ProjectBrowserItemModel | null) => void;
  onOpenRename: () => void;
  onDiscardAll: () => void;
  onApply: () => void;
  t: (key: string) => string;
};

export function ProjectBrowserItemTable({
  items,
  pendingByID,
  selectedItem,
  rootFolderID,
  searchQuery,
  applyLoading,
  pendingCount,
  onSearchChange,
  onSelectItem,
  onOpenRename,
  onDiscardAll,
  onApply,
  t,
}: ProjectBrowserItemTableProps) {
  const columns: ColumnsType<ProjectBrowserItemModel> = [
    {
      title: t('itemName'),
      dataIndex: 'name',
      key: 'name',
      render: (_name: string, item) => displayNameForItem(item, pendingByID),
    },
    { title: t('itemType'), dataIndex: 'type', key: 'type', width: 110 },
    { title: t('fileSize'), dataIndex: 'size', key: 'size', width: 130 },
    { title: t('childCount'), dataIndex: 'childCount', key: 'childCount', width: 130 },
    { title: t('modifiedTime'), dataIndex: 'modifiedAt', key: 'modifiedAt', width: 180 },
    {
      title: t('contentStatus'),
      dataIndex: 'contentAvailable',
      key: 'contentAvailable',
      width: 150,
      render: (value: boolean) => (value ? t('available') : t('unavailable')),
    },
    {
      title: t('pendingState'),
      key: 'pendingState',
      width: 150,
      render: (_, item) => (pendingByID.has(item.id) ? t('pendingRename') : ''),
    },
  ];

  return (
    <div className="project-browser-items">
      <Flex justify="space-between" align="center" gap={12} wrap>
        <Typography.Title level={5}>{t('currentFolderItems')}</Typography.Title>
        <Space>
          <Button onClick={onOpenRename} disabled={!selectedItem || selectedItem.id === rootFolderID}>
            {t('renameItem')}
          </Button>
          <Button onClick={onDiscardAll} disabled={pendingCount === 0}>
            {t('discardChanges')}
          </Button>
          <Button type="primary" loading={applyLoading} disabled={pendingCount === 0} onClick={onApply}>
            {t('applyChanges')}
          </Button>
        </Space>
      </Flex>
      <Input.Search
        allowClear
        value={searchQuery}
        placeholder={t('searchItems')}
        onChange={(event) => onSearchChange(event.target.value)}
      />
      <Table
        rowKey="id"
        columns={columns}
        dataSource={items}
        pagination={false}
        size="small"
        scroll={{ x: 720 }}
        rowSelection={{
          type: 'radio',
          selectedRowKeys: selectedItem ? [selectedItem.id] : [],
          onChange: (_, rows) => onSelectItem(rows[0] ?? null),
        }}
        onRow={(item) => ({
          onClick: () => onSelectItem(item),
          onDoubleClick: () => {
            if (item.id !== rootFolderID) {
              onSelectItem(item);
              onOpenRename();
            }
          },
        })}
      />
    </div>
  );
}
