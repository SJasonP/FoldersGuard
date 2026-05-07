import { Checkbox, Form, Input, Modal } from 'antd';
import { PathInput } from '../common/PathInput';

type ExportProjectValues = {
  password: string;
  outputPath: string;
  force: boolean;
};

type ExportProjectModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: ExportProjectValues) => void;
  t: (key: string) => string;
};

export function ExportProjectModal({ open, loading, onCancel, onSubmit, t }: ExportProjectModalProps) {
  const [form] = Form.useForm<ExportProjectValues>();

  return (
    <Modal
      title={t('exportProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('exportProject')}
      confirmLoading={loading}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ force: false }}
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="current-password" />
        </Form.Item>
        <Form.Item
          name="outputPath"
          label={t('exportOutputPath')}
          rules={[{ required: true, message: t('exportOutputPath') }]}
        >
          <PathInput
            dialogKind="save-file"
            dialogTitle={t('exportOutputPath')}
            defaultFilename="project.fg"
            filters={[{ displayName: 'FoldersGuard Project (*.fg)', pattern: '*.fg' }]}
            placeholder="/path/to/project.fg"
            t={t}
          />
        </Form.Item>
        <Form.Item name="force" valuePropName="checked">
          <Checkbox>{t('forceOverwrite')}</Checkbox>
        </Form.Item>
      </Form>
    </Modal>
  );
}
