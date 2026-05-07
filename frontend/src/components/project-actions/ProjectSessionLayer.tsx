import { CreateProjectModal } from './CreateProjectModal';
import { DeleteProjectModal } from './DeleteProjectModal';
import { ExportProjectModal } from './ExportProjectModal';
import { ImportProjectModal } from './ImportProjectModal';
import { InspectProjectDrawer } from './InspectProjectDrawer';
import { InspectProjectModal } from './InspectProjectModal';
import { ProjectActionsDrawer } from './ProjectActionsDrawer';
import { VerifyProjectDrawer } from './VerifyProjectDrawer';
import { VerifyProjectModal } from './VerifyProjectModal';
import type { InspectProjectResultModel, LocalProjectSummary, SettingsModel, VerifyProjectResultModel } from '../../types';

type ProjectSessionLayerProps = {
  createDialogOpen: boolean;
  createLoading: boolean;
  settings: SettingsModel | null;
  defaultSourceCleanup: string;
  onCloseCreate: () => void;
  onCreateProject: (values: {
    sourcePath: string;
    contentOutput: string;
    password: string;
    passwordConfirm: string;
    maxPartSize?: number;
    useDefaultMaxPartSize: boolean;
    force: boolean;
    sourceCleanup: string;
    databaseExport?: string;
  }) => void;
  importDialogOpen: boolean;
  importLoading: boolean;
  onCloseImport: () => void;
  onImportProject: (values: { inputPath: string; password: string; force: boolean }) => void;
  projectActionsOpen: boolean;
  selectedProject: LocalProjectSummary | null;
  onCloseProjectActions: () => void;
  onOpenInspect: () => void;
  onOpenVerify: () => void;
  onOpenExport: () => void;
  onOpenDelete: () => void;
  inspectDialogOpen: boolean;
  inspectLoading: boolean;
  onCloseInspect: () => void;
  onInspectProject: (password: string) => void;
  inspectResultOpen: boolean;
  inspectResult: InspectProjectResultModel | null;
  onCloseInspectResult: () => void;
  verifyDialogOpen: boolean;
  verifyLoading: boolean;
  onCloseVerify: () => void;
  onVerifyProject: (values: { password: string; encryptedPath: string }) => void;
  verifyResultOpen: boolean;
  verifyResult: VerifyProjectResultModel | null;
  onCloseVerifyResult: () => void;
  exportDialogOpen: boolean;
  exportLoading: boolean;
  onCloseExport: () => void;
  onExportProject: (values: { password: string; outputPath: string; force: boolean }) => void;
  deleteDialogOpen: boolean;
  deleteLoading: boolean;
  onCloseDelete: () => void;
  onDeleteProject: (password: string) => void;
  t: (key: string) => string;
};

export function ProjectSessionLayer({
  createDialogOpen,
  createLoading,
  settings,
  defaultSourceCleanup,
  onCloseCreate,
  onCreateProject,
  importDialogOpen,
  importLoading,
  onCloseImport,
  onImportProject,
  projectActionsOpen,
  selectedProject,
  onCloseProjectActions,
  onOpenInspect,
  onOpenVerify,
  onOpenExport,
  onOpenDelete,
  inspectDialogOpen,
  inspectLoading,
  onCloseInspect,
  onInspectProject,
  inspectResultOpen,
  inspectResult,
  onCloseInspectResult,
  verifyDialogOpen,
  verifyLoading,
  onCloseVerify,
  onVerifyProject,
  verifyResultOpen,
  verifyResult,
  onCloseVerifyResult,
  exportDialogOpen,
  exportLoading,
  onCloseExport,
  onExportProject,
  deleteDialogOpen,
  deleteLoading,
  onCloseDelete,
  onDeleteProject,
  t,
}: ProjectSessionLayerProps) {
  return (
    <>
      <ProjectActionsDrawer
        open={projectActionsOpen}
        project={selectedProject}
        onClose={onCloseProjectActions}
        onInspect={onOpenInspect}
        onVerify={onOpenVerify}
        onExport={onOpenExport}
        onDelete={onOpenDelete}
        t={t}
      />
      <CreateProjectModal
        open={createDialogOpen}
        loading={createLoading}
        settings={settings}
        defaultSourceCleanup={defaultSourceCleanup}
        onCancel={onCloseCreate}
        onSubmit={onCreateProject}
        t={t}
      />
      <ImportProjectModal open={importDialogOpen} loading={importLoading} onCancel={onCloseImport} onSubmit={onImportProject} t={t} />
      <InspectProjectModal open={inspectDialogOpen} loading={inspectLoading} onCancel={onCloseInspect} onSubmit={onInspectProject} t={t} />
      <InspectProjectDrawer open={inspectResultOpen} result={inspectResult} onClose={onCloseInspectResult} t={t} />
      <VerifyProjectModal open={verifyDialogOpen} loading={verifyLoading} onCancel={onCloseVerify} onSubmit={onVerifyProject} t={t} />
      <VerifyProjectDrawer open={verifyResultOpen} result={verifyResult} onClose={onCloseVerifyResult} t={t} />
      <ExportProjectModal open={exportDialogOpen} loading={exportLoading} onCancel={onCloseExport} onSubmit={onExportProject} t={t} />
      <DeleteProjectModal open={deleteDialogOpen} loading={deleteLoading} onCancel={onCloseDelete} onSubmit={onDeleteProject} t={t} />
    </>
  );
}
