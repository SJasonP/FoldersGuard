import { Checkbox, Form, Input, Modal } from 'antd';
import { PathInput } from '../common/PathInput';
import { useResetFormOnClose } from '../common/useResetFormOnClose';

type ImportProjectValues = {
  inputPath: string;
  password: string;
  force: boolean;
};

type ImportProjectModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: ImportProjectValues) => void;
  t: (key: string) => string;
};

export function ImportProjectModal({ open, loading, onCancel, onSubmit, t }: ImportProjectModalProps) {
  const [form] = Form.useForm<ImportProjectValues>();
  useResetFormOnClose(form, open);

  return (
    <Modal
      title={t('importProject')}
      open={open}
      onCancel={onCancel}
      onOk={() => void form.submit()}
      okText={t('importProject')}
      confirmLoading={loading}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ force: false }}
        onFinish={(values) => {
          onSubmit(values);
        }}
      >
        <Form.Item
          name="inputPath"
          label={t('importInputPath')}
          rules={[{ required: true, message: t('importInputPath') }]}
        >
          <PathInput
            dialogKind="open-file"
            dialogTitle={t('importInputPath')}
            filters={[{ displayName: t('fgProjectFilter'), pattern: '*.fg' }]}
            placeholder="/path/to/project.fg"
            t={t}
          />
        </Form.Item>
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item name="force" valuePropName="checked">
          <Checkbox>{t('forceOverwrite')}</Checkbox>
        </Form.Item>
      </Form>
    </Modal>
  );
}
