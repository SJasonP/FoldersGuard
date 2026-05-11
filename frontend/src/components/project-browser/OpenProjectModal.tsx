import {Form, Input, Modal} from 'antd';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type OpenProjectValues = {
    password: string;
    encryptedPath: string;
};

type OpenProjectModalProps = {
    open: boolean;
    loading: boolean;
    onCancel: () => void;
    onSubmit: (values: OpenProjectValues) => void;
    t: (key: string, values?: Record<string, string | number>) => string;
};

export function OpenProjectModal({open, loading, onCancel, onSubmit, t}: OpenProjectModalProps) {
    const [form] = Form.useForm<OpenProjectValues>();
    useResetFormOnClose(form, open);

    return (
        <Modal
            title={t('modifyProject')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('openProject')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                initialValues={{encryptedPath: ''}}
                onFinish={(values) => {
                    onSubmit(values);
                }}
            >
                <Form.Item name="password" label={t('password')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
                    <Input.Password autoComplete="current-password"/>
                </Form.Item>
                <Form.Item name="encryptedPath" label={t('encryptedContentPath')}>
                    <PathInput
                        dialogKind="open-directory"
                        dialogTitle={t('encryptedContentPath')}
                        placeholder="/path/to/encrypted-content"
                        t={t}
                    />
                </Form.Item>
            </Form>
        </Modal>
    );
}
