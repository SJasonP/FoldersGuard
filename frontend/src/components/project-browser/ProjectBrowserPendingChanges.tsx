import { Button, List, Typography } from 'antd';
import type { PendingRename } from '../../hooks/useProjectBrowser';

type ProjectBrowserPendingChangesProps = {
  pendingRenames: PendingRename[];
  onDiscardRename: (itemId: string) => void;
  t: (key: string) => string;
};

export function ProjectBrowserPendingChanges({ pendingRenames, onDiscardRename, t }: ProjectBrowserPendingChangesProps) {
  return (
    <div>
      <Typography.Title level={5}>{t('pendingChanges')}</Typography.Title>
      <List
        size="small"
        bordered
        dataSource={pendingRenames}
        locale={{ emptyText: t('noPendingChanges') }}
        renderItem={(rename) => (
          <List.Item actions={[<Button onClick={() => onDiscardRename(rename.itemId)}>{t('discard')}</Button>]}>
            <Typography.Text>
              {t('pendingRename')}: {rename.itemPath} -&gt; {rename.newName}
            </Typography.Text>
          </List.Item>
        )}
      />
    </div>
  );
}
