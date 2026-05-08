import { useEffect, useMemo, useState } from 'react';
import { App as AntApp, Breadcrumb, Button, Descriptions, Drawer, Flex, Input, Space, Table, Tree, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { ProjectBrowserItemModel, ProjectBrowserStateModel } from '../../types';
import { displayItemType } from '../../itemDisplay';
import { formatDateTime, formatFileSize, formatNumber } from '../../formatters';
import {
  buildFolderTree,
  filteredFolderItems,
  folderBreadcrumbItems,
  pendingRenameMap,
} from '../project-browser/projectBrowserView';

type ShareSelectionDrawerProps = {
  open: boolean;
  loading: boolean;
  state: ProjectBrowserStateModel | null;
  onCancel: () => void;
  onContinue: (itemPaths: string[]) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ShareSelectionDrawer({
  open,
  loading,
  state,
  onCancel,
  onContinue,
  t,
}: ShareSelectionDrawerProps) {
  const { modal } = AntApp.useApp();
  const root = state?.items.find((item) => item.id === state.rootFolderId) ?? null;
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [selectedItemIds, setSelectedItemIds] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const activeFolderId = selectedFolderId ?? root?.id ?? '';
  const pendingByID = useMemo(() => pendingRenameMap([]), []);
  const itemsByID = useMemo(() => new Map((state?.items ?? []).map((item) => [item.id, item])), [state?.items]);
  const selectedItems = useMemo(
    () =>
      selectedItemIds
        .map((itemId) => itemsByID.get(itemId))
        .filter((item): item is ProjectBrowserItemModel => Boolean(item)),
    [itemsByID, selectedItemIds],
  );
  const selectedTopLevelItems = useMemo(() => {
    const selectedIDs = new Set(selectedItemIds);
    return selectedItems.filter((item) => {
      for (let parentID = item.parentId; parentID; parentID = itemsByID.get(parentID)?.parentId ?? '') {
        if (selectedIDs.has(parentID)) {
          return false;
        }
      }
      return true;
    });
  }, [itemsByID, selectedItemIds, selectedItems]);
  const treeData = useMemo(() => buildFolderTree(state?.items ?? [], root?.id ?? '', pendingByID), [pendingByID, root?.id, state?.items]);
  const breadcrumbs = useMemo(
    () => folderBreadcrumbItems(state?.items ?? [], activeFolderId, pendingByID),
    [activeFolderId, pendingByID, state?.items],
  );
  const currentItems = useMemo(
    () => filteredFolderItems(state?.items ?? [], activeFolderId, searchQuery, pendingByID),
    [activeFolderId, pendingByID, searchQuery, state?.items],
  );
  const activeFolderSelected = activeFolderId ? selectedItemIds.includes(activeFolderId) : false;
  const columns: ColumnsType<ProjectBrowserItemModel> = [
    { title: t('itemName'), dataIndex: 'name', key: 'name' },
    { title: t('itemType'), dataIndex: 'type', key: 'type', width: 110, render: (value: string) => displayItemType(value, t) },
    { title: t('fileSize'), dataIndex: 'size', key: 'size', width: 130, render: (value: number) => formatFileSize(value) },
    { title: t('childCount'), dataIndex: 'childCount', key: 'childCount', width: 130, render: (value: number) => formatNumber(value) },
    { title: t('modifiedTime'), dataIndex: 'modifiedAt', key: 'modifiedAt', width: 180, render: (value: string) => formatDateTime(value) },
  ];

  useEffect(() => {
    if (open) {
      setSelectedFolderId(root?.id ?? null);
      setSelectedItemIds([]);
      setSearchQuery('');
    }
  }, [open, root?.id]);

  const selectFolder = (folderID: string) => {
    setSelectedFolderId(folderID);
    setSearchQuery('');
  };
  const setItemSelected = (itemId: string, selected: boolean) => {
    setSelectedItemIds((current) => {
      const next = new Set(current);
      if (selected) {
        next.add(itemId);
      } else {
        next.delete(itemId);
      }
      return [...next];
    });
  };
  const setCurrentItemsSelected = (selected: boolean) => {
    setSelectedItemIds((current) => {
      const next = new Set(current);
      for (const item of currentItems) {
        if (selected) {
          next.add(item.id);
        } else {
          next.delete(item.id);
        }
      }
      return [...next];
    });
  };
  const toggleActiveFolder = () => {
    if (activeFolderId) {
      setItemSelected(activeFolderId, !activeFolderSelected);
    }
  };
  const hasUnsavedSelection = selectedItemIds.length > 0 || searchQuery !== '' || activeFolderId !== (root?.id ?? '');
  const closeDrawer = () => {
    if (!hasUnsavedSelection) {
      onCancel();
      return;
    }
    modal.confirm({
      title: t('unsavedChanges'),
      content: t('unsavedChangesConfirm'),
      okText: t('discardAndClose'),
      cancelText: t('stay'),
      onOk: onCancel,
    });
  };

  return (
    <Drawer title={t('shareSelectionTitle')} open={open} onClose={closeDrawer} width={1120} maskClosable={false}>
      {state ? (
        <Flex vertical gap={18}>
          <Descriptions column={4} bordered size="small">
            <Descriptions.Item label={t('projectName')}>{state.projectName}</Descriptions.Item>
            <Descriptions.Item label={t('projectId')}>{state.projectId}</Descriptions.Item>
            <Descriptions.Item label={t('fileCount')}>{formatNumber(state.files)}</Descriptions.Item>
            <Descriptions.Item label={t('folderCount')}>{formatNumber(state.folders)}</Descriptions.Item>
          </Descriptions>
          <Breadcrumb
            items={breadcrumbs.map((breadcrumb) => ({
              title: (
                <button className="project-browser-breadcrumb-button" type="button" onClick={() => selectFolder(breadcrumb.key)}>
                  {breadcrumb.title}
                </button>
              ),
            }))}
          />
          <div className="project-browser-grid">
            <div className="project-browser-tree">
              <Typography.Title level={5}>{t('folderTree')}</Typography.Title>
              <Tree
                treeData={treeData}
                selectedKeys={activeFolderId ? [activeFolderId] : []}
                defaultExpandAll
                onSelect={(keys) => selectFolder((keys[0] as string | undefined) ?? root?.id ?? '')}
              />
            </div>
            <div className="project-browser-items">
              <Flex justify="space-between" align="center" gap={12} wrap>
                <Typography.Title level={5}>{t('shareSelectionItems')}</Typography.Title>
                <Space wrap>
                  <Typography.Text>{t('selectedShareItemCount', { count: selectedTopLevelItems.length })}</Typography.Text>
                  <Button onClick={toggleActiveFolder} disabled={!activeFolderId}>
                    {activeFolderSelected ? t('unselectCurrentFolder') : t('selectCurrentFolder')}
                  </Button>
                  <Button
                    type="primary"
                    loading={loading}
                    disabled={selectedTopLevelItems.length === 0}
                    onClick={() => onContinue(selectedTopLevelItems.map((item) => item.path))}
                  >
                    {t('continueAction')}
                  </Button>
                  <Button onClick={closeDrawer}>{t('close')}</Button>
                </Space>
              </Flex>
              <Input.Search
                allowClear
                value={searchQuery}
                placeholder={t('searchItems')}
                onChange={(event) => setSearchQuery(event.target.value)}
              />
              <Table
                rowKey="id"
                columns={columns}
                dataSource={currentItems}
                pagination={false}
                size="small"
                scroll={{ x: 720 }}
                rowSelection={{
                  type: 'checkbox',
                  selectedRowKeys: selectedItemIds,
                  onSelect: (item, selected) => setItemSelected(item.id, selected),
                  onSelectAll: (selected) => setCurrentItemsSelected(selected),
                }}
                onRow={(item) => ({
                  onDoubleClick: () => {
                    if (item.type === 'folder') {
                      selectFolder(item.id);
                    }
                  },
                })}
              />
            </div>
          </div>
        </Flex>
      ) : null}
    </Drawer>
  );
}
