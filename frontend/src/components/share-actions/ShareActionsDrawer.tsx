import {Button, Descriptions, Drawer, Space} from 'antd';
import {SafetyCertificateOutlined} from '@ant-design/icons';
import type {ShareSummaryModel} from '../../types';
import {formatNumber} from '../../formatters';

type ShareActionsDrawerProps = {
    open: boolean;
    share: ShareSummaryModel | null;
    onClose: () => void;
    onInspect: () => void;
    onDecrypt: () => void;
    onVerify: () => void;
    t: (key: string) => string;
};

export function ShareActionsDrawer({open, share, onClose, onInspect, onDecrypt, onVerify, t}: ShareActionsDrawerProps) {
    return (
        <Drawer title={t('shareActions')} open={open} onClose={onClose} width={540}>
            <Space direction="vertical" size="middle" className="content-stack">
                {share ? (
                    <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label={t('shareId')}>{share.shareId}</Descriptions.Item>
                        <Descriptions.Item label={t('databaseType')}>{share.databaseType}</Descriptions.Item>
                        <Descriptions.Item label={t('formatVersion')}>{share.formatVersion}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('shareSummaryTopLevelItems')}>{formatNumber(share.topLevelItems)}</Descriptions.Item>
                        <Descriptions.Item label={t('fileCount')}>{formatNumber(share.files)}</Descriptions.Item>
                        <Descriptions.Item label={t('folderCount')}>{formatNumber(share.folders)}</Descriptions.Item>
                        <Descriptions.Item label={t('partCount')}>{formatNumber(share.parts)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('storageObjects')}>{formatNumber(share.storageObjects)}</Descriptions.Item>
                        <Descriptions.Item label={t('passwordProtected')}>
                            {share.passwordProtected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
                        </Descriptions.Item>
                    </Descriptions>
                ) : null}
                <Button block type="primary" onClick={onInspect}>
                    {t('inspectShare')}
                </Button>
                <Button block onClick={onDecrypt}>
                    {t('decryptShare')}
                </Button>
                <Button block icon={<SafetyCertificateOutlined/>} onClick={onVerify}>
                    {t('verifyShare')}
                </Button>
            </Space>
        </Drawer>
    );
}
