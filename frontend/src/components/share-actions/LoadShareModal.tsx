import { Form, Input, Modal, Typography } from 'antd';

type LoadShareValues = {
  databasePath: string;
  password: string;
};

type LoadShareModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: LoadShareValues) => void;
  t: (key: string) => string;
};

export function LoadShareModal({ open, loading, onCancel, onSubmit, t }: LoadShareModalProps) {
  const [form] = Form.useForm<LoadShareValues>();

  return (
    <Modal
      title={t('loadShare')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('loadShare')}
      confirmLoading={loading}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item
          name="databasePath"
          label={t('databasePath')}
          rules={[{ required: true, message: t('databasePath') }]}
        >
          <Input placeholder="/path/to/share.fgs" />
        </Form.Item>
        <Form.Item name="password" label={t('password')}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Typography.Text type="secondary">{t('loadSharePasswordHint')}</Typography.Text>
      </Form>
    </Modal>
  );
}
