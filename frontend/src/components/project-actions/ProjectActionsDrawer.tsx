import { Button, Drawer, Space, Typography } from 'antd';
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons';
import type { LocalProjectSummary } from '../../types';

type ProjectActionsDrawerProps = {
  open: boolean;
  project: LocalProjectSummary | null;
  onClose: () => void;
  onInspect: () => void;
  onExport: () => void;
  onDelete: () => void;
  t: (key: string) => string;
};

export function ProjectActionsDrawer({
  open,
  project,
  onClose,
  onInspect,
  onExport,
  onDelete,
  t,
}: ProjectActionsDrawerProps) {
  return (
    <Drawer title={t('projectActions')} open={open} onClose={onClose} width={360}>
      <Space direction="vertical" size="middle" className="content-stack">
        {project ? (
          <Typography.Text type="secondary">
            {project.projectId} / {project.fileName}
          </Typography.Text>
        ) : null}
        <Button block type="primary" onClick={onInspect}>
          {t('inspectProject')}
        </Button>
        <Button block icon={<ExportOutlined />} onClick={onExport}>
          {t('exportProject')}
        </Button>
        <Button block danger icon={<DeleteOutlined />} onClick={onDelete}>
          {t('deleteProject')}
        </Button>
      </Space>
    </Drawer>
  );
}
