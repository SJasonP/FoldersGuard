import { Button, Drawer, Space, Typography } from 'antd';
import { DeleteOutlined, ExportOutlined, SafetyCertificateOutlined, ShareAltOutlined, UnlockOutlined } from '@ant-design/icons';
import type { LocalProjectSummary } from '../../types';

type ProjectActionsDrawerProps = {
  open: boolean;
  project: LocalProjectSummary | null;
  onClose: () => void;
  onInspect: () => void;
  onVerify: () => void;
  onDecrypt: () => void;
  onCreateShare: () => void;
  onExport: () => void;
  onDelete: () => void;
  t: (key: string) => string;
};

export function ProjectActionsDrawer({
  open,
  project,
  onClose,
  onInspect,
  onVerify,
  onDecrypt,
  onCreateShare,
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
        <Button block icon={<SafetyCertificateOutlined />} onClick={onVerify}>
          {t('verifyProject')}
        </Button>
        <Button block icon={<UnlockOutlined />} onClick={onDecrypt}>
          {t('decryptProject')}
        </Button>
        <Button block icon={<ShareAltOutlined />} onClick={onCreateShare}>
          {t('createShare')}
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
