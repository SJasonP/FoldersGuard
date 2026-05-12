import {Alert, App as AntApp, Descriptions, Drawer, List, Space, Typography} from 'antd';
import type {ApplyProjectChangesResultModel} from '../../types';
import {formatNumber} from '../../formatters';
import type {main} from '../../../wailsjs/go/models';

type ApplyChangesResultDrawerProps = {
    open: boolean;
    result: ApplyProjectChangesResultModel | null;
    onClose: () => void;
    t: (key: string, values?: Record<string, string | number>) => string;
};

export function ApplyChangesResultDrawer({open, result, onClose, t}: ApplyChangesResultDrawerProps) {
    const {modal} = AntApp.useApp();
    const manualContentGuide = Boolean(result?.manualContentGuide);
    const manualOperations = manualContentGuide ? (result?.contentOperations ?? []) : [];
    const uploadOperations = manualOperations.filter((operation) => operation.type === 'upload');
    const moveOperations = manualOperations.filter((operation) => operation.type === 'move');
    const deleteOperations = manualOperations.filter((operation) => operation.type === 'delete');
    const otherOperations = manualOperations.filter((operation) => !['upload', 'move', 'delete'].includes(operation.type));
    const closeOrConfirm = () => {
        if (!manualContentGuide) {
            onClose();
            return;
        }
        modal.confirm({
            title: t('manualContentGuideCloseTitle'),
            content: t('manualContentGuideCloseConfirm'),
            okText: t('close'),
            cancelText: t('stay'),
            onOk: onClose,
        });
    };

    return (
        <Drawer title={t('applyChangesResult')} open={open} onClose={closeOrConfirm} width={760}
                maskClosable={!manualContentGuide}>
            {result ? (
                <>
                    {manualContentGuide ? (
                        <Alert
                            type="warning"
                            showIcon
                            message={t('manualContentGuideCreated')}
                            description={
                                <Space direction="vertical" size={4}>
                                    <Typography.Text>{t('manualContentGuideCreatedDescription')}</Typography.Text>
                                    {result.stagedContentName ? (
                                        <Typography.Text strong>
                                            {result.stagedContentOnDesktop
                                                ? t('stagedContentWrittenToDesktop', {name: result.stagedContentName})
                                                : t('stagedContentWrittenToPath', {path: result.stagedContentPath})}
                                        </Typography.Text>
                                    ) : null}
                                </Space>
                            }
                            style={{marginBottom: 12}}
                        />
                    ) : result.appliedContentChanges?.length ? (
                        <Alert
                            type="success"
                            showIcon
                            message={t('contentOperationsApplied')}
                            description={t('contentOperationsAppliedDescription')}
                            style={{marginBottom: 12}}
                        />
                    ) : null}
                    <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('appliedRenames')}>{formatNumber(result.appliedRenames)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('appliedMoves')}>{formatNumber(result.appliedMoves)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('appliedRemoves')}>{formatNumber(result.appliedRemoves)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('appliedAdds')}>{formatNumber(result.appliedAdds)}</Descriptions.Item>
                        <Descriptions.Item
                            label={t('appliedCreatedFolders')}>{formatNumber(result.appliedCreatedFolders)}</Descriptions.Item>
                        {result.stagedContentPath ? <Descriptions.Item
                            label={result.stagedContentOnDesktop ? t('desktopFolder') : t('stagedContentPath')}>
                            {result.stagedContentOnDesktop && result.stagedContentName
                                ? t('desktopFolderNamed', {name: result.stagedContentName})
                                : result.stagedContentPath}
                        </Descriptions.Item> : null}
                        {manualContentGuide ? (
                            <Descriptions.Item
                                label={t('manualContentOperations')}>{formatNumber(manualOperations.length)}</Descriptions.Item>
                        ) : null}
                    </Descriptions>
                    {manualContentGuide ? (
                        <>
                            <Typography.Title level={5}>{t('manualContentGuide')}</Typography.Title>
                            <Typography.Paragraph>{t('manualContentGuideIntro')}</Typography.Paragraph>
                            <ManualOperationList
                                title={t('manualUploadTitle')}
                                description={uploadOperations.length > 0 ? t('manualUploadDescription') : ''}
                                operations={uploadOperations}
                                emptyText={t('noContentOperations')}
                                renderOperation={(operation) => operation.targetPath || operation.sourcePath}
                            />
                            <ManualOperationList
                                title={t('manualMoveTitle')}
                                description={moveOperations.length > 0 ? t('manualMoveDescription') : ''}
                                operations={moveOperations}
                                emptyText={t('noContentOperations')}
                                renderOperation={(operation) => `${operation.sourcePath} -> ${operation.targetPath}`}
                            />
                            <ManualOperationList
                                title={t('manualDeleteTitle')}
                                description={deleteOperations.length > 0 ? t('manualDeleteDescription') : ''}
                                operations={deleteOperations}
                                emptyText={t('noContentOperations')}
                                renderOperation={(operation) => operation.targetPath}
                            />
                            {otherOperations.length > 0 ? (
                                <ManualOperationList
                                    title={t('manualOtherTitle')}
                                    description=""
                                    operations={otherOperations}
                                    emptyText={t('noContentOperations')}
                                    renderOperation={(operation) =>
                                        operation.sourcePath && operation.targetPath
                                            ? `${operation.sourcePath} -> ${operation.targetPath}`
                                            : operation.targetPath || operation.sourcePath || operation.type}
                                />
                            ) : null}
                        </>
                    ) : null}
                </>
            ) : null}
        </Drawer>
    );
}

type ManualOperationListProps = {
    title: string;
    description: string;
    operations: main.ProjectContentOperation[];
    emptyText: string;
    renderOperation: (operation: main.ProjectContentOperation) => string;
};

function ManualOperationList({title, description, operations, emptyText, renderOperation}: ManualOperationListProps) {
    if (operations.length === 0) {
        return null;
    }
    return (
        <Space direction="vertical" size={6} style={{display: 'flex', marginBottom: 12}}>
            <Typography.Text strong>{title}</Typography.Text>
            {description ? <Typography.Text type="secondary">{description}</Typography.Text> : null}
            <List
                size="small"
                bordered
                dataSource={operations}
                locale={{emptyText}}
                renderItem={(operation) => (
                    <List.Item>
                        <Typography.Text code>{renderOperation(operation)}</Typography.Text>
                    </List.Item>
                )}
            />
        </Space>
    );
}
