import { Button, List, Typography } from 'antd';
import type { PendingAdd, PendingMove, PendingRemove, PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserPendingChangesProps = {
  pendingRenames: PendingRename[];
  pendingMoves: PendingMove[];
  pendingRemoves: PendingRemove[];
  pendingAdds: PendingAdd[];
  onDiscardRename: (itemId: string) => void;
  onDiscardMove: (itemId: string) => void;
  onDiscardRemove: (itemId: string) => void;
  onDiscardAdd: (itemId: string) => void;
  t: (key: string) => string;
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
  onDiscardRename,
  onDiscardMove,
  onDiscardRemove,
  onDiscardAdd,
  t,
}: ProjectBrowserPendingChangesProps) {
  const rows: PendingChangeRow[] = [
    ...pendingAdds.map((add) => ({
      key: `add:${add.itemId}`,
      itemId: add.itemId,
      label: `${t('pendingAdd')}: ${add.sourcePath} -> ${add.targetFolderPath}`,
      onDiscard: onDiscardAdd,
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
