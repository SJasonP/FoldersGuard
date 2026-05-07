import { Checkbox, Form, Input, Modal, Select } from 'antd';

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
  const [form] = Form.useForm<DecryptShareValues>();

  return (
    <Modal
      title={t('decryptShare')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
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
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item name="password" label={t('password')}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item
          name="encryptedPath"
          label={t('verifyEncryptedPath')}
          rules={[{ required: true, message: t('verifyEncryptedPath') }]}
        >
          <Input placeholder="/path/to/encrypted-content" />
        </Form.Item>
        <Form.Item name="outputPath" label={t('outputPath')} rules={[{ required: true, message: t('outputPath') }]}>
          <Input placeholder="/path/to/restored-output" />
        </Form.Item>
        <Form.Item name="sourceCleanup" label={t('sourceCleanupOperationMode')} rules={[{ required: true }]}>
          <Select
            options={[
              { value: 'keep', label: t('sourceCleanupKeep') },
              { value: 'delete', label: t('sourceCleanupDelete') },
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
