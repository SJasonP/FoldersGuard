import { Descriptions, Form, Input, Modal, Space, Typography } from 'antd';
import type { LocalProjectSummary } from '../../types';

type DeleteProjectModalProps = {
  open: boolean;
  loading: boolean;
  dataDirectory: string;
  project: LocalProjectSummary | null;
  onCancel: () => void;
  onSubmit: (password: string) => void;
  t: (key: string) => string;
};

export function DeleteProjectModal({ open, loading, dataDirectory, project, onCancel, onSubmit, t }: DeleteProjectModalProps) {
  const [form] = Form.useForm<{ password: string }>();

  return (
    <Modal
      title={t('deleteProject')}
      open={open}
      onCancel={() => {
        form.resetFields();
        onCancel();
      }}
      onOk={() => void form.submit()}
      okText={t('deleteProject')}
      okButtonProps={{ danger: true }}
      confirmLoading={loading}
    >
      <Space direction="vertical" size="middle" className="content-stack">
        <Typography.Text>{t('deleteProjectConfirm')}</Typography.Text>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('projectId')}>{project?.projectId ?? ''}</Descriptions.Item>
          <Descriptions.Item label={t('localDatabaseFileName')}>{project?.fileName ?? ''}</Descriptions.Item>
          <Descriptions.Item label={t('dataDirectory')}>{dataDirectory}</Descriptions.Item>
        </Descriptions>
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => {
            onSubmit(values.password);
            form.resetFields();
          }}
        >
          <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
            <Input.Password autoComplete="current-password" />
          </Form.Item>
        </Form>
      </Space>
    </Modal>
  );
}
