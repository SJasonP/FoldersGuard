import {Descriptions, Drawer} from 'antd';
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
    return (
        <Drawer title={title ?? t('verifyProject')} open={open} onClose={onClose} width={540}>
            {result ? (
                <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label={identityLabel ?? t('projectId')}>{result.projectId}</Descriptions.Item>
                    <Descriptions.Item label={t('verifyProjectStatus')}>
                        {result.status === 'ok' ? t('verificationOk') : t('verificationFailed')}
                    </Descriptions.Item>
                    <Descriptions.Item
                        label={t('checkedObjects')}>{formatNumber(result.checkedObjects)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('missingObjects')}>{formatNumber(result.missingObjects)}</Descriptions.Item>
                    <Descriptions.Item
                        label={t('tamperedObjects')}>{formatNumber(result.tamperedObjects)}</Descriptions.Item>
                    <Descriptions.Item label={t('extraObjects')}>{formatNumber(result.extraObjects)}</Descriptions.Item>
                </Descriptions>
            ) : null}
        </Drawer>
    );
}
