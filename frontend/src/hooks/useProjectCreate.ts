import {useState} from 'react';
import type {MessageInstance} from 'antd/es/message/interface';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';
import {CreateProject} from '../../wailsjs/go/main/App';
import type {CreateProjectResultModel} from '../types';
import {formatNumber} from '../formatters';
import {showOperationError} from '../components/common/operationError';

type CreateProjectValues = {
    sourcePath: string;
    contentOutput: string;
    password: string;
    passwordConfirm: string;
    force: boolean;
};

type UseProjectCreateArgs = {
    messageApi: MessageInstance;
    modalApi: ModalHookAPI;
    t: (key: string) => string;
    reloadProjects: () => Promise<void>;
};

export function useProjectCreate({messageApi, modalApi, t, reloadProjects}: UseProjectCreateArgs) {
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [createLoading, setCreateLoading] = useState(false);

    const handleCreateProject = async (values: CreateProjectValues) => {
        // Close the form as soon as the operation starts so only the progress
        // overlay remains while encryption runs.
        setCreateDialogOpen(false);
        setCreateLoading(true);
        try {
            const result: CreateProjectResultModel = await CreateProject({
                sourcePath: values.sourcePath,
                contentOutput: values.contentOutput,
                password: values.password,
                maxPartSize: 0,
                force: values.force,
                sourceCleanup: '',
                databaseExport: '',
            });
            await reloadProjects();
            messageApi.success(
                [
                    t('createProjectSucceeded'),
                    `${t('createSummaryProjectId')}: ${result.projectId}`,
                    `${t('createSummaryProjectName')}: ${result.projectName}`,
                    `${t('contentOutputPath')}: ${result.contentOutput}`,
                    `${t('createSummaryEncryptedFiles')}: ${formatNumber(result.encryptedFiles)}`,
                    `${t('createSummaryEncryptedFolders')}: ${formatNumber(result.encryptedFolders)}`,
                    `${t('createSummaryEncryptedParts')}: ${formatNumber(result.encryptedParts)}`,
                    `${t('createSummaryDeletedCleartextFiles')}: ${formatNumber(result.deletedCleartextFiles)}`,
                    `${t('createSummaryFailedFiles')}: ${formatNumber(result.failedFiles)}`,
                ].join(' | '),
            );
        } catch (error) {
            showOperationError(modalApi, t('createProjectFailed'), error, t);
        } finally {
            setCreateLoading(false);
        }
    };

    return {
        createDialogOpen,
        createLoading,
        setCreateDialogOpen,
        handleCreateProject,
    };
}
