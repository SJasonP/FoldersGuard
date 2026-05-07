import { Form, Input, Modal } from 'antd';
import { PathInput } from '../common/PathInput';

type VerifyShareValues = {
  password: string;
  encryptedPath: string;
};

type VerifyShareModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: VerifyShareValues) => void;
  t: (key: string) => string;
};

export function VerifyShareModal({ open, loading, onCancel, onSubmit, t }: VerifyShareModalProps) {
  const [form] = Form.useForm<VerifyShareValues>();

  return (
    <Modal
      title={t('verifyShare')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('verifyShare')}
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
        <Form.Item name="password" label={t('password')}>
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
