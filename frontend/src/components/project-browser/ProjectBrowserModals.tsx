import type { TreeSelectProps } from 'antd';
import type { ProjectBrowserItemModel } from '../../types';
import type { PendingAdd, PendingCreateFolder, PendingMove, PendingRename } from '../../hooks/useProjectBrowser';
import { AddItemModal } from './AddItemModal';
import { ApplyChangesModal } from './ApplyChangesModal';
import { CreateFolderModal } from './CreateFolderModal';
import { MoveItemModal } from './MoveItemModal';
import { RenameItemModal } from './RenameItemModal';
import { partSizeMBToOverrideBytes } from '../../partSize';

type ProjectBrowserModalsProps = {
  addOpen: boolean;
  applyConfirmOpen: boolean;
  applyLoading: boolean;
  contentConnected: boolean;
  createFolderOpen: boolean;
  moveOpen: boolean;
  pendingAddCount: number;
  pendingCreateFolderCount: number;
  pendingMoveCount: number;
  pendingRemoveCount: number;
  pendingRenameCount: number;
  renameOpen: boolean;
  blockingConflicts: string[];
  warnings: string[];
  willWriteOperationGuide: boolean;
  itemsByID: Map<string, ProjectBrowserItemModel>;
  selectedItem: ProjectBrowserItemModel | null;
  selectedItems: ProjectBrowserItemModel[];
  selectableFolderTreeData: TreeSelectProps['treeData'];
  targetFolder: ProjectBrowserItemModel | null;
  onAdd: (add: PendingAdd) => void;
  onApply: () => void;
  onCloseAdd: () => void;
  onCloseApplyConfirm: () => void;
  onCloseCreateFolder: () => void;
  onCloseMove: () => void;
  onCloseRename: () => void;
  onCreateFolder: (createFolder: PendingCreateFolder) => void;
  onMove: (move: PendingMove) => void;
  onRename: (rename: PendingRename) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function ProjectBrowserModals({
  addOpen,
  applyConfirmOpen,
  applyLoading,
  contentConnected,
  createFolderOpen,
  moveOpen,
  pendingAddCount,
  pendingCreateFolderCount,
  pendingMoveCount,
  pendingRemoveCount,
  pendingRenameCount,
  renameOpen,
  blockingConflicts,
  warnings,
  willWriteOperationGuide,
  itemsByID,
  selectedItem,
  selectedItems,
  selectableFolderTreeData,
  targetFolder,
  onAdd,
  onApply,
  onCloseAdd,
  onCloseApplyConfirm,
  onCloseCreateFolder,
  onCloseMove,
  onCloseRename,
  onCreateFolder,
  onMove,
  onRename,
  t,
}: ProjectBrowserModalsProps) {
  return (
    <>
      <AddItemModal
        open={addOpen}
        loading={applyLoading}
        onCancel={onCloseAdd}
        onSubmit={(values) => {
          if (targetFolder) {
            onAdd({
              itemId: crypto.randomUUID(),
              sourcePath: values.sourcePath,
              targetFolderPath: targetFolder.path,
              maxPartSize: values.maxPartSize === undefined ? 0 : partSizeMBToOverrideBytes(values.maxPartSize),
            });
          }
          onCloseAdd();
        }}
        t={t}
      />
      <RenameItemModal
        open={renameOpen}
        item={selectedItem}
        onCancel={onCloseRename}
        onSubmit={(newName) => {
          if (selectedItem) {
            onRename({
              itemId: selectedItem.id,
              itemPath: selectedItem.path,
              oldName: selectedItem.name,
              newName,
            });
          }
          onCloseRename();
        }}
        t={t}
      />
      <CreateFolderModal
        open={createFolderOpen}
        loading={applyLoading}
        onCancel={onCloseCreateFolder}
        onSubmit={(values) => {
          if (targetFolder) {
            onCreateFolder({
              itemId: crypto.randomUUID(),
              targetFolderPath: targetFolder.path,
              name: values.name,
            });
          }
          onCloseCreateFolder();
        }}
        t={t}
      />
      <MoveItemModal
        open={moveOpen}
        items={selectedItems}
        treeData={selectableFolderTreeData}
        onCancel={onCloseMove}
        onSubmit={(targetFolderId) => {
          const moveTarget = itemsByID.get(targetFolderId);
          if (moveTarget) {
            for (const item of selectedItems) {
              onMove({
                itemId: item.id,
                itemPath: item.path,
                targetFolderPath: moveTarget.path,
              });
            }
          }
          onCloseMove();
        }}
        t={t}
      />
      <ApplyChangesModal
        open={applyConfirmOpen}
        loading={applyLoading}
        renameCount={pendingRenameCount}
        moveCount={pendingMoveCount}
        removeCount={pendingRemoveCount}
        addCount={pendingAddCount}
        createFolderCount={pendingCreateFolderCount}
        contentConnected={contentConnected}
        willWriteOperationGuide={willWriteOperationGuide}
        blockingConflicts={blockingConflicts}
        warnings={warnings}
        onCancel={onCloseApplyConfirm}
        onConfirm={() => {
          onCloseApplyConfirm();
          onApply();
        }}
        t={t}
      />
    </>
  );
}
