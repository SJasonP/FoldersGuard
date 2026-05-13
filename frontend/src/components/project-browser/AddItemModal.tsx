import {Form, Modal} from 'antd';
import {PathInput} from '../common/PathInput';
import {useResetFormOnClose} from '../common/useResetFormOnClose';

type AddItemModalValues = {
    sourcePath: string;
};

type AddItemModalProps = {
    open: boolean;
    loading: boolean;
    onCancel: () => void;
    onSubmit: (values: AddItemModalValues) => void;
    t: (key: string, values?: Record<string, string | number>) => string;
};

export function AddItemModal({open, loading, onCancel, onSubmit, t}: AddItemModalProps) {
    const [form] = Form.useForm<AddItemModalValues>();
    useResetFormOnClose(form, open);

    return (
        <Modal
            title={t('addItem')}
            open={open}
            onCancel={onCancel}
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
                }}
            >
                <Form.Item name="sourcePath" label={t('sourcePath')}
                           rules={[{required: true, message: t('sourcePath')}]}>
                    <PathInput
                        dialogKind="open-file"
                        dialogTitle={t('sourcePath')}
                        buttonLabel={t('browseFile')}
                        secondaryDialogKind="open-directory"
                        secondaryDialogTitle={t('sourcePath')}
                        secondaryButtonLabel={t('browseFolder')}
                        placeholder="/path/to/file-or-folder"
                        t={t}
                    />
                </Form.Item>
            </Form>
        </Modal>
    );
}
