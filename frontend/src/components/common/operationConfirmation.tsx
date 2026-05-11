import type {ReactNode} from 'react';
import {Descriptions, Space, Typography} from 'antd';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';

type ConfirmationItem = {
    label: ReactNode;
    value: ReactNode;
};

type ShowOperationConfirmationArgs = {
    modalApi: ModalHookAPI;
    title: string;
    message: ReactNode;
    okText: string;
    items: ConfirmationItem[];
    danger?: boolean;
    onConfirm: () => void;
};

export function showOperationConfirmation({
                                              modalApi,
                                              title,
                                              message,
                                              okText,
                                              items,
                                              danger = false,
                                              onConfirm,
                                          }: ShowOperationConfirmationArgs) {
    modalApi.confirm({
        title,
        okText,
        okButtonProps: danger ? {danger: true} : undefined,
        content: (
            <Space direction="vertical" size="middle" style={{width: '100%'}}>
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
