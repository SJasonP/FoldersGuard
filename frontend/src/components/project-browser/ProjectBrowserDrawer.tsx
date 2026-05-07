import { useEffect, useMemo, useState } from 'react';
import { Breadcrumb, Descriptions, Drawer, Flex, Tree, Typography } from 'antd';
import type { ProjectBrowserItemModel, ProjectBrowserStateModel } from '../../types';
import { ProjectBrowserDetailsPanel } from './ProjectBrowserDetailsPanel';
import { ProjectBrowserItemTable } from './ProjectBrowserItemTable';
import { ProjectBrowserPendingChanges } from './ProjectBrowserPendingChanges';
import { RenameItemModal } from './RenameItemModal';
import type { PendingRename } from '../../hooks/useProjectBrowser';
import { buildFolderTree, filteredFolderItems, folderBreadcrumbItems, pendingRenameMap } from './projectBrowserView';

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
  const [searchQuery, setSearchQuery] = useState('');
  const activeFolderId = selectedFolderId ?? root?.id ?? '';
  const pendingByID = useMemo(() => pendingRenameMap(pendingRenames), [pendingRenames]);

  const treeData = useMemo(() => buildFolderTree(state?.items ?? [], root?.id ?? '', pendingByID), [pendingByID, root?.id, state?.items]);
  const breadcrumbs = useMemo(
    () => folderBreadcrumbItems(state?.items ?? [], activeFolderId, pendingByID),
    [activeFolderId, pendingByID, state?.items],
  );
  const currentItems = useMemo(
    () => filteredFolderItems(state?.items ?? [], activeFolderId, searchQuery, pendingByID),
    [activeFolderId, pendingByID, searchQuery, state?.items],
  );
  const selectFolder = (folderID: string) => {
    setSelectedFolderId(folderID);
    setSelectedItem(null);
    setSearchQuery('');
  };

  useEffect(() => {
    setSelectedFolderId(root?.id ?? null);
    setSelectedItem(null);
    setSearchQuery('');
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
            <ProjectBrowserItemTable
              items={currentItems}
              pendingByID={pendingByID}
              selectedItem={selectedItem}
              rootFolderID={state.rootFolderId}
              searchQuery={searchQuery}
              applyLoading={applyLoading}
              pendingCount={pendingRenames.length}
              onSearchChange={setSearchQuery}
              onSelectItem={setSelectedItem}
              onOpenRename={() => setRenameOpen(true)}
              onDiscardAll={onDiscardAll}
              onApply={onApply}
              t={t}
            />
            <ProjectBrowserDetailsPanel item={selectedItem} pendingByID={pendingByID} t={t} />
          </div>
          <ProjectBrowserPendingChanges pendingRenames={pendingRenames} onDiscardRename={onDiscardRename} t={t} />
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
