import { OpenProjectModal } from './OpenProjectModal';
import { ProjectBrowserDrawer } from './ProjectBrowserDrawer';
import type { ProjectBrowserStateModel } from '../../types';

type ProjectBrowserLayerProps = {
  openProjectDialogOpen: boolean;
  browserLoading: boolean;
  browserOpen: boolean;
  browserState: ProjectBrowserStateModel | null;
  onCloseOpenProject: () => void;
  onOpenProject: (values: { password: string; encryptedPath: string }) => void;
  onCloseBrowser: () => void;
  t: (key: string) => string;
};

export function ProjectBrowserLayer({
  openProjectDialogOpen,
  browserLoading,
  browserOpen,
  browserState,
  onCloseOpenProject,
  onOpenProject,
  onCloseBrowser,
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
      <ProjectBrowserDrawer open={browserOpen} state={browserState} onClose={onCloseBrowser} t={t} />
    </>
  );
}
