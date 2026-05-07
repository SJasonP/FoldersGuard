import { Button, Flex, Input, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { ProjectBrowserItemModel } from '../../types';
import type { PendingRename } from '../../hooks/useProjectBrowser';
import { displayNameForItem } from './projectBrowserView';
import { formatDateTime, formatFileSize, formatNumber } from '../../formatters';

type ProjectBrowserItemTableProps = {
  items: ProjectBrowserItemModel[];
  pendingByID: Map<string, PendingRename>;
  pendingStateByID: Map<string, string>;
  selectedItemIds: string[];
  rootFolderID: string;
  searchQuery: string;
  applyLoading: boolean;
  pendingCount: number;
  applyBlocked: boolean;
  onSearchChange: (value: string) => void;
  onSelectItem: (item: ProjectBrowserItemModel | null) => void;
  onSelectItems: (items: ProjectBrowserItemModel[]) => void;
  onOpenAdd: () => void;
  onOpenCreateFolder: () => void;
  onOpenRename: () => void;
  onOpenMove: () => void;
  onRemove: () => void;
  onDiscardAll: () => void;
  onApply: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ProjectBrowserItemTable({
  items,
  pendingByID,
  pendingStateByID,
  selectedItemIds,
  rootFolderID,
  searchQuery,
  applyLoading,
  pendingCount,
  applyBlocked,
  onSearchChange,
  onSelectItem,
  onSelectItems,
  onOpenAdd,
  onOpenCreateFolder,
  onOpenRename,
  onOpenMove,
  onRemove,
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
    { title: t('fileSize'), dataIndex: 'size', key: 'size', width: 130, render: (value: number) => formatFileSize(value) },
    { title: t('childCount'), dataIndex: 'childCount', key: 'childCount', width: 130, render: (value: number) => formatNumber(value) },
    { title: t('modifiedTime'), dataIndex: 'modifiedAt', key: 'modifiedAt', width: 180, render: (value: string) => formatDateTime(value) },
    {
      title: t('metadataCaptured'),
      dataIndex: 'metadataCaptured',
      key: 'metadataCaptured',
      width: 150,
      render: (value: boolean) => (value ? t('passwordProtectedYes') : t('passwordProtectedNo')),
    },
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
      render: (_, item) => pendingStateByID.get(item.id) ?? '',
    },
  ];
  const selectedItems = items.filter((item) => selectedItemIds.includes(item.id));
  const noEditableSelection = selectedItems.length === 0 || selectedItems.some((item) => item.id === rootFolderID);
  const renameDisabled = selectedItems.length !== 1 || selectedItems[0]?.id === rootFolderID;

  return (
    <div className="project-browser-items">
      <Flex justify="space-between" align="center" gap={12} wrap>
        <Typography.Title level={5}>{t('currentFolderItems')}</Typography.Title>
        <Space>
          <Button onClick={onOpenAdd}>
            {t('addItem')}
          </Button>
          <Button onClick={onOpenCreateFolder}>
            {t('createFolder')}
          </Button>
          <Button onClick={onOpenRename} disabled={renameDisabled}>
            {t('renameItem')}
          </Button>
          <Button onClick={onOpenMove} disabled={noEditableSelection}>
            {t('moveItem')}
          </Button>
          <Button danger onClick={onRemove} disabled={noEditableSelection}>
            {t('removeItem')}
          </Button>
          <Button onClick={onDiscardAll} disabled={pendingCount === 0}>
            {t('discardChanges')}
          </Button>
          <Button type="primary" loading={applyLoading} disabled={pendingCount === 0 || applyBlocked} onClick={onApply}>
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
          type: 'checkbox',
          selectedRowKeys: selectedItemIds,
          onChange: (_, rows) => {
            onSelectItems(rows);
            onSelectItem(rows[0] ?? null);
          },
        }}
        onRow={(item) => ({
          onClick: () => {
            onSelectItem(item);
            onSelectItems([item]);
          },
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
