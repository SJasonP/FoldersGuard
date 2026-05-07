import { useEffect, useMemo, useState } from 'react';
import { Button, Descriptions, Drawer, Flex, List, Space, Table, Tree, Typography } from 'antd';
import type { DataNode } from 'antd/es/tree';
import type { ColumnsType } from 'antd/es/table';
import type { ProjectBrowserItemModel, ProjectBrowserStateModel } from '../../types';
import { RenameItemModal } from './RenameItemModal';
import type { PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserDrawerProps = {
  open: boolean;
  state: ProjectBrowserStateModel | null;
  pendingRenames: PendingRename[];
  applyLoading: boolean;
  onClose: () => void;
  onRename: (rename: PendingRename) => void;
  onDiscardRename: (itemId: string) => void;
  onDiscardAll: () => void;
  onApply: () => void;
  t: (key: string) => string;
};

export function ProjectBrowserDrawer({
  open,
  state,
  pendingRenames,
  applyLoading,
  onClose,
  onRename,
  onDiscardRename,
  onDiscardAll,
  onApply,
  t,
}: ProjectBrowserDrawerProps) {
  const root = state?.items.find((item) => item.id === state.rootFolderId) ?? null;
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [selectedItem, setSelectedItem] = useState<ProjectBrowserItemModel | null>(null);
  const [renameOpen, setRenameOpen] = useState(false);
  const activeFolderId = selectedFolderId ?? root?.id ?? '';
  const pendingByID = useMemo(() => new Map(pendingRenames.map((rename) => [rename.itemId, rename])), [pendingRenames]);

  const treeData = useMemo(() => buildFolderTree(state?.items ?? [], root?.id ?? ''), [root?.id, state?.items]);
  const currentItems = useMemo(
    () =>
      (state?.items ?? [])
        .filter((item) => item.parentId === activeFolderId)
        .sort((left, right) => {
          if (left.type !== right.type) {
            return left.type === 'folder' ? -1 : 1;
          }
          return left.name.localeCompare(right.name);
        }),
    [activeFolderId, state?.items],
  );

  const columns: ColumnsType<ProjectBrowserItemModel> = [
    {
      title: t('itemName'),
      dataIndex: 'name',
      key: 'name',
      render: (name: string, item) => pendingByID.get(item.id)?.newName ?? name,
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

  useEffect(() => {
    setSelectedFolderId(root?.id ?? null);
    setSelectedItem(null);
  }, [root?.id]);

  return (
    <Drawer title={t('modifyProject')} open={open} onClose={onClose} width={1120}>
      {state ? (
        <Flex vertical gap={18}>
          <Descriptions column={4} bordered size="small">
            <Descriptions.Item label={t('projectName')}>{state.projectName}</Descriptions.Item>
            <Descriptions.Item label={t('projectId')}>{state.projectId}</Descriptions.Item>
            <Descriptions.Item label={t('fileCount')}>{state.files}</Descriptions.Item>
            <Descriptions.Item label={t('folderCount')}>{state.folders}</Descriptions.Item>
            <Descriptions.Item label={t('partCount')}>{state.parts}</Descriptions.Item>
            <Descriptions.Item label={t('createdTime')}>{state.createdAt}</Descriptions.Item>
            <Descriptions.Item label={t('updatedTime')}>{state.updatedAt}</Descriptions.Item>
            <Descriptions.Item label={t('contentConnected')}>
              {state.contentConnected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
            </Descriptions.Item>
          </Descriptions>
          <div className="project-browser-grid">
            <div className="project-browser-tree">
              <Typography.Title level={5}>{t('folderTree')}</Typography.Title>
              <Tree
                treeData={treeData}
                selectedKeys={activeFolderId ? [activeFolderId] : []}
                defaultExpandAll
                onSelect={(keys) => setSelectedFolderId((keys[0] as string | undefined) ?? root?.id ?? null)}
              />
            </div>
            <div className="project-browser-items">
              <Flex justify="space-between" align="center" gap={12}>
                <Typography.Title level={5}>{t('currentFolderItems')}</Typography.Title>
                <Space>
                  <Button
                    onClick={() => setRenameOpen(true)}
                    disabled={!selectedItem || selectedItem.id === state.rootFolderId}
                  >
                    {t('renameItem')}
                  </Button>
                  <Button onClick={onDiscardAll} disabled={pendingRenames.length === 0}>
                    {t('discardChanges')}
                  </Button>
                  <Button type="primary" loading={applyLoading} disabled={pendingRenames.length === 0} onClick={onApply}>
                    {t('applyChanges')}
                  </Button>
                </Space>
              </Flex>
              <Table
                rowKey="id"
                columns={columns}
                dataSource={currentItems}
                pagination={false}
                size="small"
                scroll={{ x: 720 }}
                rowSelection={{
                  type: 'radio',
                  selectedRowKeys: selectedItem ? [selectedItem.id] : [],
                  onChange: (_, rows) => setSelectedItem(rows[0] ?? null),
                }}
                onRow={(item) => ({
                  onClick: () => setSelectedItem(item),
                  onDoubleClick: () => {
                    if (item.id !== state.rootFolderId) {
                      setSelectedItem(item);
                      setRenameOpen(true);
                    }
                  },
                })}
              />
            </div>
          </div>
          <div>
            <Typography.Title level={5}>{t('pendingChanges')}</Typography.Title>
            <List
              size="small"
              bordered
              dataSource={pendingRenames}
              locale={{ emptyText: t('noPendingChanges') }}
              renderItem={(rename) => (
                <List.Item actions={[<Button onClick={() => onDiscardRename(rename.itemId)}>{t('discard')}</Button>]}>
                  <Typography.Text>
                    {t('pendingRename')}: {rename.itemPath} -&gt; {rename.newName}
                  </Typography.Text>
                </List.Item>
              )}
            />
          </div>
          <RenameItemModal
            open={renameOpen}
            item={selectedItem}
            onCancel={() => setRenameOpen(false)}
            onSubmit={(newName) => {
              if (selectedItem) {
                onRename({
                  itemId: selectedItem.id,
                  itemPath: selectedItem.path,
                  oldName: selectedItem.name,
                  newName,
                });
              }
              setRenameOpen(false);
            }}
            t={t}
          />
        </Flex>
      ) : null}
    </Drawer>
  );
}

function buildFolderTree(items: ProjectBrowserItemModel[], rootID: string): DataNode[] {
  const folders = items.filter((item) => item.type === 'folder');
  const childrenByParent = new Map<string, ProjectBrowserItemModel[]>();
  for (const folder of folders) {
    const children = childrenByParent.get(folder.parentId) ?? [];
    children.push(folder);
    childrenByParent.set(folder.parentId, children);
  }

  const buildNode = (folder: ProjectBrowserItemModel): DataNode => ({
    key: folder.id,
    title: folder.name,
    children: (childrenByParent.get(folder.id) ?? []).sort((left, right) => left.name.localeCompare(right.name)).map(buildNode),
  });

  const root = folders.find((folder) => folder.id === rootID);
  return root ? [buildNode(root)] : [];
}
