import type { ReactNode } from 'react';
import { Descriptions, Modal, Space, Typography } from 'antd';

type ConfirmationItem = {
  label: ReactNode;
  value: ReactNode;
};

type ShowOperationConfirmationArgs = {
  title: string;
  message: ReactNode;
  okText: string;
  items: ConfirmationItem[];
  danger?: boolean;
  onConfirm: () => void;
};

export function showOperationConfirmation({
  title,
  message,
  okText,
  items,
  danger = false,
  onConfirm,
}: ShowOperationConfirmationArgs) {
  Modal.confirm({
    title,
    okText,
    okButtonProps: danger ? { danger: true } : undefined,
    content: (
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Typography.Paragraph>{message}</Typography.Paragraph>
        <Descriptions column={1} bordered size="small">
          {items.map((item, index) => (
            <Descriptions.Item key={index} label={item.label}>
              {item.value}
            </Descriptions.Item>
          ))}
        </Descriptions>
      </Space>
    ),
    onOk: onConfirm,
  });
}
