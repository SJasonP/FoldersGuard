import { useEffect } from 'react';
import { Form, Input, Modal } from 'antd';
import type { ProjectBrowserItemModel } from '../../types';
import { projectItemNameRules } from './projectBrowserNameValidation';
import { useResetFormOnClose } from '../common/useResetFormOnClose';

type RenameItemModalProps = {
  open: boolean;
  item: ProjectBrowserItemModel | null;
  onCancel: () => void;
  onSubmit: (newName: string) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function RenameItemModal({ open, item, onCancel, onSubmit, t }: RenameItemModalProps) {
  const [form] = Form.useForm<{ newName: string }>();
  useResetFormOnClose(form, open);

  useEffect(() => {
    if (open && item) {
      form.setFieldsValue({ newName: item.name });
    }
  }, [form, item, open]);

  return (
    <Modal
      title={t('renameItem')}
      open={open}
      onCancel={onCancel}
      onOk={() => void form.submit()}
      okText={t('renameItem')}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => {
          onSubmit(values.newName);
        }}
      >
        <Form.Item name="newName" label={t('newName')} rules={projectItemNameRules(t)}>
          <Input autoComplete="off" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
