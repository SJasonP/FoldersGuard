import { Form, Input, Modal } from 'antd';
import { PathInput } from '../common/PathInput';

type VerifyProjectValues = {
  password: string;
  encryptedPath: string;
};

type VerifyProjectModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: VerifyProjectValues) => void;
  t: (key: string) => string;
};

export function VerifyProjectModal({ open, loading, onCancel, onSubmit, t }: VerifyProjectModalProps) {
  const [form] = Form.useForm<VerifyProjectValues>();

  return (
    <Modal
      title={t('verifyProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('verifyProject')}
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
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item
          name="encryptedPath"
          label={t('verifyEncryptedPath')}
          rules={[{ required: true, message: t('verifyEncryptedPath') }]}
        >
          <PathInput
            dialogKind="open-directory"
            dialogTitle={t('verifyEncryptedPath')}
            placeholder="/path/to/encrypted-content"
            t={t}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}
