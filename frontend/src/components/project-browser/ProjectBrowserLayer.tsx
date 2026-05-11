import {ApplyChangesResultDrawer} from './ApplyChangesResultDrawer';
import {OpenProjectModal} from './OpenProjectModal';
import {ProjectBrowserDrawer} from './ProjectBrowserDrawer';
import type {ApplyProjectChangesResultModel, ProjectBrowserStateModel} from '../../types';
import type {
    PendingAdd,
    PendingCreateFolder,
    PendingMove,
    PendingRemove,
    PendingRename
} from '../../hooks/useProjectBrowser';

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
    pendingCreateFolders: PendingCreateFolder[];
    onCloseOpenProject: () => void;
    onOpenProject: (values: { password: string; encryptedPath: string }) => void;
    onCloseBrowser: () => void;
    onCloseApplyResult: () => void;
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
                                        pendingCreateFolders,
                                        onCloseOpenProject,
                                        onOpenProject,
                                        onCloseBrowser,
                                        onCloseApplyResult,
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
                pendingCreateFolders={pendingCreateFolders}
                applyLoading={applyLoading}
                onClose={onCloseBrowser}
                onAdd={onAdd}
                onCreateFolder={onCreateFolder}
                onRename={onRename}
                onMove={onMove}
                onRemove={onRemove}
                onDiscardRename={onDiscardRename}
                onDiscardMove={onDiscardMove}
                onDiscardRemove={onDiscardRemove}
                onDiscardAdd={onDiscardAdd}
                onDiscardCreateFolder={onDiscardCreateFolder}
                onDiscardAll={onDiscardAll}
                onApply={onApply}
                t={t}
            />
            <ApplyChangesResultDrawer open={applyResultOpen} result={applyResult} onClose={onCloseApplyResult} t={t}/>
        </>
    );
}
