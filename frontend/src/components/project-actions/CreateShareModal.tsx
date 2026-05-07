import { Checkbox, Form, Input, Modal, Select } from 'antd';

type CreateShareValues = {
  itemPaths: string[];
  outputPath: string;
  force: boolean;
  passwordProtected: boolean;
  sharePassword?: string;
  sharePasswordConfirm?: string;
};

type CreateShareModalProps = {
  open: boolean;
  loading: boolean;
  selectableItems: Array<{ value: string; label: string }>;
  onCancel: () => void;
  onSubmit: (values: CreateShareValues) => void;
  t: (key: string) => string;
};

export function CreateShareModal({ open, loading, selectableItems, onCancel, onSubmit, t }: CreateShareModalProps) {
  const [form] = Form.useForm<CreateShareValues>();
  const passwordProtected = Form.useWatch('passwordProtected', form) ?? true;

  return (
    <Modal
      title={t('createShare')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('createShare')}
      confirmLoading={loading}
      width={720}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          force: false,
          passwordProtected: true,
        }}
        onFinish={(values) => {
          onSubmit(values);
          form.resetFields();
        }}
      >
        <Form.Item
          name="itemPaths"
          label={t('shareSelectionItems')}
          rules={[{ required: true, message: t('shareSelectionItems') }]}
        >
          <Select
            mode="multiple"
            options={selectableItems}
            placeholder={t('shareSelectionPlaceholder')}
            optionFilterProp="label"
          />
        </Form.Item>
        <Form.Item
          name="outputPath"
          label={t('shareDatabaseOutputPath')}
          rules={[{ required: true, message: t('shareDatabaseOutputPath') }]}
        >
          <Input placeholder="/path/to/share.fgs" />
        </Form.Item>
        <Form.Item name="passwordProtected" valuePropName="checked">
          <Checkbox>{t('sharePasswordProtected')}</Checkbox>
        </Form.Item>
        <Form.Item
          name="sharePassword"
          label={t('sharePassword')}
          rules={
            passwordProtected
              ? [{ required: true, message: t('sharePasswordRequired') }]
              : []
          }
        >
          <Input.Password autoComplete="new-password" disabled={!passwordProtected} />
        </Form.Item>
        <Form.Item
          name="sharePasswordConfirm"
          label={t('sharePasswordConfirm')}
          dependencies={['sharePassword', 'passwordProtected']}
          rules={
            passwordProtected
              ? [
                  { required: true, message: t('sharePasswordConfirmRequired') },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (value === getFieldValue('sharePassword')) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error(t('passwordMismatch')));
                    },
                  }),
                ]
              : []
          }
        >
          <Input.Password autoComplete="new-password" disabled={!passwordProtected} />
        </Form.Item>
        <Form.Item name="force" valuePropName="checked">
          <Checkbox>{t('forceOverwrite')}</Checkbox>
        </Form.Item>
      </Form>
    </Modal>
  );
}
