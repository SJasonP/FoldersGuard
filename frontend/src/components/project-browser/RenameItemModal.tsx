import { useEffect } from 'react';
import { Form, Input, Modal } from 'antd';
import type { ProjectBrowserItemModel } from '../../types';

type RenameItemModalProps = {
  open: boolean;
  item: ProjectBrowserItemModel | null;
  onCancel: () => void;
  onSubmit: (newName: string) => void;
  t: (key: string) => string;
};

export function RenameItemModal({ open, item, onCancel, onSubmit, t }: RenameItemModalProps) {
  const [form] = Form.useForm<{ newName: string }>();

  useEffect(() => {
    if (open && item) {
      form.setFieldsValue({ newName: item.name });
    }
  }, [form, item, open]);

  return (
    <Modal
      title={t('renameItem')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('renameItem')}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => {
          onSubmit(values.newName);
          form.resetFields();
        }}
      >
        <Form.Item name="newName" label={t('newName')} rules={[{ required: true, message: t('newName') }]}>
          <Input autoComplete="off" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
