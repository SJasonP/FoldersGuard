import type { DataNode } from 'antd/es/tree';
import type { ProjectBrowserItemModel } from '../../types';
import type { PendingRename } from '../../hooks/useProjectBrowser';

export function pendingRenameMap(pendingRenames: PendingRename[]) {
  return new Map(pendingRenames.map((rename) => [rename.itemId, rename]));
}

export function displayNameForItem(item: ProjectBrowserItemModel, pendingByID: Map<string, PendingRename>) {
  return pendingByID.get(item.id)?.newName ?? item.name;
}

export function buildFolderTree(
  items: ProjectBrowserItemModel[],
  rootID: string,
  pendingByID: Map<string, PendingRename>,
): DataNode[] {
  const folders = items.filter((item) => item.type === 'folder');
  const childrenByParent = new Map<string, ProjectBrowserItemModel[]>();
  for (const folder of folders) {
    const children = childrenByParent.get(folder.parentId) ?? [];
    children.push(folder);
    childrenByParent.set(folder.parentId, children);
  }

  const buildNode = (folder: ProjectBrowserItemModel): DataNode => ({
    key: folder.id,
    title: displayNameForItem(folder, pendingByID),
    children: (childrenByParent.get(folder.id) ?? [])
      .sort((left, right) => displayNameForItem(left, pendingByID).localeCompare(displayNameForItem(right, pendingByID)))
      .map(buildNode),
  });

  const root = folders.find((folder) => folder.id === rootID);
  return root ? [buildNode(root)] : [];
}

export function folderBreadcrumbItems(
  items: ProjectBrowserItemModel[],
  folderID: string,
  pendingByID: Map<string, PendingRename>,
) {
  const byID = new Map(items.map((item) => [item.id, item]));
  const breadcrumbs: { key: string; title: string }[] = [];
  for (let current = byID.get(folderID); current; current = current.parentId ? byID.get(current.parentId) : undefined) {
    breadcrumbs.push({
      key: current.id,
      title: displayNameForItem(current, pendingByID),
    });
  }
  return breadcrumbs.reverse();
}

export function filteredFolderItems(
  items: ProjectBrowserItemModel[],
  folderID: string,
  searchQuery: string,
  pendingByID: Map<string, PendingRename>,
) {
  const normalizedQuery = searchQuery.trim().toLowerCase();
  return items
    .filter((item) => item.parentId === folderID)
    .filter((item) => {
      if (!normalizedQuery) {
        return true;
      }
      return displayNameForItem(item, pendingByID).toLowerCase().includes(normalizedQuery) || item.path.toLowerCase().includes(normalizedQuery);
    })
    .sort((left, right) => {
      if (left.type !== right.type) {
        return left.type === 'folder' ? -1 : 1;
      }
      return displayNameForItem(left, pendingByID).localeCompare(displayNameForItem(right, pendingByID));
    });
}
