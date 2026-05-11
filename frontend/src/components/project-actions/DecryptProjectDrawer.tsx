import {Descriptions, Drawer} from 'antd';
import type {DecryptProjectResultModel} from '../../types';
import {formatNumber} from '../../formatters';

type DecryptProjectDrawerProps = {
    open: boolean;
    result: DecryptProjectResultModel | null;
    onClose: () => void;
    t: (key: string) => string;
};

export function DecryptProjectDrawer({open, result, onClose, t}: DecryptProjectDrawerProps) {
    return (
        <Drawer title={t('decryptProject')} open={open} onClose={onClose} width={540}>
            {result ? (
                <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
                    <Descriptions.Item label={t('outputPath')}>{result.outputPath}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('decryptedFiles')}>{formatNumber(result.decryptedFiles)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('restoredFolders')}>{formatNumber(result.restoredFolders)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('skippedFolders')}>{formatNumber(result.skippedFolders)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('deletedEncryptedFiles')}>{formatNumber(result.deletedEncryptedFiles)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('failedEncryptedFiles')}>{formatNumber(result.failedEncryptedFiles)}</Descriptions.Item>
                </Descriptions>
            ) : null}
        </Drawer>
    );
}
