import { useEffect, useMemo, useState } from 'react';
import { Descriptions, Drawer, Flex, Table, Tree, Typography } from 'antd';
import type { DataNode } from 'antd/es/tree';
import type { ColumnsType } from 'antd/es/table';
import type { ProjectBrowserItemModel, ProjectBrowserStateModel } from '../../types';

type ProjectBrowserDrawerProps = {
  open: boolean;
  state: ProjectBrowserStateModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function ProjectBrowserDrawer({ open, state, onClose, t }: ProjectBrowserDrawerProps) {
  const root = state?.items.find((item) => item.id === state.rootFolderId) ?? null;
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const activeFolderId = selectedFolderId ?? root?.id ?? '';

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
    { title: t('itemName'), dataIndex: 'name', key: 'name' },
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
  ];

  useEffect(() => {
    setSelectedFolderId(root?.id ?? null);
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
              <Typography.Title level={5}>{t('currentFolderItems')}</Typography.Title>
              <Table
                rowKey="id"
                columns={columns}
                dataSource={currentItems}
                pagination={false}
                size="small"
                scroll={{ x: 720 }}
              />
            </div>
          </div>
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
