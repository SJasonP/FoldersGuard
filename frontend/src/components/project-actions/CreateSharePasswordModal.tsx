import { Form, Input, Modal } from 'antd';

type CreateSharePasswordModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (password: string) => void;
  t: (key: string) => string;
};

export function CreateSharePasswordModal({ open, loading, onCancel, onSubmit, t }: CreateSharePasswordModalProps) {
  const [form] = Form.useForm<{ password: string }>();

  return (
    <Modal
      title={t('createShare')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
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
          form.resetFields();
        }}
      >
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
