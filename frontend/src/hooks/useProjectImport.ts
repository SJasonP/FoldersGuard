import {useState} from 'react';
import type {MessageInstance} from 'antd/es/message/interface';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';
import {ImportProject} from '../../wailsjs/go/main/App';
import type {ImportProjectResultModel} from '../types';
import {showOperationError} from '../components/common/operationError';

type UseProjectImportArgs = {
    messageApi: MessageInstance;
    modalApi: ModalHookAPI;
    t: (key: string) => string;
    reloadProjects: () => Promise<void>;
};

export function useProjectImport({messageApi, modalApi, t, reloadProjects}: UseProjectImportArgs) {
    const [importDialogOpen, setImportDialogOpen] = useState(false);
    const [importLoading, setImportLoading] = useState(false);

    const handleImportProject = async (values: { inputPath: string; password: string; force: boolean }) => {
        setImportDialogOpen(false);
        setImportLoading(true);
        try {
            const result: ImportProjectResultModel = await ImportProject({
                inputPath: values.inputPath,
                password: values.password,
                force: values.force,
            });
            await reloadProjects();
            messageApi.success(`${t('importProjectSucceeded')}: ${result.projectId}`);
        } catch (error) {
            showOperationError(modalApi, t('importProjectFailed'), error, t);
        } finally {
            setImportLoading(false);
        }
    };

    return {
        importDialogOpen,
        importLoading,
        setImportDialogOpen,
        handleImportProject,
    };
}
