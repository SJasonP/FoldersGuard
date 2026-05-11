import {App as AntApp, Checkbox, Form, Input, Modal, Select} from 'antd';
import {showOperationConfirmation} from '../common/operationConfirmation';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type DecryptShareValues = {
    password: string;
    encryptedPath: string;
    outputPath: string;
    force: boolean;
    sourceCleanup: string;
};

type DecryptShareModalProps = {
    open: boolean;
    loading: boolean;
    defaultSourceCleanup: string;
    onCancel: () => void;
    onSubmit: (values: DecryptShareValues) => void;
    t: (key: string) => string;
};

export function DecryptShareModal({
                                      open,
                                      loading,
                                      defaultSourceCleanup,
                                      onCancel,
                                      onSubmit,
                                      t,
                                  }: DecryptShareModalProps) {
    const {modal} = AntApp.useApp();
    const [form] = Form.useForm<DecryptShareValues>();
    useResetFormOnClose(form, open);
    const sourceCleanupLabel = (value: string) => {
        if (value === 'delete') {
            return t('sourceCleanupDelete');
        }
        return t('sourceCleanupKeep');
    };

    const confirmSubmit = (values: DecryptShareValues) => {
        showOperationConfirmation({
            modalApi: modal,
            title: t('decryptShare'),
            message: t('decryptShareConfirm'),
            okText: t('decryptShare'),
            danger: values.sourceCleanup === 'delete',
            items: [
                {label: t('verifyEncryptedPath'), value: values.encryptedPath},
                {label: t('outputPath'), value: values.outputPath},
                {label: t('sourceCleanupOperationMode'), value: sourceCleanupLabel(values.sourceCleanup)},
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
            title={t('decryptShare')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('decryptShare')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                initialValues={{
                    force: false,
                    sourceCleanup: defaultSourceCleanup,
                }}
                onFinish={confirmSubmit}
            >
                <Form.Item name="password" label={t('password')}>
                    <Input.Password autoComplete="current-password"/>
                </Form.Item>
                <Form.Item
                    name="encryptedPath"
                    label={t('verifyEncryptedPath')}
                    rules={[{required: true, message: t('verifyEncryptedPath')}]}
                >
                    <PathInput
                        dialogKind="open-directory"
                        dialogTitle={t('verifyEncryptedPath')}
                        placeholder="/path/to/encrypted-content"
                        t={t}
                    />
                </Form.Item>
                <Form.Item name="outputPath" label={t('outputPath')}
                           rules={[{required: true, message: t('outputPath')}]}>
                    <PathInput
                        dialogKind="open-directory"
                        dialogTitle={t('outputPath')}
                        placeholder="/path/to/restored-output"
                        t={t}
                    />
                </Form.Item>
                <Form.Item name="sourceCleanup" label={t('sourceCleanupOperationMode')} rules={[{required: true}]}>
                    <Select
                        options={[
                            {value: 'keep', label: t('sourceCleanupKeep')},
                            {value: 'delete', label: t('sourceCleanupDelete')},
                        ]}
                    />
                </Form.Item>
                <Form.Item name="force" valuePropName="checked">
                    <Checkbox>{t('forceOverwrite')}</Checkbox>
                </Form.Item>
            </Form>
        </Modal>
    );
}
