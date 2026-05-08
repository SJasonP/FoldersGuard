import { Form, Input, Modal } from 'antd';
import { projectItemNameRules } from './projectBrowserNameValidation';
import { useResetFormOnClose } from '../common/useResetFormOnClose';

type CreateFolderModalValues = {
  name: string;
};

type CreateFolderModalProps = {
  open: boolean;
  loading: boolean;
  onCancel: () => void;
  onSubmit: (values: CreateFolderModalValues) => void;
  t: (key: string, values?: Record<string, string | number>) => string;
};

export function CreateFolderModal({ open, loading, onCancel, onSubmit, t }: CreateFolderModalProps) {
  const [form] = Form.useForm<CreateFolderModalValues>();
  useResetFormOnClose(form, open);

  return (
    <Modal
      title={t('createFolder')}
      open={open}
      onCancel={onCancel}
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
        }}
      >
        <Form.Item name="name" label={t('folderName')} rules={projectItemNameRules(t)}>
          <Input autoComplete="off" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
