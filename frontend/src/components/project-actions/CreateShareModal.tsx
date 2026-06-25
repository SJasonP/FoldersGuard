import {App as AntApp, Checkbox, Form, Input, Modal, Typography} from 'antd';
import {showOperationConfirmation} from '../common/operationConfirmation';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type CreateShareValues = {
    outputPath: string;
    force: boolean;
    passwordProtected: boolean;
    sharePassword?: string;
    sharePasswordConfirm?: string;
};

type CreateShareModalProps = {
    open: boolean;
    loading: boolean;
    selectedItemCount: number;
    onCancel: () => void;
    onSubmit: (values: CreateShareValues) => void;
    t: (key: string, values?: Record<string, string | number>) => string;
};

export function CreateShareModal({open, loading, selectedItemCount, onCancel, onSubmit, t}: CreateShareModalProps) {
    const {modal} = AntApp.useApp();
    const [form] = Form.useForm<CreateShareValues>();
    const passwordProtected = Form.useWatch('passwordProtected', form) ?? true;
    useResetFormOnClose(form, open);

    const confirmSubmit = (values: CreateShareValues) => {
        showOperationConfirmation({
            modalApi: modal,
            title: t('createShare'),
            message: values.passwordProtected ? t('createShareConfirm') : t('createUnprotectedShareConfirm'),
            okText: t('createShare'),
            items: [
                {label: t('shareSelectionItems'), value: selectedItemCount},
                {label: t('shareDatabaseOutputPath'), value: values.outputPath},
                {
                    label: t('passwordProtected'),
                    value: values.passwordProtected ? t('passwordProtectedYes') : t('passwordProtectedNo'),
                },
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
            title={t('createShare')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('createShare')}
            confirmLoading={loading}
            destroyOnHidden
            width={720}
        >
            <Form
                form={form}
                layout="vertical"
                initialValues={{
                    force: false,
                    passwordProtected: true,
                }}
                onFinish={confirmSubmit}
            >
                <Typography.Paragraph>{t('selectedShareItemCount', {count: selectedItemCount})}</Typography.Paragraph>
                <Form.Item
                    name="outputPath"
                    label={t('shareDatabaseOutputPath')}
                    rules={[{required: true, message: t('shareDatabaseOutputPath')}]}
                >
                    <PathInput
                        dialogKind="save-file"
                        dialogTitle={t('shareDatabaseOutputPath')}
                        defaultFilename="share.fgs"
                        filters={[{displayName: t('fgShareFilter'), pattern: '*.fgs'}]}
                        placeholder="/path/to/share.fgs"
                        t={t}
                    />
                </Form.Item>
                <Form.Item name="passwordProtected" valuePropName="checked">
                    <Checkbox>{t('sharePasswordProtected')}</Checkbox>
                </Form.Item>
                <Form.Item
                    name="sharePassword"
                    label={t('sharePassword')}
                    rules={passwordProtected ? [{required: true, message: t('sharePasswordRequired')}] : []}
                >
                    <Input.Password autoComplete="new-password" disabled={!passwordProtected}/>
                </Form.Item>
                <Form.Item
                    name="sharePasswordConfirm"
                    label={t('sharePasswordConfirm')}
                    dependencies={['sharePassword', 'passwordProtected']}
                    rules={
                        passwordProtected
                            ? [
                                {required: true, message: t('sharePasswordConfirmRequired')},
                                ({getFieldValue}) => ({
                                    validator(_, value) {
                                        if (value === getFieldValue('sharePassword')) {
                                            return Promise.resolve();
                                        }
                                        return Promise.reject(new Error(t('passwordMismatch')));
                                    },
                                }),
                            ]
                            : []
                    }
                >
                    <Input.Password autoComplete="new-password" disabled={!passwordProtected}/>
                </Form.Item>
                <Form.Item name="force" valuePropName="checked">
                    <Checkbox>{t('forceOverwrite')}</Checkbox>
                </Form.Item>
            </Form>
        </Modal>
    );
}
