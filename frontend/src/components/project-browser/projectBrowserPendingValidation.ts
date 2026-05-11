import type {ProjectBrowserStateModel} from '../../types';
import type {
    PendingAdd,
    PendingCreateFolder,
    PendingMove,
    PendingRemove,
    PendingRename
} from '../../hooks/useProjectBrowser';
import {validateProjectItemName} from './projectBrowserNameValidation';

type PendingValidationArgs = {
    state: ProjectBrowserStateModel;
    pendingRenames: PendingRename[];
    pendingMoves: PendingMove[];
    pendingRemoves: PendingRemove[];
    pendingAdds: PendingAdd[];
    pendingCreateFolders: PendingCreateFolder[];
    t: (key: string, values?: Record<string, string | number>) => string;
};

type PendingValidationResult = {
    blockingConflicts: string[];
    warnings: string[];
};

type ExistingNode = {
    id: string;
    parentId: string;
    name: string;
    path: string;
    type: string;
};

export function validatePendingProjectChanges({
                                                  state,
                                                  pendingRenames,
                                                  pendingMoves,
                                                  pendingRemoves,
                                                  pendingAdds,
                                                  pendingCreateFolders,
                                                  t,
                                              }: PendingValidationArgs): PendingValidationResult {
    const nodes = state.items.map<ExistingNode>((item) => ({
        id: item.id,
        parentId: item.parentId,
        name: item.name,
        path: item.path,
        type: item.type,
    }));
    const nodesByID = new Map(nodes.map((node) => [node.id, node]));
    const nodesByPath = new Map(nodes.map((node) => [node.path, node]));
    const parentByID = new Map(nodes.map((node) => [node.id, node.parentId]));
    const renameByID = new Map(pendingRenames.map((rename) => [rename.itemId, rename]));
    const moveByID = new Map(pendingMoves.map((move) => [move.itemId, move]));
    const removeByID = new Map(pendingRemoves.map((remove) => [remove.itemId, remove]));
    const renameIDs = new Set(pendingRenames.map((rename) => rename.itemId));
    const moveIDs = new Set(pendingMoves.map((move) => move.itemId));
    const removeIDs = new Set(pendingRemoves.map((remove) => remove.itemId));
    const blockingConflicts = new Set<string>();
    const warnings = new Set<string>();

    const ancestorIDsOf = (itemId: string) => {
        const ancestors: string[] = [];
        for (let current = parentByID.get(itemId); current; current = parentByID.get(current) ?? '') {
            ancestors.push(current);
        }
        return ancestors;
    };

    const hasAncestorStructuralChange = (itemId: string) =>
        ancestorIDsOf(itemId).some((ancestorId) => renameIDs.has(ancestorId) || moveIDs.has(ancestorId) || removeIDs.has(ancestorId));

    const hasTargetPathConflict = (targetNode: ExistingNode | undefined) => {
        if (!targetNode) {
            return true;
        }
        if (removeIDs.has(targetNode.id) || renameIDs.has(targetNode.id) || moveIDs.has(targetNode.id)) {
            return true;
        }
        return ancestorIDsOf(targetNode.id).some((ancestorId) => renameIDs.has(ancestorId) || moveIDs.has(ancestorId) || removeIDs.has(ancestorId));
    };

    for (const rename of pendingRenames) {
        if (!validateProjectItemName(rename.newName)) {
            blockingConflicts.add(t('conflictInvalidItemName', {name: rename.newName}));
        }
        if (hasAncestorStructuralChange(rename.itemId)) {
            blockingConflicts.add(t('conflictRenameAncestorChanged', {path: rename.itemPath}));
        }
    }

    for (const move of pendingMoves) {
        const targetNode = nodesByPath.get(move.targetFolderPath);
        if (renameByID.has(move.itemId)) {
            blockingConflicts.add(t('conflictMoveItemRenamed', {path: move.itemPath}));
        }
        if (hasAncestorStructuralChange(move.itemId)) {
            blockingConflicts.add(t('conflictMoveAncestorChanged', {path: move.itemPath}));
        }
        if (hasTargetPathConflict(targetNode)) {
            blockingConflicts.add(t('conflictMoveTargetChanged', {path: move.targetFolderPath}));
        }
    }

    for (const remove of pendingRemoves) {
        if (hasAncestorStructuralChange(remove.itemId)) {
            blockingConflicts.add(t('conflictRemoveAncestorChanged', {path: remove.itemPath}));
        }
    }

    for (const add of pendingAdds) {
        const targetNode = nodesByPath.get(add.targetFolderPath);
        if (hasTargetPathConflict(targetNode)) {
            blockingConflicts.add(t('conflictAddTargetChanged', {path: add.targetFolderPath}));
        }
    }

    for (const createFolder of pendingCreateFolders) {
        if (!validateProjectItemName(createFolder.name)) {
            blockingConflicts.add(t('conflictInvalidItemName', {name: createFolder.name}));
        }
        const targetNode = nodesByPath.get(createFolder.targetFolderPath);
        if (hasTargetPathConflict(targetNode)) {
            blockingConflicts.add(t('conflictCreateFolderTargetChanged', {path: createFolder.targetFolderPath}));
        }
    }

    const effectiveParentByID = new Map(nodes.map((node) => [node.id, node.parentId]));
    for (const move of pendingMoves) {
        const targetNode = nodesByPath.get(move.targetFolderPath);
        if (targetNode) {
            effectiveParentByID.set(move.itemId, targetNode.id);
        }
    }

    const childrenByParent = new Map<string, string[]>();
    for (const [itemId, parentId] of effectiveParentByID.entries()) {
        const children = childrenByParent.get(parentId) ?? [];
        children.push(itemId);
        childrenByParent.set(parentId, children);
    }

    const visiting = new Set<string>();
    const visited = new Set<string>();
    const visit = (itemId: string): boolean => {
        if (visiting.has(itemId)) {
            return true;
        }
        if (visited.has(itemId)) {
            return false;
        }
        visiting.add(itemId);
        const parentId = effectiveParentByID.get(itemId);
        if (parentId && visit(parentId)) {
            return true;
        }
        visiting.delete(itemId);
        visited.add(itemId);
        return false;
    };
    for (const node of nodes) {
        if (visit(node.id)) {
            blockingConflicts.add(t('conflictMoveCycle'));
            break;
        }
    }

    const removedIDs = new Set<string>();
    const markRemoved = (itemId: string) => {
        if (removedIDs.has(itemId)) {
            return;
        }
        removedIDs.add(itemId);
        for (const childId of childrenByParent.get(itemId) ?? []) {
            markRemoved(childId);
        }
    };
    for (const remove of pendingRemoves) {
        markRemoved(remove.itemId);
    }

    const nameGroups = new Map<string, string[]>();
    for (const node of nodes) {
        if (removedIDs.has(node.id)) {
            continue;
        }
        const effectiveParent = effectiveParentByID.get(node.id) ?? '';
        const effectiveName = renameByID.get(node.id)?.newName ?? node.name;
        const key = `${effectiveParent}\x00${effectiveName}`;
        const members = nameGroups.get(key) ?? [];
        members.push(node.path);
        nameGroups.set(key, members);
    }
    for (const add of pendingAdds) {
        const targetNode = nodesByPath.get(add.targetFolderPath);
        if (!targetNode) {
            continue;
        }
        const key = `${targetNode.id}\x00${basename(add.sourcePath)}`;
        const members = nameGroups.get(key) ?? [];
        members.push(add.sourcePath);
        nameGroups.set(key, members);
    }
    for (const createFolder of pendingCreateFolders) {
        const targetNode = nodesByPath.get(createFolder.targetFolderPath);
        if (!targetNode) {
            continue;
        }
        const key = `${targetNode.id}\x00${createFolder.name}`;
        const members = nameGroups.get(key) ?? [];
        members.push(`${createFolder.targetFolderPath}/${createFolder.name}`);
        nameGroups.set(key, members);
    }
    for (const [key, members] of nameGroups.entries()) {
        if (members.length < 2) {
            continue;
        }
        const [, name] = key.split('\x00');
        blockingConflicts.add(t('conflictDuplicateSiblingName', {name}));
    }

    if (!state.contentConnected && (pendingMoves.length > 0 || pendingRemoves.length > 0 || pendingAdds.length > 0 || pendingCreateFolders.length > 0)) {
        warnings.add(t('warningManualContentOperations'));
    }
    if (!state.contentConnected && (pendingAdds.length > 0 || pendingCreateFolders.length > 0)) {
        warnings.add(t('warningStagedEncryptedContent'));
    }

    return {
        blockingConflicts: [...blockingConflicts],
        warnings: [...warnings],
    };
}

function basename(path: string) {
    const normalized = path.replace(/\\/g, '/').replace(/\/+$/g, '');
    const parts = normalized.split('/');
    return parts[parts.length - 1] ?? normalized;
}
