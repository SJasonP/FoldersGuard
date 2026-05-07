import { ApplyChangesResultDrawer } from './ApplyChangesResultDrawer';
import { OpenProjectModal } from './OpenProjectModal';
import { ProjectBrowserDrawer } from './ProjectBrowserDrawer';
import type { ApplyProjectChangesResultModel, ProjectBrowserStateModel } from '../../types';
import type { PendingAdd, PendingMove, PendingRemove, PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserLayerProps = {
  openProjectDialogOpen: boolean;
  browserLoading: boolean;
  applyLoading: boolean;
  browserOpen: boolean;
  browserState: ProjectBrowserStateModel | null;
  applyResult: ApplyProjectChangesResultModel | null;
  applyResultOpen: boolean;
  pendingRenames: PendingRename[];
  pendingMoves: PendingMove[];
  pendingRemoves: PendingRemove[];
  pendingAdds: PendingAdd[];
  onCloseOpenProject: () => void;
  onOpenProject: (values: { password: string; encryptedPath: string }) => void;
  onCloseBrowser: () => void;
  onCloseApplyResult: () => void;
  onAdd: (add: PendingAdd) => void;
  onRename: (rename: PendingRename) => void;
  onMove: (move: PendingMove) => void;
  onRemove: (remove: PendingRemove) => void;
  onDiscardRename: (itemId: string) => void;
  onDiscardMove: (itemId: string) => void;
  onDiscardRemove: (itemId: string) => void;
  onDiscardAdd: (itemId: string) => void;
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
  applyResult,
  applyResultOpen,
  pendingRenames,
  pendingMoves,
  pendingRemoves,
  pendingAdds,
  onCloseOpenProject,
  onOpenProject,
  onCloseBrowser,
  onCloseApplyResult,
  onAdd,
  onRename,
  onMove,
  onRemove,
  onDiscardRename,
  onDiscardMove,
  onDiscardRemove,
  onDiscardAdd,
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
        pendingAdds={pendingAdds}
        applyLoading={applyLoading}
        onClose={onCloseBrowser}
        onAdd={onAdd}
        onRename={onRename}
        onMove={onMove}
        onRemove={onRemove}
        onDiscardRename={onDiscardRename}
        onDiscardMove={onDiscardMove}
        onDiscardRemove={onDiscardRemove}
        onDiscardAdd={onDiscardAdd}
        onDiscardAll={onDiscardAll}
        onApply={onApply}
        t={t}
      />
      <ApplyChangesResultDrawer open={applyResultOpen} result={applyResult} onClose={onCloseApplyResult} t={t} />
    </>
  );
}
