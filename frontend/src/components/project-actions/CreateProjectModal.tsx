import { Checkbox, Form, Input, InputNumber, Modal, Select } from 'antd';
import type { SettingsModel } from '../../types';

type CreateProjectValues = {
  sourcePath: string;
  contentOutput: string;
  password: string;
  passwordConfirm: string;
  maxPartSize?: number;
  useDefaultMaxPartSize: boolean;
  force: boolean;
  sourceCleanup: string;
  databaseExport?: string;
};

type CreateProjectModalProps = {
  open: boolean;
  loading: boolean;
  settings: SettingsModel | null;
  defaultSourceCleanup: string;
  onCancel: () => void;
  onSubmit: (values: CreateProjectValues) => void;
  t: (key: string) => string;
};

export function CreateProjectModal({
  open,
  loading,
  settings,
  defaultSourceCleanup,
  onCancel,
  onSubmit,
  t,
}: CreateProjectModalProps) {
  const [form] = Form.useForm<CreateProjectValues>();
  const useDefaultMaxPartSize = Form.useWatch('useDefaultMaxPartSize', form) ?? true;
  const effectiveDefaultMaxPartSize = settings?.defaultMaxPartSize ?? 0;

  return (
    <Modal
      title={t('createProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('createProject')}
      confirmLoading={loading}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          useDefaultMaxPartSize: true,
          force: false,
          sourceCleanup: defaultSourceCleanup,
        }}
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item
          name="sourcePath"
          label={t('createSourcePath')}
          rules={[{ required: true, message: t('createSourcePath') }]}
        >
          <Input placeholder="/path/to/source-folder" />
        </Form.Item>
        <Form.Item
          name="contentOutput"
          label={t('contentOutputPath')}
          rules={[{ required: true, message: t('contentOutputPath') }]}
        >
          <Input placeholder="/path/to/encrypted-content" />
        </Form.Item>
        <Form.Item name="databaseExport" label={t('databaseExportPath')}>
          <Input placeholder="/path/to/exported-project.fg" />
        </Form.Item>
        <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
          <Input.Password autoComplete="new-password" />
        </Form.Item>
        <Form.Item
          name="passwordConfirm"
          label={t('passwordConfirm')}
          dependencies={['password']}
          rules={[
            { required: true, message: t('passwordConfirmRequired') },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (value === getFieldValue('password')) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error(t('passwordMismatch')));
              },
            }),
          ]}
        >
          <Input.Password autoComplete="new-password" />
        </Form.Item>
        <Form.Item name="sourceCleanup" label={t('sourceCleanupOperationMode')} rules={[{ required: true }]}>
          <Select
            options={[
              { value: 'keep', label: t('sourceCleanupKeep') },
              { value: 'delete', label: t('sourceCleanupDelete') },
            ]}
          />
        </Form.Item>
        <Form.Item name="useDefaultMaxPartSize" valuePropName="checked">
          <Checkbox>{t('createUseDefaultMaxPartSize')}</Checkbox>
        </Form.Item>
        <Form.Item
          name="maxPartSize"
          label={t('defaultMaxPartSize')}
          rules={
            useDefaultMaxPartSize
              ? []
              : [{ required: true, message: t('defaultMaxPartSize') }, { type: 'number', min: 1 }]
          }
        >
          <InputNumber
            min={1}
            style={{ width: '100%' }}
            disabled={useDefaultMaxPartSize}
            placeholder={effectiveDefaultMaxPartSize > 0 ? String(effectiveDefaultMaxPartSize) : undefined}
          />
        </Form.Item>
        <Form.Item name="force" valuePropName="checked">
          <Checkbox>{t('forceOverwrite')}</Checkbox>
        </Form.Item>
      </Form>
    </Modal>
  );
}
