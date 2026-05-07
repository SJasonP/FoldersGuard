import { Form, Input, Modal } from 'antd';

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

export function OpenProjectModal({ open, loading, onCancel, onSubmit, t }: OpenProjectModalProps) {
  const [form] = Form.useForm<OpenProjectValues>();

  return (
    <Modal
      title={t('modifyProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('openProject')}
      confirmLoading={loading}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ encryptedPath: '' }}
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item name="encryptedPath" label={t('encryptedContentPath')}>
          <Input placeholder="/path/to/encrypted-content" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
