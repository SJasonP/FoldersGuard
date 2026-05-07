import { Checkbox, Form, Input, Modal, Select } from 'antd';
import { showOperationConfirmation } from '../common/operationConfirmation';
import { PathInput } from '../common/PathInput';

type DecryptProjectValues = {
  password: string;
  encryptedPath: string;
  outputPath: string;
  force: boolean;
  sourceCleanup: string;
};

type DecryptProjectModalProps = {
  open: boolean;
  loading: boolean;
  defaultSourceCleanup: string;
  onCancel: () => void;
  onSubmit: (values: DecryptProjectValues) => void;
  t: (key: string) => string;
};

export function DecryptProjectModal({
  open,
  loading,
  defaultSourceCleanup,
  onCancel,
  onSubmit,
  t,
}: DecryptProjectModalProps) {
  const [form] = Form.useForm<DecryptProjectValues>();
  const sourceCleanupLabel = (value: string) => {
    if (value === 'delete') {
      return t('sourceCleanupDelete');
    }
    return t('sourceCleanupKeep');
  };

  const confirmSubmit = (values: DecryptProjectValues) => {
    showOperationConfirmation({
      title: t('decryptProject'),
      message: t('decryptProjectConfirm'),
      okText: t('decryptProject'),
      danger: values.sourceCleanup === 'delete',
      items: [
        { label: t('verifyEncryptedPath'), value: values.encryptedPath },
        { label: t('outputPath'), value: values.outputPath },
        { label: t('sourceCleanupOperationMode'), value: sourceCleanupLabel(values.sourceCleanup) },
        { label: t('forceOverwrite'), value: values.force ? t('passwordProtectedYes') : t('passwordProtectedNo') },
      ],
      onConfirm: () => {
        onSubmit(values);
        form.resetFields();
      },
    });
  };

  return (
    <Modal
      title={t('decryptProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('decryptProject')}
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
        onFinish={confirmSubmit}
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
        <Form.Item name="outputPath" label={t('outputPath')} rules={[{ required: true, message: t('outputPath') }]}>
          <PathInput
            dialogKind="open-directory"
            dialogTitle={t('outputPath')}
            placeholder="/path/to/restored-output"
            t={t}
          />
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
