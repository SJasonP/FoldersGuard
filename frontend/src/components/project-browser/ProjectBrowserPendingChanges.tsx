import { Alert, Button, List, Typography } from 'antd';
import type { PendingAdd, PendingCreateFolder, PendingMove, PendingRemove, PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserPendingChangesProps = {
  pendingRenames: PendingRename[];
  pendingMoves: PendingMove[];
  pendingRemoves: PendingRemove[];
  pendingAdds: PendingAdd[];
  pendingCreateFolders: PendingCreateFolder[];
  blockingConflicts: string[];
  warnings: string[];
  onDiscardRename: (itemId: string) => void;
  onDiscardMove: (itemId: string) => void;
  onDiscardRemove: (itemId: string) => void;
  onDiscardAdd: (itemId: string) => void;
  onDiscardCreateFolder: (itemId: string) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

type PendingChangeRow = {
  key: string;
  itemId: string;
  label: string;
  onDiscard: (itemId: string) => void;
};

export function ProjectBrowserPendingChanges({
  pendingRenames,
  pendingMoves,
  pendingRemoves,
  pendingAdds,
  pendingCreateFolders,
  blockingConflicts,
  warnings,
  onDiscardRename,
  onDiscardMove,
  onDiscardRemove,
  onDiscardAdd,
  onDiscardCreateFolder,
  t,
}: ProjectBrowserPendingChangesProps) {
  const rows: PendingChangeRow[] = [
    ...pendingAdds.map((add) => ({
      key: `add:${add.itemId}`,
      itemId: add.itemId,
      label: `${t('pendingAdd')}: ${add.sourcePath} -> ${add.targetFolderPath}`,
      onDiscard: onDiscardAdd,
    })),
    ...pendingCreateFolders.map((createFolder) => ({
      key: `create-folder:${createFolder.itemId}`,
      itemId: createFolder.itemId,
      label: `${t('pendingCreateFolder')}: ${createFolder.targetFolderPath} -> ${createFolder.name}`,
      onDiscard: onDiscardCreateFolder,
    })),
    ...pendingRenames.map((rename) => ({
      key: `rename:${rename.itemId}`,
      itemId: rename.itemId,
      label: `${t('pendingRename')}: ${rename.itemPath} -> ${rename.newName}`,
      onDiscard: onDiscardRename,
    })),
    ...pendingMoves.map((move) => ({
      key: `move:${move.itemId}`,
      itemId: move.itemId,
      label: `${t('pendingMove')}: ${move.itemPath} -> ${move.targetFolderPath}`,
      onDiscard: onDiscardMove,
    })),
    ...pendingRemoves.map((remove) => ({
      key: `remove:${remove.itemId}`,
      itemId: remove.itemId,
      label: `${t('pendingRemove')}: ${remove.itemPath}`,
      onDiscard: onDiscardRemove,
    })),
  ];

  return (
    <div>
      <Typography.Title level={5}>{t('pendingChanges')}</Typography.Title>
      {blockingConflicts.length > 0 ? (
        <Alert
          type="error"
          showIcon
          message={t('blockingConflicts')}
          description={
            <List
              size="small"
              dataSource={blockingConflicts}
              renderItem={(conflict) => <List.Item>{conflict}</List.Item>}
            />
          }
          style={{ marginBottom: 12 }}
        />
      ) : null}
      {warnings.length > 0 ? (
        <Alert
          type="warning"
          showIcon
          message={t('applyWarnings')}
          description={
            <List
              size="small"
              dataSource={warnings}
              renderItem={(warning) => <List.Item>{warning}</List.Item>}
            />
          }
          style={{ marginBottom: 12 }}
        />
      ) : null}
      <List
        size="small"
        bordered
        dataSource={rows}
        locale={{ emptyText: t('noPendingChanges') }}
        renderItem={(row) => (
          <List.Item actions={[<Button onClick={() => row.onDiscard(row.itemId)}>{t('discard')}</Button>]}>
            <Typography.Text>
              {row.label}
            </Typography.Text>
          </List.Item>
        )}
      />
    </div>
  );
}
