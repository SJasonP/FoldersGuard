import {Button, Empty, Flex, Input, Space, Table, Typography} from 'antd';
import type {ColumnsType} from 'antd/es/table';
import {FolderAddOutlined, ImportOutlined, ReloadOutlined, ShareAltOutlined} from '@ant-design/icons';
import type {LocalProjectRow} from '../types';

type HomeViewProps = {
    columns: ColumnsType<LocalProjectRow>;
    tableScrollX: number;
    loading: boolean;
    projects: LocalProjectRow[];
    projectSearch: string;
    projectsError: string | null;
    selectedProjectId: string | null;
    disabled: boolean;
    operationActive: boolean;
    onCreateProject: () => void;
    onImportProject: () => void;
    onLoadShare: () => void;
    onProjectSearchChange: (value: string) => void;
    onRefresh: () => void;
    onSelectProject: (projectId: string | null) => void;
    onOpenProjectActions: (projectId?: string) => void;
    t: (key: string) => string;
};

export function HomeView({
                             columns,
                             tableScrollX,
                             loading,
                             projects,
                             projectSearch,
                             projectsError,
                             selectedProjectId,
                             disabled,
                             operationActive,
                             onCreateProject,
                             onImportProject,
                             onLoadShare,
                             onProjectSearchChange,
                             onRefresh,
                             onSelectProject,
                             onOpenProjectActions,
                             t,
                         }: HomeViewProps) {
    // While an operation runs (or data services are unavailable) the whole list
    // is locked: no refreshing, no selecting, no opening project actions. This
    // prevents accidental interaction with project state mid-operation.
    const locked = disabled || operationActive;
    return (
        <Space direction="vertical" size="large" className="content-stack">
            <Flex justify="space-between" align="center" gap={16}>
                <Typography.Title level={2}>{t('localProjects')}</Typography.Title>
                <Space wrap>
                    <Button icon={<FolderAddOutlined/>} onClick={onCreateProject}
                            disabled={locked}>
                        {t('createProject')}
                    </Button>
                    <Button icon={<ImportOutlined/>} onClick={onImportProject} disabled={locked}>
                        {t('importProject')}
                    </Button>
                    <Button icon={<ShareAltOutlined/>} onClick={onLoadShare} disabled={locked}>
                        {t('loadShare')}
                    </Button>
                    <Button onClick={() => onOpenProjectActions()} disabled={locked || !selectedProjectId}>
                        {t('viewProjectActions')}
                    </Button>
                    <Input.Search
                        placeholder={t('searchProjects')}
                        value={projectSearch}
                        onChange={(event) => onProjectSearchChange(event.target.value)}
                        disabled={locked}
                    />
                    <Button icon={<ReloadOutlined/>} onClick={onRefresh} disabled={locked}>
                        {t('refresh')}
                    </Button>
                </Space>
            </Flex>
            {projectsError ? <Typography.Text type="danger">{projectsError}</Typography.Text> : null}
            <Table
                columns={columns}
                dataSource={projects}
                loading={loading}
                scroll={{x: tableScrollX}}
                tableLayout="fixed"
                rowSelection={{
                    type: 'radio',
                    selectedRowKeys: selectedProjectId ? [selectedProjectId] : [],
                    getCheckboxProps: () => ({disabled: locked}),
                    onChange: (selectedRowKeys) => onSelectProject((selectedRowKeys[0] as string | undefined) ?? null),
                }}
                onRow={(record) => ({
                    onClick: () => {
                        if (!locked) {
                            onSelectProject(record.projectId);
                            onOpenProjectActions(record.projectId);
                        }
                    },
                    onDoubleClick: () => {
                        if (!locked) {
                            onOpenProjectActions(record.projectId);
                        }
                    },
                })}
                locale={{emptyText: <Empty description={t('noProjects')}/>}}
                pagination={false}
            />
        </Space>
    );
}
