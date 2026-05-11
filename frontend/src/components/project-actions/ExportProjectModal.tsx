import {App as AntApp, Checkbox, Form, Input, Modal} from 'antd';
import {showOperationConfirmation} from '../common/operationConfirmation';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type ExportProjectValues = {
    password: string;
    outputPath: string;
    force: boolean;
};

type ExportProjectModalProps = {
    open: boolean;
    loading: boolean;
    onCancel: () => void;
    onSubmit: (values: ExportProjectValues) => void;
    t: (key: string) => string;
};

export function ExportProjectModal({open, loading, onCancel, onSubmit, t}: ExportProjectModalProps) {
    const {modal} = AntApp.useApp();
    const [form] = Form.useForm<ExportProjectValues>();
    useResetFormOnClose(form, open);
    const confirmSubmit = (values: ExportProjectValues) => {
        showOperationConfirmation({
            modalApi: modal,
            title: t('exportProject'),
            message: t('exportProjectConfirm'),
            okText: t('exportProject'),
            items: [
                {label: t('exportOutputPath'), value: values.outputPath},
                {
                    label: t('forceOverwrite'),
                    value: values.force ? t('passwordProtectedYes') : t('passwordProtectedNo')
                },
            ],
            onConfirm: () => {
                onSubmit(values);
            },
        });
    };

    return (
        <Modal
            title={t('exportProject')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('exportProject')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                initialValues={{force: false}}
                onFinish={confirmSubmit}
            >
                <Form.Item name="password" label={t('password')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
                    <Input.Password autoComplete="current-password"/>
                </Form.Item>
                <Form.Item
                    name="outputPath"
                    label={t('exportOutputPath')}
                    rules={[{required: true, message: t('exportOutputPath')}]}
                >
                    <PathInput
                        dialogKind="save-file"
                        dialogTitle={t('exportOutputPath')}
                        defaultFilename="project.fg"
                        filters={[{displayName: t('fgProjectFilter'), pattern: '*.fg'}]}
                        placeholder="/path/to/project.fg"
                        t={t}
                    />
                </Form.Item>
                <Form.Item name="force" valuePropName="checked">
                    <Checkbox>{t('forceOverwrite')}</Checkbox>
                </Form.Item>
            </Form>
        </Modal>
    );
}
