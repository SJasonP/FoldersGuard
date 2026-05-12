import {Descriptions, Drawer, List, Space, Typography} from 'antd';
import type {VerifyProjectResultModel} from '../../types';
import {formatNumber} from '../../formatters';

type VerifyProjectDrawerProps = {
    open: boolean;
    result: VerifyProjectResultModel | null;
    onClose: () => void;
    title?: string;
    identityLabel?: string;
    t: (key: string) => string;
};

export function VerifyProjectDrawer({open, result, onClose, title, identityLabel, t}: VerifyProjectDrawerProps) {
    const detailSections = result ? [
        {key: 'missing', title: t('missingObjectPaths'), paths: result.missingPaths ?? []},
        {key: 'tampered', title: t('tamperedObjectPaths'), paths: result.tamperedPaths ?? []},
        {key: 'extra', title: t('extraObjectPaths'), paths: result.extraPaths ?? []},
    ].filter((section) => section.paths.length > 0) : [];

    return (
        <Drawer title={title ?? t('verifyProject')} open={open} onClose={onClose} width={720}>
            {result ? (
                <Space direction="vertical" size="middle" style={{width: '100%'}}>
                    <Descriptions column={1} bordered size="small">
                        <Descriptions.Item
                            label={identityLabel ?? t('projectId')}>{result.projectId}</Descriptions.Item>
                        <Descriptions.Item label={t('verifyProjectStatus')}>
                            {result.status === 'ok' ? t('verificationOk') : t('verificationFailed')}
                        </Descriptions.Item>
                        <Descriptions.Item
                            label={t('checkedObjects')}>{formatNumber(result.checkedObjects)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('missingObjects')}>{formatNumber(result.missingObjects)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('tamperedObjects')}>{formatNumber(result.tamperedObjects)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('extraObjects')}>{formatNumber(result.extraObjects)}</Descriptions.Item>
                    </Descriptions>
                    {detailSections.map((section) => (
                        <div key={section.key}>
                            <Typography.Title level={5}>{section.title}</Typography.Title>
                            <List
                                size="small"
                                bordered
                                dataSource={section.paths}
                                renderItem={(path) => (
                                    <List.Item>
                                        <Typography.Text code copyable={{text: path}}>{path}</Typography.Text>
                                    </List.Item>
                                )}
                            />
                        </div>
                    ))}
                </Space>
            ) : null}
        </Drawer>
    );
}
