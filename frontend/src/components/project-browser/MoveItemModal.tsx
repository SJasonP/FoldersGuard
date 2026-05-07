import { useEffect } from 'react';
import { Form, Modal, TreeSelect } from 'antd';
import type { TreeSelectProps } from 'antd';
import type { ProjectBrowserItemModel } from '../../types';

type MoveItemModalProps = {
  open: boolean;
  items: ProjectBrowserItemModel[];
  treeData: TreeSelectProps['treeData'];
  onCancel: () => void;
  onSubmit: (targetFolderId: string) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function MoveItemModal({ open, items, treeData, onCancel, onSubmit, t }: MoveItemModalProps) {
  const [form] = Form.useForm<{ targetFolderId: string }>();

  useEffect(() => {
    if (open) {
      form.resetFields();
    }
  }, [form, open]);

  return (
    <Modal
      title={t('moveItem')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('moveItem')}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(values) => {
          onSubmit(values.targetFolderId);
          form.resetFields();
        }}
      >
        <Form.Item name="targetFolderId" label={t('targetFolder')} rules={[{ required: true, message: t('targetFolder') }]}>
          <TreeSelect
            treeData={treeData}
            treeDefaultExpandAll
            disabled={items.length === 0}
            placeholder={t('targetFolder')}
            showSearch
            treeNodeFilterProp="title"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}
