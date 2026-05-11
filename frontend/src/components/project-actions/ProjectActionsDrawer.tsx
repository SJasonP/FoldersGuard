import {useEffect} from 'react';
import {Button, Drawer, Form, Input, Space, Typography} from 'antd';
import {
    DeleteOutlined,
    EditOutlined,
    ExportOutlined,
    SafetyCertificateOutlined,
    ShareAltOutlined,
    UnlockOutlined,
} from '@ant-design/icons';
import type {LocalProjectSummary} from '../../types';

type ProjectActionsDrawerProps = {
    open: boolean;
    project: LocalProjectSummary | null;
    projectNameSaving: boolean;
    onClose: () => void;
    onSaveProjectName: (projectName: string) => void;
    onInspect: () => void;
    onModify: () => void;
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
                                         projectNameSaving,
                                         onClose,
                                         onSaveProjectName,
                                         onInspect,
                                         onModify,
                                         onVerify,
                                         onDecrypt,
                                         onCreateShare,
                                         onExport,
                                         onDelete,
                                         t,
                                     }: ProjectActionsDrawerProps) {
    const [form] = Form.useForm<{ projectName: string }>();

    useEffect(() => {
        if (open) {
            form.setFieldsValue({projectName: project?.projectName ?? ''});
        }
    }, [form, open, project?.projectName]);

    return (
        <Drawer
            title={t('projectActions')}
            open={open}
            onClose={onClose}
            width={360}
            afterOpenChange={(nextOpen) => {
                if (!nextOpen) {
                    form.resetFields();
                }
            }}
        >
            <Space direction="vertical" size="middle" className="content-stack">
                {project ? (
                    <>
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{projectName: project.projectName}}
                            onFinish={(values) => onSaveProjectName(values.projectName)}
                        >
                            <Form.Item name="projectName" label={t('projectName')}
                                       rules={[{required: true, whitespace: true, message: t('projectNameRequired')}]}>
                                <Input.Search enterButton={t('save')} loading={projectNameSaving}
                                              onSearch={() => form.submit()}/>
                            </Form.Item>
                        </Form>
                        <Typography.Text type="secondary">
                            {t('projectId')}: {project.projectId}
                        </Typography.Text>
                    </>
                ) : null}
                <Button block onClick={onInspect}>
                    {t('inspectProject')}
                </Button>
                <Button block type="primary" ghost icon={<EditOutlined/>} onClick={onModify}>
                    {t('modifyProject')}
                </Button>
                <Button block icon={<SafetyCertificateOutlined/>} onClick={onVerify}>
                    {t('verifyProject')}
                </Button>
                <Button block type="primary" ghost icon={<UnlockOutlined/>} onClick={onDecrypt}>
                    {t('decryptProject')}
                </Button>
                <Button block type="primary" ghost icon={<ShareAltOutlined/>} onClick={onCreateShare}>
                    {t('createShare')}
                </Button>
                <Button block icon={<ExportOutlined/>} onClick={onExport}>
                    {t('exportProject')}
                </Button>
                <Button block danger icon={<DeleteOutlined/>} onClick={onDelete}>
                    {t('deleteProject')}
                </Button>
            </Space>
        </Drawer>
    );
}
