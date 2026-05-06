import { Button, Empty, Flex, Input, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ReloadOutlined } from '@ant-design/icons';
import type { LocalProjectRow } from '../types';

type HomeViewProps = {
  columns: ColumnsType<LocalProjectRow>;
  loading: boolean;
  projects: LocalProjectRow[];
  projectSearch: string;
  projectsError: string | null;
  onProjectSearchChange: (value: string) => void;
  onRefresh: () => void;
  t: (key: string) => string;
};

export function HomeView({
  columns,
  loading,
  projects,
  projectSearch,
  projectsError,
  onProjectSearchChange,
  onRefresh,
  t,
}: HomeViewProps) {
  return (
    <Space direction="vertical" size="large" className="content-stack">
      <Flex justify="space-between" align="center" gap={16}>
        <Typography.Title level={2}>{t('localProjects')}</Typography.Title>
        <Space>
          <Input.Search
            placeholder={t('searchProjects')}
            value={projectSearch}
            onChange={(event) => onProjectSearchChange(event.target.value)}
          />
          <Button icon={<ReloadOutlined />} onClick={onRefresh}>
            {t('refresh')}
          </Button>
        </Space>
      </Flex>
      {projectsError ? <Typography.Text type="danger">{projectsError}</Typography.Text> : null}
      <Table
        columns={columns}
        dataSource={projects}
        loading={loading}
        locale={{ emptyText: <Empty description={t('noProjects')} /> }}
        pagination={false}
      />
    </Space>
  );
}
