import { Form, Input, InputNumber, Modal } from 'antd';

type AddItemModalValues = {
  sourcePath: string;
  maxPartSize?: number;
};

type AddItemModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: AddItemModalValues) => void;
  t: (key: string) => string;
};

export function AddItemModal({ open, loading, onCancel, onSubmit, t }: AddItemModalProps) {
  const [form] = Form.useForm<AddItemModalValues>();

  return (
    <Modal
      title={t('addItem')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('addItem')}
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
        <Form.Item name="sourcePath" label={t('sourcePath')} rules={[{ required: true, message: t('sourcePath') }]}>
          <Input placeholder="/path/to/file-or-folder" />
        </Form.Item>
        <Form.Item name="maxPartSize" label={t('defaultMaxPartSize')}>
          <InputNumber min={1} style={{ width: '100%' }} placeholder={t('createUseDefaultMaxPartSize')} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
