import {App as AntApp, Checkbox, Form, Input, Modal} from 'antd';
import {showOperationConfirmation} from '../common/operationConfirmation';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type DecryptProjectValues = {
    password: string;
    encryptedPath: string;
    outputPath: string;
    force: boolean;
    resume: boolean;
    continueOnError: boolean;
};

type DecryptProjectModalProps = {
    open: boolean;
    loading: boolean;
    sourceCleanupMode: string;
    defaultFailureHandling: string;
    onCancel: () => void;
    onSubmit: (values: DecryptProjectValues) => void;
    t: (key: string) => string;
};

export function DecryptProjectModal({
                                        open,
                                        loading,
                                        sourceCleanupMode,
                                        defaultFailureHandling,
                                        onCancel,
                                        onSubmit,
                                        t,
                                    }: DecryptProjectModalProps) {
    const {modal} = AntApp.useApp();
    const [form] = Form.useForm<DecryptProjectValues>();
    useResetFormOnClose(form, open);
    const sourceCleanupLabel = (value: string) => {
        if (value === 'delete') {
            return t('sourceCleanupDelete');
        }
        return t('sourceCleanupKeep');
    };

    const confirmSubmit = (values: DecryptProjectValues) => {
        showOperationConfirmation({
            modalApi: modal,
            title: t('decryptProject'),
            message: t('decryptProjectConfirm'),
            okText: t('decryptProject'),
            items: [
                {label: t('verifyEncryptedPath'), value: values.encryptedPath},
                {label: t('outputPath'), value: values.outputPath},
                {label: t('sourceCleanupMode'), value: sourceCleanupLabel(sourceCleanupMode)},
                {
                    label: t('forceOverwrite'),
                    value: values.force ? t('passwordProtectedYes') : t('passwordProtectedNo')
                },
                {
                    label: t('resumeDecryption'),
                    value: values.resume ? t('passwordProtectedYes') : t('passwordProtectedNo')
                },
                {
                    label: t('continueOnError'),
                    value: values.continueOnError ? t('passwordProtectedYes') : t('passwordProtectedNo')
                },
            ],
            onConfirm: () => {
                onSubmit(values);
            },
        });
    };

    return (
        <Modal
            title={t('decryptProject')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('decryptProject')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                initialValues={{
                    force: false,
                    resume: false,
                    continueOnError: defaultFailureHandling === 'continue',
                }}
                onFinish={confirmSubmit}
            >
                <Form.Item name="password" label={t('password')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
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
                <Form.Item name="force" valuePropName="checked">
                    <Checkbox>{t('forceOverwrite')}</Checkbox>
                </Form.Item>
                <Form.Item name="resume" valuePropName="checked" extra={t('resumeDecryptionHint')}>
                    <Checkbox>{t('resumeDecryption')}</Checkbox>
                </Form.Item>
                <Form.Item name="continueOnError" valuePropName="checked" extra={t('continueOnErrorHint')}>
                    <Checkbox>{t('continueOnError')}</Checkbox>
                </Form.Item>
            </Form>
        </Modal>
    );
}
