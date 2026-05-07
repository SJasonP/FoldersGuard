import { OpenProjectModal } from './OpenProjectModal';
import { ProjectBrowserDrawer } from './ProjectBrowserDrawer';
import type { ProjectBrowserStateModel } from '../../types';
import type { PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserLayerProps = {
  openProjectDialogOpen: boolean;
  browserLoading: boolean;
  applyLoading: boolean;
  browserOpen: boolean;
  browserState: ProjectBrowserStateModel | null;
  pendingRenames: PendingRename[];
  onCloseOpenProject: () => void;
  onOpenProject: (values: { password: string; encryptedPath: string }) => void;
  onCloseBrowser: () => void;
  onRename: (rename: PendingRename) => void;
  onDiscardRename: (itemId: string) => void;
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
  onCloseOpenProject,
  onOpenProject,
  onCloseBrowser,
  onRename,
  onDiscardRename,
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
        applyLoading={applyLoading}
        onClose={onCloseBrowser}
        onRename={onRename}
        onDiscardRename={onDiscardRename}
        onDiscardAll={onDiscardAll}
        onApply={onApply}
        t={t}
      />
    </>
  );
}
