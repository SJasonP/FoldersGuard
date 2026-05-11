import {Descriptions, Drawer, List, Typography} from 'antd';
import type {CreateShareResultModel} from '../../types';
import {formatNumber} from '../../formatters';

type CreateShareResultDrawerProps = {
    open: boolean;
    result: CreateShareResultModel | null;
    onClose: () => void;
    t: (key: string) => string;
};

export function CreateShareResultDrawer({open, result, onClose, t}: CreateShareResultDrawerProps) {
    return (
        <Drawer title={t('createShare')} open={open} onClose={onClose} width={620}>
            {result ? (
                <>
                    <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
                        <Descriptions.Item label={t('shareId')}>{result.shareId}</Descriptions.Item>
                        <Descriptions.Item label={t('shareDatabaseOutputPath')}>{result.outputPath}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('shareSummaryTopLevelItems')}>{formatNumber(result.topLevelItems)}</Descriptions.Item>
                        <Descriptions.Item label={t('fileCount')}>{formatNumber(result.files)}</Descriptions.Item>
                        <Descriptions.Item label={t('folderCount')}>{formatNumber(result.folders)}</Descriptions.Item>
                        <Descriptions.Item label={t('partCount')}>{formatNumber(result.parts)}</Descriptions.Item>
                        <Descriptions.Item label={t('passwordProtected')}>
                            {result.passwordProtected ? t('passwordProtectedYes') : t('passwordProtectedNo')}
                        </Descriptions.Item>
                    </Descriptions>
                    <Typography.Title level={5} style={{marginTop: 16}}>
                        {t('shareContentLocations')}
                    </Typography.Title>
                    <List
                        size="small"
                        bordered
                        dataSource={result.contentLocations}
                        renderItem={(location) => (
                            <List.Item>
                                <Typography.Text code>{location.sourcePath}</Typography.Text>
                                <Typography.Text type="secondary"> -&gt; </Typography.Text>
                                <Typography.Text code>{location.targetPath}</Typography.Text>
                            </List.Item>
                        )}
                    />
                </>
            ) : null}
        </Drawer>
    );
}
