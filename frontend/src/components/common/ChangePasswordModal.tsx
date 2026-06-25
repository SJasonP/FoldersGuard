import {Form, Input, Modal} from 'antd';
import {useResetFormOnClose} from './useResetFormOnClose';

type ChangePasswordValues = {
    oldPassword: string;
    newPassword: string;
    newPasswordConfirm: string;
};

type ChangePasswordModalProps = {
    open: boolean;
    loading: boolean;
    title: string;
    onCancel: () => void;
    onSubmit: (values: {oldPassword: string; newPassword: string}) => void;
    t: (key: string) => string;
};

export function ChangePasswordModal({open, loading, title, onCancel, onSubmit, t}: ChangePasswordModalProps) {
    const [form] = Form.useForm<ChangePasswordValues>();
    useResetFormOnClose(form, open);

    return (
        <Modal
            title={title}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('changePassword')}
            confirmLoading={loading}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                onFinish={(values) => onSubmit({oldPassword: values.oldPassword, newPassword: values.newPassword})}
            >
                <Form.Item name="oldPassword" label={t('currentPassword')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
                    <Input.Password autoComplete="current-password"/>
                </Form.Item>
                <Form.Item name="newPassword" label={t('newPassword')}
                           rules={[{required: true, message: t('passwordRequired')}]}>
                    <Input.Password autoComplete="new-password"/>
                </Form.Item>
                <Form.Item
                    name="newPasswordConfirm"
                    label={t('newPasswordConfirm')}
                    dependencies={['newPassword']}
                    rules={[
                        {required: true, message: t('passwordConfirmRequired')},
                        ({getFieldValue}) => ({
                            validator(_, value) {
                                if (value === getFieldValue('newPassword')) {
                                    return Promise.resolve();
                                }
                                return Promise.reject(new Error(t('passwordMismatch')));
                            },
                        }),
                    ]}
                >
                    <Input.Password autoComplete="new-password"/>
                </Form.Item>
            </Form>
        </Modal>
    );
}
