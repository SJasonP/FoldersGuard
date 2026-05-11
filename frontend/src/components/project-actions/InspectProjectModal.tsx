import {Form, Input, Modal} from 'antd';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type InspectProjectModalProps = {
    open: boolean;
    loading: boolean;
    onCancel: () => void;
    onSubmit: (password: string) => void;
    t: (key: string) => string;
};

export function InspectProjectModal({open, loading, onCancel, onSubmit, t}: InspectProjectModalProps) {
    const [form] = Form.useForm<{ password: string }>();
    useResetFormOnClose(form, open);

    return (
        <Modal
            title={t('inspectProject')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('inspectProject')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                onFinish={(values) => {
                    onSubmit(values.password);
                }}
            >
                <Form.Item name="password" label={t('password')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
                    <Input.Password autoComplete="current-password"/>
                </Form.Item>
            </Form>
        </Modal>
    );
}
