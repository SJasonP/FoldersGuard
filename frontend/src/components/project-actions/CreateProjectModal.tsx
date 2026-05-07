import { Checkbox, Form, Input, InputNumber, Modal, Select } from 'antd';
import type { SettingsModel } from '../../types';
import { showOperationConfirmation } from '../common/operationConfirmation';
import { PathInput } from '../common/PathInput';

type CreateProjectValues = {
  sourcePath: string;
  contentOutput: string;
  password: string;
  passwordConfirm: string;
  maxPartSize?: number;
  useDefaultMaxPartSize: boolean;
  force: boolean;
  sourceCleanup: string;
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
  const sourceCleanupLabel = (value: string) => {
    if (value === 'delete') {
      return t('sourceCleanupDelete');
    }
    return t('sourceCleanupKeep');
  };

  const confirmSubmit = (values: CreateProjectValues) => {
    showOperationConfirmation({
      title: t('createProject'),
      message: t('createProjectConfirm'),
      okText: t('createProject'),
      danger: values.sourceCleanup === 'delete',
      items: [
        { label: t('createSourcePath'), value: values.sourcePath },
        { label: t('contentOutputPath'), value: values.contentOutput },
        {
          label: t('defaultMaxPartSize'),
          value: values.useDefaultMaxPartSize ? t('createUseDefaultMaxPartSize') : values.maxPartSize,
        },
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
        onFinish={confirmSubmit}
      >
        <Form.Item
          name="sourcePath"
          label={t('createSourcePath')}
          rules={[{ required: true, message: t('createSourcePath') }]}
        >
          <PathInput
            dialogKind="open-directory"
            dialogTitle={t('createSourcePath')}
            placeholder="/path/to/source-folder"
            t={t}
          />
        </Form.Item>
        <Form.Item
          name="contentOutput"
          label={t('contentOutputPath')}
          rules={[{ required: true, message: t('contentOutputPath') }]}
        >
          <PathInput
            dialogKind="open-directory"
            dialogTitle={t('contentOutputPath')}
            placeholder="/path/to/encrypted-content"
            t={t}
          />
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
