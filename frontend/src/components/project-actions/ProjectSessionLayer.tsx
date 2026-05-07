import { CreateProjectModal } from './CreateProjectModal';
import { CreateShareModal } from './CreateShareModal';
import { CreateSharePasswordModal } from './CreateSharePasswordModal';
import { CreateShareResultDrawer } from './CreateShareResultDrawer';
import { DecryptProjectDrawer } from './DecryptProjectDrawer';
import { DecryptProjectModal } from './DecryptProjectModal';
import { DeleteProjectModal } from './DeleteProjectModal';
import { ExportProjectModal } from './ExportProjectModal';
import { ImportProjectModal } from './ImportProjectModal';
import { InspectProjectDrawer } from './InspectProjectDrawer';
import { InspectProjectModal } from './InspectProjectModal';
import { ProjectActionsDrawer } from './ProjectActionsDrawer';
import { VerifyProjectDrawer } from './VerifyProjectDrawer';
import { VerifyProjectModal } from './VerifyProjectModal';
import type {
  CreateShareResultModel,
  DecryptProjectResultModel,
  InspectProjectResultModel,
  LocalProjectSummary,
  SettingsModel,
  VerifyProjectResultModel,
} from '../../types';

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
  onOpenDecrypt: () => void;
  onOpenCreateShare: () => void;
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
  decryptDialogOpen: boolean;
  decryptLoading: boolean;
  onCloseDecrypt: () => void;
  onDecryptProject: (values: {
    password: string;
    encryptedPath: string;
    outputPath: string;
    force: boolean;
    sourceCleanup: string;
  }) => void;
  decryptResultOpen: boolean;
  decryptResult: DecryptProjectResultModel | null;
  onCloseDecryptResult: () => void;
  createSharePasswordDialogOpen: boolean;
  createShareDialogOpen: boolean;
  createShareLoading: boolean;
  selectableShareItems: Array<{ value: string; label: string }>;
  createShareResultOpen: boolean;
  createShareResult: CreateShareResultModel | null;
  onCloseCreateSharePassword: () => void;
  onLoadShareableItems: (password: string) => void;
  onCloseCreateShare: () => void;
  onCreateShare: (values: {
    itemPaths: string[];
    outputPath: string;
    force: boolean;
    passwordProtected: boolean;
    sharePassword?: string;
    sharePasswordConfirm?: string;
  }) => void;
  onCloseCreateShareResult: () => void;
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
  onOpenDecrypt,
  onOpenCreateShare,
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
  decryptDialogOpen,
  decryptLoading,
  onCloseDecrypt,
  onDecryptProject,
  decryptResultOpen,
  decryptResult,
  onCloseDecryptResult,
  createSharePasswordDialogOpen,
  createShareDialogOpen,
  createShareLoading,
  selectableShareItems,
  createShareResultOpen,
  createShareResult,
  onCloseCreateSharePassword,
  onLoadShareableItems,
  onCloseCreateShare,
  onCreateShare,
  onCloseCreateShareResult,
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
        onDecrypt={onOpenDecrypt}
        onCreateShare={onOpenCreateShare}
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
      <DecryptProjectModal
        open={decryptDialogOpen}
        loading={decryptLoading}
        defaultSourceCleanup={defaultSourceCleanup}
        onCancel={onCloseDecrypt}
        onSubmit={onDecryptProject}
        t={t}
      />
      <DecryptProjectDrawer open={decryptResultOpen} result={decryptResult} onClose={onCloseDecryptResult} t={t} />
      <CreateSharePasswordModal
        open={createSharePasswordDialogOpen}
        loading={createShareLoading}
        onCancel={onCloseCreateSharePassword}
        onSubmit={onLoadShareableItems}
        t={t}
      />
      <CreateShareModal
        open={createShareDialogOpen}
        loading={createShareLoading}
        selectableItems={selectableShareItems}
        onCancel={onCloseCreateShare}
        onSubmit={onCreateShare}
        t={t}
      />
      <CreateShareResultDrawer open={createShareResultOpen} result={createShareResult} onClose={onCloseCreateShareResult} t={t} />
      <ExportProjectModal open={exportDialogOpen} loading={exportLoading} onCancel={onCloseExport} onSubmit={onExportProject} t={t} />
      <DeleteProjectModal open={deleteDialogOpen} loading={deleteLoading} onCancel={onCloseDelete} onSubmit={onDeleteProject} t={t} />
    </>
  );
}
