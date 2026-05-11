import {useEffect} from 'react';
import type {TreeSelectProps} from 'antd';
import {Form, Modal, TreeSelect} from 'antd';
import type {ProjectBrowserItemModel} from '../../types';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type MoveItemModalProps = {
    open: boolean;
    items: ProjectBrowserItemModel[];
    treeData: TreeSelectProps['treeData'];
    onCancel: () => void;
    onSubmit: (targetFolderId: string) => void;
    t: (key: string, values?: Record<string, string | number>) => string;
};

export function MoveItemModal({open, items, treeData, onCancel, onSubmit, t}: MoveItemModalProps) {
    const [form] = Form.useForm<{ targetFolderId: string }>();
    useResetFormOnClose(form, open);

    useEffect(() => {
        if (open) {
            form.setFieldsValue({targetFolderId: ''});
        }
    }, [form, open]);

    return (
        <Modal
            title={t('moveItem')}
            open={open}
            onCancel={onCancel}
            onOk={() => void form.submit()}
            okText={t('moveItem')}
            destroyOnHidden
        >
            <Form
                form={form}
                layout="vertical"
                onFinish={(values) => {
                    onSubmit(values.targetFolderId);
                }}
            >
                <Form.Item name="targetFolderId" label={t('targetFolder')}
                           rules={[{required: true, message: t('targetFolder')}]}>
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
