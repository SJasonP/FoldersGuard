import { Form, Input, Modal, Typography } from 'antd';
import { PathInput } from '../common/PathInput';
import { useResetFormOnClose } from '../common/useResetFormOnClose';

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
  useResetFormOnClose(form, open);

  return (
    <Modal
      title={t('loadShare')}
      open={open}
      onCancel={onCancel}
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
        }}
      >
        <Form.Item
          name="databasePath"
          label={t('databasePath')}
          rules={[{ required: true, message: t('databasePath') }]}
        >
          <PathInput
            dialogKind="open-file"
            dialogTitle={t('databasePath')}
            filters={[{ displayName: t('fgShareFilter'), pattern: '*.fgs' }]}
            placeholder="/path/to/share.fgs"
            t={t}
          />
        </Form.Item>
        <Form.Item name="password" label={t('password')}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Typography.Text type="secondary">{t('loadSharePasswordHint')}</Typography.Text>
      </Form>
    </Modal>
  );
}
