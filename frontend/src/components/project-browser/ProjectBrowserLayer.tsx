import { OpenProjectModal } from './OpenProjectModal';
import { ProjectBrowserDrawer } from './ProjectBrowserDrawer';
import type { ProjectBrowserStateModel } from '../../types';
import type { PendingMove, PendingRemove, PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserLayerProps = {
  openProjectDialogOpen: boolean;
  browserLoading: boolean;
  applyLoading: boolean;
  browserOpen: boolean;
  browserState: ProjectBrowserStateModel | null;
  pendingRenames: PendingRename[];
  pendingMoves: PendingMove[];
  pendingRemoves: PendingRemove[];
  onCloseOpenProject: () => void;
  onOpenProject: (values: { password: string; encryptedPath: string }) => void;
  onCloseBrowser: () => void;
  onRename: (rename: PendingRename) => void;
  onMove: (move: PendingMove) => void;
  onRemove: (remove: PendingRemove) => void;
  onDiscardRename: (itemId: string) => void;
  onDiscardMove: (itemId: string) => void;
  onDiscardRemove: (itemId: string) => void;
  onDiscardAll: () => void;
  onApply: () => void;
  t: (key: string) => string;
};

export function ProjectBrowserLayer({
  openProjectDialogOpen,
  browserLoading,
  applyLoading,
  browserOpen,
  browserState,
  pendingRenames,
  pendingMoves,
  pendingRemoves,
  onCloseOpenProject,
  onOpenProject,
  onCloseBrowser,
  onRename,
  onMove,
  onRemove,
  onDiscardRename,
  onDiscardMove,
  onDiscardRemove,
  onDiscardAll,
  onApply,
  t,
}: ProjectBrowserLayerProps) {
  return (
    <>
      <OpenProjectModal
        open={openProjectDialogOpen}
        loading={browserLoading}
        onCancel={onCloseOpenProject}
        onSubmit={onOpenProject}
        t={t}
      />
      <ProjectBrowserDrawer
        open={browserOpen}
        state={browserState}
        pendingRenames={pendingRenames}
        pendingMoves={pendingMoves}
        pendingRemoves={pendingRemoves}
        applyLoading={applyLoading}
        onClose={onCloseBrowser}
        onRename={onRename}
        onMove={onMove}
        onRemove={onRemove}
        onDiscardRename={onDiscardRename}
        onDiscardMove={onDiscardMove}
        onDiscardRemove={onDiscardRemove}
        onDiscardAll={onDiscardAll}
        onApply={onApply}
        t={t}
      />
    </>
  );
}
