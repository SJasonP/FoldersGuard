import { Form, Input, Modal } from 'antd';

type InspectProjectModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (password: string) => void;
  t: (key: string) => string;
};

export function InspectProjectModal({ open, loading, onCancel, onSubmit, t }: InspectProjectModalProps) {
  const [form] = Form.useForm<{ password: string }>();

  return (
    <Modal
      title={t('inspectProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('inspectProject')}
      confirmLoading={loading}
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
