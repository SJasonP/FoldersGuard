import {createPortal} from 'react-dom';
import {Progress, Typography} from 'antd';
import {LoadingOutlined} from '@ant-design/icons';
import type {OperationProgress as OperationProgressData} from '../../hooks/useOperationProgress';
import {formatDuration, formatFileSize, formatThroughput} from '../../formatters';

type OperationProgressProps = {
    label: string;
    progress: OperationProgressData | null;
    resolvedTheme: 'light' | 'dark';
    t: (key: string, options?: Record<string, unknown>) => string;
};

const phaseLabelKeys: Record<string, string> = {
    preparing: 'operationPhasePreparing',
    encrypting: 'operationPhaseEncrypting',
    decrypting: 'operationPhaseDecrypting',
    verifying: 'operationPhaseVerifying',
    copying: 'operationPhaseCopying',
    finalizing: 'operationPhaseFinalizing',
};

function phaseLabel(phase: string, t: OperationProgressProps['t']) {
    const key = phaseLabelKeys[phase];
    return key ? t(key) : t('operationPreparing');
}

export function OperationProgress({label, progress, resolvedTheme, t}: OperationProgressProps) {
    const determinate = Boolean(progress?.determinate && progress.totalBytes > 0);
    const percent = determinate
        ? Math.min(100, Math.floor((progress!.processedBytes / progress!.totalBytes) * 100))
        : 0;

    const phaseText = progress ? phaseLabel(progress.phase, t) : t('operationPreparing');

    const detailParts: string[] = [];
    if (progress && determinate) {
        detailParts.push(
            t('operationProcessedOfTotal', {
                processed: formatFileSize(progress.processedBytes),
                total: formatFileSize(progress.totalBytes),
            }),
        );
        if (progress.bytesPerSecond > 0) {
            detailParts.push(formatThroughput(progress.bytesPerSecond));
        }
        if (progress.etaSeconds > 0) {
            detailParts.push(t('operationEtaRemaining', {eta: formatDuration(progress.etaSeconds)}));
        }
    }
    const detail = detailParts.join(' · ');

    const card = (
        <div
            className={`operation-overlay operation-overlay--${resolvedTheme}`}
            role="status"
            aria-live="polite"
        >
            <div className="operation-overlay-head">
                <LoadingOutlined className="operation-overlay-spinner" spin/>
                <Typography.Text strong ellipsis className="operation-overlay-title">
                    {t('operationRunning')}: {label}
                </Typography.Text>
            </div>
            <Progress
                className="operation-overlay-bar"
                percent={percent}
                status="active"
                showInfo={determinate}
                format={(value) => `${value ?? 0}%`}
            />
            <Typography.Text type="secondary" ellipsis className="operation-overlay-detail">
                {detail ? `${phaseText} · ${detail}` : phaseText}
            </Typography.Text>
        </div>
    );

    return createPortal(card, document.body);
}
