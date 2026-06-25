import {App as AntApp, Button, Empty, Modal, Table, Typography} from 'antd';
import type {ColumnsType} from 'antd/es/table';
import type {ProjectBackupInfoModel} from '../../types';
import {formatDateTime, formatFileSize} from '../../formatters';

type RestoreBackupModalProps = {
    open: boolean;
    loading: boolean;
    restoreLoading: boolean;
    backups: ProjectBackupInfoModel[];
    onRestore: (backupId: string) => void;
    onCancel: () => void;
    t: (key: string, options?: Record<string, unknown>) => string;
};

const reasonKeys: Record<string, string> = {
    apply: 'backupReasonApply',
    delete: 'backupReasonDelete',
    rekey: 'backupReasonRekey',
    restore: 'backupReasonRestore',
    manual: 'backupReasonManual',
};

export function RestoreBackupModal({
                                       open,
                                       loading,
                                       restoreLoading,
                                       backups,
                                       onRestore,
                                       onCancel,
                                       t,
                                   }: RestoreBackupModalProps) {
    const {modal} = AntApp.useApp();

    const reasonLabel = (reason: string) => {
        const key = reasonKeys[reason];
        return key ? t(key) : reason;
    };

    const confirmRestore = (backup: ProjectBackupInfoModel) => {
        modal.confirm({
            title: t('restoreBackup'),
            content: t('restoreBackupConfirm', {time: formatDateTime(backup.createdAt)}),
            okText: t('restoreBackup'),
            onOk: () => onRestore(backup.id),
        });
    };

    const columns: ColumnsType<ProjectBackupInfoModel> = [
        {
            title: t('backupCreatedAt'),
            dataIndex: 'createdAt',
            key: 'createdAt',
            render: (value: string) => formatDateTime(value),
        },
        {
            title: t('backupReason'),
            dataIndex: 'reason',
            key: 'reason',
            render: (value: string) => reasonLabel(value),
        },
        {
            title: t('backupSize'),
            dataIndex: 'size',
            key: 'size',
            render: (value: number) => formatFileSize(value),
        },
        {
            title: '',
            key: 'action',
            align: 'right',
            render: (_, backup) => (
                <Button size="small" onClick={() => confirmRestore(backup)} disabled={restoreLoading}>
                    {t('restore')}
                </Button>
            ),
        },
    ];

    return (
        <Modal title={t('restoreBackup')} open={open} onCancel={onCancel} footer={null} width={640}>
            <Typography.Paragraph type="secondary">{t('restoreBackupIntro')}</Typography.Paragraph>
            <Table<ProjectBackupInfoModel>
                rowKey="id"
                size="small"
                loading={loading}
                columns={columns}
                dataSource={backups}
                pagination={false}
                locale={{emptyText: <Empty description={t('noBackups')}/>}}
            />
        </Modal>
    );
}
