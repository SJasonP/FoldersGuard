import { useEffect, useMemo, useState } from 'react';
import { Breadcrumb, Descriptions, Drawer, Flex, Modal, Tree, Typography } from 'antd';
import type { ProjectBrowserItemModel, ProjectBrowserStateModel } from '../../types';
import { ProjectBrowserCloseGuardModal } from './ProjectBrowserCloseGuardModal';
import { ProjectBrowserDetailsPanel } from './ProjectBrowserDetailsPanel';
import { ProjectBrowserItemTable } from './ProjectBrowserItemTable';
import { ProjectBrowserModals } from './ProjectBrowserModals';
import { ProjectBrowserPendingChanges } from './ProjectBrowserPendingChanges';
import type { PendingAdd, PendingCreateFolder, PendingMove, PendingRemove, PendingRename } from '../../hooks/useProjectBrowser';
import {
  buildFolderTree,
  buildSelectableFolderTree,
  descendantFolderIDs,
  filteredFolderItems,
  folderBreadcrumbItems,
  pendingRenameMap,
} from './projectBrowserView';
import { validatePendingProjectChanges } from './projectBrowserPendingValidation';

type ProjectBrowserDrawerProps = {
  open: boolean;
  state: ProjectBrowserStateModel | null;
  pendingRenames: PendingRename[];
  pendingMoves: PendingMove[];
  pendingRemoves: PendingRemove[];
  pendingAdds: PendingAdd[];
  pendingCreateFolders: PendingCreateFolder[];
  applyLoading: boolean;
  onClose: () => void;
  onAdd: (add: PendingAdd) => void;
  onCreateFolder: (createFolder: PendingCreateFolder) => void;
  onRename: (rename: PendingRename) => void;
  onMove: (move: PendingMove) => void;
  onRemove: (remove: PendingRemove) => void;
  onDiscardRename: (itemId: string) => void;
  onDiscardMove: (itemId: string) => void;
  onDiscardRemove: (itemId: string) => void;
  onDiscardAdd: (itemId: string) => void;
  onDiscardCreateFolder: (itemId: string) => void;
  onDiscardAll: () => void;
  onApply: () => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ProjectBrowserDrawer({
  open,
  state,
  pendingRenames,
  pendingMoves,
  pendingRemoves,
  pendingAdds,
  pendingCreateFolders,
  applyLoading,
  onClose,
  onAdd,
  onCreateFolder,
  onRename,
  onMove,
  onRemove,
  onDiscardRename,
  onDiscardMove,
  onDiscardRemove,
  onDiscardAdd,
  onDiscardCreateFolder,
  onDiscardAll,
  onApply,
  t,
}: ProjectBrowserDrawerProps) {
  const root = state?.items.find((item) => item.id === state.rootFolderId) ?? null;
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [selectedItem, setSelectedItem] = useState<ProjectBrowserItemModel | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [applyConfirmOpen, setApplyConfirmOpen] = useState(false);
  const [closeGuardOpen, setCloseGuardOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [moveOpen, setMoveOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const activeFolderId = selectedFolderId ?? root?.id ?? '';
  const pendingByID = useMemo(() => pendingRenameMap(pendingRenames), [pendingRenames]);
  const itemsByID = useMemo(() => new Map((state?.items ?? []).map((item) => [item.id, item])), [state?.items]);
  const pendingStateByID = useMemo(() => {
    const next = new Map<string, string>();
    for (const rename of pendingRenames) {
      next.set(rename.itemId, t('pendingRename'));
    }
    for (const move of pendingMoves) {
      next.set(move.itemId, t('pendingMove'));
    }
    for (const remove of pendingRemoves) {
      next.set(remove.itemId, t('pendingRemove'));
    }
    return next;
  }, [pendingMoves, pendingRemoves, pendingRenames, t]);

  const treeData = useMemo(() => buildFolderTree(state?.items ?? [], root?.id ?? '', pendingByID), [pendingByID, root?.id, state?.items]);
  const selectableFolderTreeData = useMemo(
    () => {
      const disabledIDs = selectedItem ? descendantFolderIDs(state?.items ?? [], selectedItem.id) : new Set<string>();
      if (selectedItem?.parentId) {
        disabledIDs.add(selectedItem.parentId);
      }
      return buildSelectableFolderTree(state?.items ?? [], root?.id ?? '', pendingByID, disabledIDs);
    },
    [pendingByID, root?.id, selectedItem, state?.items],
  );
  const breadcrumbs = useMemo(
    () => folderBreadcrumbItems(state?.items ?? [], activeFolderId, pendingByID),
    [activeFolderId, pendingByID, state?.items],
  );
  const currentItems = useMemo(
    () => filteredFolderItems(state?.items ?? [], activeFolderId, searchQuery, pendingByID),
    [activeFolderId, pendingByID, searchQuery, state?.items],
  );
  const pendingCount = pendingRenames.length + pendingMoves.length + pendingRemoves.length + pendingAdds.length + pendingCreateFolders.length;
  const pendingValidation = useMemo(
    () =>
      state
        ? validatePendingProjectChanges({
            state,
            pendingRenames,
            pendingMoves,
            pendingRemoves,
            pendingAdds,
            pendingCreateFolders,
            t,
          })
        : { blockingConflicts: [], warnings: [] },
    [pendingAdds, pendingCreateFolders, pendingMoves, pendingRemoves, pendingRenames, state, t],
  );
  const selectFolder = (folderID: string) => {
    setSelectedFolderId(folderID);
    setSelectedItem(null);
    setSearchQuery('');
  };
  const closeOrConfirm = () => {
    if (pendingCount === 0) {
      onClose();
      return;
    }
    setCloseGuardOpen(true);
  };
  const discardAndClose = () => {
    onDiscardAll();
    setCloseGuardOpen(false);
    onClose();
  };
  const confirmApplyBeforeClose = () => {
    setCloseGuardOpen(false);
    setApplyConfirmOpen(true);
  };

  useEffect(() => {
    setSelectedFolderId(root?.id ?? null);
    setSelectedItem(null);
    setSearchQuery('');
  }, [root?.id]);

  return (
    <Drawer title={t('modifyProject')} open={open} onClose={closeOrConfirm} width={1120}>
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
              pendingStateByID={pendingStateByID}
              selectedItem={selectedItem}
              rootFolderID={state.rootFolderId}
              searchQuery={searchQuery}
              applyLoading={applyLoading}
              pendingCount={pendingCount}
              applyBlocked={pendingValidation.blockingConflicts.length > 0}
              onSearchChange={setSearchQuery}
              onSelectItem={setSelectedItem}
              onOpenAdd={() => setAddOpen(true)}
              onOpenCreateFolder={() => setCreateFolderOpen(true)}
              onOpenRename={() => setRenameOpen(true)}
              onOpenMove={() => setMoveOpen(true)}
              onRemove={() => {
                if (!selectedItem) {
                  return;
                }
                Modal.confirm({
                  title: t('removeItem'),
                  content: selectedItem.path,
                  okText: t('removeItem'),
                  okButtonProps: { danger: true },
                  onOk: () => onRemove({ itemId: selectedItem.id, itemPath: selectedItem.path }),
                });
              }}
              onDiscardAll={onDiscardAll}
              onApply={() => setApplyConfirmOpen(true)}
              t={t}
            />
            <ProjectBrowserDetailsPanel item={selectedItem} pendingByID={pendingByID} pendingStateByID={pendingStateByID} t={t} />
          </div>
          <ProjectBrowserPendingChanges
            pendingRenames={pendingRenames}
            pendingMoves={pendingMoves}
            pendingRemoves={pendingRemoves}
            pendingAdds={pendingAdds}
            pendingCreateFolders={pendingCreateFolders}
            blockingConflicts={pendingValidation.blockingConflicts}
            warnings={pendingValidation.warnings}
            onDiscardRename={onDiscardRename}
            onDiscardMove={onDiscardMove}
            onDiscardRemove={onDiscardRemove}
            onDiscardAdd={onDiscardAdd}
            onDiscardCreateFolder={onDiscardCreateFolder}
            t={t}
          />
          <ProjectBrowserModals
            addOpen={addOpen}
            applyConfirmOpen={applyConfirmOpen}
            applyLoading={applyLoading}
            contentConnected={state.contentConnected}
            createFolderOpen={createFolderOpen}
            moveOpen={moveOpen}
            pendingAddCount={pendingAdds.length}
            pendingCreateFolderCount={pendingCreateFolders.length}
            pendingMoveCount={pendingMoves.length}
            pendingRemoveCount={pendingRemoves.length}
            pendingRenameCount={pendingRenames.length}
            renameOpen={renameOpen}
            blockingConflicts={pendingValidation.blockingConflicts}
            warnings={pendingValidation.warnings}
            itemsByID={itemsByID}
            selectedItem={selectedItem}
            selectableFolderTreeData={selectableFolderTreeData}
            targetFolder={itemsByID.get(activeFolderId) ?? null}
            onAdd={onAdd}
            onApply={onApply}
            onCloseCreateFolder={() => setCreateFolderOpen(false)}
            onCloseAdd={() => setAddOpen(false)}
            onCloseApplyConfirm={() => setApplyConfirmOpen(false)}
            onCloseMove={() => setMoveOpen(false)}
            onCloseRename={() => setRenameOpen(false)}
            onCreateFolder={onCreateFolder}
            onMove={onMove}
            onRename={onRename}
            t={t}
          />
          <ProjectBrowserCloseGuardModal
            open={closeGuardOpen}
            applyBlocked={pendingValidation.blockingConflicts.length > 0}
            applyLoading={applyLoading}
            onApply={confirmApplyBeforeClose}
            onCancel={() => setCloseGuardOpen(false)}
            onDiscard={discardAndClose}
            t={t}
          />
        </Flex>
      ) : null}
    </Drawer>
  );
}
