import { Form, Input, Modal } from 'antd';
import { useResetFormOnClose } from '../common/useResetFormOnClose';

type CreateSharePasswordModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (password: string) => void;
  t: (key: string) => string;
};

export function CreateSharePasswordModal({ open, loading, onCancel, onSubmit, t }: CreateSharePasswordModalProps) {
  const [form] = Form.useForm<{ password: string }>();
  useResetFormOnClose(form, open);

  return (
    <Modal
      title={t('createShare')}
      open={open}
      onCancel={onCancel}
      onOk={() => void form.submit()}
      okText={t('continueAction')}
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
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
