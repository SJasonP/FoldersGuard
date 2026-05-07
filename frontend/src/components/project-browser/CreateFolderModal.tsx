import { Form, Input, Modal } from 'antd';

type CreateFolderModalValues = {
  name: string;
};

type CreateFolderModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: CreateFolderModalValues) => void;
  t: (key: string) => string;
};

export function CreateFolderModal({ open, loading, onCancel, onSubmit, t }: CreateFolderModalProps) {
  const [form] = Form.useForm<CreateFolderModalValues>();

  return (
    <Modal
      title={t('createFolder')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('createFolder')}
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
        <Form.Item name="name" label={t('folderName')} rules={[{ required: true, message: t('folderName') }]}>
          <Input autoComplete="off" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
